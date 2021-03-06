package usecases_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
	"github.com/savannahghi/converterandformatter"
	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/exceptions"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases"
	"github.com/savannahghi/profileutils"
	"github.com/savannahghi/scalarutils"
)

func TestProfileUseCaseImpl_UpdateVerifiedUIDS(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	type args struct {
		ctx  context.Context
		uids []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:_update_profile_uids",
			args: args{
				ctx: ctx,
				uids: []string{
					"f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					"5d46d3bd-a482-4787-9b87-3c94510c8b53",
				},
			},
			wantErr: false,
		},

		{
			name: "invalid:_unable_to_get_logged_in_user",
			args: args{
				ctx: ctx,
				uids: []string{
					"f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					"5d46d3bd-a482-4787-9b87-3c94510c8b53",
				},
			},
			wantErr: true,
		},

		{
			name: "invalid:_unable_to_get_profile_of_logged_in_user",
			args: args{
				ctx: ctx,
				uids: []string{
					"f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					"5d46d3bd-a482-4787-9b87-3c94510c8b53",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:_update_profile_uids" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					}, nil
				}
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}
				fakeInfraRepo.UpdateVerifiedUIDSFn = func(ctx context.Context, id string, uids []string) error {
					return nil
				}
			}

			if tt.name == "invalid:_unable_to_get_logged_in_user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("unable to get logged user")
				}
			}

			if tt.name == "invalid:_unable_to_get_profile_of_logged_in_user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get profile")
				}
			}

			err := i.UpdateVerifiedUIDS(tt.args.ctx, tt.args.uids)

			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}

		})
	}
}

func TestProfileUseCaseImpl_UpdateSecondaryEmailAddresses(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	type args struct {
		ctx            context.Context
		emailAddresses []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:_update_profile_secondary_email",
			args: args{
				ctx:            ctx,
				emailAddresses: []string{"me4@gmail.com", "kalulu@gmail.com"},
			},
			wantErr: false,
		},
		{
			name: "invalid:_update_profile_secondary_email", // no primary email
			args: args{
				ctx:            ctx,
				emailAddresses: []string{"me4@gmail.com", "kalulu@gmail.com"},
			},
			wantErr: true,
		},
		{
			name: "invalid:_unable_to_get_logged_in_user",
			args: args{
				ctx:            ctx,
				emailAddresses: []string{"me4@gmail.com", "kalulu@gmail.com"},
			},
			wantErr: true,
		},
		{
			name: "invalid:_unable_to_get_profile_of_logged_in_user",
			args: args{
				ctx:            ctx,
				emailAddresses: []string{"me4@gmail.com", "kalulu@gmail.com"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:_update_profile_secondary_email" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					email := firebasetools.TestUserEmail
					return &profileutils.UserProfile{
						ID:                  "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
						PrimaryEmailAddress: &email,
					}, nil
				}
				fakeInfraRepo.UpdateSecondaryEmailAddressesFn = func(ctx context.Context, id string, uids []string) error {
					return nil
				}

				fakeInfraRepo.CheckIfEmailExistsFn = func(ctx context.Context, email string) (bool, error) {
					return false, nil
				}
			}

			if tt.name == "invalid:_update_profile_secondary_email" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					}, nil
				}
				fakeInfraRepo.UpdateSecondaryEmailAddressesFn = func(ctx context.Context, id string, uids []string) error {
					return fmt.Errorf("unable to update secondary email")
				}
			}

			if tt.name == "invalid:_unable_to_get_logged_in_user" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "", fmt.Errorf("unable to get logged user")
				}
			}

			if tt.name == "invalid:_unable_to_get_profile_of_logged_in_user" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get profile")
				}
			}

			err := i.UpdateSecondaryEmailAddresses(tt.args.ctx, tt.args.emailAddresses)

			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_UpdateUserName(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	type args struct {
		ctx      context.Context
		userName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:_update_name_succeeds",
			args: args{
				ctx:      ctx,
				userName: "kamau",
			},
			wantErr: false,
		},
		{
			name: "invalid:_unable_to_get_logged_in_user",
			args: args{
				ctx:      ctx,
				userName: "mwas",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:_update_name_succeeds" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					}, nil
				}
				fakeInfraRepo.UpdateUserNameFn = func(ctx context.Context, id string, phoneNumber string) error {
					return nil
				}
			}

			if tt.name == "invalid:_unable_to_get_logged_in_user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("unable to get logged user")
				}
			}
			err := i.UpdateUserName(tt.args.ctx, tt.args.userName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_UpdateVerifiedIdentifiers(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	type args struct {
		ctx         context.Context
		identifiers []profileutils.VerifiedIdentifier
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:_update_name_succeeds",
			args: args{
				ctx: ctx,
				identifiers: []profileutils.VerifiedIdentifier{
					{
						UID:           "a4f39af7-5b64-4c2f-91bd-42b3af315a5h",
						LoginProvider: "Facebook",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid:_unable_to_get_logged_in_user",
			args: args{
				ctx: ctx,
				identifiers: []profileutils.VerifiedIdentifier{
					{
						UID:           "j4f39af7-5b64-4c2f-91bd-42b3af225a5c",
						LoginProvider: "Phone",
					},
				},
			},
			wantErr: true,
		},

		{
			name: "invalid:_unable_to_get_profile_of_logged_in_user",
			args: args{
				ctx: ctx,
				identifiers: []profileutils.VerifiedIdentifier{
					{
						UID:           "p4f39af7-5b64-4c2f-91bd-42b3af315a5c",
						LoginProvider: "Google",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:_update_name_succeeds" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					}, nil
				}
				fakeInfraRepo.UpdateVerifiedIdentifiersFn = func(ctx context.Context, id string, identifiers []profileutils.VerifiedIdentifier) error {
					return nil
				}
			}

			if tt.name == "invalid:_unable_to_get_logged_in_user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("unable to get logged user")
				}
			}

			if tt.name == "invalid:_unable_to_get_profile_of_logged_in_user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get profile")
				}
			}

			err := i.UpdateVerifiedIdentifiers(tt.args.ctx, tt.args.identifiers)
			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_UpdatePrimaryEmailAddress(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	primaryEmail := "me@gmail.com"
	primaryPhone := "0711223344"

	type args struct {
		ctx          context.Context
		emailAddress string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:_update_email_succeeds",
			args: args{
				ctx:          ctx,
				emailAddress: primaryEmail,
			},
			wantErr: false,
		},
		{
			name: "invalid:_unable_to_get_logged_in_user",
			args: args{
				ctx:          ctx,
				emailAddress: "kalulu@gmail.com",
			},
			wantErr: true,
		},
		{
			name: "invalid:_unable_to_get_profile_of_logged_in_user",
			args: args{
				ctx:          ctx,
				emailAddress: "juha@gmail.com",
			},
			wantErr: true,
		},
		{
			name: "invalid:_unable_to_update_primary_email_address",
			args: args{
				ctx:          ctx,
				emailAddress: "juha@gmail.com",
			},
			wantErr: true,
		},
		{
			name: "invalid:_unable_to_update_secondary_email_address",
			args: args{
				ctx:          ctx,
				emailAddress: "juha@gmail.com",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "valid:_update_email_succeeds" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:                  "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
						PrimaryEmailAddress: &primaryEmail,
						PrimaryPhone:        &primaryPhone,
					}, nil
				}
				fakeInfraRepo.UpdatePrimaryEmailAddressFn = func(ctx context.Context, id string, emailAddress string) error {
					return nil
				}

				fakeInfraRepo.UpdateSecondaryEmailAddressesFn = func(ctx context.Context, id string, emailAddresses []string) error {
					return nil
				}
			}

			if tt.name == "invalid:_unable_to_update_primary_email_address" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:                  "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
						PrimaryEmailAddress: &primaryEmail,
						PrimaryPhone:        &primaryPhone,
					}, nil
				}
				fakeInfraRepo.UpdatePrimaryEmailAddressFn = func(ctx context.Context, id string, emailAddress string) error {
					return fmt.Errorf("unable to update primary address")
				}
			}

			if tt.name == "invalid:_unable_to_update_secondary_email_address" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:                  "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
						PrimaryEmailAddress: &primaryEmail,
						PrimaryPhone:        &primaryPhone,
						SecondaryEmailAddresses: []string{
							"", "lulu@gmail.com",
						},
					}, nil
				}
				fakeInfraRepo.UpdatePrimaryEmailAddressFn = func(ctx context.Context, id string, emailAddress string) error {
					return nil
				}

				fakeInfraRepo.UpdateSecondaryEmailAddressesFn = func(ctx context.Context, id string, emailAddresses []string) error {
					return fmt.Errorf("unable to update secondary email")
				}
			}

			if tt.name == "invalid:_unable_to_get_logged_in_user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("unable to get logged user")
				}
			}

			if tt.name == "invalid:_unable_to_get_profile_of_logged_in_user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get profile")
				}
			}

			err := i.UpdatePrimaryEmailAddress(tt.args.ctx, tt.args.emailAddress)
			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}

		})
	}
}

func TestProfileUseCaseImpl_SetPrimaryEmailAddress(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	primaryEmail := "juha@gmail.com"
	phone := gofakeit.Phone()

	type args struct {
		ctx          context.Context
		emailAddress string
		otp          string
		UID          string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:_set_primary_address_succeeds",
			args: args{
				ctx:          ctx,
				emailAddress: primaryEmail,
				otp:          "689552",
			},
			wantErr: false,
		},
		{
			name: "invalid:_failed_to_get_logged_in_uid",
			args: args{
				ctx:          ctx,
				emailAddress: "kichwa@gmail.com",
				otp:          "453852",
			},
			wantErr: true,
		},
		{
			name: "invalid:_verify_otp_fails",
			args: args{
				ctx:          ctx,
				emailAddress: "kichwa@gmail.com",
				otp:          "453852",
			},
			wantErr: true,
		},
		{
			name: "invalid:_verify_otp_returns_false",
			args: args{
				ctx:          ctx,
				emailAddress: "kalu@gmail.com",
				otp:          "235789",
			},
			wantErr: true,
		},
		{
			name: "invalid:_update_primary_address_fails",
			args: args{
				ctx:          ctx,
				emailAddress: "mwendwapole@gmail.com",
				otp:          "897523",
			},
			wantErr: true,
		},
		{
			name: "invalid:_resolving_the_consumer_nudge_fails",
			args: args{
				ctx:          ctx,
				emailAddress: "mwendwapole@gmail.com",
				otp:          "897523",
			},
			wantErr: false,
		},
		{
			name: "invalid:_resolving_the_pro_nudge_fails",
			args: args{
				ctx:          ctx,
				emailAddress: "mwendwapole@gmail.com",
				otp:          "897523",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:_set_primary_address_succeeds" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return uuid.New().String(), nil
				}

				fakeEngagementSvs.VerifyEmailOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}

				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:                  uuid.New().String(),
						PrimaryEmailAddress: &primaryEmail,
						PrimaryPhone:        &phone,
					}, nil
				}

				fakeInfraRepo.UpdatePrimaryEmailAddressFn = func(ctx context.Context, id string, emailAddress string) error {
					return nil
				}

				fakeInfraRepo.UpdateSecondaryEmailAddressesFn = func(ctx context.Context, id string, emailAddress []string) error {
					return nil
				}

				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID: uuid.NewString(),
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:                  uuid.New().String(),
						PrimaryEmailAddress: &primaryEmail,
						PrimaryPhone:        &phone,
					}, nil
				}

				fakePubSub.TopicIDsFn = func() []string {
					return []string{uuid.New().String()}
				}

				fakePubSub.AddPubSubNamespaceFn = func(topicName string) string {
					return uuid.New().String()
				}

				fakePubSub.PublishToPubsubFn = func(ctx context.Context, topicID string, payload []byte) error {
					return nil
				}

				fakeEngagementSvs.ResolveDefaultNudgeByTitleFn = func(
					ctx context.Context,
					UID string,
					flavour feedlib.Flavour,
					nudgeTitle string,
				) error {
					return nil
				}

				// Resolve the second nudge
				fakeEngagementSvs.ResolveDefaultNudgeByTitleFn = func(
					ctx context.Context,
					UID string,
					flavour feedlib.Flavour,
					nudgeTitle string,
				) error {
					return nil
				}
			}

			if tt.name == "invalid:_failed_to_get_logged_in_uid" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("an error has occurred")
				}
			}

			if tt.name == "invalid:_verify_otp_fails" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:                  uuid.New().String(),
						PrimaryEmailAddress: &primaryEmail,
					}, nil
				}
				fakeEngagementSvs.VerifyEmailOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return false, fmt.Errorf("unable to verify email otp")
				}
			}

			if tt.name == "invalid:_verify_otp_returns_false" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:                  uuid.New().String(),
						PrimaryEmailAddress: &primaryEmail,
					}, nil
				}
				fakeEngagementSvs.VerifyEmailOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return false, nil
				}
			}

			if tt.name == "invalid:_update_primary_address_fails" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:                  uuid.New().String(),
						PrimaryEmailAddress: &primaryEmail,
					}, nil
				}
				fakeEngagementSvs.VerifyEmailOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.UpdatePrimaryEmailAddressFn = func(ctx context.Context, id string, emailAddress string) error {
					return fmt.Errorf("unable to update primary email")
				}
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "", fmt.Errorf("unable to get loggedin user")
				}
			}

			if tt.name == "invalid:_resolving_the_consumer_nudge_fails" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:                  uuid.New().String(),
						PrimaryEmailAddress: &primaryEmail,
						PrimaryPhone:        &phone,
					}, nil
				}
				fakeEngagementSvs.VerifyEmailOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.UpdatePrimaryEmailAddressFn = func(ctx context.Context, id string, emailAddress string) error {
					return nil
				}
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return uuid.New().String(), nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:                  uuid.New().String(),
						PrimaryEmailAddress: &primaryEmail,
						PrimaryPhone:        &phone,
					}, nil
				}
				fakeInfraRepo.UpdateSecondaryEmailAddressesFn = func(ctx context.Context, id string, emailAddresses []string) error {
					return nil
				}

				fakePubSub.TopicIDsFn = func() []string {
					return []string{uuid.New().String()}
				}

				fakePubSub.AddPubSubNamespaceFn = func(topicName string) string {
					return uuid.New().String()
				}

				fakePubSub.PublishToPubsubFn = func(ctx context.Context, topicID string, payload []byte) error {
					return nil
				}

				fakeEngagementSvs.ResolveDefaultNudgeByTitleFn = func(
					ctx context.Context,
					UID string,
					flavour feedlib.Flavour,
					nudgeTitle string,
				) error {
					return fmt.Errorf("an error occurred")
				}
			}

			if tt.name == "invalid:_resolving_the_pro_nudge_fails" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:                  uuid.New().String(),
						PrimaryEmailAddress: &primaryEmail,
						PrimaryPhone:        &phone,
					}, nil
				}
				fakeEngagementSvs.VerifyEmailOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.UpdatePrimaryEmailAddressFn = func(ctx context.Context, id string, emailAddress string) error {
					return nil
				}
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return uuid.New().String(), nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:                  uuid.New().String(),
						PrimaryEmailAddress: &primaryEmail,
						PrimaryPhone:        &phone,
					}, nil
				}
				fakeInfraRepo.UpdateSecondaryEmailAddressesFn = func(ctx context.Context, id string, emailAddresses []string) error {
					return nil
				}

				fakePubSub.TopicIDsFn = func() []string {
					return []string{uuid.New().String()}
				}

				fakePubSub.AddPubSubNamespaceFn = func(topicName string) string {
					return uuid.New().String()
				}

				fakePubSub.PublishToPubsubFn = func(ctx context.Context, topicID string, payload []byte) error {
					return nil
				}

				fakeEngagementSvs.ResolveDefaultNudgeByTitleFn = func(
					ctx context.Context,
					UID string,
					flavour feedlib.Flavour,
					nudgeTitle string,
				) error {
					return nil
				}

				// Resolve the second nudge
				fakeEngagementSvs.ResolveDefaultNudgeByTitleFn = func(
					ctx context.Context,
					UID string,
					flavour feedlib.Flavour,
					nudgeTitle string,
				) error {
					return fmt.Errorf("an error occurred")
				}
			}

			err := i.SetPrimaryEmailAddress(
				tt.args.ctx,
				tt.args.emailAddress,
				tt.args.otp,
			)
			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}

		})
	}
}

func TestProfileUseCaseImpl_UpdatePermissions(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	type args struct {
		ctx   context.Context
		perms []profileutils.PermissionType
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid: successfully updates permissions",
			args: args{
				ctx:   ctx,
				perms: []profileutils.PermissionType{profileutils.PermissionTypeSuperAdmin},
			},
			wantErr: false,
		},
		{
			name: "invalid: get logged in user uid fails",
			args: args{
				ctx:   ctx,
				perms: []profileutils.PermissionType{profileutils.PermissionTypeSuperAdmin},
			},
			wantErr: true,
		},
		{
			name: "invalid: get user profile by UID fails",
			args: args{
				ctx:   ctx,
				perms: []profileutils.PermissionType{profileutils.PermissionTypeSuperAdmin},
			},
			wantErr: true,
		},
		{
			name: "invalid: update permissions fails",
			args: args{
				ctx:   ctx,
				perms: []profileutils.PermissionType{profileutils.PermissionTypeSuperAdmin},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "valid: successfully updates permissions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: "12334"}, nil
				}
				fakeInfraRepo.UpdatePermissionsFn = func(ctx context.Context, id string, perms []profileutils.PermissionType) error {
					return nil
				}
			}

			if tt.name == "invalid: get logged in user uid fails" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("failed to get loggeg in user UID")
				}
			}

			if tt.name == "invalid: get user profile by UID fails" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("failed to get user profile by UID")
				}
			}

			if tt.name == "invalid: update permissions fails" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: "12334"}, nil
				}
				fakeInfraRepo.UpdatePermissionsFn = func(ctx context.Context, id string, perms []profileutils.PermissionType) error {
					return fmt.Errorf("unable to update permissions")
				}
			}

			err := i.UpdatePermissions(tt.args.ctx, tt.args.perms)
			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_AddRoleToUser(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	type args struct {
		ctx   context.Context
		phone string
		role  profileutils.RoleType
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid: successfully updates role",
			args: args{
				ctx:   ctx,
				phone: "+254721123123",
				role:  profileutils.RoleTypeEmployee,
			},
			wantErr: false,
		},
		{
			name: "invalid: get profile by primary phone number failed",
			args: args{
				ctx:   ctx,
				phone: "+254721123123",
				role:  profileutils.RoleTypeEmployee,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid: successfully updates role" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}
				fakeInfraRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "123",
						PrimaryPhone: &phoneNumber,
						SecondaryPhoneNumbers: []string{
							"0721521456", "0721856741",
						},
					}, nil
				}

				fakeInfraRepo.UpdateRoleFn = func(ctx context.Context, id string, role profileutils.RoleType) error {
					return nil
				}
			}

			if tt.name == "invalid: get profile by primary phone number failed" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}
				fakeBaseExt.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phone string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("UserProfile matching PhoneNumber not found")
				}
				fakeInfraRepo.UpdateRoleFn = func(ctx context.Context, id string, role profileutils.RoleType) error {
					return fmt.Errorf("User Roles not updated")
				}
			}

			err := i.AddRoleToUser(tt.args.ctx, tt.args.phone, tt.args.role)

			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_RemoveRoleToUser(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	type args struct {
		ctx   context.Context
		phone string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:successfully_removed_role",
			args: args{
				ctx:   ctx,
				phone: "+254721123123",
			},
			wantErr: false,
		},
		{
			name: "invalid:failed_to_remove_role_invalid_profile",
			args: args{
				ctx:   ctx,
				phone: "+254721123123",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:successfully_removed_role" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}
				fakeInfraRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "123",
						PrimaryPhone: &phoneNumber,
						SecondaryPhoneNumbers: []string{
							"0721521456", "0721856741",
						},
					}, nil
				}

				fakeInfraRepo.UpdateRoleFn = func(ctx context.Context, id string, role profileutils.RoleType) error {
					return nil
				}
			}

			if tt.name == "invalid:failed_to_remove_role_invalid_profile" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}
				fakeBaseExt.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phone string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("UserProfile matching PhoneNumber not found")
				}
				fakeInfraRepo.UpdateRoleFn = func(ctx context.Context, id string, role profileutils.RoleType) error {
					return fmt.Errorf("User Roles not updated")
				}
			}

			err := i.RemoveRoleToUser(tt.args.ctx, tt.args.phone)

			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_GetUserProfileAttributes(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	type args struct {
		ctx       context.Context
		UIDs      []string
		attribute string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]string
		wantErr bool
	}{
		{
			name: "valid:_get_user_profile_emails",
			args: args{
				ctx:       ctx,
				UIDs:      []string{uuid.New().String()},
				attribute: usecases.EmailsAttribute,
			},
			wantErr: false,
		},
		{
			name: "valid:_get_user_profile_phone_numbers",
			args: args{
				ctx:       ctx,
				UIDs:      []string{uuid.New().String()},
				attribute: usecases.PhoneNumbersAttribute,
			},
			wantErr: false,
		},
		{
			name: "valid:_get_user_profile_fcm_tokens",
			args: args{
				ctx:       ctx,
				UIDs:      []string{uuid.New().String()},
				attribute: usecases.FCMTokensAttribute,
			},
			wantErr: false,
		},
		{
			name: "invalid:_failed_get_user_profile_attribute",
			args: args{
				ctx:       ctx,
				UIDs:      []string{uuid.New().String()},
				attribute: "not-an-attribute",
			},
			wantErr: true,
		},
		{
			name: "invalid:_failed_get_user_profile",
			args: args{
				ctx:       ctx,
				UIDs:      []string{uuid.New().String()},
				attribute: usecases.FCMTokensAttribute,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:_get_user_profile_emails" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					email := converterandformatter.GenerateRandomEmail()
					return &profileutils.UserProfile{
						PrimaryEmailAddress: &email,
						SecondaryEmailAddresses: []string{
							converterandformatter.GenerateRandomEmail(),
						},
					}, nil
				}
			}

			if tt.name == "valid:_get_user_profile_phone_numbers" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					phone := interserviceclient.TestUserPhoneNumber
					return &profileutils.UserProfile{
						PrimaryPhone:          &phone,
						SecondaryPhoneNumbers: []string{"+254700000000"},
					}, nil
				}
			}

			if tt.name == "valid:_get_user_profile_fcm_tokens" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						PushTokens: []string{uuid.New().String()},
					}, nil
				}
			}

			if tt.name == "invalid:_failed_get_user_profile" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					email := converterandformatter.GenerateRandomEmail()
					phone := interserviceclient.TestUserPhoneNumber
					return &profileutils.UserProfile{
						PrimaryEmailAddress: &email,
						SecondaryEmailAddresses: []string{
							converterandformatter.GenerateRandomEmail(),
						},
						PrimaryPhone:          &phone,
						SecondaryPhoneNumbers: []string{"+254700000000"},
						PushTokens:            []string{uuid.New().String()},
					}, nil
				}
			}

			if tt.name == "invalid:_failed_get_user_profile" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					return nil, exceptions.ProfileNotFoundError(
						fmt.Errorf("user profile not found"),
					)
				}
			}

			attribute, err := i.GetUserProfileAttributes(
				tt.args.ctx,
				tt.args.UIDs,
				tt.args.attribute,
			)

			if tt.wantErr && attribute != nil {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}

			if !tt.wantErr && attribute == nil {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_ConfirmedEmailAddresses(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}
	type args struct {
		ctx  context.Context
		UIDs []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:_get_confirmed_emails",
			args: args{
				ctx:  ctx,
				UIDs: []string{uuid.New().String()},
			},
			wantErr: false,
		},
		{
			name: "invalid:_failed_get_user_profile",
			args: args{
				ctx:  ctx,
				UIDs: []string{uuid.New().String()},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:_get_confirmed_emails" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					email := converterandformatter.GenerateRandomEmail()
					return &profileutils.UserProfile{
						PrimaryEmailAddress: &email,
						SecondaryEmailAddresses: []string{
							converterandformatter.GenerateRandomEmail(),
						},
					}, nil
				}
			}

			if tt.name == "invalid:_failed_get_user_profile" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					return nil, exceptions.ProfileNotFoundError(
						fmt.Errorf("user profile not found"),
					)
				}
			}

			confirmedEmails, err := i.ConfirmedEmailAddresses(
				tt.args.ctx,
				tt.args.UIDs,
			)
			if tt.wantErr && confirmedEmails != nil {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}

			if !tt.wantErr && confirmedEmails == nil {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_ConfirmedPhoneNumbers(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}
	type args struct {
		ctx  context.Context
		UIDs []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:_get_confirmed_emails",
			args: args{
				ctx:  ctx,
				UIDs: []string{uuid.New().String()},
			},
			wantErr: false,
		},
		{
			name: "invalid:_failed_get_user_profile",
			args: args{
				ctx:  ctx,
				UIDs: []string{uuid.New().String()},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:_get_confirmed_emails" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					phone := interserviceclient.TestUserPhoneNumber
					return &profileutils.UserProfile{
						PrimaryPhone:          &phone,
						SecondaryPhoneNumbers: []string{"+254700000000"},
					}, nil
				}
			}

			if tt.name == "invalid:_failed_get_user_profile" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					return nil, exceptions.ProfileNotFoundError(
						fmt.Errorf("user profile not found"),
					)
				}
			}

			confirmedEmails, err := i.ConfirmedPhoneNumbers(
				tt.args.ctx,
				tt.args.UIDs,
			)
			if tt.wantErr && confirmedEmails != nil {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}

			if !tt.wantErr && confirmedEmails == nil {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_validFCM(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}
	type args struct {
		ctx  context.Context
		UIDs []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:_valid_fcm_tokens",
			args: args{
				ctx:  ctx,
				UIDs: []string{uuid.New().String()},
			},
			wantErr: false,
		},
		{
			name: "invalid:_failed_get_user_profile",
			args: args{
				ctx:  ctx,
				UIDs: []string{uuid.New().String()},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:_valid_fcm_tokens" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						PushTokens: []string{uuid.New().String()},
					}, nil
				}
			}

			if tt.name == "invalid:_failed_get_user_profile" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					return nil, exceptions.ProfileNotFoundError(
						fmt.Errorf("user profile not found"),
					)
				}
			}

			validFCM, err := i.ValidFCMTokens(
				tt.args.ctx,
				tt.args.UIDs,
			)
			if tt.wantErr && validFCM != nil {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}

			if !tt.wantErr && validFCM == nil {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_ProfileAttributes(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	type args struct {
		ctx       context.Context
		UIDs      []string
		attribute string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]string
		wantErr bool
	}{
		{
			name: "valid:_get_user_profile_emails",
			args: args{
				ctx:       ctx,
				UIDs:      []string{uuid.New().String()},
				attribute: usecases.EmailsAttribute,
			},
			wantErr: false,
		},
		{
			name: "valid:_get_user_profile_phone_numbers",
			args: args{
				ctx:       ctx,
				UIDs:      []string{uuid.New().String()},
				attribute: usecases.PhoneNumbersAttribute,
			},
			wantErr: false,
		},
		{
			name: "valid:_get_user_profile_fcm_tokens",
			args: args{
				ctx:       ctx,
				UIDs:      []string{uuid.New().String()},
				attribute: usecases.FCMTokensAttribute,
			},
			wantErr: false,
		},
		{
			name: "invalid:_failed_get_user_profile_attribute",
			args: args{
				ctx:       ctx,
				UIDs:      []string{uuid.New().String()},
				attribute: "not-an-attribute",
			},
			wantErr: true,
		},
		{
			name: "invalid:_failed_get_user_profile",
			args: args{
				ctx:       ctx,
				UIDs:      []string{uuid.New().String()},
				attribute: usecases.FCMTokensAttribute,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:_get_user_profile_emails" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					email := converterandformatter.GenerateRandomEmail()
					return &profileutils.UserProfile{
						PrimaryEmailAddress: &email,
						SecondaryEmailAddresses: []string{
							converterandformatter.GenerateRandomEmail(),
						},
					}, nil
				}
			}

			if tt.name == "valid:_get_user_profile_phone_numbers" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					phone := interserviceclient.TestUserPhoneNumber
					return &profileutils.UserProfile{
						PrimaryPhone:          &phone,
						SecondaryPhoneNumbers: []string{"+254700000000"},
					}, nil
				}
			}

			if tt.name == "valid:_get_user_profile_fcm_tokens" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						PushTokens: []string{uuid.New().String()},
					}, nil
				}
			}

			if tt.name == "invalid:_failed_get_user_profile" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					email := converterandformatter.GenerateRandomEmail()
					phone := interserviceclient.TestUserPhoneNumber
					return &profileutils.UserProfile{
						PrimaryEmailAddress: &email,
						SecondaryEmailAddresses: []string{
							converterandformatter.GenerateRandomEmail(),
						},
						PrimaryPhone:          &phone,
						SecondaryPhoneNumbers: []string{"+254700000000"},
						PushTokens:            []string{uuid.New().String()},
					}, nil
				}
			}

			if tt.name == "invalid:_failed_get_user_profile" {
				fakeInfraRepo.GetUserProfileByUIDFn = func(
					ctx context.Context,
					uid string,
					suspended bool,
				) (*profileutils.UserProfile, error) {
					return nil, exceptions.ProfileNotFoundError(
						fmt.Errorf("user profile not found"),
					)
				}
			}

			attribute, err := i.ProfileAttributes(
				tt.args.ctx,
				tt.args.UIDs,
				tt.args.attribute,
			)

			if tt.wantErr && attribute != nil {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}

			if !tt.wantErr && attribute == nil {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_UpdateSuspended(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}
	type args struct {
		ctx        context.Context
		status     bool
		phone      string
		useContext bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:_suspend_with_use_context_false",
			args: args{
				ctx:        ctx,
				status:     true,
				phone:      "0721152896",
				useContext: false,
			},
			wantErr: false,
		},
		{
			name: "invalid:_suspend_with_use_context_false_update_user_fails",
			args: args{
				ctx:        ctx,
				status:     true,
				phone:      "0721152896",
				useContext: false,
			},
			wantErr: true,
		},
		{
			name: "valid:_suspend_with_use_context_true",
			args: args{
				ctx:        ctx,
				status:     true,
				phone:      "0721152896",
				useContext: true,
			},
			wantErr: false,
		},
		{
			name: "invalid:_suspend_with_use_context_true_get_user_profile_fails",
			args: args{
				ctx:        ctx,
				status:     true,
				phone:      "0721152896",
				useContext: true,
			},
			wantErr: true,
		},
		{
			name: "invalid:_suspend_with_use_context_true_get_loggedin_user_fails",
			args: args{
				ctx:        ctx,
				status:     true,
				phone:      "0721152896",
				useContext: true,
			},
			wantErr: true,
		},
		{
			name: "invalid:_normalize_msisdn_fails",
			args: args{
				ctx:        ctx,
				status:     true,
				phone:      "0721152896",
				useContext: false,
			},
			wantErr: true,
		},
		{
			name: "invalid:_suspend_use_context_false_get_user_profile_fails",
			args: args{
				ctx:        ctx,
				status:     true,
				phone:      "0721152896",
				useContext: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:_suspend_with_use_context_false" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}

				fakeInfraRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "123",
						PrimaryPhone: &phoneNumber,
						SecondaryPhoneNumbers: []string{
							"0721521456", "0721856741",
						},
					}, nil
				}

				fakeInfraRepo.UpdateSuspendedFn = func(ctx context.Context, id string, status bool) error {
					return nil
				}
			}

			if tt.name == "invalid:_suspend_with_use_context_false_update_user_fails" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}

				fakeInfraRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "123",
						PrimaryPhone: &phoneNumber,
						SecondaryPhoneNumbers: []string{
							"0721521456", "0721856741",
						},
					}, nil
				}

				fakeInfraRepo.UpdateSuspendedFn = func(ctx context.Context, id string, status bool) error {
					return fmt.Errorf("unable to update user profile")
				}
			}

			if tt.name == "valid:_suspend_with_use_context_true" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}

				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					}, nil
				}

				fakeInfraRepo.UpdateSuspendedFn = func(ctx context.Context, id string, status bool) error {
					return nil
				}
			}

			if tt.name == "invalid:_suspend_with_use_context_true_get_loggedin_user_fails" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}

				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "", fmt.Errorf("unable to get loggedin user")
				}

			}

			if tt.name == "invalid:_suspend_with_use_context_true_get_user_profile_fails" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}

				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get userprofile")
				}

			}

			if tt.name == "invalid:_normalize_msisdn_fails" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("unable to normalize phone")
				}
			}

			if tt.name == "invalid:_suspend_use_context_false_get_user_profile_fails" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}

				fakeInfraRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get user profile")
				}
			}

			err := i.UpdateSuspended(
				tt.args.ctx,
				tt.args.status,
				tt.args.phone,
				tt.args.useContext,
			)

			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}

		})
	}
}

func TestProfileUseCaseImpl_UpdatePrimaryPhoneNumber(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	primaryPhone := "+254719789543"
	primaryPhone1 := "+254765739201"
	type args struct {
		ctx        context.Context
		phone      string
		useContext bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:_update_primaryPhoneNumber_with_useContext_false",
			args: args{
				ctx:        ctx,
				phone:      primaryPhone,
				useContext: false,
			},
			wantErr: false,
		},

		{
			name: "valid:_update_primaryPhoneNumber_with_useContext_true",
			args: args{
				ctx:        ctx,
				phone:      primaryPhone1,
				useContext: true,
			},
			wantErr: false,
		},
		{
			name: "invalid:_missing_primaryPhoneNumber",
			args: args{
				ctx:        ctx,
				phone:      " ",
				useContext: true,
			},
			wantErr: true,
		},
		{
			name: "invalid:_unable_to_get_logged_in_user",
			args: args{
				ctx:        ctx,
				phone:      "+25463728192",
				useContext: true,
			},
			wantErr: true,
		},
		{
			name: "invalid:_unable_to_get_userProfile_by_phonenumber",
			args: args{
				ctx:        ctx,
				phone:      "+254736291036",
				useContext: false,
			},
			wantErr: true,
		},
		{
			name: "invalid:_unable_to_get_profile_of_logged_in_user",
			args: args{
				ctx:        ctx,
				phone:      "+25463728192",
				useContext: true,
			},
			wantErr: true,
		},
		{
			name: "invalid:_unable_to_update_secondary_phonenumber",
			args: args{
				ctx:        ctx,
				phone:      "+25463728192",
				useContext: false,
			},
			wantErr: true,
		},
		{
			name: "invalid:_unable_to_update_primary_phonenumber",
			args: args{
				ctx:        ctx,
				phone:      "+25463728192",
				useContext: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "valid:_update_primaryPhoneNumber_with_useContext_false" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254755889922"
					return &phone, nil
				}

				fakeInfraRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "ABCDE",
						PrimaryPhone: &phoneNumber,
						SecondaryPhoneNumbers: []string{
							"0765839203", "0789437282",
						},
					}, nil
				}

				fakeInfraRepo.UpdatePrimaryPhoneNumberFn = func(ctx context.Context, id string, phoneNumber string) error {
					return nil
				}

				fakeInfraRepo.UpdateSecondaryPhoneNumbersFn = func(ctx context.Context, id string, phoneNumbers []string) error {
					return nil
				}
			}

			if tt.name == "valid:_update_primaryPhoneNumber_with_useContext_true" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254789029156"
					return &phone, nil
				}

				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "f4f39af7--91bd-42b3af315a4e",
						PrimaryPhone: &primaryPhone1,
						SecondaryPhoneNumbers: []string{
							"0765839203", "0789437282",
						},
					}, nil
				}

				fakeInfraRepo.UpdatePrimaryPhoneNumberFn = func(ctx context.Context, id string, phoneNumber string) error {
					return nil
				}

				fakeInfraRepo.UpdateSecondaryPhoneNumbersFn = func(ctx context.Context, id string, phoneNumbers []string) error {
					return nil
				}
			}

			if tt.name == "invalid:_missing_primaryPhoneNumber" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("empty phone number provided")
				}
			}

			if tt.name == "invalid:_unable_to_get_logged_in_user" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254789029156"
					return &phone, nil
				}

				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("unable to get logged in user")
				}
			}

			if tt.name == "invalid:_unable_to_get_userProfile_by_phonenumber" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254799774466"
					return &phone, nil
				}

				fakeInfraRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get user profile by phonenumber")
				}
			}

			if tt.name == "invalid:_unable_to_get_profile_of_logged_in_user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get profile")
				}
			}

			if tt.name == "invalid:_unable_to_update_secondary_phonenumber" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254755889922"
					return &phone, nil
				}

				fakeInfraRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "ABCDE",
						PrimaryPhone: &phoneNumber,
						SecondaryPhoneNumbers: []string{
							"0765839203", "0789437282",
						},
					}, nil
				}

				fakeInfraRepo.UpdatePrimaryPhoneNumberFn = func(ctx context.Context, id string, phoneNumber string) error {
					return nil
				}

				fakeInfraRepo.UpdateSecondaryPhoneNumbersFn = func(ctx context.Context, id string, phoneNumbers []string) error {
					return fmt.Errorf("unable to update secondary phonenumber")
				}
			}

			if tt.name == "invalid:_unable_to_update_secondary_phonenumber" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254755889922"
					return &phone, nil
				}

				fakeInfraRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "ABCDE",
						PrimaryPhone: &phoneNumber,
						SecondaryPhoneNumbers: []string{
							"0765839203", "0789437282",
						},
					}, nil
				}

				fakeInfraRepo.UpdatePrimaryPhoneNumberFn = func(ctx context.Context, id string, phoneNumber string) error {
					return fmt.Errorf("unable to update primary phonenumber")
				}

			}

			err := i.UpdatePrimaryPhoneNumber(
				tt.args.ctx,
				tt.args.phone,
				tt.args.useContext,
			)
			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCase_UpdateBioData(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}
	phone := gofakeit.Phone()
	dateOfBirth := scalarutils.Date{
		Day:   12,
		Year:  2000,
		Month: 2,
	}

	firstName := "Jatelo"
	lastName := "Omera"
	bioData := profileutils.BioData{
		FirstName:   &firstName,
		LastName:    &lastName,
		DateOfBirth: &dateOfBirth,
	}

	var gender enumutils.Gender = "female"
	updateGender := profileutils.BioData{
		Gender: gender,
	}

	updateDOB := profileutils.BioData{
		DateOfBirth: &dateOfBirth,
	}

	updateFirstName := profileutils.BioData{
		FirstName: &firstName,
	}

	updateLastName := profileutils.BioData{
		LastName: &lastName,
	}
	type args struct {
		ctx  context.Context
		data profileutils.BioData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid: update primary biodata of a specific user profile",
			args: args{
				ctx:  ctx,
				data: bioData,
			},
			wantErr: false,
		},
		{
			name: "valid: update primary biodata of a specific user profile - gender",
			args: args{
				ctx:  ctx,
				data: updateGender,
			},
			wantErr: false,
		},
		{
			name: "valid: update primary biodata of a specific user profile - DOB",
			args: args{
				ctx:  ctx,
				data: updateDOB,
			},
			wantErr: false,
		},
		{
			name: "valid: update primary biodata of a specific user profile - First Name",
			args: args{
				ctx:  ctx,
				data: updateFirstName,
			},
			wantErr: false,
		},
		{
			name: "valid: update primary biodata of a specific user profile - Last Name",
			args: args{
				ctx:  ctx,
				data: updateLastName,
			},
			wantErr: false,
		},
		{
			name: "invalid: get logged in user uid fails",
			args: args{
				ctx:  ctx,
				data: bioData,
			},
			wantErr: true,
		},
		{
			name: "invalid: get user profile by UID fails",
			args: args{
				ctx:  ctx,
				data: bioData,
			},
			wantErr: true,
		},
		{
			name: "invalid: update primary biodata of a specific user profile",
			args: args{
				ctx:  ctx,
				data: bioData,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid: update primary biodata of a specific user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
						PrimaryPhone: &phone,
					}, nil
				}

				fakeInfraRepo.UpdateBioDataFn = func(ctx context.Context, id string, data profileutils.BioData) error {
					return nil
				}

				fakePubSub.TopicIDsFn = func() []string {
					return []string{uuid.New().String()}
				}

				fakePubSub.AddPubSubNamespaceFn = func(topicName string) string {
					return uuid.New().String()
				}

				fakePubSub.PublishToPubsubFn = func(ctx context.Context, topicID string, payload []byte) error {
					return nil
				}

			}
			if tt.name == "valid: update primary biodata of a specific user profile - gender" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
						PrimaryPhone: &phone,
					}, nil
				}

				fakeInfraRepo.UpdateBioDataFn = func(ctx context.Context, id string, data profileutils.BioData) error {
					return nil
				}

				fakePubSub.TopicIDsFn = func() []string {
					return []string{uuid.New().String()}
				}

				fakePubSub.AddPubSubNamespaceFn = func(topicName string) string {
					return uuid.New().String()
				}

				fakePubSub.PublishToPubsubFn = func(ctx context.Context, topicID string, payload []byte) error {
					return nil
				}

			}
			if tt.name == "valid: update primary biodata of a specific user profile - DOB" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
						PrimaryPhone: &phone,
					}, nil
				}

				fakeInfraRepo.UpdateBioDataFn = func(ctx context.Context, id string, data profileutils.BioData) error {
					return nil
				}

			}
			if tt.name == "valid: update primary biodata of a specific user profile - First Name" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
						PrimaryPhone: &phone,
					}, nil
				}

				fakeInfraRepo.UpdateBioDataFn = func(ctx context.Context, id string, data profileutils.BioData) error {
					return nil
				}

				fakePubSub.TopicIDsFn = func() []string {
					return []string{uuid.New().String()}
				}

				fakePubSub.AddPubSubNamespaceFn = func(topicName string) string {
					return uuid.New().String()
				}

				fakePubSub.PublishToPubsubFn = func(ctx context.Context, topicID string, payload []byte) error {
					return nil
				}

			}
			if tt.name == "valid: update primary biodata of a specific user profile - Last Name" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
						PrimaryPhone: &phone,
					}, nil
				}

				fakeInfraRepo.UpdateBioDataFn = func(ctx context.Context, id string, data profileutils.BioData) error {
					return nil
				}

				fakePubSub.TopicIDsFn = func() []string {
					return []string{uuid.New().String()}
				}

				fakePubSub.AddPubSubNamespaceFn = func(topicName string) string {
					return uuid.New().String()
				}

				fakePubSub.PublishToPubsubFn = func(ctx context.Context, topicID string, payload []byte) error {
					return nil
				}

			}
			if tt.name == "invalid: get logged in user uid fails" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("failed to get loggeg in user UID")
				}
			}

			if tt.name == "invalid: get user profile by UID fails" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("failed to get user profile by UID")
				}
			}
			if tt.name == "invalid: update primary biodata of a specific user profile" {

				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
						PrimaryPhone: &phone,
					}, nil
				}

				fakeInfraRepo.UpdateBioDataFn = func(ctx context.Context, id string, data profileutils.BioData) error {
					return fmt.Errorf("failed update primary biodata of a user profile")
				}

			}

			err := i.UpdateBioData(tt.args.ctx, tt.args.data)
			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCase_CheckPhoneExists(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}
	type args struct {
		ctx   context.Context
		phone string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "valid:_check phone number exists",
			args: args{
				ctx:   ctx,
				phone: "+254711223344",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "invalid:_check phone number exists - empty phone number provided",
			args: args{
				ctx:   ctx,
				phone: "",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "invalid:_check phone number exists",
			args: args{
				ctx:   ctx,
				phone: "+254711223344",
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:_check phone number exists" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254711223344"
					return &phone, nil
				}
				fakeInfraRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return false, nil
				}
			}
			if tt.name == "invalid:_check phone number exists - empty phone number provided" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("empty phone number provided")
				}
			}
			if tt.name == "invalid:_check phone number exists" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254711223344"
					return &phone, nil
				}
				fakeInfraRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return false, fmt.Errorf("error checking if phone number exists")
				}
			}
			_, err := i.CheckPhoneExists(tt.args.ctx, tt.args.phone)
			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCase_CheckEmailExists(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	validEmail := "me4@gmail.com"
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "valid:_check email exists",
			args: args{
				ctx:   ctx,
				email: validEmail,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "invalid:_check email exists",
			args: args{
				ctx:   ctx,
				email: "",
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:_check email exists" {
				fakeInfraRepo.CheckIfEmailExistsFn = func(ctx context.Context, email string) (bool, error) {
					return false, nil
				}
			}
			if tt.name == "invalid:_check email exists" {
				fakeInfraRepo.CheckIfEmailExistsFn = func(ctx context.Context, email string) (bool, error) {
					return false, fmt.Errorf("failed to if email exists")
				}
			}
			_, err := i.CheckEmailExists(tt.args.ctx, tt.args.email)
			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_UpdatePhotoUploadID(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	type args struct {
		ctx      context.Context
		uploadID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:successfully_updatePhotoUploadID",
			args: args{
				ctx:      ctx,
				uploadID: "some-upload-id",
			},
			wantErr: false,
		},
		{
			name: "invalid:fail_to_update_photoUploadID",
			args: args{
				ctx:      ctx,
				uploadID: "some-upload-id",
			},
			wantErr: true,
		},
		{
			name: "invalid:fail_to_get_loggedInUser",
			args: args{
				ctx:      ctx,
				uploadID: "some-upload-id",
			},
			wantErr: true,
		},
		{
			name: "invalid:fail_to_get_userProfile",
			args: args{
				ctx:      ctx,
				uploadID: "some-upload-id",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:successfully_updatePhotoUploadID" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					}, nil
				}

				fakeInfraRepo.UpdatePhotoUploadIDFn = func(ctx context.Context, id string, uploadID string) error {
					return nil
				}
			}

			if tt.name == "invalid:fail_to_get_loggedInUser" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "", fmt.Errorf("failed to get logged in user")
				}
			}

			if tt.name == "invalid:fail_to_get_userProfile" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("failed to get user profile")
				}
			}

			if tt.name == "invalid:fail_to_update_photoUploadID" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					}, nil
				}

				fakeInfraRepo.UpdatePhotoUploadIDFn = func(ctx context.Context, id string, uploadID string) error {
					return fmt.Errorf("failed to update photo upload ID")
				}
			}
			err := i.UpdatePhotoUploadID(tt.args.ctx, tt.args.uploadID)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"ProfileUseCaseImpl.UpdatePhotoUploadID() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_AddAddress(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	addr := dto.UserAddressInput{
		Latitude:  1.2,
		Longitude: -34.001,
	}
	type args struct {
		ctx         context.Context
		input       dto.UserAddressInput
		addressType enumutils.AddressType
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy:) add home address",
			args: args{
				ctx:         ctx,
				input:       addr,
				addressType: enumutils.AddressTypeHome,
			},
			wantErr: false,
		},
		{
			name: "happy:) add work address",
			args: args{
				ctx:         ctx,
				input:       addr,
				addressType: enumutils.AddressTypeWork,
			},
			wantErr: false,
		},
		{
			name: "sad:( failed to get logged in user",
			args: args{
				ctx:         ctx,
				input:       addr,
				addressType: enumutils.AddressTypeWork,
			},
			wantErr: true,
		},
		{
			name: "sad:( failed to get user profile",
			args: args{
				ctx:         ctx,
				input:       addr,
				addressType: enumutils.AddressTypeWork,
			},
			wantErr: true,
		},
		{
			name: "sad:( failed to update user profile",
			args: args{
				ctx:         ctx,
				input:       addr,
				addressType: enumutils.AddressTypeWork,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "happy:) add home address" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}

				fakeInfraRepo.UpdateAddressesFn = func(ctx context.Context, id string, address profileutils.Address, addressType enumutils.AddressType) error {
					return nil
				}
			}

			if tt.name == "happy:) add work address" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}

				fakeInfraRepo.UpdateAddressesFn = func(ctx context.Context, id string, address profileutils.Address, addressType enumutils.AddressType) error {
					return nil
				}
			}

			if tt.name == "sad:( failed to get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("an error occurred")
				}
			}

			if tt.name == "sad:( failed to get user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("an error occurred")
				}
			}

			if tt.name == "sad:( failed to update user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}

				fakeInfraRepo.UpdateAddressesFn = func(ctx context.Context, id string, address profileutils.Address, addressType enumutils.AddressType) error {
					return fmt.Errorf("an error occurred")
				}
			}

			_, err := i.AddAddress(
				tt.args.ctx,
				tt.args.input,
				tt.args.addressType,
			)

			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_GetAddresses(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "happy:) get addresses",
			args:    args{ctx: ctx},
			wantErr: false,
		},
		{
			name:    "sad:( failed to get logged in user",
			args:    args{ctx: ctx},
			wantErr: true,
		},
		{
			name:    "sad:( failed to get user profile",
			args:    args{ctx: ctx},
			wantErr: true,
		},
		{
			name:    "sad:/ failed to get the home addresses",
			args:    args{ctx: ctx},
			wantErr: true,
		},
		{
			name:    "sad:/ failed to get the work addresses",
			args:    args{ctx: ctx},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "happy:) get addresses" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
						HomeAddress: &profileutils.Address{
							Latitude:  "123",
							Longitude: "-1.2",
						},
						WorkAddress: &profileutils.Address{
							Latitude:  "123",
							Longitude: "-1.2",
						},
					}, nil
				}
			}

			if tt.name == "sad:( failed to get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("an error occurred")
				}
			}

			if tt.name == "sad:( failed to get user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("an error occurred")
				}
			}

			if tt.name == "sad:/ failed to get the home addresses" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:          uuid.New().String(),
						HomeAddress: &profileutils.Address{},
						WorkAddress: &profileutils.Address{
							Latitude:  "123",
							Longitude: "-1.2",
						},
					}, nil
				}
			}

			if tt.name == "sad:/ failed to get the work addresses" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
						HomeAddress: &profileutils.Address{
							Latitude:  "123",
							Longitude: "-1.2",
						},
						WorkAddress: &profileutils.Address{},
					}, nil
				}
			}

			_, err := i.GetAddresses(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"ProfileUseCaseImpl.GetAddresses() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_GetUserCommunicationsSettings(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "valid: get comms settings",
			args:    args{ctx: ctx},
			wantErr: false,
		},
		{
			name:    "invalid: failed to get logged in user",
			args:    args{ctx: ctx},
			wantErr: true,
		},
		{
			name:    "invalid: failed to get user profile",
			args:    args{ctx: ctx},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid: get comms settings" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserCommunicationsSettingsFn = func(ctx context.Context, profileID string) (*profileutils.UserCommunicationsSetting, error) {
					return &profileutils.UserCommunicationsSetting{
						ID:            uuid.New().String(),
						AllowWhatsApp: true,
						AllowTextSMS:  true,
						AllowEmail:    true,
						AllowPush:     true,
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}
			}

			if tt.name == "invalid: failed to get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("an error occurred")
				}
			}

			if tt.name == "invalid: failed to get user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("an error occurred")
				}
			}

			_, err := i.GetUserCommunicationsSettings(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"ProfileUseCaseImpl.GetUserCommunicationsSettings() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_SetUserCommunicationsSettings(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	type args struct {
		ctx           context.Context
		allowWhatsApp bool
		allowTextSms  bool
		allowPush     bool
		allowEmail    bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid: set comms settings",
			args: args{
				ctx:           ctx,
				allowWhatsApp: true,
				allowTextSms:  true,
				allowPush:     true,
				allowEmail:    true,
			},
			wantErr: false,
		},
		{
			name: "invalid: failed to get logged in user",
			args: args{
				ctx:           ctx,
				allowWhatsApp: true,
				allowTextSms:  true,
				allowPush:     true,
				allowEmail:    true,
			},
			wantErr: true,
		},
		{
			name: "invalid: failed to get user profile",
			args: args{
				ctx:           ctx,
				allowWhatsApp: true,
				allowTextSms:  true,
				allowPush:     true,
				allowEmail:    true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid: set comms settings" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.SetUserCommunicationsSettingsFn = func(ctx context.Context, profileID string,
					allowWhatsApp *bool, allowTextSms *bool, allowPush *bool, allowEmail *bool) (*profileutils.UserCommunicationsSetting, error) {
					return &profileutils.UserCommunicationsSetting{
						ID:            uuid.New().String(),
						AllowWhatsApp: true,
						AllowTextSMS:  true,
						AllowEmail:    true,
						AllowPush:     true,
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}
			}

			if tt.name == "invalid: failed to get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("an error occurred")
				}
			}

			if tt.name == "invalid: failed to get user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("an error occurred")
				}
			}

			_, err := i.SetUserCommunicationsSettings(
				tt.args.ctx,
				&tt.args.allowWhatsApp,
				&tt.args.allowTextSms,
				&tt.args.allowEmail,
				&tt.args.allowPush,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"ProfileUseCaseImpl.SetUserCommunicationsSettings() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_SaveFavoriteNavActions(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	type args struct {
		ctx   context.Context
		title string
	}

	initialFavActions := []string{"agents", "consumers"}
	finalfavActions := []string{"home", "agents", "consumers"}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    bool
	}{
		{
			name: "invalid: unable get loggedIn user",
			args: args{
				ctx:   ctx,
				title: "home",
			},
			wantErr: true,
			want:    false,
		},
		{
			name: "invalid: unable get user profile",
			args: args{
				ctx:   ctx,
				title: "home",
			},
			wantErr: true,
			want:    false,
		},
		{
			name: "invalid: unable to add favorite navigation actions",
			args: args{
				ctx:   ctx,
				title: "home",
			},
			wantErr: true,
			want:    false,
		},
		{
			name: "invalid: unable to update user favorite navactions",
			args: args{
				ctx:   ctx,
				title: "home",
			},
			wantErr: true,
			want:    false,
		},
		{
			name: "valid: saved user navactions",
			args: args{
				ctx:   ctx,
				title: "home",
			},
			wantErr: false,
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "invalid: unable get loggedIn user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("unable to get logged in user")
				}
			}
			if tt.name == "invalid: unable get user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get user profile")
				}
			}
			if tt.name == "invalid: unable to add favorite navigation actions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:            uuid.New().String(),
						FavNavActions: finalfavActions,
					}, nil
				}
				fakeInfraRepo.UpdateFavNavActionsFn = func(ctx context.Context, id string, favActions []string) error {
					return nil
				}
			}
			if tt.name == "invalid: unable to update user favorite navactions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:            uuid.New().String(),
						FavNavActions: initialFavActions,
					}, nil
				}
				fakeInfraRepo.UpdateFavNavActionsFn = func(ctx context.Context, id string, favActions []string) error {
					return fmt.Errorf("unable to update user favorite navactions")
				}
			}
			if tt.name == "valid: saved user navactions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:            uuid.New().String(),
						FavNavActions: initialFavActions,
					}, nil
				}
				fakeInfraRepo.UpdateFavNavActionsFn = func(ctx context.Context, id string, favActions []string) error {
					return nil
				}
			}
			got, err := i.SaveFavoriteNavActions(tt.args.ctx, tt.args.title)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"ProfileUseCaseImpl.SaveFavoriteNavActions() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if got != tt.want {
				t.Errorf("ProfileUseCaseImpl.SaveFavoriteNavActions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProfileUseCaseImpl_DeleteFavoriteNavActions(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	initialFavActions := []string{"home", "agents", "consumers"}
	finalFavActions := []string{"agents", "consumers"}
	type args struct {
		ctx   context.Context
		title string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    bool
	}{
		{
			name: "invalid: unable get loggedIn user",
			args: args{
				ctx:   ctx,
				title: "home",
			},
			wantErr: true,
			want:    false,
		},
		{
			name: "invalid: unable get user profile",
			args: args{
				ctx:   ctx,
				title: "home",
			},
			wantErr: true,
			want:    false,
		},
		{
			name: "invalid: unable to remove favorite navigation action",
			args: args{
				ctx:   ctx,
				title: "home",
			},
			wantErr: true,
			want:    false,
		},
		{
			name: "invalid: unable to update user favorite navactions",
			args: args{
				ctx:   ctx,
				title: "home",
			},
			wantErr: true,
			want:    false,
		},
		{
			name: "valid: deleted user navactions",
			args: args{
				ctx:   ctx,
				title: "home",
			},
			wantErr: false,
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "invalid: unable get loggedIn user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("unable to get logged in user")
				}
			}
			if tt.name == "invalid: unable get user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get user profile")
				}
			}
			if tt.name == "invalid: unable to remove favorite navigation action" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:            uuid.New().String(),
						FavNavActions: finalFavActions,
					}, nil
				}
				fakeInfraRepo.UpdateFavNavActionsFn = func(ctx context.Context, id string, favActions []string) error {
					return nil
				}
			}
			if tt.name == "invalid: unable to update user favorite navactions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:            uuid.New().String(),
						FavNavActions: initialFavActions,
					}, nil
				}
				fakeInfraRepo.UpdateFavNavActionsFn = func(ctx context.Context, id string, favActions []string) error {
					return fmt.Errorf("unable to update user favorite navactions")
				}
			}
			if tt.name == "valid: deleted user navactions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:            uuid.New().String(),
						FavNavActions: initialFavActions,
					}, nil
				}
				fakeInfraRepo.UpdateFavNavActionsFn = func(ctx context.Context, id string, favActions []string) error {
					return nil
				}
			}
			got, err := i.DeleteFavoriteNavActions(tt.args.ctx, tt.args.title)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"ProfileUseCaseImpl.DeleteFavoriteNavActions() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if got != tt.want {
				t.Errorf(
					"ProfileUseCaseImpl.DeleteFavoriteNavActions() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestProfileUseCaseImpl_RefreshNavigationActions(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
		wantErr bool
	}{
		{
			name:    "sad: failed to get logged in user",
			args:    args{ctx: ctx},
			wantNil: true,
			wantErr: true,
		}, {
			name:    "sad: failed to get logged in user profile",
			args:    args{ctx: ctx},
			wantNil: true,
			wantErr: true,
		}, {
			name:    "sad: failed to get user navigation actions",
			args:    args{ctx: ctx},
			wantNil: true,
			wantErr: true,
		}, {
			name:    "happy: got user navigation actions",
			args:    args{ctx: ctx},
			wantNil: false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "sad: failed to get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("unable to get logged in user")
				}
			}
			if tt.name == "sad: failed to get logged in user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get user profile")
				}
			}
			if tt.name == "happy: got user navigation actions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}

				fakeInfraRepo.GetRolesByIDsFn = func(ctx context.Context, roleIDs []string) (*[]profileutils.Role, error) {
					roles := []profileutils.Role{
						{
							ID: uuid.NewString(),
						},
					}
					return &roles, nil
				}
			}

			got, err := i.RefreshNavigationActions(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"ProfileUseCaseImpl.RefreshNavigationActions() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if !tt.wantNil {
				if got == nil {
					t.Errorf(
						"ProfileUseCaseImpl.RefreshNavigationActions() = %v, want %v",
						got,
						tt.wantNil,
					)
				}
			}
		})
	}
}

func TestProfileUseCaseImpl_FindUserByPhone(t *testing.T) {

	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	type args struct {
		ctx         context.Context
		phoneNumber string
	}
	tests := []struct {
		name    string
		args    args
		want    *profileutils.UserProfile
		wantErr bool
	}{
		{
			name: "fail: cannot normalize  phone number",
			args: args{
				ctx:         ctx,
				phoneNumber: interserviceclient.TestUserPhoneNumber,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "fail: cannot finder user with phone number",
			args: args{
				ctx:         ctx,
				phoneNumber: interserviceclient.TestUserPhoneNumber,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success: find user using phone number",
			args: args{
				ctx:         ctx,
				phoneNumber: interserviceclient.TestUserPhoneNumber,
			},
			want: &profileutils.UserProfile{
				ID: "3029c544-78ea-4e2e-841a-82fed3af4e94",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "fail: cannot normalize  phone number" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("cannot normalize phone number")
				}
			}

			if tt.name == "fail: cannot finder user with phone number" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					p := interserviceclient.TestUserPhoneNumber
					return &p, nil
				}
				fakeInfraRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("cannot retrieve user profile")
				}
			}

			if tt.name == "success: find user using phone number" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					p := interserviceclient.TestUserPhoneNumber
					return &p, nil
				}
				fakeInfraRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					p := &profileutils.UserProfile{
						ID: "3029c544-78ea-4e2e-841a-82fed3af4e94",
					}
					return p, nil
				}
			}
			got, err := i.FindUserByPhone(tt.args.ctx, tt.args.phoneNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProfileUseCaseImpl.FindUserByPhone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProfileUseCaseImpl.FindUserByPhone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProfileUseCaseImpl_FindUsersByPhone(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	profile := profileutils.UserProfile{ID: "3029c544-78ea-4e2e-841a-82fed3af4e94"}
	type args struct {
		ctx         context.Context
		phoneNumber string
	}
	tests := []struct {
		name    string
		args    args
		want    []*profileutils.UserProfile
		wantErr bool
	}{

		{
			name: "fail: cannot normalize  phone number",
			args: args{
				ctx:         ctx,
				phoneNumber: interserviceclient.TestUserPhoneNumber,
			},
			want:    []*profileutils.UserProfile{},
			wantErr: false,
		},
		{
			name: "fail: cannot finder user with phone number",
			args: args{
				ctx:         ctx,
				phoneNumber: interserviceclient.TestUserPhoneNumber,
			},
			want:    []*profileutils.UserProfile{},
			wantErr: false,
		},
		{
			name: "success: find user using phone number",
			args: args{
				ctx:         ctx,
				phoneNumber: interserviceclient.TestUserPhoneNumber,
			},
			want: []*profileutils.UserProfile{
				&profile,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "fail: cannot normalize  phone number" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("cannot normalize phone number")
				}
			}

			if tt.name == "fail: cannot finder user with phone number" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					p := interserviceclient.TestUserPhoneNumber
					return &p, nil
				}
				fakeInfraRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("cannot retrieve user profile")
				}
			}

			if tt.name == "success: find user using phone number" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					p := interserviceclient.TestUserPhoneNumber
					return &p, nil
				}
				fakeInfraRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					p := &profileutils.UserProfile{
						ID: "3029c544-78ea-4e2e-841a-82fed3af4e94",
					}
					return p, nil
				}
			}
			got, err := i.FindUsersByPhone(tt.args.ctx, tt.args.phoneNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProfileUseCaseImpl.FindUsersByPhone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProfileUseCaseImpl.FindUsersByPhone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProfileUseCaseImpl_GetNavigationActions(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	type args struct {
		ctx context.Context
	}

	tests := []struct {
		name    string
		args    args
		want    *dto.GroupedNavigationActions
		wantErr bool
	}{
		{
			name:    "sad unable to get logged in user",
			args:    args{ctx: ctx},
			want:    nil,
			wantErr: true,
		},

		{
			name:    "sad unable to get user profile",
			args:    args{ctx: ctx},
			want:    nil,
			wantErr: true,
		},

		{
			name:    "sad unable to get user roles",
			args:    args{ctx: ctx},
			want:    nil,
			wantErr: true,
		},
		{
			name: "happy got user navigation actions",
			args: args{ctx: ctx},
			want: &dto.GroupedNavigationActions{
				Primary: []domain.NavigationAction{
					domain.HomeNavAction,
					domain.HelpNavAction,
				},
				Secondary: []domain.NavigationAction{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "sad unable to get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("cannot get logged in user")
				}
			}

			if tt.name == "sad unable to get user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get user profile")
				}
			}

			if tt.name == "sad unable to get user roles" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}
				fakeInfraRepo.GetRolesByIDsFn = func(ctx context.Context, roleIDs []string) (*[]profileutils.Role, error) {
					return nil, fmt.Errorf("unable to get user roles")
				}
			}

			if tt.name == "happy got user navigation actions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}
				fakeInfraRepo.GetRolesByIDsFn = func(ctx context.Context, roleIDs []string) (*[]profileutils.Role, error) {
					return &[]profileutils.Role{}, nil
				}
			}
			got, err := i.GetNavigationActions(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"ProfileUseCaseImpl.GetNavigationActions() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProfileUseCaseImpl.GetNavigationActions() = %v, want %v", got, tt.want)
			}
		})
	}
}
