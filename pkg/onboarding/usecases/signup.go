package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/exceptions"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/utils"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure"
	"github.com/savannahghi/profileutils"
	"github.com/savannahghi/pubsubtools"
	"github.com/savannahghi/scalarutils"
)

// SignUpUseCases represents all the business logic involved in setting up a user
type SignUpUseCases interface {
	// VerifyPhoneNumber checks validity of a phone number by sending an OTP to it
	VerifyPhoneNumber(ctx context.Context, phone string, appID *string) (*profileutils.OtpResponse, error)

	// creates an account for the user, setting the provided phone number as the PRIMARY PHONE
	// NUMBER
	CreateUserByPhone(ctx context.Context, input *dto.SignUpInput) (*profileutils.UserResponse, error)

	// updates the user profile of the currently logged in user
	UpdateUserProfile(
		ctx context.Context,
		input *dto.UserProfileInput,
	) (*profileutils.UserProfile, error)

	// adds a new push token in the users profile if the push token does not exist
	RegisterPushToken(ctx context.Context, token string) (bool, error)

	CompleteSignup(ctx context.Context, flavour feedlib.Flavour) (bool, error)

	// removes a push token from the users profile
	RetirePushToken(ctx context.Context, token string) (bool, error)

	// fetches the phone numbers of a user for the purposes of recoverying an account.
	// the returned phone numbers should be masked
	GetUserRecoveryPhoneNumbers(
		ctx context.Context,
		phoneNumber string,
	) (*dto.AccountRecoveryPhonesResponse, error)

	// called to set the provided phone number as the PRIMARY PHONE NUMBER in the user profile of
	// the user
	// where the phone number is associated with.
	SetPhoneAsPrimary(ctx context.Context, phone, otp string) (bool, error)

	RemoveUserByPhoneNumber(ctx context.Context, phone string) error

	RegisterUser(ctx context.Context, input dto.RegisterUserInput) (*profileutils.UserProfile, error)
}

// SignUpUseCasesImpl represents usecase implementation object
type SignUpUseCasesImpl struct {
	infrastructure infrastructure.Infrastructure
	profileUsecase ProfileUseCase
	pinUsecase     UserPINUseCases
	baseExt        extension.BaseExtension
}

// NewSignUpUseCases returns a new a onboarding usecase
func NewSignUpUseCases(
	infrastructure infrastructure.Infrastructure,
	profile ProfileUseCase,
	pin UserPINUseCases,
	ext extension.BaseExtension,
) SignUpUseCases {
	return &SignUpUseCasesImpl{
		infrastructure: infrastructure,
		profileUsecase: profile,
		pinUsecase:     pin,
		baseExt:        ext,
	}
}

// VerifyPhoneNumber checks validity of a phone number by sending an OTP to it
func (s *SignUpUseCasesImpl) VerifyPhoneNumber(
	ctx context.Context,
	phone string,
	appID *string,
) (*profileutils.OtpResponse, error) {
	ctx, span := tracer.Start(ctx, "VerifyPhoneNumber")
	defer span.End()

	phoneNumber, err := s.baseExt.NormalizeMSISDN(phone)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.NormalizeMSISDNError(err)
	}
	// check if phone number exists
	exists, err := s.profileUsecase.CheckPhoneExists(ctx, *phoneNumber)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}
	// if phone exists return early
	if exists {
		return nil, exceptions.CheckPhoneNumberExistError()
	}
	// generate and send otp to the phone number
	otpResp, err := s.infrastructure.Engagement.GenerateAndSendOTP(ctx, *phoneNumber, appID)

	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.GenerateAndSendOTPError(err)
	}
	// return the generated otp
	return otpResp, nil
}

// CreateUserByPhone creates an account for the user, setting the provided phone number as the
// PRIMARY PHONE NUMBER
func (s *SignUpUseCasesImpl) CreateUserByPhone(
	ctx context.Context,
	input *dto.SignUpInput,
) (*profileutils.UserResponse, error) {
	ctx, span := tracer.Start(ctx, "CreateUserByPhone")
	defer span.End()

	userData, err := utils.ValidateSignUpInput(input)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}
	verified, err := s.infrastructure.Engagement.VerifyOTP(
		ctx,
		*userData.PhoneNumber,
		*userData.OTP,
	)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.VerifyOTPError(err)
	}

	if !verified {
		return nil, exceptions.VerifyOTPError(nil)
	}

	// get or create user via their phone number
	user, err := s.infrastructure.Database.GetOrCreatePhoneNumberUser(ctx, *userData.PhoneNumber)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	// create a user profile
	profile, err := s.infrastructure.Database.CreateUserProfile(
		ctx,
		*userData.PhoneNumber,
		user.UID,
	)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}
	// generate auth credentials
	auth, err := s.infrastructure.Database.GenerateAuthCredentials(
		ctx,
		*userData.PhoneNumber,
		profile,
	)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}
	// save the user pin
	_, err = s.pinUsecase.SetUserPIN(
		ctx,
		*userData.PIN,
		profile.ID,
	)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}
	// set the user default communications settings
	defaultCommunicationSetting := true
	comms, err := s.infrastructure.Database.SetUserCommunicationsSettings(
		ctx,
		profile.ID,
		&defaultCommunicationSetting,
		&defaultCommunicationSetting,
		&defaultCommunicationSetting,
		&defaultCommunicationSetting,
	)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	// get navigation actions
	roles, err := s.infrastructure.Database.GetRolesByIDs(ctx, profile.Roles)
	if err != nil {
		return nil, err
	}

	navActions, err := utils.GetUserNavigationActions(ctx, *profile, *roles)
	if err != nil {
		return nil, err
	}

	return &profileutils.UserResponse{
		Profile:               profile,
		CommunicationSettings: comms,
		Auth:                  *auth,
		NavActions:            utils.NewActionsMapper(ctx, navActions),
	}, nil
}

// UpdateUserProfile  updates the user profile of the currently logged in user
func (s *SignUpUseCasesImpl) UpdateUserProfile(
	ctx context.Context,
	input *dto.UserProfileInput,
) (*profileutils.UserProfile, error) {
	ctx, span := tracer.Start(ctx, "UpdateUserProfile")
	defer span.End()

	// get the old user profile
	pr, err := s.profileUsecase.UserProfile(ctx)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return nil, err
	}

	if input.PhotoUploadID != nil {
		if err := s.profileUsecase.UpdatePhotoUploadID(ctx, *input.PhotoUploadID); err != nil {
			utils.RecordSpanError(span, err)
			return nil, err
		}
	}

	if err := s.profileUsecase.UpdateBioData(ctx, profileutils.BioData{
		FirstName: func(n *string) *string {
			if n != nil {
				return n
			}
			return pr.UserBioData.FirstName
		}(input.FirstName),
		LastName: func(n *string) *string {
			if n != nil {
				return n
			}
			return pr.UserBioData.LastName
		}(input.LastName),
		DateOfBirth: func(n *scalarutils.Date) *scalarutils.Date {
			if n != nil {
				return n
			}
			return pr.UserBioData.DateOfBirth
		}(input.DateOfBirth),
		Gender: func(n *enumutils.Gender) enumutils.Gender {
			if n != nil {
				return *n
			}
			return pr.UserBioData.Gender
		}(input.Gender),
	}); err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}
	return s.profileUsecase.UserProfile(ctx)
}

// RegisterPushToken adds a new push token in the users profile if the push token does not exist
func (s *SignUpUseCasesImpl) RegisterPushToken(ctx context.Context, token string) (bool, error) {
	ctx, span := tracer.Start(ctx, "RegisterPushToken")
	defer span.End()

	if len(token) < 5 {
		return false, exceptions.InValidPushTokenLengthError()
	}
	if err := s.profileUsecase.UpdatePushTokens(ctx, token, false); err != nil {
		utils.RecordSpanError(span, err)
		return false, err
	}
	return true, nil
}

// CompleteSignup is not implemented but maintains backward compatibility
//  This API is only valid for `BEWELL CONSUMER`
func (s *SignUpUseCasesImpl) CompleteSignup(
	ctx context.Context,
	flavour feedlib.Flavour,
) (bool, error) {
	_, span := tracer.Start(ctx, "CompleteSignup")
	defer span.End()

	return true, nil
}

// RetirePushToken removes a push token from the users profile
func (s *SignUpUseCasesImpl) RetirePushToken(ctx context.Context, token string) (bool, error) {
	ctx, span := tracer.Start(ctx, "RetirePushToken")
	defer span.End()

	if len(token) < 5 {
		return false, exceptions.InValidPushTokenLengthError()
	}
	if err := s.profileUsecase.UpdatePushTokens(ctx, token, true); err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.InternalServerError(err)
	}
	return true, nil
}

// GetUserRecoveryPhoneNumbers fetches the phone numbers of a user for the purposes of recoverying
// an account.
func (s *SignUpUseCasesImpl) GetUserRecoveryPhoneNumbers(
	ctx context.Context,
	phone string,
) (*dto.AccountRecoveryPhonesResponse, error) {
	ctx, span := tracer.Start(ctx, "GetUserRecoveryPhoneNumbers")
	defer span.End()

	phoneNumber, err := s.baseExt.NormalizeMSISDN(phone)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.NormalizeMSISDNError(err)
	}

	pr, err := s.infrastructure.Database.GetUserProfileByPhoneNumber(ctx, *phoneNumber, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return nil, err
	}
	// cherrypick the phone numbers and mask them
	phones := func(p *profileutils.UserProfile) []string {
		phs := []string{}
		phs = append(phs, *p.PrimaryPhone)
		phs = append(phs, p.SecondaryPhoneNumbers...)
		return phs

	}(pr)
	masked := s.profileUsecase.MaskPhoneNumbers(phones)
	return &dto.AccountRecoveryPhonesResponse{
		MaskedPhoneNumbers:   masked,
		UnMaskedPhoneNumbers: phones,
	}, nil
}

// SetPhoneAsPrimary called to set the provided phone number as the PRIMARY PHONE NUMBER in the user
// profile of the user
// where the phone number is associated with.
func (s *SignUpUseCasesImpl) SetPhoneAsPrimary(
	ctx context.Context,
	phone, otp string,
) (bool, error) {
	ctx, span := tracer.Start(ctx, "SetPhoneAsPrimary")
	defer span.End()

	phoneNumber, err := s.baseExt.NormalizeMSISDN(phone)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.NormalizeMSISDNError(err)
	}

	err = s.profileUsecase.SetPrimaryPhoneNumber(ctx, *phoneNumber, otp, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, err
	}
	return true, nil
}

// RemoveUserByPhoneNumber removes the record of a user using the provided phone number. This method
// will ONLY be called
// in testing environment.
func (s *SignUpUseCasesImpl) RemoveUserByPhoneNumber(ctx context.Context, phone string) error {
	ctx, span := tracer.Start(ctx, "RemoveUserByPhoneNumber")
	defer span.End()

	phoneNumber, err := s.baseExt.NormalizeMSISDN(phone)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.NormalizeMSISDNError(err)
	}
	return s.infrastructure.Database.PurgeUserByPhoneNumber(ctx, *phoneNumber)
}

// RegisterUser creates a new userprofile
func (s *SignUpUseCasesImpl) RegisterUser(ctx context.Context, input dto.RegisterUserInput) (*profileutils.UserProfile, error) {
	ctx, span := tracer.Start(ctx, "RegisterUser")
	defer span.End()

	uid, err := s.baseExt.GetLoggedInUserUID(ctx)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.UserNotFoundError(err)
	}

	// create a user profile
	//make createdByID optional only if the profile of the creating user is found
	profile, err := s.infrastructure.Database.GetUserProfileByUID(ctx, uid, false)
	var profileID string
	if err == nil {
		profileID = profile.ID
	}

	phoneNumber, err := s.baseExt.NormalizeMSISDN(*input.PhoneNumber)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.NormalizeMSISDNError(err)
	}

	timestamp := time.Now().In(pubsubtools.TimeLocation)

	userProfile := profileutils.UserProfile{
		PrimaryEmailAddress: input.Email,
		UserBioData: profileutils.BioData{
			FirstName:   input.FirstName,
			LastName:    input.LastName,
			Gender:      enumutils.Gender(*input.Gender),
			DateOfBirth: input.DateOfBirth,
		},
		CreatedByID: &profileID,
		Created:     &timestamp,
		Roles:       input.RoleIDs,
	}

	createdProfile, err := s.infrastructure.Database.CreateDetailedUserProfile(ctx, *phoneNumber, userProfile)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	// set the user default communications settings
	defaultCommunicationSetting := true
	_, err = s.infrastructure.Database.SetUserCommunicationsSettings(
		ctx,
		createdProfile.ID,
		&defaultCommunicationSetting,
		&defaultCommunicationSetting,
		&defaultCommunicationSetting,
		&defaultCommunicationSetting,
	)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	otp, err := s.pinUsecase.SetUserTempPIN(ctx, createdProfile.ID)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	message := input.WelcomeMessage
	if message == nil {
		message = &domain.WelcomeMessage
	}

	formartedMessage := fmt.Sprintf(*message, *input.FirstName, otp)

	if err := s.infrastructure.Engagement.SendSMS(ctx, []string{*phoneNumber}, formartedMessage); err != nil {
		return nil, fmt.Errorf("unable to send consumer registration message: %w", err)
	}

	return createdProfile, nil
}
