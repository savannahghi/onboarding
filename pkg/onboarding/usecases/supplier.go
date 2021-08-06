package usecases

import (
	"context"
	"fmt"

	"github.com/cenkalti/backoff"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/profileutils"
	"github.com/sirupsen/logrus"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/authorization"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/authorization/permission"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/exceptions"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/utils"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/engagement"
	"github.com/savannahghi/onboarding/pkg/onboarding/repository"

	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/messaging"
	pubsubmessaging "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/pubsub"
)

// Supplier constants
const (
	active                 = true
	country                = "KEN" // Anticipate worldwide expansion
	supplierCollectionName = "suppliers"
	futureHours            = 878400
	SavannahSladeCode      = "1"
	SavannahOrgName        = "Savannah Informatics"

	// Supplier Suspension EmailSubject Title
	supplierSuspensionEmailSubjectTitle = "Suspension from Be.Well"
	// PartnerAccountSetupNudgeTitle is the title defined in the `engagement service`
	// for the `PartnerAccountSetupNudge`
	PartnerAccountSetupNudgeTitle = "Setup your partner account"
)

// SupplierUseCases represent the business logic required for management of suppliers
type SupplierUseCases interface {
	AddPartnerType(ctx context.Context, name *string, partnerType *profileutils.PartnerType) (bool, error)

	FindSupplierByID(ctx context.Context, id string) (*profileutils.Supplier, error)

	FindSupplierByUID(ctx context.Context) (*profileutils.Supplier, error)

	SetUpSupplier(ctx context.Context, accountType profileutils.AccountType) (*profileutils.Supplier, error)

	SuspendSupplier(ctx context.Context, suspensionReason *string) (bool, error)
	SupplierSetDefaultLocation(ctx context.Context, locationID string) (*profileutils.Supplier, error)
}

// SupplierUseCasesImpl represents usecase implementation object
type SupplierUseCasesImpl struct {
	repo       repository.OnboardingRepository
	profile    ProfileUseCase
	engagement engagement.ServiceEngagement
	messaging  messaging.ServiceMessaging
	baseExt    extension.BaseExtension
	pubsub     pubsubmessaging.ServicePubSub
}

// NewSupplierUseCases returns a new a onboarding usecase
func NewSupplierUseCases(
	r repository.OnboardingRepository,
	p ProfileUseCase,
	eng engagement.ServiceEngagement,
	messaging messaging.ServiceMessaging,
	ext extension.BaseExtension,
	pubsub pubsubmessaging.ServicePubSub,
) SupplierUseCases {

	return &SupplierUseCasesImpl{
		repo:       r,
		profile:    p,
		engagement: eng,
		messaging:  messaging,
		baseExt:    ext,
		pubsub:     pubsub,
	}
}

// AddPartnerType create the initial supplier record
func (s SupplierUseCasesImpl) AddPartnerType(
	ctx context.Context,
	name *string,
	partnerType *profileutils.PartnerType,
) (bool, error) {
	ctx, span := tracer.Start(ctx, "AddPartnerType")
	defer span.End()

	if name == nil || partnerType == nil {
		return false, fmt.Errorf("expected `name` to be defined and `partnerType` to be valid")
	}

	if !partnerType.IsValid() {
		return false, exceptions.InvalidPartnerTypeError()
	}

	if *partnerType == profileutils.PartnerTypeConsumer {
		return false, exceptions.WrongEnumTypeError(partnerType.String())
	}

	user, err := s.baseExt.GetLoggedInUser(ctx)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, fmt.Errorf("can't get user: %w", err)
	}
	isAuthorized, err := authorization.IsAuthorized(user, permission.PartnerTypeCreate)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, err
	}
	if !isAuthorized {
		return false, fmt.Errorf("user not authorized to access this resource")
	}

	profile, err := s.repo.GetUserProfileByUID(ctx, user.UID, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, err
	}

	v, err := s.repo.AddPartnerType(ctx, profile.ID, name, partnerType)
	if !v || err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.AddPartnerTypeError(err)
	}

	return true, nil
}

// FindSupplierByID fetches a supplier by their id
func (s SupplierUseCasesImpl) FindSupplierByID(
	ctx context.Context,
	id string,
) (*profileutils.Supplier, error) {
	ctx, span := tracer.Start(ctx, "FindSupplierByID")
	defer span.End()

	return s.repo.GetSupplierProfileByID(ctx, id)
}

// FindSupplierByUID fetches a supplier by logged in user uid
func (s SupplierUseCasesImpl) FindSupplierByUID(ctx context.Context) (*profileutils.Supplier, error) {
	ctx, span := tracer.Start(ctx, "FindSupplierByUID")
	defer span.End()

	pr, err := s.profile.UserProfile(ctx)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}
	return s.repo.GetSupplierProfileByProfileID(ctx, pr.ID)
}

// SetUpSupplier performs initial account set up during onboarding
func (s SupplierUseCasesImpl) SetUpSupplier(
	ctx context.Context,
	accountType profileutils.AccountType,
) (*profileutils.Supplier, error) {
	ctx, span := tracer.Start(ctx, "SetUpSupplier")
	defer span.End()

	validAccountType := accountType.IsValid()
	if !validAccountType {
		return nil, fmt.Errorf("%v is not an allowed AccountType choice", accountType.String())
	}

	user, err := s.baseExt.GetLoggedInUser(ctx)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, fmt.Errorf("can't get user: %w", err)
	}
	isAuthorized, err := authorization.IsAuthorized(user, permission.SupplierCreate)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}
	if !isAuthorized {
		return nil, fmt.Errorf("user not authorized to access this resource")
	}

	profile, err := s.repo.GetUserProfileByUID(ctx, user.UID, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return nil, err
	}

	sup, err := s.repo.AddSupplierAccountType(ctx, profile.ID, accountType)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}
	if *sup.AccountType == profileutils.AccountTypeOrganisation ||
		*sup.AccountType == profileutils.AccountTypeIndividual {
		sup.OrganizationName = sup.SupplierName
		err := s.repo.UpdateSupplierProfile(ctx, profile.ID, sup)
		if err != nil {
			utils.RecordSpanError(span, err)
			return nil, err
		}
	}

	// go func(u string, pnt profileutils.PartnerType, acnt profileutils.AccountType) {
	// 	op := func() error {
	// 		return s.PublishKYCNudge(ctx, u, &pnt, &acnt)
	// 	}

	// 	if err := backoff.Retry(op, backoff.NewExponentialBackOff()); err != nil {
	// 		utils.RecordSpanError(span, err)
	// 		logrus.Error(err)
	// 	}
	// }(user.UID, sup.PartnerType, *sup.AccountType)

	go func() {
		pro := func() error {
			return s.engagement.ResolveDefaultNudgeByTitle(
				ctx,
				user.UID,
				feedlib.FlavourPro,
				PartnerAccountSetupNudgeTitle,
			)
		}
		if err := backoff.Retry(
			pro,
			backoff.NewExponentialBackOff(),
		); err != nil {
			utils.RecordSpanError(span, err)
			logrus.Error(err)
		}
	}()

	return sup, nil
}

// SuspendSupplier flips the active boolean on the erp partner from true to false
func (s SupplierUseCasesImpl) SuspendSupplier(ctx context.Context, suspensionReason *string) (bool, error) {
	ctx, span := tracer.Start(ctx, "SuspendSupplier")
	defer span.End()

	uid, err := s.baseExt.GetLoggedInUserUID(ctx)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.UserNotFoundError(err)
	}
	profile, err := s.repo.GetUserProfileByUID(ctx, uid, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return false, err
	}
	sup, err := s.FindSupplierByUID(ctx)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return false, err
	}
	sup.Active = false

	if err := s.repo.UpdateSupplierProfile(ctx, profile.ID, sup); err != nil {
		utils.RecordSpanError(span, err)
		return false, err
	}

	supplierEmailPayload := dto.EmailNotificationPayload{
		SupplierName: *profile.UserBioData.FirstName,
		SubjectTitle: supplierSuspensionEmailSubjectTitle,
		EmailBody:    *suspensionReason,
		EmailAddress: *profile.PrimaryEmailAddress,
		PrimaryPhone: *profile.PrimaryPhone,
	}
	err = s.engagement.NotifySupplierOnSuspension(ctx, supplierEmailPayload)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, err
	}

	return true, nil

}

// SupplierSetDefaultLocation updates the default location ot the supplier by the given location id
func (s SupplierUseCasesImpl) SupplierSetDefaultLocation(
	ctx context.Context,
	locationID string,
) (*profileutils.Supplier, error) {
	_, span := tracer.Start(ctx, "SupplierSetDefaultLocation")
	defer span.End()

	return nil, fmt.Errorf("unable to get location of id %v : %v", locationID, nil)
}

// SaveProfileNudge stages nudges published from this service. These nudges will be
// referenced later to support some specialized use-case. A nudge will be uniquely
// identified by its id and sequenceNumber
func (s *SupplierUseCasesImpl) SaveProfileNudge(
	ctx context.Context,
	nudge *feedlib.Nudge,
) error {
	ctx, span := tracer.Start(ctx, "SaveProfileNudge")
	defer span.End()

	return s.repo.StageProfileNudge(ctx, nudge)
}
