package extension_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/stretchr/testify/assert"
)

// var baseExt mock.FakeBaseExtensionImpl

func TestMain(m *testing.M) {
	log.Printf("Setting tests up ...")
	envOriginalValue := os.Getenv("ENVIRONMENT")
	os.Setenv("ENVIRONMENT", "staging")
	debugEnvValue := os.Getenv("DEBUG")
	os.Setenv("DEBUG", "true")

	// do clean up
	log.Printf("Running tests ...")
	code := m.Run()

	log.Printf("Tearing tests down ...")

	// restore environment varibles to original values
	os.Setenv(envOriginalValue, "ENVIRONMENT")
	os.Setenv("DEBUG", debugEnvValue)

	os.Exit(code)
}

func TestGetLoggedInUser(t *testing.T) {
	baseExt := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})
	ctx := firebasetools.GetAuthenticatedContext(t)
	invalidCtx := context.Background()

	userInfo, err := baseExt.GetLoggedInUser(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, userInfo)

	userInfo, err = baseExt.GetLoggedInUser(invalidCtx)
	assert.NotNil(t, err)
	assert.Empty(t, userInfo)
}

func TestGetLoggedInUserUID(t *testing.T) {
	baseExt := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})
	ctx := firebasetools.GetAuthenticatedContext(t)
	invalidCtx := context.Background()

	uid, err := baseExt.GetLoggedInUserUID(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, uid)

	uid, err = baseExt.GetLoggedInUserUID(invalidCtx)
	assert.NotNil(t, err)
	assert.Empty(t, uid)
}

func TestNormalizeMSISDN(t *testing.T) {
	baseExt := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})
	validMSISDN := interserviceclient.TestUserPhoneNumber
	invalidMSISDN := "INVALID"

	msisdn, err := baseExt.NormalizeMSISDN(validMSISDN)
	assert.Nil(t, err)
	assert.NotNil(t, msisdn)

	msisdn, err = baseExt.NormalizeMSISDN(invalidMSISDN)
	assert.NotNil(t, err)
	assert.Empty(t, msisdn)
}

// func TestFetchDefaultCurrency(t *testing.T) {
// 	_, token := firebasetools.GetAuthenticatedContextAndToken(t)
// 	// uid, err := firebasetools.GetLoggedInUserUID(ctx)
// 	// if err != nil {
// 	// 	t.Errorf("Failed to create client: %v", err)
// 	// }
// 	// baseExt := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})
// 	fakeBaseExt := mock.FakeBaseExtensionImpl{}
// 	fakeBaseExt.FetchDefaultCurrencyFn()
// 	clientID := token.UID
// 	clientSecret := "OEOWOEOWPEOOWPEOPWEOWPEO"
// 	apiTokenURL := "https://test.com"
// 	apiHost := "test.com"
// 	apiScheme := "https"
// 	grantType := "password"
// 	username := "testUser"
// 	password := "testPassword"
// 	extraHeaders := map[string]string{}
// 	client, err := fakeBaseExt.NewServerClientFn(clientID, clientSecret, apiTokenURL, apiHost, apiScheme, grantType, username, password, extraHeaders)
// 	// client, err := apiclient.NewServerClient(clientID, clientSecret, apiTokenURL, apiHost, apiScheme, grantType, username, password, extraHeaders)
// 	if err != nil {
// 		t.Errorf("Failed to create client: %v", err)
// 	}
// 	assert.NotNil(t, client)
// 	// invalidMSISDN := "INVALID"
// 	// msisdn, err := baseExt.NormalizeMSISDN(validMSISDN)
// 	// assert.Nil(t, err)
// 	// assert.NotNil(t, msisdn)
// 	// msisdn, err = baseExt.NormalizeMSISDN(invalidMSISDN)
// 	// assert.NotNil(t, err)
// 	// assert.Empty(t, msisdn)
// }

// func TestLoginClient(t *testing.T) {
// }

// func TestFetchUserProfile(t *testing.T) {
// 	baseExt := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})
// 	// _, token := firebasetools.GetAuthenticatedContextAndToken(t)
// 	// uid, err := firebasetools.GetLoggedInUserUID(ctx)
// 	// if err != nil {
// 	// 	t.Errorf("Failed to create client: %v", err)
// 	// }
// 	ediUser := profileutils.EDIUserProfile{
// 		ID: 12345,
// 	}
// 	profile, err := baseExt.FetchUserProfile(ediUser)
// }

func TestLoadDepsFromYAML(t *testing.T) {
	baseExt := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})

	loadDeps, err := baseExt.LoadDepsFromYAML()
	assert.NotNil(t, loadDeps)
	assert.Nil(t, err)

}

func TestSetupISCclient(t *testing.T) {
	baseExt := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})

	service := "test"

	loadDeps, err := baseExt.LoadDepsFromYAML()
	assert.NotNil(t, loadDeps)
	assert.Nil(t, err)

	setupIsc, err := baseExt.SetupISCclient(*loadDeps, service)
	assert.NotNil(t, setupIsc)
	assert.Nil(t, err)
}

func TestGetEnvVar(t *testing.T) {
	baseExt := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})

	envVar := "DEBUG"
	nonexistentEnvVar := "NON_EXISTENT_ENV"
	emptyEnvVar := ""

	env, err := baseExt.GetEnvVar(envVar)
	assert.NotNil(t, env)
	assert.Nil(t, err)

	env, err = baseExt.GetEnvVar(nonexistentEnvVar)
	assert.NotNil(t, err)
	assert.Empty(t, env)

	env, err = baseExt.GetEnvVar(emptyEnvVar)
	assert.NotNil(t, err)
	assert.Empty(t, env)
}

func TestGetLoginFunc(t *testing.T) {
	baseExt := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})

	ctx := firebasetools.GetAuthenticatedContext(t)

	handler := baseExt.GetLoginFunc(ctx)
	assert.NotNil(t, handler)
}

func TestGetRefreshFunc(t *testing.T) {
	baseExt := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})

	handler := baseExt.GetRefreshFunc()
	assert.NotNil(t, handler)
}

func TestGetVerifyTokenFunc(t *testing.T) {
	baseExt := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})

	ctx := firebasetools.GetAuthenticatedContext(t)

	handler := baseExt.GetVerifyTokenFunc(ctx)
	assert.NotNil(t, handler)
}
