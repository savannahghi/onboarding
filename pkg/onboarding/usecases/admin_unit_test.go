package usecases_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/profileutils"
	erp "gitlab.slade360emr.com/go/commontools/accounting/pkg/usecases"
	erpMock "gitlab.slade360emr.com/go/commontools/accounting/pkg/usecases/mock"
	"gitlab.slade360emr.com/go/commontools/crm/pkg/infrastructure/services/hubspot"

	extMock "github.com/savannahghi/onboarding/pkg/onboarding/application/extension/mock"
	crmExt "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/crm"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/engagement"
	engagementMock "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/engagement/mock"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/messaging"
	messagingMock "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/messaging/mock"
	pubsubmessaging "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/pubsub"
	pubsubmessagingMock "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/pubsub/mock"
	"github.com/savannahghi/onboarding/pkg/onboarding/presentation/interactor"
	"github.com/savannahghi/onboarding/pkg/onboarding/repository"
	mockRepo "github.com/savannahghi/onboarding/pkg/onboarding/repository/mock"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases"
	adminSrv "github.com/savannahghi/onboarding/pkg/onboarding/usecases/admin"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases/ussd"
	hubspotRepo "gitlab.slade360emr.com/go/commontools/crm/pkg/infrastructure/database/fs"
	hubspotUsecases "gitlab.slade360emr.com/go/commontools/crm/pkg/usecases"
)

var fakeRepo mockRepo.FakeOnboardingRepository
var fakeBaseExt extMock.FakeBaseExtensionImpl
var fakePinExt extMock.PINExtensionImpl
var fakeEngagementSvs engagementMock.FakeServiceEngagement
var fakeMessagingSvc messagingMock.FakeServiceMessaging
var fakeEPRSvc erpMock.FakeServiceCommonTools
var fakePubSub pubsubmessagingMock.FakeServicePubSub

// InitializeFakeOnboaridingInteractor represents a fakeonboarding interactor
func InitializeFakeOnboardingInteractor() (*interactor.Interactor, error) {
	var r repository.OnboardingRepository = &fakeRepo
	var erpSvc erp.AccountingUsecase = &fakeEPRSvc
	var engagementSvc engagement.ServiceEngagement = &fakeEngagementSvs
	var messagingSvc messaging.ServiceMessaging = &fakeMessagingSvc
	var ext extension.BaseExtension = &fakeBaseExt
	var pinExt extension.PINExtension = &fakePinExt
	var ps pubsubmessaging.ServicePubSub = &fakePubSub

	// hubspot usecases
	hubspotService := hubspot.NewHubSpotService()
	hubspotfr, err := hubspotRepo.NewHubSpotFirebaseRepository(context.Background(), hubspotService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize hubspot crm repository: %w", err)
	}
	hubspotUsecases := hubspotUsecases.NewHubSpotUsecases(hubspotfr)
	crmExt := crmExt.NewCrmService(hubspotUsecases)
	profile := usecases.NewProfileUseCase(r, ext, engagementSvc, ps, crmExt)
	survey := usecases.NewSurveyUseCases(r, ext)
	supplier := usecases.NewSupplierUseCases(
		r, profile, erpSvc, engagementSvc, messagingSvc, ext, ps,
	)
	userpin := usecases.NewUserPinUseCase(r, profile, ext, pinExt, engagementSvc)
	su := usecases.NewSignUpUseCases(r, profile, userpin, supplier, ext, engagementSvc, ps)
	nhif := usecases.NewNHIFUseCases(r, profile, ext, engagementSvc)
	aitUssd := ussd.NewUssdUsecases(r, ext, profile, userpin, su, pinExt, ps, crmExt)
	adminSrv := adminSrv.NewService(ext)
	sms := usecases.NewSMSUsecase(r, ext)
	admin := usecases.NewAdminUseCases(r, engagementSvc, ext, userpin)
	agent := usecases.NewAgentUseCases(r, engagementSvc, ext, userpin)
	role := usecases.NewRoleUseCases(r, ext)

	i, err := interactor.NewOnboardingInteractor(
		r, profile, su, supplier,
		survey, userpin, erpSvc,
		engagementSvc, messagingSvc, nhif, ps, sms,
		aitUssd, agent, admin, adminSrv, crmExt,
		role,
	)
	if err != nil {
		return nil, fmt.Errorf("can't instantiate service : %w", err)
	}
	return i, nil

}

func TestAdminUseCaseImpl_FetchAdmins(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    []*dto.Admin
		wantErr bool
	}{
		{
			name: "success:_non_empty_list_of_user_admins",
			args: args{
				ctx: ctx,
			},
			want: []*dto.Admin{
				{
					ID:                  "c9d62c7e-93e5-44a6-b503-6fc159c1782f",
					PrimaryPhone:        interserviceclient.TestUserPhoneNumber,
					PrimaryEmailAddress: firebasetools.TestUserEmail,
				},
				{
					ID:                  "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					PrimaryPhone:        interserviceclient.TestUserPhoneNumber,
					PrimaryEmailAddress: firebasetools.TestUserEmail,
				},
			},
			wantErr: false,
		},
		{
			name: "success:_empty_list_of_user_admins",
			args: args{
				ctx: ctx,
			},
			want:    []*dto.Admin{},
			wantErr: false,
		},
		{
			name: "fail:error_fetching_list_of_user_admins",
			args: args{
				ctx: ctx,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "success:_non_empty_list_of_user_admins" {
				fakeRepo.ListUserProfilesFn = func(ctx context.Context, role profileutils.RoleType) ([]*profileutils.UserProfile, error) {
					p := interserviceclient.TestUserPhoneNumber
					e := firebasetools.TestUserEmail
					s := []*profileutils.UserProfile{
						{
							ID:                  "c9d62c7e-93e5-44a6-b503-6fc159c1782f",
							PrimaryPhone:        &p,
							PrimaryEmailAddress: &e,
							VerifiedUIDS:        []string{"f4f39af7-5b64-4c2f-91bd-42b3af315a4e"},
							Role:                profileutils.RoleTypeEmployee,
						},
						{
							ID:                  "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
							PrimaryPhone:        &p,
							PrimaryEmailAddress: &e,
							VerifiedUIDS:        []string{"c9d62c7e-93e5-44a6-b503-6fc159c1782f"},
							Role:                profileutils.RoleTypeEmployee,
						},
					}
					return s, nil
				}
			}
			if tt.name == "success:_empty_list_of_user_admins" {
				fakeRepo.ListUserProfilesFn = func(ctx context.Context, role profileutils.RoleType) ([]*profileutils.UserProfile, error) {
					return []*profileutils.UserProfile{}, nil
				}
			}
			if tt.name == "fail:error_fetching_list_of_user_admins" {
				fakeRepo.ListUserProfilesFn = func(ctx context.Context, role profileutils.RoleType) ([]*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("cannot fetch list of user profiles")
				}
			}
			got, err := i.Admin.FetchAdmins(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("AdminUseCaseImpl.FetchAdmins() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AdminUseCaseImpl.FetchAdmins() = %v, want %v", got, tt.want)
			}
		})
	}
}
