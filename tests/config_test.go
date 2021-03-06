package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"github.com/google/uuid"
	"github.com/imroc/req"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/database/fb"
	"github.com/savannahghi/onboarding/pkg/onboarding/presentation/interactor"
	"github.com/savannahghi/profileutils"
	"github.com/savannahghi/serverutils"
	"github.com/sirupsen/logrus"

	"github.com/savannahghi/onboarding/pkg/onboarding/presentation"
)

const (
	testHTTPClientTimeout = 180
)

const (
	engagementService = "engagement"
	ediService        = "edi"
	testRoleName      = "Test Role"
	testPIN           = "2030"
)

/// these are set up once in TestMain and used by all the acceptance tests in
// this package
var (
	srv            *http.Server
	baseURL        string
	serverErr      error
	testInteractor interactor.Usecases
	testInfra      infrastructure.Infrastructure
)

func mapToJSONReader(m map[string]interface{}) (io.Reader, error) {
	bs, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal map to JSON: %w", err)
	}

	buf := bytes.NewBuffer(bs)
	return buf, nil
}

func initializeAcceptanceTestFirebaseClient(ctx context.Context) (*firestore.Client, *auth.Client) {
	fc := firebasetools.FirebaseClient{}
	fa, err := fc.InitFirebase()
	if err != nil {
		log.Panicf("unable to initialize Firestore for the Feed: %s", err)
	}

	fsc, err := fa.Firestore(ctx)
	if err != nil {
		log.Panicf("unable to initialize Firestore: %s", err)
	}

	fbc, err := fa.Auth(ctx)
	if err != nil {
		log.Panicf("can't initialize Firebase auth when setting up profile service: %s", err)
	}
	return fsc, fbc
}

func InitializeTestService(ctx context.Context, infra infrastructure.Infrastructure) (interactor.Usecases, error) {
	ext := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})

	pinExt := extension.NewPINExtensionImpl()

	usecases := interactor.NewUsecasesInteractor(
		infra, ext, pinExt,
	)
	return usecases, nil
}

func InitializeTestInfrastructure(ctx context.Context) (infrastructure.Infrastructure, error) {

	return infrastructure.NewInfrastructureInteractor(), nil
}

func composeInValidUserPayload(t *testing.T) *dto.SignUpInput {
	phone := interserviceclient.TestUserPhoneNumber
	pin := "" // empty string
	flavour := feedlib.FlavourPro
	payload := &dto.SignUpInput{
		PhoneNumber: &phone,
		PIN:         &pin,
		Flavour:     flavour,
	}
	return payload
}

func composeValidUserPayload(t *testing.T, phone string) (*dto.SignUpInput, error) {
	pin := testPIN
	flavour := feedlib.FlavourPro
	otp, err := generateTestOTP(t, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to generate test OTP: %v", err)
	}
	return &dto.SignUpInput{
		PhoneNumber: &phone,
		PIN:         &pin,
		Flavour:     flavour,
		OTP:         &otp.OTP,
	}, nil
}

func composeValidRolePayload(t *testing.T, roleName string) *dto.RoleInput {
	ctx := context.Background()

	var allScopes []string

	perms, _ := profileutils.AllPermissions(ctx)
	for _, perm := range perms {
		allScopes = append(allScopes, perm.Scope)
	}

	role := dto.RoleInput{
		Name:        roleName,
		Description: "Role for running tests",
		Scopes:      allScopes,
	}

	return &role
}

func CreateTestUserByPhone(t *testing.T, phone string) (*profileutils.UserResponse, error) {
	validPayload, err := composeValidUserPayload(t, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to compose a valid payload: %v", err)
	}

	bs, err := json.Marshal(validPayload)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal test item to JSON: %s", err)
	}
	payload := bytes.NewBuffer(bs)

	url := fmt.Sprintf("%s/create_user_by_phone", baseURL)

	r, err := http.NewRequest(
		http.MethodPost,
		url,
		payload,
	)
	if err != nil {
		return nil, fmt.Errorf("can't create new request: %v", err)

	}
	if r == nil {
		return nil, fmt.Errorf("nil request")
	}

	r.Header.Add("Accept", "application/json")
	r.Header.Add("Content-Type", "application/json")

	client := http.DefaultClient

	resp, err := client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("HTTP error: %v", err)

	}
	// if resp.StatusCode != http.StatusCreated {
	// 	return nil, fmt.Errorf("failed to create user: expected status to be %v got %v ", http.StatusCreated, resp.StatusCode)
	// }

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("HTTP error: %v", err)
	}

	var userResponse profileutils.UserResponse
	err = json.Unmarshal(data, &userResponse)
	if err != nil {
		return nil, fmt.Errorf("unable to marshall response: %v", err)
	}
	return &userResponse, nil
}

func CreateTestRole(t *testing.T, roleName string) (*dto.RoleOutput, error) {
	ctx := getTestAuthenticatedContext(t)

	validPayload := composeValidRolePayload(t, roleName)

	role, err := testInteractor.CreateUnauthorizedRole(ctx, *validPayload)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func AssignTestRole(t *testing.T, profileID, roleID string) (bool, error) {
	ctx := getTestAuthenticatedContext(t)

	_, err := testInteractor.AssignRole(ctx, profileID, roleID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func RemoveTestRole(t *testing.T, name string) (bool, error) {
	ctx := getTestAuthenticatedContext(t)

	role, err := testInteractor.GetRoleByName(ctx, testRoleName)
	if err != nil {
		return false, err
	}

	_, err = testInteractor.DeleteRole(ctx, role.ID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func TestCreateTestUserByPhone(t *testing.T) {
	userResponse, err := CreateTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
	if err != nil {
		t.Errorf("failed to create test user")
		return
	}
	if userResponse == nil {
		t.Errorf("got a nil user response")
		return
	}

	role, err := CreateTestRole(t, testRoleName)
	if err != nil {
		t.Errorf("cannot create test role with err: %v", err)
		return
	}

	_, err = AssignTestRole(t, userResponse.Profile.ID, role.ID)
	if err != nil {
		t.Errorf("cannot assign test role with err: %v", err)
		return
	}

	// perform tear down; remove user
	_, err = RemoveTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}

func RemoveTestUserByPhone(t *testing.T, phone string) (bool, error) {
	_, err := RemoveTestRole(t, testRoleName)
	if err != nil {
		return false, fmt.Errorf("unable to remove test role: %s", err)
	}

	validPayload := &dto.PhoneNumberPayload{PhoneNumber: &phone}
	bs, err := json.Marshal(validPayload)
	if err != nil {
		return false, fmt.Errorf("unable to marshal test item to JSON: %s", err)
	}
	payload := bytes.NewBuffer(bs)

	url := fmt.Sprintf("%s/remove_user", baseURL)

	r, err := http.NewRequest(
		http.MethodPost,
		url,
		payload,
	)
	if err != nil {
		return false, fmt.Errorf("can't create new request: %v", err)

	}
	if r == nil {
		return false, fmt.Errorf("nil request")
	}

	r.Header.Add("Accept", "application/json")
	r.Header.Add("Content-Type", "application/json")

	client := http.DefaultClient

	resp, err := client.Do(r)
	if err != nil {
		return false, fmt.Errorf("HTTP error: %v", err)

	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf(
			"failed to remove user: expected status to be %v got %v ",
			http.StatusCreated,
			resp.StatusCode,
		)
	}

	return true, nil
}

func TestRemoveTestUserByPhone(t *testing.T) {
	phone := interserviceclient.TestUserPhoneNumber
	userResponse, err := CreateTestUserByPhone(t, phone)
	if err != nil {
		t.Errorf("failed to create test user")
		return
	}
	if userResponse == nil {
		t.Errorf("got a nil user response")
		return
	}

	role, err := CreateTestRole(t, testRoleName)
	if err != nil {
		t.Errorf("cannot create test role with err: %v", err)
		return
	}

	_, err = AssignTestRole(t, userResponse.Profile.ID, role.ID)
	if err != nil {
		t.Errorf("cannot assign test role with err: %v", err)
		return
	}

	removed, err := RemoveTestUserByPhone(t, phone)
	if err != nil {
		t.Errorf("an error occurred: %v", err)
		return
	}
	if !removed {
		t.Errorf("user was not removed")
		return
	}
}

func generateTestOTP(t *testing.T, phone string) (*profileutils.OtpResponse, error) {
	infrastructure := infrastructure.NewInfrastructureInteractor()
	ctx := context.Background()
	testAppID := uuid.New().String()
	return infrastructure.Engagement.GenerateAndSendOTP(ctx, phone, &testAppID)
}

func getTestUserCredentials(t *testing.T) (*profileutils.UserResponse, error) {
	ctx := context.Background()

	phone := interserviceclient.TestUserPhoneNumber
	pin := testPIN
	flavour := feedlib.FlavourPro
	userResponse, err := testInteractor.LoginByPhone(ctx, phone, pin, flavour)
	if err != nil {
		return nil, fmt.Errorf("unable to get test user credentials: %v", err)
	}

	return userResponse, nil
}

func getRoleByName(t *testing.T, name string) (*dto.RoleOutput, error) {
	ctx := context.Background()

	phone := interserviceclient.TestUserPhoneNumber
	pin := testPIN
	flavour := feedlib.FlavourPro
	userResponse, err := testInteractor.LoginByPhone(ctx, phone, pin, flavour)
	if err != nil {
		return nil, fmt.Errorf("unable to get role by name: %v", err)
	}

	authCred := &auth.Token{UID: userResponse.Auth.UID}
	authenticatedContext := context.WithValue(
		ctx,
		firebasetools.AuthTokenContextKey,
		authCred,
	)

	role, err := testInteractor.GetRoleByName(authenticatedContext, name)
	if err != nil {
		return nil, fmt.Errorf("unable to get role by name: %v", err)
	}

	return role, nil
}

func setUpLoggedInTestUserGraphHeaders(t *testing.T) map[string]string {
	// create a user and their profile
	phoneNumber := interserviceclient.TestUserPhoneNumber
	userResponse, err := CreateTestUserByPhone(t, phoneNumber)
	if err != nil {
		t.Errorf("unable to create a test user: %s", err)
		return nil
	}

	if userResponse.Profile.ID == "" {
		t.Errorf(" user profile id should not be empty")
		return nil
	}

	if len(userResponse.Profile.VerifiedUIDS) == 0 {
		t.Errorf(" user profile VerifiedUIDS should not be empty")
		return nil
	}

	logrus.Infof("profile from create user : %v", userResponse.Profile)

	logrus.Infof("uid from create user : %v", userResponse.Auth.UID)

	role, err := CreateTestRole(t, testRoleName)
	if err != nil {
		t.Errorf("cannot create test role with err: %v", err)
		return nil
	}

	_, err = AssignTestRole(t, userResponse.Profile.ID, role.ID)
	if err != nil {
		t.Errorf("cannot assign test role with err: %v", err)
		return nil
	}

	headers := getGraphHeaders(*userResponse.Auth.IDToken)

	return headers
}

func getGraphHeaders(idToken string) map[string]string {
	return req.Header{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", idToken),
	}
}

func getTestAuthenticatedContext(t *testing.T) context.Context {
	ctx := context.Background()
	user, err := getTestUserCredentials(t)
	if err != nil {
		t.Errorf("error getting test user credentials:%v", err)
	}

	authCred := &auth.Token{UID: user.Auth.UID}
	authenticatedContext := context.WithValue(
		ctx,
		firebasetools.AuthTokenContextKey,
		authCred,
	)

	return authenticatedContext
}

func TestMain(m *testing.M) {
	// setup
	os.Setenv("ENVIRONMENT", "staging")
	os.Setenv("ROOT_COLLECTION_SUFFIX", "onboarding_testing")
	os.Setenv("SAVANNAH_ADMIN_EMAIL", "test@bewell.co.ke")
	os.Setenv("REPOSITORY", "firebase")

	ctx := context.Background()
	srv, baseURL, serverErr = serverutils.StartTestServer(
		ctx,
		presentation.PrepareServer,
		presentation.AllowedOrigins,
	) // set the globals
	if serverErr != nil {
		log.Printf("unable to start test server: %s", serverErr)
		return
	}

	fsc, _ := initializeAcceptanceTestFirebaseClient(ctx)

	infra, err := InitializeTestInfrastructure(ctx)
	if err != nil {
		log.Printf("unable to initialize test infrastructure: %v", err)
		return
	}

	testInfra = infra

	i, err := InitializeTestService(ctx, infra)
	if err != nil {
		log.Printf("unable to initialize test service: %v", err)
		return
	}

	testInteractor = i

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

	purgeRecords()

	// run the tests
	log.Printf("about to run tests")
	code := m.Run()
	log.Printf("finished running tests")

	// cleanup here
	defer func() {
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Printf("test server shutdown error: %s", err)
		}
	}()
	os.Exit(code)
}

func TestRouter(t *testing.T) {
	ctx := context.Background()
	router, err := presentation.Router(ctx)
	if err != nil {
		t.Errorf("can't initialize router: %v", err)
		return
	}

	if router == nil {
		t.Errorf("nil router")
		return
	}
}

func TestHealthStatusCheck(t *testing.T) {
	client := http.DefaultClient

	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantErr    bool
	}{
		{
			name: "successful health check",
			args: args{
				url: fmt.Sprintf(
					"%s/health",
					baseURL,
				),
				httpMethod: http.MethodPost,
				body:       nil,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := http.NewRequest(
				tt.args.httpMethod,
				tt.args.url,
				tt.args.body,
			)

			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			if r == nil {
				t.Errorf("nil request")
				return
			}

			for k, v := range interserviceclient.GetDefaultHeaders(t, baseURL, "profile") {
				r.Header.Add(k, v)
			}

			resp, err := client.Do(r)
			if err != nil {
				t.Errorf("HTTP error: %v", err)
				return
			}

			if !tt.wantErr && resp == nil {
				t.Errorf("unexpected nil response (did not expect an error)")
				return
			}

			if tt.wantErr {
				return
			}

			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("can't read response body: %v", err)
				return
			}

			if data == nil {
				t.Errorf("nil response body data")
				return
			}

			if tt.wantStatus != resp.StatusCode {
				t.Errorf(
					"expected status %d, got %d and response %s",
					tt.wantStatus,
					resp.StatusCode,
					string(data),
				)
				return
			}

			if !tt.wantErr && resp == nil {
				t.Errorf("unexpected nil response (did not expect an error)")
				return
			}
		})
	}
}
