package usecases_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/profileutils"
	"github.com/savannahghi/scalarutils"
)

func TestSignUpUseCasesImpl_RetirePushToken(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	type args struct {
		ctx   context.Context
		token string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "valid:_successfully_retire_pushtoken",
			args: args{
				ctx:   ctx,
				token: "VAL1IDT0K3N",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "invalid:_fail_to_retire_pushtoken",
			args: args{
				ctx:   ctx,
				token: "VAL1IDT0K3N",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "invalid:_fail_to_retire_pushtoken_invalid_length",
			args: args{
				ctx:   ctx,
				token: "*",
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "valid:_successfully_retire_pushtoken" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:         "f4f39af7--91bd-42b3af315a4e",
						PushTokens: []string{"token12", "token23", "token34"},
					}, nil
				}

				fakeInfraRepo.UpdatePushTokensFn = func(ctx context.Context, id string, pushToken []string) error {
					return nil
				}
			}

			if tt.name == "invalid:_fail_to_retire_pushtoken" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:         "f4f39af7--91bd-42b3af315a4e",
						PushTokens: []string{"token12", "token23", "token34"},
					}, nil
				}

				fakeInfraRepo.UpdatePushTokensFn = func(ctx context.Context, id string, pushToken []string) error {
					return fmt.Errorf("failed to retire push token")
				}
			}

			if tt.name == "invalid:_fail_to_retire_pushtoken_invalid_length" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:         "f4f39af7--91bd-42b3af315a4e",
						PushTokens: []string{"token12", "token23", "token34"},
					}, nil
				}

				fakeInfraRepo.UpdatePushTokensFn = func(ctx context.Context, id string, pushToken []string) error {
					return fmt.Errorf("failed to retire push token")
				}
			}

			got, err := i.RetirePushToken(tt.args.ctx, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"SignUpUseCasesImpl.RetirePushToken() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if got != tt.want {
				t.Errorf("SignUpUseCasesImpl.RetirePushToken() = %v, want %v", got, tt.want)
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

func TestSignUpUseCasesImpl_CreateUserByPhone(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	phoneNumber := "+254777886622"
	pin := "1234"
	otp := "678251"

	validSignUpInput := &dto.SignUpInput{
		PhoneNumber: &phoneNumber,
		PIN:         &pin,
		Flavour:     feedlib.FlavourConsumer,
		OTP:         &otp,
	}

	invalidPhoneNumber := "+254"
	invalidPin := ""
	invalidOTP := ""

	invalidSignUpInput := &dto.SignUpInput{
		PhoneNumber: &invalidPhoneNumber,
		PIN:         &invalidPin,
		Flavour:     feedlib.FlavourConsumer,
		OTP:         &invalidOTP,
	}
	phone := gofakeit.Phone()

	type args struct {
		ctx   context.Context
		input *dto.SignUpInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:successfully_create_user_by_phone",
			args: args{
				ctx:   ctx,
				input: validSignUpInput,
			},
			wantErr: false,
		},
		{
			name: "invalid:fail_to_verifyOTP",
			args: args{
				ctx:   ctx,
				input: validSignUpInput,
			},
			wantErr: true,
		},
		{
			name: "invalid:use_invalid_input",
			args: args{
				ctx:   ctx,
				input: invalidSignUpInput,
			},
			wantErr: true,
		},
		{
			name: "invalid:fail_to_create_user_profile",
			args: args{
				ctx:   ctx,
				input: validSignUpInput,
			},
			wantErr: true,
		},
		{
			name: "invalid:fail_to_generate_auth_credentials",
			args: args{
				ctx:   ctx,
				input: validSignUpInput,
			},
			wantErr: true,
		},
		{
			name: "invalid:fail_to_set_userPin",
			args: args{
				ctx:   ctx,
				input: validSignUpInput,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "valid:successfully_create_user_by_phone" {
				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetOrCreatePhoneNumberUserFn = func(ctx context.Context, phone string) (*dto.CreatedUserResponse, error) {
					return &dto.CreatedUserResponse{
						UID:         "5cf354a2-1d3e-400d-8716-7e2aead29f2c",
						DisplayName: "John Doe",
						Email:       "johndoe@gmail.com",
						PhoneNumber: phoneNumber,
					}, nil
				}

				fakeInfraRepo.CreateUserProfileFn = func(ctx context.Context, phoneNumber, uid string) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "5cf354a2-1d3e-400d-8716-7e2aead29f2c",
						PrimaryPhone: &phone,
					}, nil
				}

				fakeInfraRepo.GenerateAuthCredentialsFn = func(ctx context.Context, phone string, profile *profileutils.UserProfile) (*profileutils.AuthCredentialResponse, error) {
					customToken := uuid.New().String()
					idToken := uuid.New().String()
					refreshToken := uuid.New().String()
					return &profileutils.AuthCredentialResponse{
						CustomToken:  &customToken,
						IDToken:      &idToken,
						RefreshToken: refreshToken,
					}, nil
				}

				// Mock SetUserPin begins here
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254777886622"
					return &phone, nil
				}

				fakeInfraRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "123",
						PrimaryPhone: &phoneNumber,
					}, nil
				}

				fakePinExt.EncryptPINFn = func(rawPwd string, options *extension.Options) (string, string) {
					return "salt", "password"
				}

				fakeInfraRepo.SavePINFn = func(ctx context.Context, pin *domain.PIN) (bool, error) {
					return true, nil
				}
				// Finished mocking SetUserPin
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

				fakeInfraRepo.GetRolesByIDsFn = func(ctx context.Context, roleIDs []string) (*[]profileutils.Role, error) {
					roles := []profileutils.Role{}
					return &roles, nil
				}
			}

			if tt.name == "invalid:fail_to_verifyOTP" {
				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return false, nil
				}
			}

			if tt.name == "invalid:use_invalid_input" {
				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return false, nil
				}
			}

			if tt.name == "invalid:fail_to_check_ifPhoneNumberExists" {
				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return true, nil
				}
			}

			if tt.name == "invalid:fail_to_create_user_profile" {
				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetOrCreatePhoneNumberUserFn = func(ctx context.Context, phone string) (*dto.CreatedUserResponse, error) {
					return &dto.CreatedUserResponse{
						UID:         "5cf354a2-1d3e-400d-8716-7e2aead29f2c",
						DisplayName: "John Doe",
						Email:       "johndoe@gmail.com",
						PhoneNumber: phoneNumber,
					}, nil
				}

				fakeInfraRepo.CreateUserProfileFn = func(ctx context.Context, phoneNumber, uid string) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("fail to create user profile")
				}
			}

			if tt.name == "invalid:fail_to_generate_auth_credentials" {
				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetOrCreatePhoneNumberUserFn = func(ctx context.Context, phone string) (*dto.CreatedUserResponse, error) {
					return &dto.CreatedUserResponse{
						UID:         "5cf354a2-1d3e-400d-8716-7e2aead29f2c",
						DisplayName: "John Doe",
						Email:       "johndoe@gmail.com",
						PhoneNumber: phoneNumber,
					}, nil
				}

				fakeInfraRepo.CreateUserProfileFn = func(ctx context.Context, phoneNumber, uid string) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "5cf354a2-1d3e-400d-8716-7e2aead29f2c",
					}, nil
				}

				fakeInfraRepo.GenerateAuthCredentialsFn = func(ctx context.Context, phone string, profile *profileutils.UserProfile) (*profileutils.AuthCredentialResponse, error) {
					return nil, fmt.Errorf("failed to generate auth credentials")
				}
			}

			if tt.name == "invalid:fail_to_set_userPin" {
				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetOrCreatePhoneNumberUserFn = func(ctx context.Context, phone string) (*dto.CreatedUserResponse, error) {
					return &dto.CreatedUserResponse{
						UID:         "5cf354a2-1d3e-400d-8716-7e2aead29f2c",
						DisplayName: "John Doe",
						Email:       "johndoe@gmail.com",
						PhoneNumber: phoneNumber,
					}, nil
				}

				fakeInfraRepo.CreateUserProfileFn = func(ctx context.Context, phoneNumber, uid string) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "5cf354a2-1d3e-400d-8716-7e2aead29f2c",
						PrimaryPhone: &phone,
					}, nil
				}

				fakeInfraRepo.GenerateAuthCredentialsFn = func(ctx context.Context, phone string, profile *profileutils.UserProfile) (*profileutils.AuthCredentialResponse, error) {
					customToken := uuid.New().String()
					idToken := uuid.New().String()
					refreshToken := uuid.New().String()
					return &profileutils.AuthCredentialResponse{
						CustomToken:  &customToken,
						IDToken:      &idToken,
						RefreshToken: refreshToken,
					}, nil
				}

				// Mock SetUserPin begins here
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254777886622"
					return &phone, nil
				}

				fakeInfraRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "123",
						PrimaryPhone: &phoneNumber,
					}, nil
				}

				fakePinExt.EncryptPINFn = func(rawPwd string, options *extension.Options) (string, string) {
					return "salt", "password"
				}

				fakeInfraRepo.SavePINFn = func(ctx context.Context, pin *domain.PIN) (bool, error) {
					return false, fmt.Errorf("failed to save user pin")
				}
			}

			_, err := i.CreateUserByPhone(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"SignUpUseCasesImpl.CreateUserByPhone() error = %v, wantErr %v",
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

func TestSignUpUseCasesImpl_VerifyPhoneNumber(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	type args struct {
		ctx   context.Context
		phone string
		appID *string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:successfully_verify_a_phonenumber",
			args: args{
				ctx:   ctx,
				phone: "+254777886622",
			},
			wantErr: false,
		},
		{
			name: "invalid:_phone_number_is_empty",
			args: args{
				ctx:   ctx,
				phone: "+",
			},
			wantErr: true,
		},
		{
			name: "invalid:_user_phone_already_exists",
			args: args{
				ctx:   ctx,
				phone: "+254777886622",
			},
			wantErr: true,
		},
		{
			name: "invalid:fail_to_generate_and_send_otp",
			args: args{
				ctx:   ctx,
				phone: "+254777886622",
			},
			wantErr: true,
		},
		{
			name: "invalid:_unable_to_check_if_phone_exists",
			args: args{
				ctx:   ctx,
				phone: "+254777886622",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "valid:successfully_verify_a_phonenumber" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254777886622"
					return &phone, nil
				}

				fakeInfraRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return false, nil
				}

				fakeEngagementSvs.GenerateAndSendOTPFn = func(ctx context.Context, phone string, appID *string) (*profileutils.OtpResponse, error) {
					return &profileutils.OtpResponse{OTP: "1234"}, nil
				}
			}

			if tt.name == "invalid:_phone_number_is_empty" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("empty phone number")
				}
			}

			if tt.name == "invalid:_user_phone_already_exists" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254777886622"
					return &phone, nil
				}

				fakeInfraRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return true, nil
				}
			}

			if tt.name == "invalid:_unable_to_check_if_phone_exists" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254777886622"
					return &phone, nil
				}

				fakeInfraRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return false, fmt.Errorf("unable to check if phone exists")
				}
			}

			if tt.name == "invalid:fail_to_generate_and_send_otp" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254777886622"
					return &phone, nil
				}

				fakeInfraRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return false, nil
				}

				fakeEngagementSvs.GenerateAndSendOTPFn = func(ctx context.Context, phone string, appID *string) (*profileutils.OtpResponse, error) {
					return nil, fmt.Errorf("failed to generate and send otp")
				}
			}

			_, err := i.VerifyPhoneNumber(tt.args.ctx, tt.args.phone, tt.args.appID)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"SignUpUseCasesImpl.VerifyPhoneNumber() error = %v, wantErr %v",
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

func TestSignUpUseCasesImpl_RemoveUserByPhoneNumber(t *testing.T) {
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
			name: "valid:successfully_RemoveUserByPhoneNumber",
			args: args{
				ctx:   ctx,
				phone: "+254799739102",
			},
			wantErr: false,
		},
		{
			name: "invalid:fail_to_RemoveUserByPhoneNumber",
			args: args{
				ctx:   ctx,
				phone: "+254799739102",
			},
			wantErr: true,
		},
		{
			name: "invalid:fail_to_normalize_phone",
			args: args{
				ctx:   ctx,
				phone: "+",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "valid:successfully_RemoveUserByPhoneNumber" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254777886622"
					return &phone, nil
				}

				fakeInfraRepo.PurgeUserByPhoneNumberFn = func(ctx context.Context, phone string) error {
					return nil
				}
			}

			if tt.name == "invalid:fail_to_RemoveUserByPhoneNumber" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254777886622"
					return &phone, nil
				}

				fakeInfraRepo.PurgeUserByPhoneNumberFn = func(ctx context.Context, phone string) error {
					return fmt.Errorf("failed to purge user by phonenumber")
				}
			}

			if tt.name == "invalid:fail_to_normalize_phone" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("failed to normalize phonenumber")
				}
			}

			err := i.RemoveUserByPhoneNumber(tt.args.ctx, tt.args.phone)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"SignUpUseCasesImpl.RemoveUserByPhoneNumber() error = %v, wantErr %v",
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

func TestSignUpUseCasesImpl_SetPhoneAsPrimary(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	type args struct {
		ctx   context.Context
		phone string
		otp   string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "valid:set_primary_phoneNumber",
			args: args{
				ctx:   ctx,
				phone: "+254795941530",
				otp:   "567291",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "invalid:fail_to_normalize_phoneNumber",
			args: args{
				ctx:   ctx,
				phone: "+",
				otp:   "567291",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "invalid:_unable_to_update_primary_phonenumber",
			args: args{
				ctx:   ctx,
				phone: "+25463728192",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "valid:set_primary_phoneNumber" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254777886622"
					return &phone, nil
				}

				// Begin Mocking SetPrimaryPhoneNumber
				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}

				// Begin Mocking UpdatePrimaryPhoneNumber
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

			if tt.name == "invalid:fail_to_normalize_phoneNumber" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("failed to normalize phonenumber")
				}
			}

			if tt.name == "invalid:_unable_to_update_primary_phonenumber" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254777886622"
					return &phone, nil
				}

				// Begin Mocking SetPrimaryPhoneNumber
				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}

				// Begin Mocking UpdatePrimaryPhoneNumber
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
					return fmt.Errorf("failed to update primary phone")
				}
			}

			got, err := i.SetPhoneAsPrimary(tt.args.ctx, tt.args.phone, tt.args.otp)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"SignUpUseCasesImpl.SetPhoneAsPrimary() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if got != tt.want {
				t.Errorf("SignUpUseCasesImpl.SetPhoneAsPrimary() = %v, want %v", got, tt.want)
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

func TestSignUpUseCasesImpl_RegisterPushToken(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	type args struct {
		ctx   context.Context
		token string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "valid:register_pushtoken",
			args: args{
				ctx:   ctx,
				token: uuid.New().String(),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "invalid:nil_token",
			args: args{
				ctx:   ctx,
				token: "",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "invalid:fail_to_get_userProfile",
			args: args{
				ctx:   ctx,
				token: uuid.New().String(),
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "invalid:fail_to_get_loggedInUser",
			args: args{
				ctx:   ctx,
				token: uuid.New().String(),
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "valid:register_pushtoken" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:        "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
						Suspended: false,
					}, nil
				}
				fakeInfraRepo.UpdatePushTokensFn = func(ctx context.Context, id string, pushToken []string) error {
					return nil
				}
			}

			if tt.name == "invalid:nil_token" {
				fakeInfraRepo.UpdatePushTokensFn = func(ctx context.Context, id string, pushToken []string) error {
					return fmt.Errorf("failed to register push token")
				}
			}

			if tt.name == "invalid:fail_to_get_userProfile" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("failed to get user profile")
				}
			}

			if tt.name == "invalid:fail_to_get_loggedInUser" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "", fmt.Errorf("failed to get logged in user")
				}
			}

			got, err := i.RegisterPushToken(tt.args.ctx, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"SignUpUseCasesImpl.RegisterPushToken() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if got != tt.want {
				t.Errorf("SignUpUseCasesImpl.RegisterPushToken() = %v, want %v", got, tt.want)
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

func TestSignUpUseCasesImpl_CompleteSignup(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	userFirstName := "John"
	userLastName := "Doe"

	type args struct {
		ctx     context.Context
		flavour feedlib.Flavour
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "valid:successfully_complete_signup",
			args: args{
				ctx:     ctx,
				flavour: feedlib.FlavourConsumer,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:successfully_complete_signup" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return uuid.New().String(), nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
						UserBioData: profileutils.BioData{
							FirstName: &userFirstName,
							LastName:  &userLastName,
						},
					}, nil
				}

				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "5cf354a2-1d3e-400d-8716-7e2aead29f2c",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
						VerifiedIdentifiers: []profileutils.VerifiedIdentifier{
							{
								UID: uuid.New().String(),
							},
						},
						UserBioData: profileutils.BioData{
							FirstName: &userFirstName,
							LastName:  &userLastName,
						},
					}, nil
				}

				fakePubSub.TopicIDsFn = func() []string {
					return []string{uuid.New().String()}
				}

				fakePubSub.EnsureTopicsExistFn = func(ctx context.Context, topicIDs []string) error {
					return nil
				}

				fakePubSub.AddPubSubNamespaceFn = func(topicName string) string {
					return uuid.New().String()
				}

				fakePubSub.PublishToPubsubFn = func(ctx context.Context, topicID string, payload []byte) error {
					return nil
				}
			}

			if tt.name == "invalid:fail_to_get_userProfile" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return uuid.New().String(), nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("failed to get user profile")
				}
			}

			if tt.name == "invalid:fail_to_get_loggedInUser" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "", fmt.Errorf("failed to get logged in user")
				}
			}

			if tt.name == "invalid:missing_bioData" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return uuid.New().String(), nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}

				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return uuid.New().String(), nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
						VerifiedIdentifiers: []profileutils.VerifiedIdentifier{
							{
								UID: uuid.New().String(),
							},
						},
					}, nil
				}

			}

			if tt.name == "invalid:invalid_flavour" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "", fmt.Errorf("invalid flavour defined")
				}
			}

			got, err := i.CompleteSignup(tt.args.ctx, tt.args.flavour)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"SignUpUseCasesImpl.CompleteSignup() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if got != tt.want {
				t.Errorf("SignUpUseCasesImpl.CompleteSignup() = %v, want %v", got, tt.want)
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

func TestSignUpUseCasesImpl_GetUserRecoveryPhoneNumbers(t *testing.T) {
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
		want    *dto.AccountRecoveryPhonesResponse
		wantErr bool
	}{
		{
			name: "valid:successfully_GetUserRecoveryPhoneNumbers",
			args: args{
				ctx:   ctx,
				phone: "+254766228822",
			},
			wantErr: false,
		},
		{
			name: "invalid:fail_to_normalize_phone",
			args: args{
				ctx:   ctx,
				phone: "+254766228822",
			},
			wantErr: true,
		},
		{
			name: "invalid:fail_to_get_userProfile",
			args: args{
				ctx:   ctx,
				phone: "+254766228822",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "valid:successfully_GetUserRecoveryPhoneNumbers" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}

				fakeInfraRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "123",
						PrimaryPhone: &phoneNumber,
						SecondaryPhoneNumbers: []string{
							"0744610111", "0794959697",
						},
					}, nil
				}
			}

			if tt.name == "invalid:fail_to_normalize_phone" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("failed to normalize phone")
				}
			}

			if tt.name == "invalid:fail_to_get_userProfile" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}

				fakeInfraRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("failed to get user profile")
				}
			}
			got, err := i.GetUserRecoveryPhoneNumbers(tt.args.ctx, tt.args.phone)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"SignUpUseCasesImpl.GetUserRecoveryPhoneNumbers() error = %v, wantErr %v",
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
				if got == nil {
					t.Errorf("returned a nil account recovery phone response")
					return
				}
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestSignUpUseCasesImpl_UpdateUserProfile(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	photoUploadID := "somePhotoUploadID"
	firstName := "John"
	lastName := "Doe"
	gender := enumutils.GenderMale
	dateOfBirth := scalarutils.Date{
		Year:  1990,
		Month: 3,
		Day:   10,
	}
	phone := gofakeit.Phone()
	validInput := &dto.UserProfileInput{
		PhotoUploadID: &photoUploadID,
		DateOfBirth:   &dateOfBirth,
		Gender:        &gender,
		FirstName:     &firstName,
		LastName:      &lastName,
	}
	invalidInput := &dto.UserProfileInput{
		PhotoUploadID: nil,
		DateOfBirth:   nil,
		Gender:        nil,
		FirstName:     nil,
		LastName:      nil,
	}
	type args struct {
		ctx   context.Context
		input *dto.UserProfileInput
	}
	tests := []struct {
		name    string
		args    args
		want    *profileutils.UserProfile
		wantErr bool
	}{
		{
			name: "valid:successfully_update_userProfile",
			args: args{
				ctx:   ctx,
				input: validInput,
			},
			wantErr: false,
		},
		{
			name: "invalid:missing_biodata",
			args: args{
				ctx:   ctx,
				input: invalidInput,
			},
			wantErr: true,
		},
		{
			name: "invalid:fail_to_updatePhotoUploadID",
			args: args{
				ctx:   ctx,
				input: validInput,
			},
			wantErr: true,
		},
		{
			name: "invalid:fail_to_getUserProfile",
			args: args{
				ctx:   ctx,
				input: validInput,
			},
			wantErr: true,
		},
		{
			name: "invalid:fail_to_getLoggedInUser",
			args: args{
				ctx:   ctx,
				input: validInput,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "valid:successfully_update_userProfile" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           uuid.New().String(),
						PrimaryPhone: &phone,
					}, nil
				}

				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           uuid.New().String(),
						PrimaryPhone: &phone,
					}, nil
				}

				fakeInfraRepo.UpdatePhotoUploadIDFn = func(ctx context.Context, id string, uploadID string) error {
					return nil
				}

				// Begin mocking UpdateBioData
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

			if tt.name == "invalid:missing_biodata" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}

				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}

				fakeInfraRepo.UpdatePhotoUploadIDFn = func(ctx context.Context, id string, uploadID string) error {
					return nil
				}

				// Begin mocking UpdateBioData
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}

				fakeInfraRepo.UpdateBioDataFn = func(ctx context.Context, id string, data profileutils.BioData) error {
					return fmt.Errorf("failed to update biodata")
				}
			}

			if tt.name == "invalid:fail_to_updatePhotoUploadID" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}

				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}

				fakeInfraRepo.UpdatePhotoUploadIDFn = func(ctx context.Context, id string, uploadID string) error {
					return fmt.Errorf("failed to update the photo upload ID")
				}
			}

			if tt.name == "invalid:fail_to_getUserProfile" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("failed to get a user profile")
				}
			}

			if tt.name == "invalid:fail_to_getLoggedInUser" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "", fmt.Errorf("failed to get logged in user")
				}
			}

			got, err := i.UpdateUserProfile(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"SignUpUseCasesImpl.UpdateUserProfile() error = %v, wantErr %v",
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
				if got == nil {
					t.Errorf("returned a nil account recovery phone response")
					return
				}
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}
		})
	}
}

func TestSignUpUseCasesImpl_RegisterUser(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	phoneNumber := interserviceclient.TestUserPhoneNumber
	fName := "Test"
	lName := "Test"
	email := "test@email.com"
	gender := "male"
	input := dto.RegisterUserInput{
		PhoneNumber: &phoneNumber,
		FirstName:   &fName,
		LastName:    &lName,
		Email:       &email,
		Gender:      (*enumutils.Gender)(&gender),
	}

	type args struct {
		ctx   context.Context
		input dto.RegisterUserInput
	}
	tests := []struct {
		name    string
		args    args
		want    *profileutils.UserProfile
		wantErr bool
	}{
		{
			name: "sad: unable to get logged in user",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad: unable to normalize phonenumber",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad: unable to create user profile",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad: unable to create communication settings",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad: unable to create otp",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad: unable to send otp sms",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "happy: registered consumer",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    &profileutils.UserProfile{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "sad: unable to get logged in user" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "", fmt.Errorf("unable to get logged in user")
				}
			}

			if tt.name == "sad: unable to normalize phonenumber" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return uuid.NewString(), nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: uuid.NewString()}, nil
				}
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("unable to normalize phone number")
				}
			}

			if tt.name == "sad: unable to create user profile" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return uuid.NewString(), nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: uuid.NewString()}, nil
				}
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return &phoneNumber, nil
				}
				fakeInfraRepo.CreateDetailedUserProfileFn = func(ctx context.Context, phoneNumber string, profile profileutils.UserProfile) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to create user profile")
				}
			}
			if tt.name == "sad: unable to create communication settings" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return uuid.NewString(), nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: uuid.NewString()}, nil
				}
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return &phoneNumber, nil
				}
				fakeInfraRepo.CreateDetailedUserProfileFn = func(ctx context.Context, phoneNumber string, profile profileutils.UserProfile) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           uuid.NewString(),
						PrimaryPhone: &phoneNumber,
					}, nil
				}
				fakeInfraRepo.SetUserCommunicationsSettingsFn = func(ctx context.Context, profileID string, allowWhatsApp, allowTextSms, allowPush, allowEmail *bool) (*profileutils.UserCommunicationsSetting, error) {
					return nil, fmt.Errorf("unable to create communication settings")
				}
			}

			if tt.name == "sad: unable to create otp" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return uuid.NewString(), nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: uuid.NewString()}, nil
				}
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return &phoneNumber, nil
				}
				fakeInfraRepo.CreateDetailedUserProfileFn = func(ctx context.Context, phoneNumber string, profile profileutils.UserProfile) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           uuid.NewString(),
						PrimaryPhone: &phoneNumber,
					}, nil
				}
				fakeInfraRepo.SetUserCommunicationsSettingsFn = func(ctx context.Context, profileID string, allowWhatsApp, allowTextSms, allowPush, allowEmail *bool) (*profileutils.UserCommunicationsSetting, error) {
					return &profileutils.UserCommunicationsSetting{}, nil
				}
				fakePinExt.GenerateTempPINFn = func(ctx context.Context) (string, error) {
					return "123", nil
				}
				fakePinExt.EncryptPINFn = func(rawPwd string, options *extension.Options) (string, string) {
					return "pin", "sha"
				}
				fakeInfraRepo.SavePINFn = func(ctx context.Context, pin *domain.PIN) (bool, error) {
					return false, fmt.Errorf("unable to create otp")
				}
			}

			if tt.name == "sad: unable to send otp sms" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return uuid.NewString(), nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: uuid.NewString()}, nil
				}
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return &phoneNumber, nil
				}
				fakeInfraRepo.CreateDetailedUserProfileFn = func(ctx context.Context, phoneNumber string, profile profileutils.UserProfile) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           uuid.NewString(),
						PrimaryPhone: &phoneNumber,
					}, nil
				}
				fakeInfraRepo.SetUserCommunicationsSettingsFn = func(ctx context.Context, profileID string, allowWhatsApp, allowTextSms, allowPush, allowEmail *bool) (*profileutils.UserCommunicationsSetting, error) {
					return &profileutils.UserCommunicationsSetting{}, nil
				}
				fakePinExt.GenerateTempPINFn = func(ctx context.Context) (string, error) {
					return "123", nil
				}
				fakePinExt.EncryptPINFn = func(rawPwd string, options *extension.Options) (string, string) {
					return "pin", "sha"
				}
				fakeInfraRepo.SavePINFn = func(ctx context.Context, pin *domain.PIN) (bool, error) {
					return true, nil
				}
				fakeEngagementSvs.SendSMSFn = func(ctx context.Context, phoneNumbers []string, message string) error {
					return fmt.Errorf("unable to send otp sms")
				}
			}

			if tt.name == "happy: registered consumer" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return uuid.NewString(), nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get user profile")
				}
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return &phoneNumber, nil
				}
				fakeInfraRepo.CreateDetailedUserProfileFn = func(ctx context.Context, phoneNumber string, profile profileutils.UserProfile) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.NewString(),
						UserBioData: profileutils.BioData{
							FirstName: &fName,
							LastName:  &lName,
						},
						PrimaryPhone:        &phoneNumber,
						PrimaryEmailAddress: &email,
					}, nil
				}
				fakeInfraRepo.SetUserCommunicationsSettingsFn = func(ctx context.Context, profileID string, allowWhatsApp, allowTextSms, allowPush, allowEmail *bool) (*profileutils.UserCommunicationsSetting, error) {
					return &profileutils.UserCommunicationsSetting{}, nil
				}
				fakePinExt.GenerateTempPINFn = func(ctx context.Context) (string, error) {
					return "123", nil
				}
				fakePinExt.EncryptPINFn = func(rawPwd string, options *extension.Options) (string, string) {
					return "pin", "sha"
				}
				fakeInfraRepo.SavePINFn = func(ctx context.Context, pin *domain.PIN) (bool, error) {
					return true, nil
				}
				fakeEngagementSvs.SendSMSFn = func(ctx context.Context, phoneNumbers []string, message string) error {
					return nil
				}
			}

			got, err := i.RegisterUser(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SignUpUseCasesImpl.RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("SignUpUseCasesImpl.RegisterUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
