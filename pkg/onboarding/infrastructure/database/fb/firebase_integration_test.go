package fb_test

import (
	"context"
	"log"
	"testing"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure"

	"fmt"

	"os"
	"reflect"

	"time"

	"github.com/google/uuid"
	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/database/fb"
	"github.com/savannahghi/profileutils"
	"github.com/savannahghi/scalarutils"
	"github.com/stretchr/testify/assert"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"

	"github.com/savannahghi/onboarding/pkg/onboarding/presentation/interactor"
)

var (
	firestoreClient *firestore.Client
	firebaseAuth    *auth.Client
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
		log.Printf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		log.Printf("failed to initialize test FireBase client")
		return
	}
	firestoreClient = fsc
	firebaseAuth = fbc

	purgeRecords := func() {
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

	// try clean up first
	purgeRecords()

	// do clean up
	log.Printf("Running tests ...")
	code := m.Run()

	log.Printf("Tearing tests down ...")
	purgeRecords()

	// restore environment variables to original values
	os.Setenv(envOriginalValue, "ENVIRONMENT")
	os.Setenv("DEBUG", debugEnvValue)
	os.Setenv("ROOT_COLLECTION_SUFFIX", collectionEnvValue)

	os.Exit(code)
}

func InitializeTestService(ctx context.Context) (interactor.Usecases, error) {
	infrastructure := infrastructure.NewInfrastructureInteractor()

	ext := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})

	pinExt := extension.NewPINExtensionImpl()

	usecases := interactor.NewUsecasesInteractor(
		infrastructure, ext, pinExt,
	)
	return usecases, nil
}

func generateTestOTP(t *testing.T, phone string) (*profileutils.OtpResponse, error) {
	ctx := context.Background()
	s := infrastructure.NewInfrastructureInteractor()
	testAppID := uuid.New().String()
	return s.Engagement.GenerateAndSendOTP(ctx, phone, &testAppID)
}

// CreateTestUserByPhone creates a user that is to be used in
// running of our test cases.
// If the test user already exists then they are logged in
// to get their auth credentials
func CreateOrLoginTestUserByPhone(t *testing.T) (*auth.Token, error) {
	ctx := context.Background()
	s, err := InitializeTestService(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize test service")
	}
	phone := interserviceclient.TestUserPhoneNumber
	flavour := feedlib.FlavourConsumer
	pin := interserviceclient.TestUserPin
	exists, err := s.CheckPhoneExists(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to check if test phone exists: %v", err)
	}
	if !exists {
		otp, err := generateTestOTP(t, phone)
		log.Println("The otp is:", otp.OTP)
		if err != nil {
			return nil, fmt.Errorf("failed to generate test OTP: %v", err)
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
		authCred := &auth.Token{
			UID: u.Auth.UID,
		} // We add the test user UID to the expected auth.Token
		return authCred, nil
	}
	logInCreds, err := s.LoginByPhone(
		ctx,
		phone,
		interserviceclient.TestUserPin,
		flavour,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to log in test user: %v", err)
	}
	authCred := &auth.Token{
		UID: logInCreds.Auth.UID,
	}
	return authCred, nil

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

func TestPurgeUserByPhoneNumber(t *testing.T) {
	s, err := InitializeTestService(context.Background())
	assert.Nil(t, err)
	// clean up
	_ = s.RemoveUserByPhoneNumber(
		context.Background(),
		interserviceclient.TestUserPhoneNumber,
	)
	ctx, auth, err := GetTestAuthenticatedContext(t)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}

	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)
	profile, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	assert.Nil(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, interserviceclient.TestUserPhoneNumber, *profile.PrimaryPhone)

	// fetch the same profile but now using the primary phone number
	profile, err = fr.GetUserProfileByPrimaryPhoneNumber(
		ctx,
		interserviceclient.TestUserPhoneNumber,
		false,
	)
	assert.Nil(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, interserviceclient.TestUserPhoneNumber, *profile.PrimaryPhone)

	// purge the record. this should not fail
	err = fr.PurgeUserByPhoneNumber(ctx, interserviceclient.TestUserPhoneNumber)
	assert.Nil(t, err)

	// try purging the record again. this should fail since not user profile will be found with the phone number
	err = fr.PurgeUserByPhoneNumber(ctx, interserviceclient.TestUserPhoneNumber)
	assert.NotNil(t, err)

	// create an invalid user profile
	fakeUID := uuid.New().String()
	invalidpr1, err := fr.CreateUserProfile(
		context.Background(),
		interserviceclient.TestUserPhoneNumber,
		fakeUID,
	)
	assert.Nil(t, err)
	assert.NotNil(t, invalidpr1)

	// fetch the pins related to invalidpr1. this should fail since no pin has been associated with invalidpr1
	pin, err := fr.GetPINByProfileID(ctx, invalidpr1.ID)
	assert.NotNil(t, err)
	assert.Nil(t, pin)

	// now set a  pin. this should not fail
	userpin := "1234"
	pset, err := s.SetUserPIN(ctx, userpin, invalidpr1.ID)
	assert.Nil(t, err)
	assert.NotNil(t, pset)
	assert.Equal(t, true, pset)

	// retrieve the pin and assert it matches the one set
	pin, err = fr.GetPINByProfileID(ctx, invalidpr1.ID)
	assert.Nil(t, err)
	assert.NotNil(t, pin)
	var pinExt extension.PINExtensionImpl
	matched := pinExt.ComparePIN(userpin, pin.Salt, pin.PINNumber, nil)
	assert.Equal(t, true, matched)

	// now remove. What must be removed is the pins
	err = fr.PurgeUserByPhoneNumber(ctx, interserviceclient.TestUserPhoneNumber)
	assert.Nil(t, err)

	// assert the pin has been removed
	pin, err = fr.GetPINByProfileID(ctx, invalidpr1.ID)
	assert.NotNil(t, err)
	assert.Nil(t, pin)

}

func TestRepository_ExchangeRefreshTokenForIDToken(t *testing.T) {
	ctx, token, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	userProfile, err := fr.GetUserProfileByUID(ctx, token.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	user, err := fr.GenerateAuthCredentials(
		ctx,
		interserviceclient.TestUserPhoneNumber,
		userProfile,
	)
	if err != nil {
		t.Errorf("failed to generate auth credentials: %v", err)
		return
	}

	type args struct {
		ctx          context.Context
		refreshToken string
	}
	tests := []struct {
		name    string
		args    args
		want    *auth.Token
		wantErr bool
	}{
		{
			name: "valid firebase refresh token",
			args: args{
				ctx:          ctx,
				refreshToken: user.RefreshToken,
			},
			want:    token,
			wantErr: false,
		},
		{
			name: "invalid firebase refresh token",
			args: args{
				ctx:          ctx,
				refreshToken: "",
			},
			wantErr: true,
		},
		{
			name: "invalid firebase refresh token",
			args: args{
				ctx:          ctx,
				refreshToken: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fr.ExchangeRefreshTokenForIDToken(tt.args.ctx, tt.args.refreshToken)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.ExchangeRefreshTokenForIDToken() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}

			if !tt.wantErr {
				// obtain auth token details from the id token string
				auth, err := firebasetools.ValidateBearerToken(ctx, *got.IDToken)
				if err != nil {
					t.Errorf("invalid token: %w", err)
					return
				}
				if auth.UID != tt.want.UID {
					t.Errorf(
						"Repository.ExchangeRefreshTokenForIDToken() = %v, want %v",
						got.UID,
						tt.want.UID,
					)
				}
			}
		})
	}
}

func TestRepository_GetUserProfileByPhoneNumber(t *testing.T) {
	ctx, _, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	type args struct {
		ctx         context.Context
		phoneNumber string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Case - Get a user by valid phonenumber",
			args: args{
				ctx:         ctx,
				phoneNumber: interserviceclient.TestUserPhoneNumber,
			},
			wantErr: false,
		},
		{
			name: "Sad Case - Get a user by an invalid phonenumber",
			args: args{
				ctx:         ctx,
				phoneNumber: "+254",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fr.GetUserProfileByPhoneNumber(tt.args.ctx, tt.args.phoneNumber, false)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.GetUserProfileByPhoneNumber() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("returned a nil user")
				return
			}
		})
	}
}

func TestRepository_GetUserProfileByPrimaryPhoneNumber(t *testing.T) {
	ctx, _, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	type args struct {
		ctx         context.Context
		phoneNumber string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid : primary phone number in context",
			args: args{
				ctx:         ctx,
				phoneNumber: interserviceclient.TestUserPhoneNumber,
			},
			wantErr: false,
		},
		{
			name: "invalid : non-existent wrong phone number format",
			args: args{
				ctx:         ctx,
				phoneNumber: "+254712qwe234",
			},
			wantErr: true,
		},
		{
			name: "invalid : non existent phone number",
			args: args{
				ctx:         ctx,
				phoneNumber: "+254712098765",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fr.GetUserProfileByPrimaryPhoneNumber(
				tt.args.ctx,
				tt.args.phoneNumber,
				false,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.GetUserProfileByPrimaryPhoneNumber() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("returned a nil user")
				return
			}
		})
	}
}

func TestRepository_GetUserProfileByUID(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	type args struct {
		ctx context.Context
		uid string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Case - Get user profile by a valid UID",
			args: args{
				ctx: ctx,
				uid: auth.UID,
			},
			wantErr: false,
		},
		{
			name: "Sad Case - Get user profile by a non-existent UID",
			args: args{
				ctx: context.Background(),
				uid: "random",
			},
			wantErr: true,
		},
		{
			name: "Sad Case: Get user profile using an empty UID",
			args: args{
				ctx: ctx,
				uid: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fr.GetUserProfileByUID(tt.args.ctx, tt.args.uid, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.GetUserProfileByUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Errorf("returned a nil user")
				return
			}
		})
	}
}

func TestRepository_GetUserProfileByID(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	user, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		want    *profileutils.UserProfile
		wantErr bool
	}{
		{
			name: "Happy Case - Get user profile using a valid ID",
			args: args{
				ctx: ctx,
				id:  user.ID,
			},
			want:    user,
			wantErr: false,
		},
		{
			name: "Sad Case - Get user profile using an invalid ID",
			args: args{
				ctx: ctx,
				id:  "invalid",
			},
			wantErr: true,
		},
		{
			name: "Sad Case - Get user profile using an empty ID",
			args: args{
				ctx: ctx,
				id:  "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fr.GetUserProfileByID(tt.args.ctx, tt.args.id, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.GetUserProfileByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Repository.GetUserProfileByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_CheckIfPhoneNumberExists(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	user, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	type args struct {
		ctx         context.Context
		phoneNumber string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Happy Case - Check for a valid number that does not exist",
			args: args{
				ctx:         ctx,
				phoneNumber: "+254721524371",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Happy Case - Check for a number that exists",
			args: args{
				ctx:         ctx,
				phoneNumber: *user.PrimaryPhone,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fr.CheckIfPhoneNumberExists(tt.args.ctx, tt.args.phoneNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.CheckIfPhoneNumberExists() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if got != tt.want {
				t.Errorf("Repository.CheckIfPhoneNumberExists() = %v, want %v", got, tt.want)
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
		})
	}
}

func TestRepository_CheckIfUsernameExists(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	user, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	type args struct {
		ctx      context.Context
		userName string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Happy Case - Check for a nonexistent username",
			args: args{
				ctx:      ctx,
				userName: "Jatelo Jakom",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Happy Case - Check for an existing username",
			args: args{
				ctx:      ctx,
				userName: *user.UserName,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fr.CheckIfUsernameExists(tt.args.ctx, tt.args.userName)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.CheckIfUsernameExists() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if got != tt.want {
				t.Errorf("Repository.CheckIfUsernameExists() = %v, want %v", got, tt.want)
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
		})
	}
}

func TestRepository_GetPINByProfileID(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	user, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	pin, err := fr.GetPINByProfileID(ctx, user.ID)
	if err != nil {
		t.Errorf("failed to get pin")
		return
	}

	type args struct {
		ctx       context.Context
		profileID string
	}
	tests := []struct {
		name    string
		args    args
		want    *domain.PIN
		wantErr bool
	}{
		{
			name: "Happy Case - Get pin using a valid profileID",
			args: args{
				ctx:       ctx,
				profileID: pin.ProfileID,
			},
			want:    pin,
			wantErr: false,
		},
		{
			name: "Sad Case - Get pin using an invalid profileID",
			args: args{
				ctx:       ctx,
				profileID: "invalidID",
			},
			wantErr: true,
		},
		{
			name: "Sad Case - Get pin using an empty profileID",
			args: args{
				ctx:       ctx,
				profileID: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fr.GetPINByProfileID(tt.args.ctx, tt.args.profileID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.GetPINByProfileID() error = %v, wantErr %v", err, tt.wantErr)
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
				t.Errorf("Repository.GetPINByProfileID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_SavePIN(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	user, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	validPin := interserviceclient.TestUserPin

	var pin extension.PINExtensionImpl
	salt, encryptedPin := pin.EncryptPIN(validPin, nil)

	validSavePinPayload := &domain.PIN{
		ID:        uuid.New().String(),
		ProfileID: user.ID,
		PINNumber: encryptedPin,
		Salt:      salt,
	}

	type args struct {
		ctx context.Context
		pin *domain.PIN
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "happy case: save pin with valid pin payload",
			args: args{
				ctx: ctx,
				pin: validSavePinPayload,
			},
			want:    true,
			wantErr: false,
		},

		{
			name: "sad case: save pin with pin no payload",
			args: args{
				ctx: ctx,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fr.SavePIN(tt.args.ctx, tt.args.pin)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.SavePIN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Repository.SavePIN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_UpdatePIN(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	user, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	validPin := interserviceclient.TestUserPin

	var pin extension.PINExtensionImpl
	salt, encryptedPin := pin.EncryptPIN(validPin, nil)

	validSavePinPayload := &domain.PIN{
		ID:        uuid.New().String(),
		ProfileID: user.ID,
		PINNumber: encryptedPin,
		Salt:      salt,
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
			name: "happy case: update pin with valid pin payload",
			args: args{
				ctx: ctx,
				id:  user.ID,
				pin: validSavePinPayload,
			},
			want:    true,
			wantErr: false,
		},

		{
			name: "sad case: update pin with invalid payload",
			args: args{
				ctx: ctx,
				id:  "", // empty user profile
				pin: validSavePinPayload,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fr.UpdatePIN(tt.args.ctx, tt.args.id, tt.args.pin)
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

func TestRepository_RecordPostVisitSurvey(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	type args struct {
		ctx   context.Context
		input dto.PostVisitSurveyInput
		UID   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Case - Successfully record a post visit survey",
			args: args{
				ctx: ctx,
				input: dto.PostVisitSurveyInput{
					LikelyToRecommend: 10,
					Criticism:         "Nothing at all. Good job.",
					Suggestions:       "Can't think of anything.",
				},
				UID: auth.UID,
			},
			wantErr: false,
		},
		{
			name: "Sad Case - Invalid input",
			args: args{
				ctx: ctx,
				input: dto.PostVisitSurveyInput{
					LikelyToRecommend: 100,
					Criticism:         "Nothing at all. Good job.",
					Suggestions:       "Can't think of anything.",
				},
				UID: auth.UID,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fr.RecordPostVisitSurvey(tt.args.ctx, tt.args.input, tt.args.UID); (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.RecordPostVisitSurvey() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestRepository_UpdateSuspended(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	user, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	type args struct {
		ctx    context.Context
		id     string
		status bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Case - Successfully update the suspend status",
			args: args{
				ctx:    ctx,
				id:     user.ID,
				status: true,
			},
			wantErr: false,
		},
		{
			name: "Happy Case - Successfully update the suspend status",
			args: args{
				ctx:    ctx,
				id:     user.ID,
				status: false,
			},
			wantErr: false,
		},
		{
			name: "Sad Case - Use an invalid id",
			args: args{
				ctx:    ctx,
				id:     "invalid id",
				status: true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fr.UpdateSuspended(tt.args.ctx, tt.args.id, tt.args.status); (err != nil) != tt.wantErr {
				t.Errorf("Repository.UpdateSuspended() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepository_UpdateVerifiedUIDS(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	user, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	type args struct {
		ctx  context.Context
		id   string
		uids []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Case - Successfully update profile UIDs",
			args: args{
				ctx: ctx,
				id:  user.ID,
				uids: []string{
					"f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					"5d46d3bd-a482-4787-9b87-3c94510c8b53",
				},
			},
			wantErr: false,
		},
		{
			name: "Sad Case - Invalid ID",
			args: args{
				ctx: ctx,
				id:  "invalidid",
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
			if err := fr.UpdateVerifiedUIDS(tt.args.ctx, tt.args.id, tt.args.uids); (err != nil) != tt.wantErr {
				t.Errorf("Repository.UpdateVerifiedUIDS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepository_UpdateVerifiedIdentifiers(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	userProfile, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	presentIdentifiers := []profileutils.VerifiedIdentifier{
		{
			UID:           "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
			LoginProvider: "Facebook",
		},
	}

	type args struct {
		ctx         context.Context
		id          string
		identifiers []profileutils.VerifiedIdentifier
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Case - Successfully update the user's verified identifiers",
			args: args{
				ctx: ctx,
				id:  userProfile.ID,
				identifiers: []profileutils.VerifiedIdentifier{
					{
						UID:           auth.UID,
						LoginProvider: "Facebook",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Happy Case - Use a different UID",
			args: args{
				ctx:         ctx,
				id:          userProfile.ID,
				identifiers: presentIdentifiers,
			},
			wantErr: false,
		},
		{
			name: "Happy Case - Adding a new identifier",
			args: args{
				ctx: ctx,
				id:  userProfile.ID,
				identifiers: []profileutils.VerifiedIdentifier{
					{
						UID:           "5d46d3bd-a482-4787-9b87-3c94510c8b53",
						LoginProvider: "Google",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Sad Case - Use an invalid id",
			args: args{
				ctx: ctx,
				id:  "invalidid",
				identifiers: []profileutils.VerifiedIdentifier{
					{
						UID:           auth.UID,
						LoginProvider: "Facebook",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fr.UpdateVerifiedIdentifiers(tt.args.ctx, tt.args.id, tt.args.identifiers); (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.UpdateVerifiedIdentifiers() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestRepository_UpdateSecondaryEmailAddresses(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	userProfile, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	type args struct {
		ctx            context.Context
		id             string
		emailAddresses []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Case - Update Profile Secondary Email",
			args: args{
				ctx:            ctx,
				id:             userProfile.ID,
				emailAddresses: []string{"jatelo@gmail.com", "nyaras@gmail.com"},
			},
			wantErr: false,
		},
		{
			name: "Sad Case - Update Profile Secondary Email using an invalid ID",
			args: args{
				ctx:            ctx,
				id:             "invalid id",
				emailAddresses: []string{"jatelo@gmail.com", "nyaras@gmail.com"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fr.UpdateSecondaryEmailAddresses(tt.args.ctx, tt.args.id, tt.args.emailAddresses)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.UpdateSecondaryEmailAddresses() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestRepository_UpdatePrimaryEmailAddress(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	userProfile, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	newPrimaryEmail := "johndoe@gmail.com"

	type args struct {
		ctx          context.Context
		id           string
		emailAddress string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Case - Update using a valid email",
			args: args{
				ctx:          ctx,
				id:           userProfile.ID,
				emailAddress: newPrimaryEmail,
			},
			wantErr: false,
		},
		{
			name: "Sad Case - Unable to get logged in user",
			args: args{
				ctx:          ctx,
				id:           "invalidid",
				emailAddress: newPrimaryEmail,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fr.UpdatePrimaryEmailAddress(tt.args.ctx, tt.args.id, tt.args.emailAddress); (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.UpdatePrimaryEmailAddress() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected, got %v", err)
					return
				}
			}
		})
	}
}

func TestRepository_UpdatePrimaryPhoneNumber(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	userProfile, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	newPrimaryPhoneNumber := "+254711111111"
	type args struct {
		ctx         context.Context
		id          string
		phoneNumber string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Case - Update using a valid email",
			args: args{
				ctx:         ctx,
				id:          userProfile.ID,
				phoneNumber: newPrimaryPhoneNumber,
			},
			wantErr: false,
		},
		{
			name: "Sad Case - Unable to get logged in user",
			args: args{
				ctx:         ctx,
				id:          "invalidid",
				phoneNumber: newPrimaryPhoneNumber,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fr.UpdatePrimaryPhoneNumber(tt.args.ctx, tt.args.id, tt.args.phoneNumber); (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.UpdatePrimaryPhoneNumber() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected, got %v", err)
					return
				}
			}
		})
	}
}

func TestRepository_UpdateSecondaryPhoneNumbers(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	userProfile, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	newSecondaryPhoneNumbers := []string{"+254744556677", "+254700998877"}

	type args struct {
		ctx          context.Context
		id           string
		phoneNumbers []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Case - Update secondary phonenumbers",
			args: args{
				ctx:          ctx,
				id:           userProfile.ID,
				phoneNumbers: newSecondaryPhoneNumbers,
			},
			wantErr: false,
		},
		{
			name: "Sad Case - Update secondary phonenumbers using an invalid ID",
			args: args{
				ctx:          ctx,
				id:           "invalid id",
				phoneNumbers: newSecondaryPhoneNumbers,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fr.UpdateSecondaryPhoneNumbers(tt.args.ctx, tt.args.id, tt.args.phoneNumbers); (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.UpdateSecondaryPhoneNumbers() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected, got %v", err)
					return
				}
			}
		})
	}
}

func TestRepository_UpdateBioData(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	userProfile, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	firstName := "Jatelo"
	lastName := "Mzima"
	dateOfBirth := scalarutils.Date{
		Year:  2000,
		Month: 12,
		Day:   17,
	}
	var gender enumutils.Gender = "male"

	updateAllData := profileutils.BioData{
		FirstName:   &firstName,
		LastName:    &lastName,
		DateOfBirth: &dateOfBirth,
		Gender:      gender,
	}

	updateFirstName := profileutils.BioData{
		FirstName: &firstName,
	}
	updateLastName := profileutils.BioData{
		LastName: &lastName,
	}
	updateDateOfBirth := profileutils.BioData{
		DateOfBirth: &dateOfBirth,
	}
	updateGender := profileutils.BioData{
		Gender: gender,
	}
	type args struct {
		ctx  context.Context
		id   string
		data profileutils.BioData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Case - Update all biodata",
			args: args{
				ctx:  ctx,
				id:   userProfile.ID,
				data: updateAllData,
			},
			wantErr: false,
		},
		{
			name: "Happy Case - Update firstname only",
			args: args{
				ctx:  ctx,
				id:   userProfile.ID,
				data: updateFirstName,
			},
			wantErr: false,
		},
		{
			name: "Happy Case - Update lastname only",
			args: args{
				ctx:  ctx,
				id:   userProfile.ID,
				data: updateLastName,
			},
			wantErr: false,
		},
		{
			name: "Happy Case - Update date of birth only",
			args: args{
				ctx:  ctx,
				id:   userProfile.ID,
				data: updateDateOfBirth,
			},
			wantErr: false,
		},
		{
			name: "Happy Case - Update gender only",
			args: args{
				ctx:  ctx,
				id:   userProfile.ID,
				data: updateGender,
			},
			wantErr: false,
		},
		{
			name: "Sad Case - Use an invalid ID",
			args: args{
				ctx:  ctx,
				id:   "invalid id",
				data: updateAllData,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fr.UpdateBioData(tt.args.ctx, tt.args.id, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Repository.UpdateBioData() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected, got %v", err)
					return
				}
			}
		})
	}
}

func TestRepositoryGenerateAuthCredentialsForAnonymousUser(t *testing.T) {
	ctx := context.Background()

	fsc, fbc := firestoreClient, firebaseAuth
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
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

func TestRepositoryGenerateAuthCredentials(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	userProfile, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	customToken, err := firebasetools.CreateFirebaseCustomToken(ctx, auth.UID)
	if err != nil {
		t.Errorf("failed to create a custom auth token for the user")
		return
	}

	userToken, err := firebasetools.AuthenticateCustomFirebaseToken(customToken)
	if err != nil {
		t.Errorf("failed to fetch an ID token")
		return
	}

	validCredentials := &profileutils.AuthCredentialResponse{
		CustomToken:  &customToken,
		IDToken:      &userToken.IDToken,
		ExpiresIn:    userToken.ExpiresIn,
		RefreshToken: userToken.RefreshToken,
		UID:          auth.UID,
		IsAnonymous:  false,
		IsAdmin:      false,
	}

	type args struct {
		ctx     context.Context
		phone   string
		profile *profileutils.UserProfile
	}
	tests := []struct {
		name    string
		args    args
		want    *profileutils.AuthCredentialResponse
		wantErr bool
	}{
		{
			name: "Happy Case - Successfully generate valid auth credentials",
			args: args{
				ctx:     ctx,
				phone:   *userProfile.PrimaryPhone,
				profile: userProfile,
			},
			wantErr: false,
		},
		{
			name: "Sad Case - Use an invalid phonenumber",
			args: args{
				ctx:     ctx,
				phone:   "invalidphone",
				profile: nil,
			},
			want:    validCredentials,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authResponse, err := fr.GenerateAuthCredentials(
				tt.args.ctx,
				tt.args.phone,
				tt.args.profile,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Repository.GenerateAuthCredentials() error = %v, wantErr %v",
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

			}
		})
	}
}

func TestRepositoryFetchAdminUsers(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	userProfile, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	permissions := profileutils.DefaultAdminPermissions

	err = fr.UpdatePermissions(ctx, userProfile.ID, permissions)
	if err != nil {
		t.Errorf("failed to update user permissions: %v", err)
		return
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    []*profileutils.UserProfile
		wantErr bool
	}{
		{
			name: "Happy Case - Fetch admin users",
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adminResponse, err := fr.FetchAdminUsers(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.FetchAdminUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(adminResponse) == 0 {
					t.Errorf("nil admin response")
					return
				}

			}
		})
	}
}

func TestUpdateAddresses(t *testing.T) {
	ctx, auth, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	userProfile, err := fr.GetUserProfileByUID(ctx, auth.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	address := profileutils.Address{
		Latitude:  "-1.2349035671",
		Longitude: "36.79329309999994",
	}
	type args struct {
		ctx         context.Context
		id          string
		address     profileutils.Address
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
				id:          userProfile.ID,
				address:     address,
				addressType: enumutils.AddressTypeHome,
			},
			wantErr: false,
		},
		{
			name: "happy:) add work address",
			args: args{
				ctx:         ctx,
				id:          userProfile.ID,
				address:     address,
				addressType: enumutils.AddressTypeWork,
			},
			wantErr: false,
		},
		{
			name: "sad:( failed to add",
			args: args{
				ctx:         ctx,
				id:          "not-a-uid",
				address:     address,
				addressType: enumutils.AddressTypeWork,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fr.UpdateAddresses(
				tt.args.ctx,
				tt.args.id,
				tt.args.address,
				tt.args.addressType,
			); (err != nil) != tt.wantErr {
				t.Errorf("Repository.UpdateAddresses() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}
func TestRepository_UpdatePIN_IntegrationTest(t *testing.T) {
	ctx := context.Background()
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
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

func TestRepository_GetUserProfileByPhoneOrEmail_Integration(t *testing.T) {
	ctx := context.Background()
	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	firestoreDB := fb.NewFirebaseRepository(firestoreExtension, fbc)

	testPhone := "+254706060060"
	invalidTestPhone := "+2547060660"
	testEmail := "test@kathurima.com"
	invalidTestEmail := "+test@"

	_, err := firestoreDB.CreateUserProfile(ctx, testPhone, uuid.NewString())
	if err != nil {
		t.Errorf("unable to create phone number user")
		return
	}

	err = firestoreDB.UpdateUserProfileEmail(ctx, testPhone, testEmail)
	if err != nil {
		t.Errorf("unable to create phone number user")
		return
	}

	type args struct {
		ctx     context.Context
		payload *dto.RetrieveUserProfileInput
	}
	tests := []struct {
		name    string
		args    args
		want    *profileutils.UserProfile
		wantErr bool
	}{
		{
			name: "Happy case:phone",
			args: args{
				ctx: ctx,
				payload: &dto.RetrieveUserProfileInput{
					PhoneNumber: &testPhone,
				},
			},
			wantErr: false,
		},
		{
			name: "Sad case:phone",
			args: args{
				ctx: ctx,
				payload: &dto.RetrieveUserProfileInput{
					PhoneNumber: &invalidTestPhone,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Happy case:email",
			args: args{
				ctx: ctx,
				payload: &dto.RetrieveUserProfileInput{
					Email: &testEmail,
				},
			},
			wantErr: false,
		},
		{
			name: "Sad case:email",
			args: args{
				ctx: ctx,
				payload: &dto.RetrieveUserProfileInput{
					Email: &invalidTestEmail,
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := firestoreDB.GetUserProfileByPhoneOrEmail(tt.args.ctx, tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.GetUserProfileByPhoneOrEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("Repository.GetUserProfileByPhoneOrEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRepository_UpdateUserRoleIDs_Integration(t *testing.T) {
	ctx, token, err := GetTestAuthenticatedContext(t)
	if err != nil {
		t.Errorf("failed to get test authenticated context: %v", err)
		return
	}

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}
	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	userProfile, err := fr.GetUserProfileByUID(ctx, token.UID, false)
	if err != nil {
		t.Errorf("failed to get a user profile")
		return
	}

	type args struct {
		ctx     context.Context
		id      string
		roleIDs []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "pass:success update user profile IDs",
			args: args{
				ctx:     ctx,
				id:      userProfile.ID,
				roleIDs: []string{uuid.NewString()},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fr.UpdateUserRoleIDs(tt.args.ctx, tt.args.id, tt.args.roleIDs); (err != nil) != tt.wantErr {
				t.Errorf("Repository.UpdateUserRoleIDs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepository_CreateRole_Integration(t *testing.T) {
	ctx := context.Background()

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}

	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}

	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	userID := uuid.NewString()

	type args struct {
		ctx       context.Context
		profileID string
		input     dto.RoleInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:create a new role",
			args: args{
				ctx:       ctx,
				profileID: userID,
				input: dto.RoleInput{
					Name:        "Created Test Role",
					Description: "Can run tests",
				},
			},
			wantErr: false,
		},
		{
			name: "fail:create using existing role name",
			args: args{
				ctx:       ctx,
				profileID: userID,
				input: dto.RoleInput{
					Name:        "Created Test Role",
					Description: "Can run tests",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fr.CreateRole(tt.args.ctx, tt.args.profileID, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.CreateRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("Repository.CreateRole() = %v", got)
				return
			}
		})
	}
}

func TestRepository_UpdateRoleDetails_Integration(t *testing.T) {
	ctx := context.Background()

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}

	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}

	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	profileID := uuid.NewString()
	input := dto.RoleInput{
		Name:        "Updated Test Role",
		Description: "Can run tests",
	}

	role, err := fr.CreateRole(ctx, profileID, input)
	if err != nil {
		t.Errorf("failed to create test role")
		return
	}

	// Update some details in roles
	role.Description = " Can still run tests"

	type args struct {
		ctx       context.Context
		profileID string
		role      profileutils.Role
	}
	tests := []struct {
		name    string
		args    args
		want    *profileutils.Role
		wantErr bool
	}{
		{
			name: "valid:success update role",
			args: args{
				ctx:       ctx,
				profileID: profileID,
				role:      *role,
			},
			want:    role,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := fr.UpdateRoleDetails(tt.args.ctx, tt.args.profileID, tt.args.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.UpdateRoleDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got.Description, tt.want.Description) {
				t.Errorf("Repository.UpdateRoleDetails().Description = %v, want %v", got.Description, tt.want.Description)
			}

			if !reflect.DeepEqual(got.UpdatedBy, profileID) {
				t.Errorf("Repository.UpdateRoleDetails().UpdatedBy = %v, want %v", got.UpdatedBy, profileID)
			}
		})
	}
}

func TestRepository_GetRoleByID_Integration(t *testing.T) {
	ctx := context.Background()

	fsc, fbc := InitializeTestFirebaseClient(ctx)
	if fsc == nil {
		t.Errorf("failed to initialize test FireStore client")
		return
	}

	if fbc == nil {
		t.Errorf("failed to initialize test FireBase client")
		return
	}

	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

	profileID := uuid.NewString()
	input := dto.RoleInput{
		Name:        "Retrieved Role",
		Description: "Can run tests",
	}

	role, err := fr.CreateRole(ctx, profileID, input)
	if err != nil {
		t.Errorf("failed to create test role")
		return
	}

	type args struct {
		ctx    context.Context
		roleID string
	}
	tests := []struct {
		name    string
		args    args
		want    profileutils.Role
		wantErr bool
	}{
		{
			name: "success: retrieve role",
			args: args{
				ctx:    ctx,
				roleID: role.ID,
			},
			want:    *role,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := fr.GetRoleByID(tt.args.ctx, tt.args.roleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.GetRoleByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.ID, tt.want.ID) {
				t.Errorf("Repository.GetRoleByID() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got.Name, tt.want.Name) {
				t.Errorf("Repository.GetRoleByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestRepository_CheckIfUserHasPermission_Integration(t *testing.T) {
// 	ctx, token, err := GetTestAuthenticatedContext(t)
// 	if err != nil {
// 		t.Errorf("failed to get test authenticated context: %v", err)
// 		return
// 	}

// 	fsc, fbc := InitializeTestFirebaseClient(ctx)
// 	if fsc == nil {
// 		t.Errorf("failed to initialize test FireStore client")
// 		return
// 	}
// 	if fbc == nil {
// 		t.Errorf("failed to initialize test FireBase client")
// 		return
// 	}

// 	firestoreExtension := fb.NewFirestoreClientExtension(fsc)
// 	fr := fb.NewFirebaseRepository(firestoreExtension, fbc)

// 	userProfile, err := fr.GetUserProfileByUID(ctx, token.UID, false)
// 	if err != nil {
// 		t.Errorf("failed to get a user profile")
// 		return
// 	}

// 	input := dto.RoleInput{
// 		Name:        "Check Permission Role",
// 		Description: "Can run tests",
// 		Scopes:      []string{profileutils.CanAssignRole.Scope},
// 	}

// 	role, err := fr.CreateRole(ctx, uuid.NewString(), input)
// 	if err != nil {
// 		t.Errorf("failed to create test role")
// 		return
// 	}

// 	err = fr.UpdateUserRoleIDs(ctx, userProfile.ID, []string{role.ID})
// 	if err != nil {
// 		t.Errorf("failed to add role to user")
// 		return
// 	}

// 	type args struct {
// 		ctx                context.Context
// 		UID                string
// 		requiredPermission profileutils.Permission
// 	}

// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    bool
// 		wantErr bool
// 	}{
// 		{
// 			name: "fail: user doesn't have required permission",
// 			args: args{
// 				ctx:                ctx,
// 				UID:                token.UID,
// 				requiredPermission: profileutils.CanEditRole,
// 			},
// 			want:    false,
// 			wantErr: false,
// 		},
// 		{
// 			name: "pass: user has required permission",
// 			args: args{
// 				ctx:                ctx,
// 				UID:                token.UID,
// 				requiredPermission: profileutils.CanAssignRole,
// 			},
// 			want:    true,
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := fr.CheckIfUserHasPermission(tt.args.ctx, tt.args.UID, tt.args.requiredPermission)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Repository.CheckIfUserHasPermission() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got != tt.want {
// 				t.Errorf("Repository.CheckIfUserHasPermission() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
