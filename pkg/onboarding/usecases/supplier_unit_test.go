package usecases_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/utils"
	"github.com/savannahghi/profileutils"
	"gitlab.slade360emr.com/go/apiclient"
)

const (
	testEmail = "test@bewell.co.ke"
)

func TestProfileUseCaseImpl_FindSupplierByID(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:_find_supplier_by_id",
			args: args{
				ctx: ctx,
				id:  "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
			},
			wantErr: false,
		},
		{
			name: "invalid:_find_supplier_by_id_fails",
			args: args{
				ctx: ctx,
				id:  "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "valid:_find_supplier_by_id" {
				fakeRepo.GetSupplierProfileByIDFn = func(ctx context.Context, id string) (*profileutils.Supplier, error) {
					return &profileutils.Supplier{
						ID: "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					}, nil
				}

			}

			if tt.name == "invalid:_find_supplier_by_id_fails" {
				fakeRepo.GetSupplierProfileByIDFn = func(ctx context.Context, id string) (*profileutils.Supplier, error) {
					return nil, fmt.Errorf("unable to get supp;ier profile")
				}

			}

			sup, err := i.Supplier.FindSupplierByID(tt.args.ctx, tt.args.id)
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

				if sup.ID == "" {
					t.Errorf("empty ID returned %v", sup.ID)
					return
				}
			}

		})
	}
}

func TestSupplierUseCasesImpl_SuspendSupplier(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	suspensionReason := `
	"This email is to inform you that as a result of your actions on April 12th, 2021, you have been issued a suspension for 1 week (7 days)"
	`

	type args struct {
		ctx              context.Context
		suspensionReason *string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "valid:successfully_suspend_supplier",
			args: args{
				ctx:              ctx,
				suspensionReason: &suspensionReason,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "invalid:fail_to_suspend_supplier",
			args: args{
				ctx:              ctx,
				suspensionReason: &suspensionReason,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "invalid:fail_to_get_user_profile",
			args: args{
				ctx:              ctx,
				suspensionReason: &suspensionReason,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "invalid:fail_to_get_supplier_profile",
			args: args{
				ctx:              ctx,
				suspensionReason: &suspensionReason,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "invalid:fail_to_get_logged_in_user",
			args: args{
				ctx:              ctx,
				suspensionReason: &suspensionReason,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "valid:successfully_suspend_supplier" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "FSO798-AD3-bvihjskdn", nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					}, nil
				}

				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "7e2aead29f2c", nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					email := testEmail
					firstName := "Makmende"
					primaryPhoneNumber := interserviceclient.TestUserPhoneNumber
					return &profileutils.UserProfile{
						ID:                  "400d-8716--91bd-42b3af315a4e",
						PrimaryPhone:        &primaryPhoneNumber,
						PrimaryEmailAddress: &email,
						UserBioData: profileutils.BioData{
							FirstName: &firstName,
							LastName:  &firstName,
						},
						VerifiedIdentifiers: []profileutils.VerifiedIdentifier{
							{
								UID: "f4f39af7-91bd-42b3af315a4e",
							},
						},
					}, nil
				}
				fakeRepo.GetSupplierProfileByProfileIDFn = func(ctx context.Context, profileID string) (*profileutils.Supplier, error) {
					return &profileutils.Supplier{
						ProfileID:    &profileID,
						KYCSubmitted: false,
					}, nil
				}

				fakeRepo.GetSupplierProfileByUIDFn = func(ctx context.Context, uid string) (*profileutils.Supplier, error) {
					return &profileutils.Supplier{
						ID: "-91bd-42b3af315a4e",
					}, nil
				}

				fakeRepo.UpdateSupplierProfileFn = func(ctx context.Context, profileID string, data *profileutils.Supplier) error {
					return nil
				}
				fakeEngagementSvs.NotifySupplierOnSuspensionFn = func(ctx context.Context, input dto.EmailNotificationPayload) error {
					return nil
				}
			}

			if tt.name == "invalid:fail_to_suspend_supplier" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "FSO798-AD3-bvihjskdn", nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					}, nil
				}

				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "7e2aead29f2c", nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "400d-8716--91bd-42b3af315a4e",
						VerifiedIdentifiers: []profileutils.VerifiedIdentifier{
							{
								UID: "f4f39af7-91bd-42b3af315a4e",
							},
						},
					}, nil
				}
				fakeRepo.GetSupplierProfileByProfileIDFn = func(ctx context.Context, profileID string) (*profileutils.Supplier, error) {
					return &profileutils.Supplier{
						ProfileID: &profileID,
					}, nil
				}

				fakeRepo.GetSupplierProfileByUIDFn = func(ctx context.Context, uid string) (*profileutils.Supplier, error) {
					return &profileutils.Supplier{
						ID: "-91bd-42b3af315a4e",
					}, nil
				}

				fakeRepo.UpdateSupplierProfileFn = func(ctx context.Context, profileID string, data *profileutils.Supplier) error {
					return fmt.Errorf("failed tp suspend supplier")
				}
			}

			if tt.name == "invalid:fail_to_get_user_profile" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "FSO798-AD3", nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("failed to get a user profile")
				}
			}

			if tt.name == "invalid:fail_to_get_supplier_profile" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "FSO798-AD3-bvihjskdn", nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					}, nil
				}

				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "7e2aead29f2c", nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "400d-8716--91bd-42b3af315a4e",
						VerifiedIdentifiers: []profileutils.VerifiedIdentifier{
							{
								UID: "f4f39af7-91bd-42b3af315a4e",
							},
						},
					}, nil
				}
				fakeRepo.GetSupplierProfileByProfileIDFn = func(ctx context.Context, profileID string) (*profileutils.Supplier, error) {
					return &profileutils.Supplier{
						ProfileID: &profileID,
					}, nil
				}

				fakeRepo.GetSupplierProfileByUIDFn = func(ctx context.Context, uid string) (*profileutils.Supplier, error) {
					return nil, fmt.Errorf("failed to get supplier profile")
				}
			}

			if tt.name == "invalid:fail_to_get_logged_in_user" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "", fmt.Errorf("failed to get logged in user")
				}
			}

			got, err := i.Supplier.SuspendSupplier(tt.args.ctx, tt.args.suspensionReason)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"SupplierUseCasesImpl.SuspendSupplier() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if got != tt.want {
				t.Errorf("SupplierUseCasesImpl.SuspendSupplier() = %v, want %v", got, tt.want)
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

func TestUnitSupplierUseCasesImpl_AddPartnerType(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	testRiderName := "Test Rider"
	rider := profileutils.PartnerTypeRider

	type args struct {
		ctx         context.Context
		name        *string
		partnerType *profileutils.PartnerType
	}
	tests := []struct {
		name        string
		args        args
		want        bool
		wantErr     bool
		expectedErr string
	}{
		{
			name: "valid:add_partner_type",
			args: args{
				ctx:         ctx,
				name:        &testRiderName,
				partnerType: &rider,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "invalid : missing name arg",
			args: args{
				ctx: ctx,
			},
			want:        false,
			wantErr:     true,
			expectedErr: "expected `name` to be defined and `partnerType` to be valid",
		},
		{
			name: "invalid:unable_to_login",
			args: args{
				ctx:         ctx,
				name:        &testRiderName,
				partnerType: &rider,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "invalid:unable_to_get_user_profile_by_id",
			args: args{
				ctx:         ctx,
				name:        &testRiderName,
				partnerType: &rider,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "invalid:unable_to_add_partner_type",
			args: args{
				ctx:         ctx,
				name:        &testRiderName,
				partnerType: &rider,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "valid:add_partner_type" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "5cf354a2-1d3e-400d-8716-7e2aead29f2c",
						Email:       testEmail,
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					}, nil
				}
				fakeRepo.AddPartnerTypeFn = func(ctx context.Context, profileID string, name *string, partnerType *profileutils.PartnerType) (bool, error) {
					return true, nil
				}
			}

			if tt.name == "invalid:unable_to_login" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("unable to login")
				}
			}

			if tt.name == "invalid:unable_to_get_user_profile_by_id" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "5cf354a2-1d3e-400d-8716-7e2aead29f2c",
						Email:       testEmail,
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get profile by uid")
				}

			}

			if tt.name == "invalid:unable_to_add_partner_type" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "5cf354a2-1d3e-400d-8716-7e2aead29f2c",
						Email:       testEmail,
						PhoneNumber: "0721568526",
					}, nil
				}
				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspend bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					}, nil
				}
				fakeRepo.AddPartnerTypeFn = func(ctx context.Context, profileID string, name *string, partnerType *profileutils.PartnerType) (bool, error) {
					return false, fmt.Errorf("unable to add partner type")
				}
			}

			got, err := i.Supplier.AddPartnerType(tt.args.ctx, tt.args.name, tt.args.partnerType)

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

				if got != tt.want {
					t.Errorf("expected %v got %v  ", tt.want, got)
					return
				}
			}

		})
	}
}

func TestProfileUseCaseImpl_FindSupplierByUID(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}
	profileID := "93ca42bb-5cfc-4499-b137-2df4d67b4a21"
	supplier := &profileutils.Supplier{
		ProfileID: &profileID,
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *profileutils.Supplier
		wantErr bool
	}{
		{
			name: "valid:_find_supplier_by_uid",
			args: args{
				ctx: ctx,
			},
			want:    supplier,
			wantErr: false,
		},
		{
			name: "invalid:_find_supplier_by_uid",
			args: args{
				ctx: ctx,
			},
			want:    supplier,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "valid:_find_supplier_by_uid" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-87167-e2aead29f2c", nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "5b64-4c2f-15a4e-f4f39af791bd-42b3af3",
					}, nil
				}
				fakeRepo.GetSupplierProfileByProfileIDFn = func(ctx context.Context, profileID string) (*profileutils.Supplier, error) {
					return &profileutils.Supplier{
						ID:        "93ca42bb-5cfc-4499-b137-2df4d67b4a21",
						ProfileID: &profileID,
					}, nil
				}

			}

			if tt.name == "invalid:_find_supplier_by_uid" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-87167-e2aead29f2c", nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "5b64-4c2f-15a4e-f4f39af791bd-42b3af3",
					}, nil
				}
				fakeRepo.GetSupplierProfileByProfileIDFn = func(ctx context.Context, profileID string) (*profileutils.Supplier, error) {
					return nil, fmt.Errorf("failed to get supplier")
				}

			}

			sup, err := i.Supplier.FindSupplierByUID(tt.args.ctx)
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

				if *sup.ProfileID == "" {
					t.Errorf("empty profileID returned %v", sup.ProfileID)
					return
				}
			}

		})
	}
}

func TestSupplierUseCasesImpl_CreateCustomerAccount(t *testing.T) {
	ctx := context.Background()

	_, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	type args struct {
		ctx         context.Context
		name        string
		partnerType profileutils.PartnerType
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy:)",
			args: args{
				ctx:         ctx,
				name:        *utils.GetRandomName(),
				partnerType: profileutils.PartnerTypeConsumer,
			},
			wantErr: false,
		},
		{
			name: "sad:( currency not found",
			args: args{
				ctx:         ctx,
				name:        *utils.GetRandomName(),
				partnerType: profileutils.PartnerTypeConsumer,
			},
			wantErr: true,
		},
		{
			name: "sad:( can't get logged in user",
			args: args{
				ctx:         ctx,
				name:        *utils.GetRandomName(),
				partnerType: profileutils.PartnerTypeConsumer,
			},
			wantErr: true,
		},
		{
			name: "sad:( failed to publish to PubSub",
			args: args{
				ctx:         ctx,
				name:        *utils.GetRandomName(),
				partnerType: profileutils.PartnerTypeConsumer,
			},
			wantErr: false, // TODO: Fix and return  to false
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "happy:)" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "5cf354a2-1d3e-400d-8716-7e2aead29f2c",
						Email:       testEmail,
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeBaseExt.FetchDefaultCurrencyFn = func(c apiclient.Client) (*apiclient.FinancialYearAndCurrency, error) {
					id := uuid.New().String()
					return &apiclient.FinancialYearAndCurrency{
						ID: &id,
					}, nil
				}

				fakePubSub.NotifyCreateCustomerFn = func(ctx context.Context, data dto.CustomerPubSubMessage) error {
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

			if tt.name == "sad:( can't get logged in user" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "", fmt.Errorf("error")
				}
			}

			if tt.name == "sad:( currency not found" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return uuid.New().String(), nil
				}

				fakeBaseExt.FetchDefaultCurrencyFn = func(c apiclient.Client) (*apiclient.FinancialYearAndCurrency, error) {
					return nil, fmt.Errorf("fail to fetch default currency")
				}
			}

			if tt.name == "sad:( failed to publish to PubSub" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return uuid.New().String(), nil
				}

				fakeBaseExt.FetchDefaultCurrencyFn = func(c apiclient.Client) (*apiclient.FinancialYearAndCurrency, error) {
					id := uuid.New().String()
					return &apiclient.FinancialYearAndCurrency{
						ID: &id,
					}, nil
				}

				fakePubSub.TopicIDsFn = func() []string {
					return []string{uuid.New().String()}
				}

				fakePubSub.AddPubSubNamespaceFn = func(topicName string) string {
					return uuid.New().String()
				}

				fakePubSub.PublishToPubsubFn = func(ctx context.Context, topicID string, payload []byte) error {
					return fmt.Errorf("error")
				}
			}

		})
	}
}
