package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"testing"
	"time"

	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/stretchr/testify/assert"
)

func composeInvalidUserPINPayload(t *testing.T) *dto.LoginPayload {
	phone := interserviceclient.TestUserPhoneNumber
	pin := "" // empty pin
	flavour := feedlib.FlavourPro
	payload := &dto.LoginPayload{
		PhoneNumber: &phone,
		PIN:         &pin,
		Flavour:     flavour,
	}
	return payload
}

func composeWrongUserPINPayload(t *testing.T) *dto.LoginPayload {
	phone := interserviceclient.TestUserPhoneNumber // This number should be the same as the
	// used to create the user
	pin := "qwer"
	flavour := feedlib.FlavourPro
	payload := &dto.LoginPayload{
		PhoneNumber: &phone,
		PIN:         &pin,
		Flavour:     flavour,
	}
	return payload
}

func composeWrongUserPhonePayload(t *testing.T) *dto.LoginPayload {
	phone := "+254700000000"
	pin := interserviceclient.TestUserPin
	flavour := feedlib.FlavourPro
	payload := &dto.LoginPayload{
		PhoneNumber: &phone,
		PIN:         &pin,
		Flavour:     flavour,
	}
	return payload
}

func composeInvalidUserPhonePayload(t *testing.T) *dto.LoginPayload {
	phone := "+254-not-a-number"
	pin := interserviceclient.TestUserPin
	flavour := feedlib.FlavourPro
	payload := &dto.LoginPayload{
		PhoneNumber: &phone,
		PIN:         &pin,
		Flavour:     flavour,
	}
	return payload
}

func composeWrongFlavourPayload(t *testing.T) *dto.LoginPayload {
	phone := interserviceclient.TestUserPhoneNumber
	pin := interserviceclient.TestUserPin
	payload := &dto.LoginPayload{
		PhoneNumber: &phone,
		PIN:         &pin,
		Flavour:     "bad-flavour-supplied",
	}
	return payload
}

func TestLoginInByPhone(t *testing.T) {
	phoneNumber := interserviceclient.TestUserPhoneNumber
	user, err := CreateTestUserByPhone(t, phoneNumber)
	if err != nil {
		t.Errorf("failed to create a user by phone %v", err)
		return
	}
	if user == nil {
		t.Errorf("nil user found")
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

	client := http.DefaultClient
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
	badPayload := composeInvalidUserPINPayload(t)
	bs2, err := json.Marshal(badPayload)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	invalidPayload := bytes.NewBuffer(bs2)

	wrongPINPayload := composeWrongUserPINPayload(t)
	wrongPINBs, err := json.Marshal(wrongPINPayload)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	badPINpayload := bytes.NewBuffer(wrongPINBs)

	wrongPhonePayload := composeWrongUserPhonePayload(t)
	wrongPhoneBs, err := json.Marshal(wrongPhonePayload)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	badPhonepayload := bytes.NewBuffer(wrongPhoneBs)

	invalidPhonePayload := composeInvalidUserPhonePayload(t)
	invalidPhoneBs, err := json.Marshal(invalidPhonePayload)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	badInvalidPhonepayload := bytes.NewBuffer(invalidPhoneBs)

	emptyData := &dto.LoginPayload{}
	emptyBs, err := json.Marshal(emptyData)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	emptyPayload := bytes.NewBuffer(emptyBs)

	invalidFlavourPayload := composeWrongFlavourPayload(t)
	invalidFlavourBs, err := json.Marshal(invalidFlavourPayload)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	badFlavourPayload := bytes.NewBuffer(invalidFlavourBs)

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
			name: "success: login user with valid payload",
			args: args{
				url:        fmt.Sprintf("%s/login_by_phone", baseURL),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "failure: login user with nil payload supplied",
			args: args{
				url:        fmt.Sprintf("%s/login_by_phone", baseURL),
				httpMethod: http.MethodPost,
				body:       emptyPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "failure: login user with invalid payload",
			args: args{
				url:        fmt.Sprintf("%s/login_by_phone", baseURL),
				httpMethod: http.MethodPost,
				body:       invalidPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "failure: login user with a wrong PIN",
			args: args{
				url:        fmt.Sprintf("%s/login_by_phone", baseURL),
				httpMethod: http.MethodPost,
				body:       badPINpayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "failure: login user with a wrong primary phone number",
			args: args{
				url:        fmt.Sprintf("%s/login_by_phone", baseURL),
				httpMethod: http.MethodPost,
				body:       badPhonepayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "failure: login user with invalid phone number",
			args: args{
				url:        fmt.Sprintf("%s/login_by_phone", baseURL),
				httpMethod: http.MethodPost,
				body:       badInvalidPhonepayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "failure: login user with invalid flavour",
			args: args{
				url:        fmt.Sprintf("%s/login_by_phone", baseURL),
				httpMethod: http.MethodPost,
				body:       badFlavourPayload,
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

			for k, v := range interserviceclient.GetDefaultHeaders(t, baseURL, "onboarding") {
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

			data := map[string]interface{}{}
			log.Printf("the data is %v", data)
			err = json.Unmarshal(dataResponse, &data)
			if err != nil {
				t.Errorf("bad data returned")
				return
			}
			if tt.wantErr {
				errMsg, ok := data["error"]
				if !ok {
					t.Errorf("Request error: %s", errMsg)
					return
				}
			}

			if !tt.wantErr {
				_, ok := data["error"]
				if ok {
					t.Errorf("error not expected")
					return
				}
			}

		})
	}
	// perform tear down; remove user
	_, err = RemoveTestUserByPhone(t, phoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}

func TestLoginAsAnonymous(t *testing.T) {
	client := http.DefaultClient

	p1, err := json.Marshal(&dto.LoginPayload{
		Flavour: feedlib.FlavourConsumer,
	})
	if err != nil {
		t.Errorf("unable to marshal payload to JSON: %s", err)
	}
	validPayload := bytes.NewBuffer(p1)

	p2, err := json.Marshal(&dto.LoginPayload{
		Flavour: feedlib.FlavourPro,
	})
	if err != nil {
		t.Errorf("unable to marshal payload to JSON: %s", err)
	}
	invalidPayload := bytes.NewBuffer(p2)

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
			name: "valid : correct flavour",
			args: args{
				url:        fmt.Sprintf("%s/login_anonymous", baseURL),
				httpMethod: http.MethodPost,
				body:       validPayload,
			},
			wantErr:    false,
			wantStatus: http.StatusOK,
		},
		{
			name: "valid : incorrect flavour",
			args: args{
				url:        fmt.Sprintf("%s/login_anonymous", baseURL),
				httpMethod: http.MethodPost,
				body:       invalidPayload,
			},
			wantErr:    true,
			wantStatus: http.StatusBadRequest,
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

			for k, v := range interserviceclient.GetDefaultHeaders(t, baseURL, "onboarding") {
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

			data := map[string]interface{}{}
			err = json.Unmarshal(dataResponse, &data)
			if err != nil {
				t.Errorf("bad data returned")
				return
			}

			if tt.wantErr {
				errMsg, ok := data["error"]
				if !ok {
					t.Errorf("Request error: %s", errMsg)
					return
				}
			}

			if !tt.wantErr {
				_, ok := data["error"]
				if ok {
					t.Errorf("error not expected")
					return
				}
			}

		})
	}

}

func TestRefreshToken(t *testing.T) {
	client := http.DefaultClient
	phoneNumber := interserviceclient.TestUserPhoneNumber
	user, err := CreateTestUserByPhone(t, phoneNumber)
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

	validToken := user.Auth.RefreshToken
	validPayload := &dto.RefreshTokenPayload{
		RefreshToken: &validToken,
	}
	bs, err := json.Marshal(validPayload)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	payload := bytes.NewBuffer(bs)

	inValidToken := "some-token"
	inValidPayload := &dto.RefreshTokenPayload{
		RefreshToken: &inValidToken,
	}
	badBs, err := json.Marshal(inValidPayload)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	badPayload := bytes.NewBuffer(badBs)

	emptyData := &dto.LoginPayload{}
	emptyBs, err := json.Marshal(emptyData)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	emptyPayload := bytes.NewBuffer(emptyBs)

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
			name: "success: refresh a token",
			args: args{
				url:        fmt.Sprintf("%s/refresh_token", baseURL),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "failure: refresh token with nil payload supplied",
			args: args{
				url:        fmt.Sprintf("%s/refresh_token", baseURL),
				httpMethod: http.MethodPost,
				body:       emptyPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "failure: refresh token with invalid payload",
			args: args{
				url:        fmt.Sprintf("%s/refresh_token", baseURL),
				httpMethod: http.MethodPost,
				body:       badPayload,
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

			for k, v := range interserviceclient.GetDefaultHeaders(t, baseURL, "onboarding") {
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

			data := map[string]interface{}{}
			err = json.Unmarshal(dataResponse, &data)
			if err != nil {
				t.Errorf("bad data returned")
				return
			}
			if tt.wantErr {
				errMsg, ok := data["error"]
				if !ok {
					t.Errorf("Request error: %s", errMsg)
					return
				}
			}

			if !tt.wantErr {
				_, ok := data["error"]
				refreshToken := data["refresh_token"]
				assert.NotNil(t, refreshToken)
				if ok {
					t.Errorf("error not expected")
					return
				}
			}
		})
	}

	// perform tear down; remove user
	_, err = RemoveTestUserByPhone(t, phoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}

func TestResumeWithPin(t *testing.T) {
	headers := setUpLoggedInTestUserGraphHeaders(t)

	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")

	graphqlMutation := `
    query resumeWithPin($pin:String!){
		resumeWithPIN(pin:$pin)
	}`

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
			name: "resume with pin successfully",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"pin": testPIN,
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
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
				t.Errorf("bad data returned")
				return
			}

			if tt.wantErr {
				_, ok := data["errors"]
				if !ok {
					t.Errorf("expected an error")
					return
				}
			}

			if !tt.wantErr {
				_, ok := data["errors"]
				if ok {
					t.Errorf("error not expected got error: %w", data["errors"])
					return
				}
			}
			if tt.wantStatus != resp.StatusCode {
				b, _ := httputil.DumpResponse(resp, true)
				t.Errorf("Bad status response returned; %v ", string(b))
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
