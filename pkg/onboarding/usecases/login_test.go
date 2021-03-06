package usecases_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"github.com/google/uuid"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/exceptions"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/database"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/database/fb"
	"github.com/savannahghi/onboarding/pkg/onboarding/presentation/interactor"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases"
	"github.com/savannahghi/profileutils"
	"github.com/savannahghi/serverutils"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"

	extMock "github.com/savannahghi/onboarding/pkg/onboarding/application/extension/mock"
	mockInfra "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/mock"

	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/engagement"
	engagementMock "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/engagement/mock"

	pubsubmessaging "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/pubsub"
	pubsubmessagingMock "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/pubsub/mock"
)

var testUsecase interactor.Usecases
var testInfrastructure infrastructure.Infrastructure

func TestMain(m *testing.M) {
	log.Printf("Setting tests up ...")
	envOriginalValue := os.Getenv("ENVIRONMENT")
	os.Setenv("ENVIRONMENT", "staging")
	emailOriginalValue := os.Getenv("SAVANNAH_ADMIN_EMAIL")
	os.Setenv("SAVANNAH_ADMIN_EMAIL", "test@bewell.co.ke")
	debugEnvValue := os.Getenv("DEBUG")
	os.Setenv("DEBUG", "true")
	os.Setenv("REPOSITORY", "firebase")
	collectionEnvValue := os.Getenv("ROOT_COLLECTION_SUFFIX")
	// !NOTE!
	// Under no circumstances should you remove this env var when testing
	// You risk purging important collections, like our prod collections
	os.Setenv("ROOT_COLLECTION_SUFFIX", fmt.Sprintf("onboarding_ci_%v", time.Now().Unix()))

	ctx := context.Background()

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		log.Panicf("failed to initialize test FireStore client")
	}
	if fbc == nil {
		log.Panicf("failed to initialize test FireBase client")
	}

	purgeRecords := func() {
		if serverutils.MustGetEnvVar(domain.Repo) == domain.FirebaseRepository {
			r := fb.Repository{}
			collections := []string{
				r.GetPINsCollectionName(),
				r.GetUserProfileCollectionName(),
				r.GetSurveyCollectionName(),
				r.GetCommunicationsSettingsCollectionName(),
				r.GetExperimentParticipantCollectionName(),
				r.GetProfileNudgesCollectionName(),
				r.GetRolesCollectionName(),
			}
			for _, collection := range collections {
				ref := fsc.Collection(collection)
				firebasetools.DeleteCollection(ctx, fsc, ref, 10)
			}
		}

	}

	// try clean up first
	purgeRecords()

	log.Printf("Initializing tests ...")
	infrastructure := infrastructure.NewInfrastructureInteractor()

	testInfrastructure = infrastructure

	s, err := InitializeTestService(ctx, infrastructure)
	if err != nil {
		log.Panicf("failed to initialize test service in package usecases_test")
	}

	testUsecase = s

	// do clean up
	log.Printf("Running tests ...")
	code := m.Run()

	log.Printf("Tearing tests down ...")
	purgeRecords()

	// restore environment variables to original values
	os.Setenv(envOriginalValue, "ENVIRONMENT")
	os.Setenv(emailOriginalValue, "SAVANNAH_ADMIN_EMAIL")
	os.Setenv("DEBUG", debugEnvValue)
	os.Setenv("ROOT_COLLECTION_SUFFIX", collectionEnvValue)

	os.Exit(code)
}

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

func InitializeTestService(ctx context.Context, infrastructure infrastructure.Infrastructure) (interactor.Usecases, error) {
	ext := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})

	pinExt := extension.NewPINExtensionImpl()

	i := interactor.NewUsecasesInteractor(infrastructure, ext, pinExt)

	return i, nil
}

func generateTestOTP(t *testing.T, phone string) (*profileutils.OtpResponse, error) {
	ctx := context.Background()
	infra := infrastructure.NewInfrastructureInteractor()
	testAppID := uuid.New().String()
	return infra.Engagement.GenerateAndSendOTP(ctx, phone, &testAppID)
}

// CreateTestUserByPhone creates a user that is to be used in
// running of our test cases.
// If the test user already exists then they are logged in
// to get their auth credentials
func CreateOrLoginTestUserByPhone(t *testing.T) (*auth.Token, error) {
	ctx := context.Background()
	s := testUsecase
	phone := interserviceclient.TestUserPhoneNumber
	flavour := feedlib.FlavourConsumer
	pin := interserviceclient.TestUserPin
	testAppID := uuid.New().String()
	otp, err := s.VerifyPhoneNumber(ctx, phone, &testAppID)
	if err != nil {
		if strings.Contains(err.Error(), exceptions.CheckPhoneNumberExistError().Error()) {
			logInCreds, err := s.LoginByPhone(
				ctx,
				phone,
				interserviceclient.TestUserPin,
				flavour,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to log in test user: %v", err)
			}

			return &auth.Token{
				UID: logInCreds.Auth.UID,
			}, nil
		}

		return nil, fmt.Errorf("failed to check if test phone exists: %v", err)
	}

	u, err := s.CreateUserByPhone(
		ctx,
		&dto.SignUpInput{
			PhoneNumber: &phone,
			PIN:         &pin,
			Flavour:     flavour,
			OTP:         &otp.OTP,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create a test user: %v", err)
	}
	if u == nil {
		return nil, fmt.Errorf("nil test user response")
	}

	return &auth.Token{
		UID: u.Auth.UID,
	}, nil
}

// TestAuthenticatedContext returns a logged in context, useful for test purposes
func GetTestAuthenticatedContext(t *testing.T) (context.Context, *auth.Token, error) {
	ctx := context.Background()
	auth, err := CreateOrLoginTestUserByPhone(t)
	if err != nil {
		return nil, nil, err
	}
	authenticatedContext := context.WithValue(
		ctx,
		firebasetools.AuthTokenContextKey,
		auth,
	)
	return authenticatedContext, auth, nil
}

func TestGetTestAuthenticatedContext(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}
	if ctx == nil {
		t.Errorf("nil context")
		return
	}
	if auth == nil {
		t.Errorf("nil auth data")
		return
	}
}

func TestLoginUseCasesImpl_LoginByPhone(t *testing.T) {
	ctx, _, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}
	flavour := feedlib.FlavourConsumer
	s := testUsecase

	type args struct {
		ctx     context.Context
		phone   string
		PIN     string
		flavour feedlib.Flavour
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy case: valid login",
			args: args{
				ctx:     ctx,
				phone:   interserviceclient.TestUserPhoneNumber,
				PIN:     interserviceclient.TestUserPin,
				flavour: flavour,
			},
			wantErr: false,
		},
		{
			name: "sad case: wrong pin number supplied",
			args: args{
				ctx:     ctx,
				phone:   interserviceclient.TestUserPhoneNumber,
				PIN:     "4567",
				flavour: flavour,
			},
			wantErr: true,
		},
		{
			name: "sad case: user profile without a primary phone number",
			args: args{
				ctx:     ctx,
				phone:   "+2547900900", // not a primary phone number
				PIN:     interserviceclient.TestUserPin,
				flavour: flavour,
			},
			wantErr: true,
		},
		{
			name: "sad case: incorrect phone number",
			args: args{
				ctx:     ctx,
				phone:   "+2541234",
				PIN:     interserviceclient.TestUserPin,
				flavour: flavour,
			},
			wantErr: true,
		},
		// {
		// 	name: "sad case: incorrect flavour",
		// 	args: args{
		// 		ctx:     ctx,
		// 		phone:   interserviceclient.TestUserPhoneNumber,
		// 		PIN:     interserviceclient.TestUserPin,
		// 		flavour: "not-a-correct-flavour",
		// 	},
		// 	wantErr: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authResponse, err := s.LoginByPhone(
				tt.args.ctx,
				tt.args.phone,
				tt.args.PIN,
				tt.args.flavour,
			)
			if tt.wantErr && authResponse != nil {
				t.Errorf("expected nil auth response but got %v, since the error %v occurred",
					authResponse,
					err,
				)
				return
			}

			if !tt.wantErr && authResponse == nil {
				t.Errorf("expected an auth response but got nil, since no error occurred")
				return
			}
		})
	}
}

var fakeBaseExt extMock.FakeBaseExtensionImpl
var fakePinExt extMock.PINExtensionImpl
var fakeEngagementSvs engagementMock.FakeServiceEngagement
var fakePubSub pubsubmessagingMock.FakeServicePubSub

var fakeInfraRepo mockInfra.FakeInfrastructure

// InitializeFakeOnboardingInteractor represents a fakeonboarding interactor
func InitializeFakeOnboardingInteractor() (usecases.Interactor, error) {
	var r database.Repository = &fakeInfraRepo
	var engagementSvc engagement.ServiceEngagement = &fakeEngagementSvs
	var ext extension.BaseExtension = &fakeBaseExt
	var pinExt extension.PINExtension = &fakePinExt
	var ps pubsubmessaging.ServicePubSub = &fakePubSub

	infra := func() infrastructure.Infrastructure {
		return infrastructure.Infrastructure{
			Database:   r,
			Engagement: engagementSvc,
			Pubsub:     ps,
		}
	}()

	i := usecases.NewUsecasesInteractor(infra, ext, pinExt)

	return i, nil

}
