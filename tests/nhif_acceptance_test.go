package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"firebase.google.com/go/auth"
	"github.com/google/uuid"
	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/profileutils"
)

func TestAddNHIFDetails(t *testing.T) {
	headers := setUpLoggedInTestUserGraphHeaders(t)

	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")

	graphqlMutation := `mutation addNHIFDetails($input: NHIFDetailsInput!) {
		addNHIFDetails(input: $input) {
		  id
		  profileID
		  membershipNumber
		  idNumber
		  idDocType
		  identificationCardPhotoID
		  NHIFCardPhotoID
		}
	}`

	membershipNo := fmt.Sprintln(time.Now().Unix())

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
			name: "success: Add NHIF Details",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"input": map[string]interface{}{
							"membershipNumber":          membershipNo,
							"idNumber":                  "123456",
							"idDocType":                 "NATIONALID",
							"identificationCardPhotoID": uuid.New().String(),
							"NHIFCardPhotoID":           uuid.New().String(),
							"employment":                "EMPLOYED",
						},
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:fail to Add NHIF Details",
			args: args{
				query: map[string]interface{}{
					"query":     `Invalid mutation query`,
					"variables": ``,
				},
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantErr:    true,
		},
		{
			name: "failure: Add existing NHIF Details",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"input": map[string]interface{}{
							"membershipNumber":          membershipNo,
							"idNumber":                  "123456",
							"idDocType":                 "NATIONALID",
							"identificationCardPhotoID": uuid.New().String(),
							"NHIFCardPhotoID":           uuid.New().String(),
							"employment":                "EMPLOYED",
						},
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
				t.Errorf("bad data returned")
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
		})
	}
	// perform tear down; remove user
	_, err := RemoveTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}

func AddTestNHIFDetails(t *testing.T, user *profileutils.UserResponse) error {
	ctx := context.Background()

	authCred := &auth.Token{UID: user.Auth.UID}
	authenticatedContext := context.WithValue(
		ctx,
		firebasetools.AuthTokenContextKey,
		authCred,
	)

	_, err := testInteractor.NHIF.AddNHIFDetails(
		authenticatedContext,
		dto.NHIFDetailsInput{
			MembershipNumber:          fmt.Sprintln(time.Now().Unix()),
			Employment:                domain.EmploymentTypeEmployed,
			NHIFCardPhotoID:           uuid.New().String(),
			IDDocType:                 enumutils.IdentificationDocTypeMilitary,
			IdentificationCardPhotoID: uuid.New().String(),
			IDNumber:                  fmt.Sprintln(time.Now().Unix()),
		},
	)
	if err != nil {
		return fmt.Errorf("an error occurred: %v", err)
	}

	return nil
}

func TestAddTestNHIFDetails(t *testing.T) {
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

	err = AddTestNHIFDetails(t, user)
	if err != nil {
		t.Errorf("an error occurred: %v", err)
		return
	}

	// perform tear down; remove user
	_, err = RemoveTestUserByPhone(t, phoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}
func TestGetNHIFDetails(t *testing.T) {
	headers := setUpLoggedInTestUserGraphHeaders(t)

	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")
	graphqlQuery := `query NHIFDetails{
		NHIFDetails{
			id
			profileID
			membershipNumber
			idNumber
			idDocType
			identificationCardPhotoID
			NHIFCardPhotoID
		}
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
			name: "success: get a user's NHIF details",
			args: args{
				query: map[string]interface{}{
					"query": graphqlQuery,
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:Fail to Find NHIF Details",
			args: args{
				query: map[string]interface{}{
					"query": "invalid query",
				},
			},
			wantStatus: http.StatusUnprocessableEntity,
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
				t.Errorf("bad data returned")
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
		})
	}
	// perform tear down; remove user
	_, err := RemoveTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}
