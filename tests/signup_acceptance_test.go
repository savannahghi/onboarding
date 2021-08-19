package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"testing"
	"time"

	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/profileutils"
)

func TestCreateUserWithPhoneNumber(t *testing.T) {
	client := http.DefaultClient

	phoneNumber := interserviceclient.TestUserPhoneNumber
	validPayload, err := composeValidUserPayload(t, phoneNumber)
	if err != nil {
		t.Errorf("failed to compose a valid payload")
		return
	}

	bs, err := json.Marshal(validPayload)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	payload := bytes.NewBuffer(bs)

	// invalid payload
	badPayload := composeInValidUserPayload(t)
	bs2, err := json.Marshal(badPayload)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	invalidPayload := bytes.NewBuffer(bs2)

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
			name: "success: signup user with valid payload",
			args: args{
				url:        fmt.Sprintf("%s/create_user_by_phone", baseURL),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name: "failure: signup user with the same valid payload again",
			args: args{
				url:        fmt.Sprintf("%s/create_user_by_phone", baseURL),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "failure: signup user with nil payload supplied",
			args: args{
				url:        fmt.Sprintf("%s/create_user_by_phone", baseURL),
				httpMethod: http.MethodPost,
				body:       nil,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "failure: signup user with invalid payload",
			args: args{
				url:        fmt.Sprintf("%s/create_user_by_phone", baseURL),
				httpMethod: http.MethodPost,
				body:       invalidPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
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
			if tt.wantStatus != resp.StatusCode {
				t.Errorf("expected status %d, got %d", tt.wantStatus, resp.StatusCode)
				return
			}
			dataResponse, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("can't read response body: %v", err)
				return
			}
			if dataResponse == nil {
				t.Errorf("nil response body data")
				return
			}

		})
	}

	// assign role before removing
	user, err := getTestUserCredentials(t)
	if err != nil {
		t.Errorf("error getting test user credentials:%v", err)
	}

	role, err := CreateTestRole(t, testRoleName)
	if err != nil {
		t.Errorf("cannot create test role with err: %v", err)
		return
	}

	_, err = AssignTestRole(t, user.Profile.ID, role.ID)
	if err != nil {
		t.Errorf("cannot assign test role with err: %v", err)
		return
	}
	// perform tear down; remove user
	_, err = RemoveTestUserByPhone(t, phoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}

func TestVerifySignUpPhoneNumber(t *testing.T) {
	client := http.DefaultClient
	// create a test user
	user, err := CreateTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
	if err != nil {
		t.Errorf("failed to create a user by phone %v", err)
		return
	}

	role, err := CreateTestRole(t, testRoleName)
	if err != nil {
		t.Errorf("cannot create test role with err: %v", err)
		return
	}

	_, err = AssignTestRole(t, user.Profile.ID, role.ID)
	if err != nil {
		t.Errorf("cannot assign test role with err: %v", err)
		return
	}

	// prepare a valid payload
	registeredPhone := struct {
		PhoneNumber string
	}{
		PhoneNumber: interserviceclient.TestUserPhoneNumberWithPin,
	}
	bs, err := json.Marshal(registeredPhone)
	if err != nil {
		t.Errorf("unable to marshal registeredPhone to JSON: %s", err)
		return
	}
	payload := bytes.NewBuffer(bs)

	// prepare an invalid payload
	unregisteredPhone := struct {
		PhoneNumber string
	}{
		PhoneNumber: interserviceclient.TestUserPhoneNumber,
	}
	bs2, err := json.Marshal(unregisteredPhone)
	if err != nil {
		t.Errorf("unable to marshal unregisteredPhone to JSON: %s", err)
	}
	unregisteredUser := bytes.NewBuffer(bs2)

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
			name: "success: verify a new user phone number",
			args: args{
				url:        fmt.Sprintf("%s/verify_phone", baseURL),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "failure: verify a phone number whose profile already exists",
			args: args{
				url:        fmt.Sprintf("%s/verify_phone", baseURL),
				httpMethod: http.MethodPost,
				body:       unregisteredUser,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "failure: verify a phone number with a nil payload",
			args: args{
				url:        fmt.Sprintf("%s/create_user_by_phone", baseURL),
				httpMethod: http.MethodPost,
				body:       nil,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
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
				t.Errorf("expected status %d, got %d and response %s", tt.wantStatus, resp.StatusCode, string(data))
				return
			}

			if !tt.wantErr && resp == nil {
				t.Errorf("unexpected nil response (did not expect an error)")
				return
			}
		})
	}
	// perform tear down; remove user
	_, err = RemoveTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}

func TestUserRecoveryPhoneNumbers(t *testing.T) {
	client := http.DefaultClient
	// create a test user
	validPhoneNumber := interserviceclient.TestUserPhoneNumber
	user, err := CreateTestUserByPhone(t, validPhoneNumber)
	if err != nil {
		t.Errorf("failed to create a user by phone %v", err)
		return
	}

	role, err := CreateTestRole(t, testRoleName)
	if err != nil {
		t.Errorf("cannot create test role with err: %v", err)
		return
	}

	_, err = AssignTestRole(t, user.Profile.ID, role.ID)
	if err != nil {
		t.Errorf("cannot assign test role with err: %v", err)
		return
	}

	validPayload := dto.PhoneNumberPayload{
		PhoneNumber: &validPhoneNumber,
	}
	bs, err := json.Marshal(validPayload)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	payload := bytes.NewBuffer(bs)

	// phone number not registered
	inValidNumber := interserviceclient.TestUserPhoneNumberWithPin
	badPayload := dto.PhoneNumberPayload{
		PhoneNumber: &inValidNumber,
	}
	bs2, err := json.Marshal(badPayload)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	invalidPayload := bytes.NewBuffer(bs2)

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
			name: "success: user recovery with valid payload",
			args: args{
				url:        fmt.Sprintf("%s/user_recovery_phonenumbers", baseURL),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "failure: user recovery with nil payload supplied",
			args: args{
				url:        fmt.Sprintf("%s/user_recovery_phonenumbers", baseURL),
				httpMethod: http.MethodPost,
				body:       nil,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "failure: user recovery with invalid payload",
			args: args{
				url:        fmt.Sprintf("%s/user_recovery_phonenumbers", baseURL),
				httpMethod: http.MethodPost,
				body:       invalidPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
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
			if tt.wantStatus != resp.StatusCode {
				dump, _ := httputil.DumpResponse(resp, true)
				t.Errorf("expected status %d, got %d with response %v", tt.wantStatus, resp.StatusCode, string(dump))
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
		})
	}

	// perform tear down; remove user
	_, err = RemoveTestUserByPhone(t, validPhoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}

func TestRegisterPushToken(t *testing.T) {
	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")
	headers := setUpLoggedInTestUserGraphHeaders(t)

	graphqlMutation := `
	mutation registerPushToken($token:String!){
		registerPushToken(token: $token)
	  }
	`
	token := "QP18DqWVyuOcPG8CcDUNcEDzU3A2" // ggignore
	type args struct {
		query map[string]interface{}
	}

	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantErr    bool
	}{
		{
			name: "success: register token with valid payload",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"phone": interserviceclient.TestUserPhoneNumber,
						"token": token,
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "failure: register token with bogus payload",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"token": "*",
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := mapToJSONReader(tt.args.query)

			if err != nil {
				t.Errorf("unable to get GQL JSON io Reader: %s", err)
				return
			}

			r, err := http.NewRequest(
				http.MethodPost,
				graphQLURL,
				body,
			)

			if err != nil {
				t.Errorf("unable to compose request: %s", err)
				return
			}

			if r == nil {
				t.Errorf("nil request")
				return
			}

			for k, v := range headers {
				r.Header.Add(k, v)
			}

			client := http.Client{
				Timeout: time.Second * testHTTPClientTimeout,
			}
			resp, err := client.Do(r)
			if err != nil {
				t.Errorf("request error: %s", err)
				return
			}

			dataResponse, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("can't read request body: %s", err)
				return
			}
			if dataResponse == nil {
				t.Errorf("nil response data")
				return
			}

			data := map[string]interface{}{}
			err = json.Unmarshal(dataResponse, &data)
			if err != nil {
				t.Errorf("bad data returned %v", data)
				return
			}
			if tt.wantErr {
				errMsg, ok := data["errors"]
				if !ok {
					t.Errorf("GraphQL error: %s", errMsg)
					return
				}
			}

			if !tt.wantErr {
				_, ok := data["errors"]
				if ok {
					t.Errorf("error not expected")
					return
				}
			}
			if tt.wantStatus != resp.StatusCode {
				t.Errorf("Bad status response returned")
				return
			}
		})
	}
	// perform tear down; remove user
	_, err := RemoveTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}

func TestCompleteSignup(t *testing.T) {
	headers := setUpLoggedInTestUserGraphHeaders(t)
	authenticatedContext := getTestAuthenticatedContext(t)

	firstName := "Be.Well"
	lastName := "Consumer"
	bioData := profileutils.BioData{
		FirstName: &firstName,
		LastName:  &lastName,
	}
	// Update the user BioData
	err := testInteractor.Onboarding.UpdateBioData(authenticatedContext, bioData)

	if err != nil {
		t.Errorf("failed to update userprofile biodata: %v", err)
		return
	}

	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")

	graphqlMutation := `
	mutation CompleteSingup($flavour:Flavour!){
		completeSignup(flavour:$flavour)
	  }
	`
	type args struct {
		query map[string]interface{}
	}

	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantErr    bool
	}{
		{
			name: "success: complete signup -  B.Well Consumer ERP account creation.",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"flavour": feedlib.FlavourConsumer,
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "failure: complete signup -  B.Well Consumer ERP account creation.",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"flavour": feedlib.FlavourPro, // invalid flavour
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := mapToJSONReader(tt.args.query)

			if err != nil {
				t.Errorf("unable to get GQL JSON io Reader: %s", err)
				return
			}

			r, err := http.NewRequest(
				http.MethodPost,
				graphQLURL,
				body,
			)

			if err != nil {
				t.Errorf("unable to compose request: %s", err)
				return
			}

			if r == nil {
				t.Errorf("nil request")
				return
			}

			for k, v := range headers {
				r.Header.Add(k, v)
			}

			client := http.Client{
				Timeout: time.Second * testHTTPClientTimeout,
			}
			resp, err := client.Do(r)
			if err != nil {
				t.Errorf("request error: %s", err)
				return
			}

			dataResponse, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("can't read request body: %s", err)
				return
			}
			if dataResponse == nil {
				t.Errorf("nil response data")
				return
			}

			data := map[string]interface{}{}
			err = json.Unmarshal(dataResponse, &data)
			if err != nil {
				t.Errorf("bad data returned %v", data)
				return
			}
			if tt.wantErr {
				errMsg, ok := data["errors"]
				if !ok {
					t.Errorf("GraphQL error: %s", errMsg)
					return
				}
			}

			if !tt.wantErr {
				_, ok := data["errors"]
				if ok {
					t.Errorf("error not expected")
					return
				}
			}
			if tt.wantStatus != resp.StatusCode {
				t.Errorf("Bad status response returned")
				return
			}
		})
	}
	// perform tear down; remove user
	_, err = RemoveTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}
