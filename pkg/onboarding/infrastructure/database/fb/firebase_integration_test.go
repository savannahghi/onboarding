package fb_test

import (
	"context"
	"log"
	"testing"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/utils"
	"gitlab.slade360emr.com/go/commontools/crm/pkg/infrastructure/services/hubspot"

	"fmt"

	"os"
	"reflect"

	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/database/fb"
	"github.com/savannahghi/profileutils"
	"github.com/savannahghi/scalarutils"
	"github.com/savannahghi/serverutils"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"firebase.google.com/go/auth"

	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/engagement"
	erp "gitlab.slade360emr.com/go/commontools/accounting/pkg/usecases"

	crmExt "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/crm"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/messaging"
	pubsubmessaging "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/pubsub"
	"github.com/savannahghi/onboarding/pkg/onboarding/presentation/interactor"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases"
	hubspotRepo "gitlab.slade360emr.com/go/commontools/crm/pkg/infrastructure/database/fs"
	hubspotUsecases "gitlab.slade360emr.com/go/commontools/crm/pkg/usecases"
)

const (
	engagementService = "engagement"
)

func TestMain(m *testing.M) {
	log.Printf("Setting tests up ...")
	envOriginalValue := os.Getenv("ENVIRONMENT")
	os.Setenv("ENVIRONMENT", "staging")
	debugEnvValue := os.Getenv("DEBUG")
	os.Setenv("DEBUG", "true")
	os.Setenv("REPOSITORY", "firebase")
	collectionEnvValue := os.Getenv("ROOT_COLLECTION_SUFFIX")

	// !NOTE!
	// Under no circumstances should you remove this env var when testing
	// You risk purging important collections, like our prod collections
	os.Setenv("ROOT_COLLECTION_SUFFIX", fmt.Sprintf("onboarding_ci_%v", time.Now().Unix()))
	ctx := context.Background()
	r := fb.Repository{} // They are nil
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}

	purgeRecords := func() {
		collections := []string{
			r.GetCustomerProfileCollectionName(),
			r.GetPINsCollectionName(),
			r.GetUserProfileCollectionName(),
			r.GetSupplierProfileCollectionName(),
			r.GetSurveyCollectionName(),
			r.GetCommunicationsSettingsCollectionName(),
			r.GetCustomerProfileCollectionName(),
			r.GetExperimentParticipantCollectionName(),
			r.GetKCYProcessCollectionName(),
			r.GetMarketingDataCollectionName(),
			r.GetNHIFDetailsCollectionName(),
			r.GetProfileNudgesCollectionName(),
			r.GetSMSCollectionName(),
			r.GetUSSDDataCollectionName(),
			r.GetRolesCollectionName(),
		}
		for _, collection := range collections {
			ref := fsc.Collection(collection)
			firebasetools.DeleteCollection(ctx, fsc, ref, 10)
		}
	}

	// try clean up first
	purgeRecords()

	// do clean up
	log.Printf("Running tests ...")
	code := m.Run()

	log.Printf("Tearing tests down ...")
	purgeRecords()

	// restore environment varibles to original values
	os.Setenv(envOriginalValue, "ENVIRONMENT")
	os.Setenv("DEBUG", debugEnvValue)
	os.Setenv("ROOT_COLLECTION_SUFFIX", collectionEnvValue)

	os.Exit(code)
}

func InitializeTestService(ctx context.Context) (*interactor.Interactor, error) {
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}

	projectID, err := serverutils.GetEnvVar(serverutils.GoogleCloudProjectIDEnvVarName)
	if err != nil {
		return nil, fmt.Errorf(
			"can't get projectID from env var `%s`: %w",
			serverutils.GoogleCloudProjectIDEnvVarName,
			err,
		)
	}
	pubSubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize pubsub client: %w", err)
	}

	ext := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})

	// Initialize ISC clients
	engagementClient := utils.NewInterServiceClient(engagementService, ext)

	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)
	erp := erp.NewAccounting()
	engage := engagement.NewServiceEngagementImpl(engagementClient, ext)
	// hubspot usecases
	hubspotService := hubspot.NewHubSpotService()
	hubspotfr, err := hubspotRepo.NewHubSpotFirebaseRepository(ctx, hubspotService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize hubspot crm repository: %w", err)
	}
	hubspotUsecases := hubspotUsecases.NewHubSpotUsecases(hubspotfr)
	crmExt := crmExt.NewCrmService(hubspotUsecases)
	ps, err := pubsubmessaging.NewServicePubSubMessaging(
		pubSubClient,
		ext,
		erp,
		crmExt,
		fr,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize new pubsub messaging service: %w", err)
	}
	mes := messaging.NewServiceMessagingImpl(ext)
	pinExt := extension.NewPINExtensionImpl()
	profile := usecases.NewProfileUseCase(fr, ext, engage, ps, crmExt)
	supplier := usecases.NewSupplierUseCases(fr, profile, erp, engage, mes, ext, ps)
	survey := usecases.NewSurveyUseCases(fr, ext)
	userpin := usecases.NewUserPinUseCase(fr, profile, ext, pinExt, engage)
	su := usecases.NewSignUpUseCases(fr, profile, userpin, supplier, ext, engage, ps)

	return &interactor.Interactor{
		Onboarding: profile,
		Signup:     su,
		Supplier:   supplier,
		Survey:     survey,
		UserPIN:    userpin,
		ERP:        erp,
		Engagement: engage,
		PubSub:     ps,
		CrmExt:     crmExt,
	}, nil
}

// func generateTestOTP(t *testing.T, phone string) (*profileutils.OtpResponse, error) {
// 	ctx := context.Background()
// 	s, err := InitializeTestService(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to initialize test service: %v", err)
// 	}
// 	return s.Engagement.GenerateAndSendOTP(ctx, phone)
// }

func InitializeTestFirebaseClient(ctx context.Context) (*firestore.Client, *auth.Client) {
	fc := firebasetools.FirebaseClient{}
	fa, err := fc.InitFirebase()
	if err != nil {
		log.Panicf("unable to initialize Firebase: %s", err)
	}

	fsc, err := fa.Firestore(ctx)
	if err != nil {
		log.Panicf("unable to initialize Firestore: %s", err)
	}

	fbc, err := fa.Auth(ctx)
	if err != nil {
		log.Panicf("can't initialize Firebase auth when setting up tests: %s", err)
	}
	return fsc, fbc
}

func TestCreateEmptyCustomerProfile(t *testing.T) {
	ctx := context.Background()
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	firestoreDB := fb.NewFirebaseRepository(firestoreExtension, fbc)

	tests := []struct {
		name      string
		profileID string
		wantErr   bool
	}{
		{
			name:      "valid case",
			profileID: uuid.New().String(),
			wantErr:   false,
		},
		{
			name:      "invalid case",
			profileID: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, err := firestoreDB.CreateEmptyCustomerProfile(ctx, tt.profileID)
			if tt.wantErr && err != nil {
				t.Errorf("error expected but returned no error")
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("error was not expected but got error: %v", err)
				return
			}

			if !tt.wantErr && customer == nil {
				t.Errorf("returned a nil customer")
				return
			}
		})
	}

}

func TestGetCustomerProfileByProfileID(t *testing.T) {
	ctx := context.Background()
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	firestoreDB := fb.NewFirebaseRepository(firestoreExtension, fbc)
	tests := []struct {
		name      string
		profileID string
		wantErr   bool
	}{
		{
			name:      "valid case",
			profileID: uuid.New().String(),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customerTest, err := firestoreDB.CreateEmptyCustomerProfile(ctx, tt.profileID)
			if err != nil {
				t.Errorf("failed to create a test Empty Customer profile err: %v", err)
				return
			}
			if customerTest.ProfileID == nil {
				t.Errorf("nil customer profile ID")
				return
			}
			customerProfile, err := firestoreDB.GetCustomerProfileByProfileID(ctx, tt.profileID)
			if err != nil && !tt.wantErr {
				t.Errorf("error not expected but got error: %v", err)
				return
			}
			if tt.wantErr && err == nil {
				t.Errorf("error expected but got no error")
				return
			}
			if !tt.wantErr && customerProfile == nil {
				t.Errorf("nil customer profile")
				return
			}

			if !tt.wantErr {
				if customerTest.ProfileID == nil {
					t.Errorf("nil customer profile ID")
					return
				}

				if customerTest.ID == "" {
					t.Errorf("nil customer ID")
					return
				}
			}
		})
	}
}

func TestRepository_GetCustomerOrSupplierProfileByProfileID(t *testing.T) {
	ctx := context.Background()
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	profileID := uuid.New().String()
	_, err := fr.CreateEmptySupplierProfile(ctx, profileID)
	if err != nil {
		t.Errorf("failed to create an empty supplier: %v", err)
	}

	_, err = fr.CreateEmptyCustomerProfile(ctx, profileID)
	if err != nil {
		t.Errorf("failed to create an empty customer: %v", err)
	}
	type args struct {
		ctx       context.Context
		flavour   feedlib.Flavour
		profileID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success: get the customer profile",
			args: args{
				ctx:       ctx,
				flavour:   feedlib.FlavourConsumer,
				profileID: profileID,
			},
			wantErr: false,
		},
		{
			name: "success: get the supplier profile",
			args: args{
				ctx:       ctx,
				flavour:   feedlib.FlavourPro,
				profileID: profileID,
			},
			wantErr: false,
		},
		{
			name: "failure: bad flavour given",
			args: args{
				ctx:       ctx,
				flavour:   "not-a-flavour-bana",
				profileID: profileID,
			},
			wantErr: true,
		},
		{
			name: "failure: profile ID that does not exist",
			args: args{
				ctx:       ctx,
				flavour:   feedlib.FlavourPro,
				profileID: "not-a-real-profile-ID",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, supplier, err := fr.GetCustomerOrSupplierProfileByProfileID(
				tt.args.ctx,
				tt.args.flavour,
				tt.args.profileID,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.GetCustomerOrSupplierProfileByProfileID() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}

			if serverutils.IsDebug() {
				log.Printf("Customer....%v", customer)
				log.Printf("Supplier....%v", supplier)
			}
		})
	}
}

func TestRepository_GetCustomerProfileByID(t *testing.T) {
	ctx := context.Background()
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	profileID := uuid.New().String()

	customer, err := fr.CreateEmptyCustomerProfile(ctx, profileID)
	if err != nil {
		t.Errorf("failed to create an empty customer: %v", err)
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
			name: "success: get a customer profile by ID",
			args: args{
				ctx: ctx,
				id:  customer.ID,
			},
			wantErr: false,
		},
		{
			name: "failure: failed to get a customer profile",
			args: args{
				ctx: ctx,
				id:  "not-a-customer-ID",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customerProfile, err := fr.GetCustomerProfileByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.GetCustomerProfileByID() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if serverutils.IsDebug() {
				log.Printf("Customer....%v", customerProfile)
			}
		})
	}
}

func TestRepository_GetSupplierProfileByProfileID(t *testing.T) {
	ctx := context.Background()

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	profileID := uuid.New().String()

	sup, err := fr.CreateEmptySupplierProfile(ctx, profileID)
	if err != nil {
		t.Errorf("failed to create an empty supplier: %v", err)
	}

	type args struct {
		ctx       context.Context
		profileID string
	}
	tests := []struct {
		name    string
		args    args
		want    *profileutils.Supplier
		wantErr bool
	}{
		{
			name: "Happy Case - Get Supplier Profile By Valid profile ID",
			args: args{
				ctx:       ctx,
				profileID: profileID,
			},
			want:    sup,
			wantErr: false,
		},
		{
			name: "Sad Case - Get Supplier Profile By a non-existent profile ID",
			args: args{
				ctx:       ctx,
				profileID: "bogus",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fr.GetSupplierProfileByProfileID(tt.args.ctx, tt.args.profileID)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.GetSupplierProfileByProfileID() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Repository.GetSupplierProfileByProfileID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_ActivateSupplierProfile(t *testing.T) {
	ctx := context.Background()

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	profileID := uuid.New().String()

	_, err := fr.CreateEmptySupplierProfile(ctx, profileID)
	if err != nil {
		t.Errorf("failed to create an empty supplier: %v", err)
	}

	sup := profileutils.Supplier{
		Active: true,
		PayablesAccount: &profileutils.PayablesAccount{
			ID: uuid.New().String(),
		},
	}

	type args struct {
		ctx       context.Context
		profileID string
		supplier  profileutils.Supplier
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Case - Activate Supplier By Valid profile ID",
			args: args{
				ctx:       ctx,
				profileID: profileID,
				supplier:  sup,
			},
			wantErr: false,
		},
		{
			name: "Sad Case - Activate Supplier By a non-existent profile ID",
			args: args{
				ctx:       ctx,
				profileID: "bogus",
				supplier:  sup,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			supp, err := fr.ActivateSupplierProfile(
				tt.args.ctx,
				tt.args.profileID,
				tt.args.supplier,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.ActivateSupplierProfile() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if supp != nil {
				if !supp.Active && supp.SupplierID == "" && supp.PayablesAccount.ID == "" {
					t.Errorf("expected an active supplier")
					return
				}
			}
		})
	}
}

func TestRepository_AddPartnerType(t *testing.T) {
	ctx := context.Background()

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	testRiderName := "Test Rider"
	rider := profileutils.PartnerTypeRider
	testPractitionerName := "Test Practitioner"
	practitioner := profileutils.PartnerTypePractitioner
	testProviderName := "Test Provider"
	provider := profileutils.PartnerTypeProvider

	profileID := uuid.New().String()

	supplier, err := fr.CreateEmptySupplierProfile(ctx, profileID)
	if err != nil {
		t.Errorf("failed to create an empty supplier: %v", err)
	}

	type args struct {
		ctx         context.Context
		profileID   string
		name        *string
		partnerType *profileutils.PartnerType
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Happy Case - Add a valid rider partner type",
			args: args{
				ctx:         ctx,
				profileID:   *supplier.ProfileID,
				name:        &testRiderName,
				partnerType: &rider,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Happy Case - Add a valid practitioner partner type",
			args: args{
				ctx:         ctx,
				profileID:   *supplier.ProfileID,
				name:        &testPractitionerName,
				partnerType: &practitioner,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Happy Case - Add a valid provider partner type",
			args: args{
				ctx:         ctx,
				profileID:   *supplier.ProfileID,
				name:        &testProviderName,
				partnerType: &provider,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Sad Case - Use an invalid ID",
			args: args{
				ctx:         ctx,
				profileID:   "invalidid",
				name:        &testProviderName,
				partnerType: &provider,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fr.AddPartnerType(
				tt.args.ctx,
				tt.args.profileID,
				tt.args.name,
				tt.args.partnerType,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.AddPartnerType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected, got %v", err)
					return
				}
			}

			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected, got %v", err)
					return
				}
			}

			if got != tt.want {
				t.Errorf("Repository.AddPartnerType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_UpdateSupplierProfile(t *testing.T) {
	ctx := context.Background()

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	profileID := uuid.New().String()

	supplier, err := fr.CreateEmptySupplierProfile(ctx, profileID)
	if err != nil {
		t.Errorf("failed to create an empty supplier: %v", err)
	}

	validPayload := &profileutils.Supplier{
		ID:        supplier.ID,
		ProfileID: supplier.ProfileID,
		Active:    true,
	}
	newprofileID := uuid.New().String()
	invalidPayload := &profileutils.Supplier{
		ID:        uuid.New().String(),
		ProfileID: &newprofileID,
		Active:    true,
	}

	type args struct {
		ctx  context.Context
		data *profileutils.Supplier
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Case - Update Supplier Profile Supplier By Valid payload",
			args: args{
				ctx:  ctx,
				data: validPayload,
			},
			wantErr: false,
		},
		{
			name: "Sad Case - Update Supplier Profile By invalid payload",
			args: args{
				ctx:  ctx,
				data: invalidPayload,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fr.UpdateSupplierProfile(tt.args.ctx, *tt.args.data.ProfileID, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.UpdateSupplierProfile() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
		})
	}
}

func TestRepositoryFetchKYCProcessingRequests(t *testing.T) {
	ctx := context.Background()

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	reqPartnerType := profileutils.PartnerTypeCoach
	organizationTypeLimitedCompany := domain.OrganizationTypeLimitedCompany
	id := uuid.New().String()
	kycReq := &domain.KYCRequest{
		ID:                  id,
		ReqPartnerType:      reqPartnerType,
		ReqOrganizationType: organizationTypeLimitedCompany,
		Status:              domain.KYCProcessStatusApproved,
	}

	err := fr.StageKYCProcessingRequest(ctx, kycReq)
	if err != nil {
		t.Errorf("failed to stage kyc: %v", err)
		return
	}

	kycRequests := []*domain.KYCRequest{}
	kycRequests = append(kycRequests, kycReq)

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    []*domain.KYCRequest
		wantErr bool
	}{
		{
			name: "Happy Case - Fetch KYC Processing Requests",
			args: args{
				ctx: ctx,
			},
			want:    kycRequests,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fr.FetchKYCProcessingRequests(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.FetchKYCProcessingRequests() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Repository.FetchKYCProcessingRequests() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepositoryFetchKYCProcessingRequestByID(t *testing.T) {
	ctx := context.Background()

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	reqPartnerType := profileutils.PartnerTypeCoach
	organizationTypeLimitedCompany := domain.OrganizationTypeLimitedCompany
	id := uuid.New().String()
	kycReq := &domain.KYCRequest{
		ID:                  id,
		ReqPartnerType:      reqPartnerType,
		ReqOrganizationType: organizationTypeLimitedCompany,
	}

	err := fr.StageKYCProcessingRequest(ctx, kycReq)
	if err != nil {
		t.Errorf("failed to stage kyc: %v", err)
		return
	}

	kycRequests := []*domain.KYCRequest{}
	kycRequests = append(kycRequests, kycReq)

	kycRequest := kycRequests[0]

	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		want    *domain.KYCRequest
		wantErr bool
	}{
		{
			name: "Happy Case - Fetch KYC Processing Requests by ID",
			args: args{
				ctx: ctx,
				id:  kycRequest.ID,
			},
			want:    kycRequest,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fr.FetchKYCProcessingRequestByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.FetchKYCProcessingRequestByID() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected, got %v", err)
					return
				}
			}

			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected, got %v", err)
					return
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Repository.FetchKYCProcessingRequestByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepositoryUpdateKYCProcessingRequest(t *testing.T) {
	ctx := context.Background()

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	reqPartnerType := profileutils.PartnerTypeCoach
	organizationTypeLimitedCompany := domain.OrganizationTypeLimitedCompany
	id := uuid.New().String()
	kycReq := &domain.KYCRequest{
		ID:                  id,
		ReqPartnerType:      reqPartnerType,
		ReqOrganizationType: organizationTypeLimitedCompany,
	}

	err := fr.StageKYCProcessingRequest(ctx, kycReq)
	if err != nil {
		t.Errorf("failed to stage kyc: %v", err)
		return
	}

	kycRequests := []*domain.KYCRequest{}
	kycRequests = append(kycRequests, kycReq)

	kycRequest := kycRequests[0]

	kycStatus := domain.KYCProcessStatusApproved

	updatedKYCReq := &domain.KYCRequest{
		ID:     kycRequest.ID,
		Status: kycStatus,
	}

	updatedKYCRequests := []*domain.KYCRequest{}
	updatedKYCRequests = append(updatedKYCRequests, updatedKYCReq)

	updatedKYCRequest := updatedKYCRequests[0]

	type args struct {
		ctx        context.Context
		kycRequest *domain.KYCRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Case - Update KYC Processing Requests",
			args: args{
				ctx:        ctx,
				kycRequest: updatedKYCRequest,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fr.UpdateKYCProcessingRequest(tt.args.ctx, tt.args.kycRequest); (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.UpdateKYCProcessingRequest() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
		if tt.wantErr {
			if err == nil {
				t.Errorf("error expected, got %v", err)
				return
			}
		}

		if !tt.wantErr {
			if err != nil {
				t.Errorf("error not expected, got %v", err)
				return
			}
		}
	}
}

func TestRepositoryGenerateAuthCredentialsForAnonymousUser(t *testing.T) {
	ctx := context.Background()

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	anonymousPhoneNumber := "+254700000000"

	user, err := fr.GetOrCreatePhoneNumberUser(ctx, anonymousPhoneNumber)
	if err != nil {
		t.Errorf("failed to create a user")
		return
	}

	customToken, err := firebasetools.CreateFirebaseCustomToken(ctx, user.UID)
	if err != nil {
		t.Errorf("failed to create a custom auth token for the user")
		return
	}

	_, err = firebasetools.AuthenticateCustomFirebaseToken(customToken)
	if err != nil {
		t.Errorf("failed to fetch an ID token")
		return
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *profileutils.AuthCredentialResponse
		wantErr bool
	}{
		{
			name: "Happy Case - Successfully generate auth credentials for anonymous user",
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authResponse, err := fr.GenerateAuthCredentialsForAnonymousUser(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.GenerateAuthCredentialsForAnonymousUser() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}

			if !tt.wantErr {
				if *authResponse.CustomToken == "" {
					t.Errorf("nil custom token")
					return
				}

				if *authResponse.IDToken == "" {
					t.Errorf("nil ID token")
					return
				}

				if authResponse.RefreshToken == "" {
					t.Errorf("nil refresh token")
					return
				}

				if authResponse.UID == "" {
					t.Errorf("returned a nil user")
					return
				}

				if !authResponse.IsAnonymous {
					t.Errorf("the user should be anonymous")
					return
				}
			}
		})
	}
}

func TestRepository_PersistIncomingSMSData(t *testing.T) {
	ctx := context.Background()
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	firestoreDB := fb.NewFirebaseRepository(firestoreExtension, fbc)

	validLinkId := uuid.New().String()
	text := "Test Covers"
	to := "3601"
	id := "60119"
	from := "+254705385894"
	date := "2021-05-17T13:20:04.490Z"

	validData := &dto.AfricasTalkingMessage{
		LinkID: validLinkId,
		Text:   text,
		To:     to,
		ID:     id,
		Date:   date,
		From:   from,
	}

	invalidData := &dto.AfricasTalkingMessage{
		LinkID: " ",
		Text:   text,
		To:     to,
		ID:     id,
		Date:   date,
		From:   " ",
	}

	type args struct {
		ctx   context.Context
		input dto.AfricasTalkingMessage
	}
	tests := []struct {
		name    string
		args    args
		want    *dto.AfricasTalkingMessage
		wantErr bool
	}{
		{
			name: "Happy :) Successfully persist sms data",
			args: args{
				ctx:   ctx,
				input: *validData,
			},
			wantErr: false,
		},
		{
			name: "Sad :) Unsuccessfully persist sms data",
			args: args{
				ctx:   ctx,
				input: *invalidData,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := firestoreDB.PersistIncomingSMSData(tt.args.ctx, &tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.PersistIncomingSMSData() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("error was not expected but got error: %v", err)
				return
			}
		})
	}
}

func TestRepository_AddAITSessionDetails(t *testing.T) {
	ctx := context.Background()
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	firestoreDB := fb.NewFirebaseRepository(firestoreExtension, fbc)

	phoneNumber := "+254700100200"
	SessionID := "151515"
	Level := 0
	Text := ""

	sessionDet := &dto.SessionDetails{
		SessionID:   SessionID,
		PhoneNumber: &phoneNumber,
		Level:       Level,
		Text:        Text,
	}

	invalidsessionDet := &dto.SessionDetails{
		SessionID:   "",
		PhoneNumber: &phoneNumber,
		Level:       Level,
	}

	type args struct {
		ctx   context.Context
		input *dto.SessionDetails
	}
	tests := []struct {
		name    string
		args    args
		want    *domain.USSDLeadDetails
		wantErr bool
	}{
		{
			name: "Happy case",
			args: args{
				ctx:   ctx,
				input: sessionDet,
			},
			wantErr: false,
		},

		{
			name: "Sad case",
			args: args{
				ctx:   ctx,
				input: invalidsessionDet,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "Happy case" {
				_, err := utils.ValidateUSSDDetails(sessionDet)
				if err != nil {
					t.Errorf("an error occurred")
				}
			}

			if tt.name == "Sad case" {
				_, err := utils.ValidateUSSDDetails(sessionDet)
				if err != nil {
					t.Errorf("an error occurred")
					return
				}
			}

			got, err := firestoreDB.AddAITSessionDetails(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.AddAITSessionDetails() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if tt.wantErr && got != nil {
				t.Errorf("the error was not expected")
				return
			}
		})
	}
}

func TestRepository_GetAITSessionDetailss(t *testing.T) {
	ctx := context.Background()
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	firestoreDB := fb.NewFirebaseRepository(firestoreExtension, fbc)

	sessionID := "151515"

	type args struct {
		ctx       context.Context
		sessionID string
	}
	tests := []struct {
		name    string
		args    args
		want    *domain.USSDLeadDetails
		wantErr bool
	}{
		{
			name: "Happy case",
			args: args{
				ctx:       ctx,
				sessionID: sessionID,
			},
			wantErr: false,
		},
		{
			name: "Sad case",
			args: args{
				ctx:       ctx,
				sessionID: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := firestoreDB.GetAITSessionDetails(tt.args.ctx, tt.args.sessionID)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.GetAITSessionDetails() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("error was not expected but got error: %v", err)
				return
			}
		})
	}
}

func TestRepository_UpdatePIN_IntegrationTest(t *testing.T) {
	ctx := context.Background()
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	firestoreDB := fb.NewFirebaseRepository(firestoreExtension, fbc)

	phoneNumber := "+254700000000"
	pin := "1234"

	user, err := firestoreDB.GetOrCreatePhoneNumberUser(ctx, phoneNumber)
	if err != nil {
		t.Errorf("unable to create phone number user")
		return
	}
	profile, err := firestoreDB.CreateUserProfile(
		ctx,
		phoneNumber,
		user.UID,
	)
	if err != nil {
		t.Errorf("unable to create phone number user")
		return
	}

	// Encrypt the PIN
	salt, encryptedPin := extension.NewPINExtensionImpl().EncryptPIN(pin, nil)

	pinPayload := &domain.PIN{
		ID:        uuid.New().String(),
		ProfileID: profile.ID,
		PINNumber: encryptedPin,
		Salt:      salt,
		IsOTP:     true,
	}

	_, err = firestoreDB.SavePIN(ctx, pinPayload)
	if err != nil {
		t.Errorf("unable to create phone number user")
		return
	}

	type args struct {
		ctx context.Context
		id  string
		pin *domain.PIN
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Happy case",
			args: args{
				ctx: ctx,
				id:  profile.ID,
				pin: pinPayload,
			},
			want:    true,
			wantErr: false,
		},

		{
			name: "Sad case",
			args: args{
				ctx: ctx,
				id:  profile.ID,
				pin: nil,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := firestoreDB.UpdatePIN(tt.args.ctx, tt.args.id, tt.args.pin)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.UpdatePIN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Repository.UpdatePIN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_UpdateSessionLevel(t *testing.T) {
	ctx := context.Background()
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	firestoreDB := fb.NewFirebaseRepository(firestoreExtension, fbc)

	phoneNumber := "+254702215783"

	sessionDet := &dto.SessionDetails{
		SessionID:   "b9839ed4-ad97-4cff-8b36-7afb0c7bf3ae",
		PhoneNumber: &phoneNumber,
		Level:       1,
		Text:        "Test",
	}

	sessionDetails, err := firestoreDB.AddAITSessionDetails(ctx, sessionDet)
	if err != nil {
		t.Errorf("unable to add data")
	}

	type args struct {
		ctx       context.Context
		sessionID string
		level     int
	}
	tests := []struct {
		name    string
		args    args
		want    *domain.USSDLeadDetails
		wantErr bool
	}{
		{
			name: "Happy case",
			args: args{
				ctx:       ctx,
				sessionID: sessionDetails.SessionID,
				level:     1,
			},
			wantErr: false,
		},
		{
			name: "Sad case",
			args: args{
				ctx:       ctx,
				sessionID: "",
				level:     1,
			},
			want:    &domain.USSDLeadDetails{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := firestoreDB.UpdateSessionLevel(
				tt.args.ctx,
				tt.args.sessionID,
				tt.args.level,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.UpdateSessionLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("Repository.UpdateSessionLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRepository_SaveUSSDEvent_IntegrationTest(t *testing.T) {
	ctx := context.Background()
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	firestoreDB := fb.NewFirebaseRepository(firestoreExtension, fbc)

	currentTime := time.Now()

	type args struct {
		ctx   context.Context
		input *dto.USSDEvent
	}
	tests := []struct {
		name    string
		args    args
		want    *dto.USSDEvent
		wantErr bool
	}{
		{
			name: "Happy case",
			args: args{
				ctx: ctx,
				input: &dto.USSDEvent{
					SessionID:         "0001000",
					PhoneNumber:       "+254700000000",
					USSDEventDateTime: &currentTime,
					Level:             10,
					USSDEventName:     "chose to reset PIN",
				},
			},
			wantErr: false,
		},

		{
			name: "Sad case",
			args: args{
				ctx: ctx,
				input: &dto.USSDEvent{
					SessionID:         "",
					PhoneNumber:       "+254700000000",
					USSDEventDateTime: &currentTime,
					Level:             10,
					USSDEventName:     "chose to reset PIN",
				},
			},
			want:    &dto.USSDEvent{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := firestoreDB.SaveUSSDEvent(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.SaveUSSDEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Errorf("Repository.SaveUSSDEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRepository_SaveCoverAutolinkingEvents_Integration_Test(t *testing.T) {
	ctx := context.Background()
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	firestoreDB := fb.NewFirebaseRepository(firestoreExtension, fbc)

	currentTime := time.Now()

	type args struct {
		ctx   context.Context
		input *dto.CoverLinkingEvent
	}
	tests := []struct {
		name    string
		args    args
		want    *dto.CoverLinkingEvent
		wantErr bool
	}{
		{
			name: "Happy case",
			args: args{
				ctx: ctx,
				input: &dto.CoverLinkingEvent{
					ID:                    uuid.NewString(),
					CoverLinkingEventTime: &currentTime,
					CoverStatus:           "started autolinking",
					MemberNumber:          "877386",
					PhoneNumber:           "+254703754685",
				},
			},
			wantErr: false,
		},

		{
			name: "Sad case",
			args: args{
				ctx: ctx,
				input: &dto.CoverLinkingEvent{
					ID:                    uuid.NewString(),
					CoverLinkingEventTime: &currentTime,
					CoverStatus:           "cover autolinking started",
					MemberNumber:          "",
					PhoneNumber:           "+254703754685",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := firestoreDB.SaveCoverAutolinkingEvents(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.SaveCoverAutolinkingEvents() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf(
					"Repository.SaveCoverAutolinkingEvents() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
		})
	}
}

func TestRepository_GetAITDetails_Integration(t *testing.T) {
	ctx := context.Background()
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	firestoreDB := fb.NewFirebaseRepository(firestoreExtension, fbc)

	phoneNumber := "+254700100200"

	sessionDet := &dto.SessionDetails{
		SessionID:   uuid.NewString(),
		PhoneNumber: &phoneNumber,
		Level:       0,
		Text:        "",
	}

	_, err := firestoreDB.AddAITSessionDetails(ctx, sessionDet)
	if err != nil {
		t.Errorf("unable to add session details")
		return
	}

	type args struct {
		ctx         context.Context
		phoneNumber string
	}
	tests := []struct {
		name    string
		args    args
		want    *domain.USSDLeadDetails
		wantErr bool
	}{
		{
			name: "Happy case",
			args: args{
				ctx:         ctx,
				phoneNumber: phoneNumber,
			},
			wantErr: false,
		},
		{
			name: "Sad case",
			args: args{
				ctx:         ctx,
				phoneNumber: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := firestoreDB.GetAITDetails(tt.args.ctx, tt.args.phoneNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.GetAITDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("Repository.GetAITDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRepository_UpdateAITSessionDetails_Integration(t *testing.T) {
	ctx := context.Background()
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	firestoreDB := fb.NewFirebaseRepository(firestoreExtension, fbc)

	phoneNumber := "+254700100200"

	contact := &domain.USSDLeadDetails{
		ID:             uuid.NewString(),
		Level:          0,
		PhoneNumber:    phoneNumber,
		SessionID:      uuid.NewString(),
		FirstName:      gofakeit.FirstName(),
		LastName:       gofakeit.LastName(),
		DateOfBirth:    scalarutils.Date{},
		IsRegistered:   false,
		ContactChannel: "USSD",
		WantCover:      false,
		PIN:            "1237",
	}

	sessionDet := &dto.SessionDetails{
		SessionID:   uuid.NewString(),
		PhoneNumber: &phoneNumber,
		Level:       0,
		Text:        "",
	}

	_, err := firestoreDB.AddAITSessionDetails(ctx, sessionDet)
	if err != nil {
		t.Errorf("unable to add session details")
		return
	}

	type args struct {
		ctx         context.Context
		phoneNumber string
		contactLead *domain.USSDLeadDetails
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy case",
			args: args{
				ctx:         ctx,
				phoneNumber: phoneNumber,
				contactLead: contact,
			},
			wantErr: false,
		},
		{
			name: "Sad case",
			args: args{
				ctx:         ctx,
				phoneNumber: "",
				contactLead: contact,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := firestoreDB.UpdateAITSessionDetails(tt.args.ctx, tt.args.phoneNumber, tt.args.contactLead); (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.UpdateAITSessionDetails() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}
