package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"testing"
	"time"

	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/profileutils"
	"github.com/savannahghi/scalarutils"
	"github.com/sirupsen/logrus"
	"gitlab.slade360emr.com/go/apiclient"
)

const (
	testEmail = "test@bewell.co.ke"
)

// CreatedUserGraphQLHeaders updates the authorization header with the
// bearer(ID) token of the created user during test
// TODO:(muchogo)  create a reusable function in base that accepts a UID
// 				or modify the apiclient.GetGraphQLHeaders(ctx) extra UID argument
func CreatedUserGraphQLHeaders(idToken *string) (map[string]string, error) {
	ctx := context.Background()

	authHeaderBearerToken := fmt.Sprintf("Bearer %v", *idToken)

	headers, err := apiclient.GetGraphQLHeaders(ctx)
	if err != nil {
		return nil, fmt.Errorf("error in getting headers: %w", err)
	}

	headers["Authorization"] = authHeaderBearerToken

	return headers, nil
}
func TestAddPartnerType(t *testing.T) {
	// create a user and their profile
	phoneNumber := interserviceclient.TestUserPhoneNumber
	user, err := CreateTestUserByPhone(t, phoneNumber)
	if err != nil {
		t.Errorf("failed to create a user by phone %v", err)
		return
	}

	idToken := user.Auth.IDToken
	headers, err := CreatedUserGraphQLHeaders(idToken)
	if err != nil {
		t.Errorf("error in getting headers: %w", err)
		return
	}

	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")
	graphqlMutation := `
	mutation addPartnerType($name:String!, $partnerType:PartnerType!){
		addPartnerType(name: $name, partnerType:$partnerType)
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
			name: "success: add partner type with valid payload",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"name":        "juha kalulu",
						"partnerType": "RIDER",
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "failure: add partner type with non existent partner type",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"name":        "juha",
						"partnerType": "NOT FOUND",
					},
				},
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantErr:    true,
		},
		{
			name: "failure: add partner type with bogus payload",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"name":        "*",
						"partnerType": "*",
					},
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
			if tt.wantStatus != resp.StatusCode {
				b, _ := httputil.DumpResponse(resp, true)
				t.Errorf("Bad status response returned; %v ", string(b))
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

func TestSetUpSupplier_acceptance(t *testing.T) {

	// create a user and their profile
	phoneNumber := interserviceclient.TestUserPhoneNumber
	user, err := CreateTestUserByPhone(t, phoneNumber)
	if err != nil {
		t.Errorf("failed to create a user by phone %v", err)
		return
	}

	idToken := user.Auth.IDToken
	headers, err := CreatedUserGraphQLHeaders(idToken)
	if err != nil {
		t.Errorf("error in getting headers: %w", err)
		return
	}

	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")

	graphqlMutation := `
	mutation setUpSupplier($input: AccountType!) {
		setUpSupplier(accountType: $input) {
		  accountType
		}
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
			name: "individual supplier setup with valid payload",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"input": "INDIVIDUAL",
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "organisation supplier setup with valid payload",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"input": "ORGANISATION",
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid case - supplier setup with invalid payload",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"input": "",
					},
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
			if tt.wantStatus != resp.StatusCode {
				b, _ := httputil.DumpResponse(resp, true)
				t.Errorf("Bad status response returned; %v ", string(b))
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

func TestAddIndividualCoachKYC(t *testing.T) {
	ctx := context.Background()
	phoneNumber := interserviceclient.TestUserPhoneNumber
	user, err := CreateTestUserByPhone(t, phoneNumber)
	log.Printf("the user is %v", user)
	if err != nil {
		t.Errorf("failed to create a user by phone %v", err)
		return
	}
	idToken := user.Auth.IDToken
	headers, err := CreatedUserGraphQLHeaders(idToken)
	if err != nil {
		t.Errorf("error in getting headers: %w", err)
		return
	}

	authToken, err := firebasetools.ValidateBearerToken(ctx, *idToken)
	if err != nil {
		t.Errorf("invalid token: %w", err)
		return
	}
	authenticatedContext := context.WithValue(ctx, firebasetools.AuthTokenContextKey, authToken)

	err = setPrimaryEmailAddress(authenticatedContext, t, testEmail)
	if err != nil {
		t.Errorf("failed to set primary email address: %v", err)
		return
	}
	dateOfBirth2 := scalarutils.Date{
		Day:   12,
		Year:  1995,
		Month: 10,
	}
	firstName2 := "makmende"
	lastName2 := "juha"

	completeUserDetails := profileutils.BioData{
		DateOfBirth: &dateOfBirth2,
		FirstName:   &firstName2,
		LastName:    &lastName2,
	}
	partnerName := "coach"
	partnerType := profileutils.PartnerTypeCoach

	_, err = addPartnerType(authenticatedContext, t, &partnerName, partnerType)
	if err != nil {
		t.Errorf("failed to add partnerType: %v", err)
		return
	}
	_, err = setUpSupplier(authenticatedContext, t, profileutils.AccountTypeIndividual)
	if err != nil {
		t.Errorf("failed to setup supplier: %v", err)
		return
	}
	err = updateBioData(authenticatedContext, t, completeUserDetails)
	if err != nil {
		t.Errorf("failed to update biodata: %v", err)
		return
	}

	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")

	graphQLMutation := `
	mutation addIndividualCoachKYC($input:IndividualCoachInput!){
		addIndividualCoachKYC(input:$input) {
		  identificationDoc {
			identificationDocType
			identificationDocNumber
			identificationDocNumberUploadID
	  }
		  KRAPIN
		  KRAPINUploadID
		  practiceLicenseID
		  accreditationID
		  accreditationUploadID
		}
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
			name: "valid mutation request",
			args: args{
				query: map[string]interface{}{
					"query": graphQLMutation,
					"variables": map[string]interface{}{
						"input": map[string]interface{}{
							"identificationDoc": map[string]interface{}{
								"identificationDocType":           "NATIONALID",
								"identificationDocNumber":         "ID-12345678",
								"identificationDocNumberUploadID": "ID-UPLOAD-12345678",
							},
							"KRAPIN":                "KRA-12345678",
							"KRAPINUploadID":        "KRA-UPLOAD-12345678",
							"practiceLicenseID":     "PRA-12345678",
							"accreditationID":       "ACR-12345678",
							"accreditationUploadID": "ACR-UPLOAD-12345678",
						},
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid mutation request - wrong input type",
			args: args{
				query: map[string]interface{}{
					"query": graphQLMutation,
					"variables": map[string]interface{}{
						"input": map[string]interface{}{
							"identificationDoc": map[string]interface{}{
								"identificationDocType":           "NATIONALID",
								"identificationDocNumber":         "ID-12345678",
								"identificationDocNumberUploadID": "ID-UPLOAD-12345678",
							},
							"KRAPIN":                12345678,
							"KRAPINUploadID":        12345678,
							"practiceLicenseID":     "PRA-12345678",
							"accreditationID":       "ACR-12345678",
							"accreditationUploadID": "ACR-UPLOAD-12345678",
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
	_, err = RemoveTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}

func TestAddOrganizationRiderKYC(t *testing.T) {
	ctx := context.Background()
	phoneNumber := interserviceclient.TestUserPhoneNumber
	user, err := CreateTestUserByPhone(t, phoneNumber)
	log.Printf("the user is %v", user)
	if err != nil {
		t.Errorf("failed to create a user by phone %v", err)
		return
	}
	idToken := user.Auth.IDToken
	headers, err := CreatedUserGraphQLHeaders(idToken)
	if err != nil {
		t.Errorf("error in getting headers: %w", err)
		return
	}

	authToken, err := firebasetools.ValidateBearerToken(ctx, *idToken)
	if err != nil {
		t.Errorf("invalid token: %w", err)
		return
	}
	authenticatedContext := context.WithValue(ctx, firebasetools.AuthTokenContextKey, authToken)

	err = setPrimaryEmailAddress(authenticatedContext, t, testEmail)
	if err != nil {
		t.Errorf("failed to set primary email address: %v", err)
		return
	}
	dateOfBirth2 := scalarutils.Date{
		Day:   12,
		Year:  1995,
		Month: 10,
	}
	firstName2 := "makmende"
	lastName2 := "juha"

	completeUserDetails := profileutils.BioData{
		DateOfBirth: &dateOfBirth2,
		FirstName:   &firstName2,
		LastName:    &lastName2,
	}
	partnerName := "rider"
	partnerType := profileutils.PartnerTypeRider

	_, err = addPartnerType(authenticatedContext, t, &partnerName, partnerType)
	if err != nil {
		t.Errorf("failed to add partnerType: %v", err)
		return
	}
	_, err = setUpSupplier(authenticatedContext, t, profileutils.AccountTypeOrganisation)
	if err != nil {
		t.Errorf("failed to setup supplier: %v", err)
		return
	}
	err = updateBioData(authenticatedContext, t, completeUserDetails)
	if err != nil {
		t.Errorf("failed to update biodata: %v", err)
		return
	}
	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")

	graphQLMutation := `
	mutation addOrganizationRiderKyc($input:OrganizationRiderInput!){
		addOrganizationRiderKYC(input:$input) {
		  organizationTypeName
		  certificateOfIncorporation
		  certificateOfInCorporationUploadID
		  directorIdentifications{
			identificationDocType
			identificationDocNumber
			identificationDocNumberUploadID
		  }
		  organizationCertificate
		  KRAPIN
		  KRAPINUploadID
		  supportingDocuments{
				supportingDocumentTitle
				supportingDocumentDescription
				supportingDocumentUpload
			}
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
			name: "valid mutation request",
			args: args{
				query: map[string]interface{}{
					"query": graphQLMutation,
					"variables": map[string]interface{}{
						"input": map[string]interface{}{
							"directorIdentifications": []map[string]interface{}{
								{
									"identificationDocType":           "NATIONALID",
									"identificationDocNumber":         "12345678",
									"identificationDocNumberUploadID": "12345678",
								},
							},
							"organizationTypeName":               "LIMITED_COMPANY",
							"certificateOfIncorporation":         "12345",
							"certificateOfInCorporationUploadID": "12345",
							"organizationCertificate":            "12345",
							"KRAPIN":                             "12345",
							"KRAPINUploadID":                     "12345",
							"supportingDocuments": []map[string]interface{}{
								{
									"supportingDocumentTitle":       "title",
									"supportingDocumentDescription": "description",
									"supportingDocumentUpload":      "upload",
								},
							},
						},
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid mutation request - wrong input",
			args: args{
				query: map[string]interface{}{
					"query":     graphQLMutation,
					"variables": map[string]interface{}{},
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
				_, ok := data["errors"]
				if !ok {
					t.Errorf("expected an error")
					return
				}
			}

			if !tt.wantErr {
				_, ok := data["errors"]
				if ok {
					t.Errorf("error not expected got error: %v", data["errors"])
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
	_, err = RemoveTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}

func TestAddIndividualRiderKYC_acceptance(t *testing.T) {
	ctx := context.Background()
	phoneNumber := interserviceclient.TestUserPhoneNumber
	user, err := CreateTestUserByPhone(t, phoneNumber)
	log.Printf("the user is %v", user)
	if err != nil {
		t.Errorf("failed to create a user by phone %v", err)
		return
	}
	idToken := user.Auth.IDToken
	headers, err := CreatedUserGraphQLHeaders(idToken)
	if err != nil {
		t.Errorf("error in getting headers: %w", err)
		return
	}

	authToken, err := firebasetools.ValidateBearerToken(ctx, *idToken)
	if err != nil {
		t.Errorf("invalid token: %w", err)
		return
	}
	authenticatedContext := context.WithValue(ctx, firebasetools.AuthTokenContextKey, authToken)

	err = setPrimaryEmailAddress(authenticatedContext, t, testEmail)
	if err != nil {
		t.Errorf("failed to set primary email address: %v", err)
		return
	}
	dateOfBirth2 := scalarutils.Date{
		Day:   12,
		Year:  1995,
		Month: 10,
	}
	firstName2 := "makmende"
	lastName2 := "juha"

	completeUserDetails := profileutils.BioData{
		DateOfBirth: &dateOfBirth2,
		FirstName:   &firstName2,
		LastName:    &lastName2,
	}
	partnerName := "rider"
	partnerType := profileutils.PartnerTypeRider

	_, err = addPartnerType(authenticatedContext, t, &partnerName, partnerType)
	if err != nil {
		t.Errorf("failed to add partnerType: %v", err)
		return
	}
	_, err = setUpSupplier(authenticatedContext, t, profileutils.AccountTypeIndividual)
	if err != nil {
		t.Errorf("failed to setup supplier: %v", err)
		return
	}
	err = updateBioData(authenticatedContext, t, completeUserDetails)
	if err != nil {
		t.Errorf("failed to update biodata: %v", err)
		return
	}

	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")
	graphqlMutationPayload := `
	mutation addIndividualRiderKYC($input: IndividualRiderInput!){
		addIndividualRiderKYC(input:$input){
		  identificationDoc{
			identificationDocType
			identificationDocNumber
			identificationDocNumberUploadID
		  }
		  KRAPIN
		  KRAPINUploadID
		  drivingLicenseID
		  drivingLicenseUploadID
		  certificateGoodConductUploadID
		  supportingDocuments{
				supportingDocumentTitle
				supportingDocumentDescription
				supportingDocumentUpload
			}
		}
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
			name: "Happy Case - Successfully Add individual rider kyc",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutationPayload,
					"variables": map[string]interface{}{
						"input": map[string]interface{}{
							"identificationDoc": map[string]interface{}{
								"identificationDocType":           "NATIONALID",
								"identificationDocNumber":         "12345678",
								"identificationDocNumberUploadID": "12345678",
							},
							"KRAPIN":                         "12345678",
							"KRAPINUploadID":                 "12345678",
							"drivingLicenseID":               "12345678",
							"certificateGoodConductUploadID": "123456",
							"supportingDocuments": []map[string]interface{}{
								{
									"supportingDocumentTitle":       "title",
									"supportingDocumentDescription": "description",
									"supportingDocumentUpload":      "upload",
								},
							},
						},
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "Sad Case - Add individual rider kyc using invalid payload",
			args: args{
				query: map[string]interface{}{
					"query":     graphqlMutationPayload,
					"variables": map[string]interface{}{},
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
			logrus.Print(resp.StatusCode)
			if tt.wantStatus != resp.StatusCode {
				b, _ := httputil.DumpResponse(resp, true)
				t.Errorf("Bad status response returned; %v ", string(b))
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

func TestFetchKYCProcessingRequests(t *testing.T) {
	// create a user and their profile
	phoneNumber := interserviceclient.TestUserPhoneNumber
	user, err := CreateTestUserByPhone(t, phoneNumber)
	if err != nil {
		t.Errorf("failed to create a user by phone %v", err)
		return
	}

	idToken := user.Auth.IDToken
	headers, err := CreatedUserGraphQLHeaders(idToken)
	if err != nil {
		t.Errorf("error in getting headers: %w", err)
		return
	}

	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")

	graphQLQuery := `
	query fetchKYCProcessingRequests {
		fetchKYCProcessingRequests {
			id
			reqPartnerType
			reqOrganizationType
			reqRaw
			processed
			supplierRecord {
				id
				active
				profileID
				partnerType
				accountType
				supplierKYC
				KYCSubmitted
				underOrganization
				isOrganizationVerified
				partnerSetupComplete
				sladeCode
				parentOrganizationID
				hasBranches
				location {
					id
					name
				}
			}
			status
		}
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
			//NB: empty data because no KYC request has been made
			name: "valid fetch request",
			args: args{
				query: map[string]interface{}{
					"query": graphQLQuery,
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
					t.Errorf("error not expected got error: %v", data["errors"])
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
	_, err = RemoveTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}

func TestSupplierSetDefaultLocation_acceptance(t *testing.T) {
	ctx := context.Background()
	s, err := InitializeTestService(ctx)
	if err != nil {
		t.Errorf("unable to initialize test service: %v", err)
		return
	}

	phoneNumber := interserviceclient.TestUserPhoneNumber
	user, err := CreateTestUserByPhone(t, phoneNumber)
	if err != nil {
		t.Errorf("failed to create a user by phone %v", err)
		return
	}

	idToken := user.Auth.IDToken
	headers, err := CreatedUserGraphQLHeaders(idToken)
	if err != nil {
		t.Errorf("error in getting headers: %w", err)
		return
	}

	authToken, err := firebasetools.ValidateBearerToken(ctx, *idToken)
	if err != nil {
		t.Errorf("invalid token: %w", err)
		return
	}
	authenticatedContext := context.WithValue(ctx, firebasetools.AuthTokenContextKey, authToken)

	name := "Makmende"
	partnerPractitioner := profileutils.PartnerTypePractitioner
	_, err = s.Supplier.AddPartnerType(authenticatedContext, &name, &partnerPractitioner)
	if err != nil {
		t.Errorf("can't create a supplier")
		return
	}

	_, err = s.Supplier.SetUpSupplier(authenticatedContext, profileutils.AccountTypeOrganisation)
	if err != nil {
		t.Errorf("can't set up a supplier")
		return
	}

	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")

	graphqlMutation := `mutation SupplierSetDefaultLocation($input: String!){
		supplierSetDefaultLocation(locationID:$input){
			id
		}
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
			name: "Sad Case - Setup supplier location with an Invalid locationID",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"input": "invalid location ID",
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "Sad Case - Setup supplier location with an empty locationID",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"input": "",
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
			if tt.wantStatus != resp.StatusCode {
				b, _ := httputil.DumpResponse(resp, true)
				t.Errorf("Bad status response returned; %v ", string(b))
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

func TestAddOrganizationCoachKYC(t *testing.T) {
	ctx := context.Background()
	phoneNumber := interserviceclient.TestUserPhoneNumber
	user, err := CreateTestUserByPhone(t, phoneNumber)
	log.Printf("the user is %v", user)
	if err != nil {
		t.Errorf("failed to create a user by phone %v", err)
		return
	}
	idToken := user.Auth.IDToken
	headers, err := CreatedUserGraphQLHeaders(idToken)
	if err != nil {
		t.Errorf("error in getting headers: %w", err)
		return
	}

	authToken, err := firebasetools.ValidateBearerToken(ctx, *idToken)
	if err != nil {
		t.Errorf("invalid token: %w", err)
		return
	}
	authenticatedContext := context.WithValue(ctx, firebasetools.AuthTokenContextKey, authToken)

	err = setPrimaryEmailAddress(authenticatedContext, t, testEmail)
	if err != nil {
		t.Errorf("failed to set primary email address: %v", err)
		return
	}
	dateOfBirth2 := scalarutils.Date{
		Day:   12,
		Year:  1995,
		Month: 10,
	}
	firstName2 := "makmende"
	lastName2 := "juha"

	completeUserDetails := profileutils.BioData{
		DateOfBirth: &dateOfBirth2,
		FirstName:   &firstName2,
		LastName:    &lastName2,
	}
	partnerName := "coach"
	partnerType := profileutils.PartnerTypeCoach

	_, err = addPartnerType(authenticatedContext, t, &partnerName, partnerType)
	if err != nil {
		t.Errorf("failed to add partnerType: %v", err)
		return
	}
	_, err = setUpSupplier(authenticatedContext, t, profileutils.AccountTypeOrganisation)
	if err != nil {
		t.Errorf("failed to setup supplier: %v", err)
		return
	}
	err = updateBioData(authenticatedContext, t, completeUserDetails)
	if err != nil {
		t.Errorf("failed to update biodata: %v", err)
		return
	}

	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")

	graphqlMutation := `
	mutation   addOrganizationCoachKYC($input:OrganizationCoachInput!){
		addOrganizationCoachKYC(input:$input) {
		    organizationTypeName        
            KRAPIN            
            certificateOfIncorporation
            certificateOfInCorporationUploadID       
            organizationCertificate       
            registrationNumber
			practiceLicenseUploadID
			practiceLicenseID
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
			name: "valid mutation request",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"input": map[string]interface{}{
							"directorIdentifications": []map[string]interface{}{
								{
									"identificationDocType":           "NATIONALID",
									"identificationDocNumber":         "12345678",
									"identificationDocNumberUploadID": "12345678",
								},
							},
							"organizationTypeName":               "LIMITED_COMPANY",
							"certificateOfIncorporation":         "CERT-123456",
							"certificateOfInCorporationUploadID": "CERT-UPLOAD-123456",
							"registrationNumber":                 "REG-123456",
							"KRAPIN":                             "KRA-123456789",
							"KRAPINUploadID":                     "KRA-UPLOAD-123456789",
							"practiceLicenseUploadID":            "PRAC-UPLOAD-123456",
							"practiceLicenseID":                  "PRACL",
						},
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid with bogus identification document type",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"input": map[string]interface{}{
							"directorIdentifications": []map[string]interface{}{
								{
									"identificationDocType":           "bogusID",
									"identificationDocNumber":         "12345678",
									"identificationDocNumberUploadID": "12345678",
								},
							},
							"organizationTypeName":               "LIMITED_COMPANY",
							"certificateOfIncorporation":         "CERT-123456",
							"certificateOfInCorporationUploadID": "CERT-UPLOAD-123456",
							"registrationNumber":                 "REG-123456",
							"KRAPIN":                             123456789,
							"KRAPINUploadID":                     "KRA-UPLOAD-123456789",
							"practiceLicenseID":                  "PRAC-123456",
							"practiceLicenseUploadID":            "PRAC-123456",
						},
					},
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
	_, err = RemoveTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}

func TestAddIndividualNutritionKYC(t *testing.T) {
	ctx := context.Background()
	phoneNumber := interserviceclient.TestUserPhoneNumber
	user, err := CreateTestUserByPhone(t, phoneNumber)
	log.Printf("the user is %v", user)
	if err != nil {
		t.Errorf("failed to create a user by phone %v", err)
		return
	}
	idToken := user.Auth.IDToken
	headers, err := CreatedUserGraphQLHeaders(idToken)
	if err != nil {
		t.Errorf("error in getting headers: %w", err)
		return
	}

	authToken, err := firebasetools.ValidateBearerToken(ctx, *idToken)
	if err != nil {
		t.Errorf("invalid token: %w", err)
		return
	}
	authenticatedContext := context.WithValue(ctx, firebasetools.AuthTokenContextKey, authToken)

	err = setPrimaryEmailAddress(authenticatedContext, t, testEmail)
	if err != nil {
		t.Errorf("failed to set primary email address: %v", err)
		return
	}
	dateOfBirth2 := scalarutils.Date{
		Day:   12,
		Year:  1995,
		Month: 10,
	}
	firstName2 := "makmende"
	lastName2 := "juha"

	completeUserDetails := profileutils.BioData{
		DateOfBirth: &dateOfBirth2,
		FirstName:   &firstName2,
		LastName:    &lastName2,
	}
	partnerName := "nutrition"
	partnerType := profileutils.PartnerTypeNutrition

	_, err = addPartnerType(authenticatedContext, t, &partnerName, partnerType)
	if err != nil {
		t.Errorf("failed to add partnerType: %v", err)
		return
	}
	_, err = setUpSupplier(authenticatedContext, t, profileutils.AccountTypeIndividual)
	if err != nil {
		t.Errorf("failed to setup supplier: %v", err)
		return
	}
	err = updateBioData(authenticatedContext, t, completeUserDetails)
	if err != nil {
		t.Errorf("failed to update biodata: %v", err)
		return
	}

	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")

	graphqlMutation := `
	mutation   addIndividualNutritionKYC($input:IndividualNutritionInput!){
		addIndividualNutritionKYC(input:$input) {    
			identificationDoc {
				identificationDocType
				identificationDocNumber
				identificationDocNumberUploadID
		  	}
			KRAPIN
			KRAPINUploadID			           
			practiceLicenseUploadID
			practiceLicenseID
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
			name: "valid mutation request",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"input": map[string]interface{}{
							"identificationDoc": map[string]interface{}{
								"identificationDocType":           "NATIONALID",
								"identificationDocNumber":         "12345",
								"identificationDocNumberUploadID": "12345",
							},
							"KRAPIN":                  "KRA-123456789",
							"KRAPINUploadID":          "KRA-UPLOAD-123456789",
							"practiceLicenseUploadID": "PRAC-UPLOAD-123456",
							"practiceLicenseID":       "PRACL",
						},
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid with bogus identification document type",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"input": map[string]interface{}{
							"identificationDoc": map[string]interface{}{
								"identificationDocType":           "bogusType",
								"identificationDocNumber":         "12345",
								"identificationDocNumberUploadID": "12345",
							},
							"KRAPIN":                  123456789,
							"KRAPINUploadID":          "KRA-UPLOAD-123456789",
							"practiceLicenseID":       "PRAC-123456",
							"practiceLicenseUploadID": "PRAC-123456",
						},
					},
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
	_, err = RemoveTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}

func TestAddOrganizationNutritionKyc(t *testing.T) {
	ctx := context.Background()
	phoneNumber := interserviceclient.TestUserPhoneNumber
	user, err := CreateTestUserByPhone(t, phoneNumber)
	log.Printf("the user is %v", user)
	if err != nil {
		t.Errorf("failed to create a user by phone %v", err)
		return
	}
	idToken := user.Auth.IDToken
	headers, err := CreatedUserGraphQLHeaders(idToken)
	if err != nil {
		t.Errorf("error in getting headers: %w", err)
		return
	}

	authToken, err := firebasetools.ValidateBearerToken(ctx, *idToken)
	if err != nil {
		t.Errorf("invalid token: %w", err)
		return
	}
	authenticatedContext := context.WithValue(ctx, firebasetools.AuthTokenContextKey, authToken)

	err = setPrimaryEmailAddress(authenticatedContext, t, testEmail)
	if err != nil {
		t.Errorf("failed to set primary email address: %v", err)
		return
	}
	dateOfBirth2 := scalarutils.Date{
		Day:   12,
		Year:  1995,
		Month: 10,
	}
	firstName2 := "makmende"
	lastName2 := "juha"

	completeUserDetails := profileutils.BioData{
		DateOfBirth: &dateOfBirth2,
		FirstName:   &firstName2,
		LastName:    &lastName2,
	}
	partnerName := "nutrition"
	partnerType := profileutils.PartnerTypeNutrition

	_, err = addPartnerType(authenticatedContext, t, &partnerName, partnerType)
	if err != nil {
		t.Errorf("failed to add partnerType: %v", err)
		return
	}
	_, err = setUpSupplier(authenticatedContext, t, profileutils.AccountTypeOrganisation)
	if err != nil {
		t.Errorf("failed to setup supplier: %v", err)
		return
	}
	err = updateBioData(authenticatedContext, t, completeUserDetails)
	if err != nil {
		t.Errorf("failed to update biodata: %v", err)
		return
	}

	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")

	graphqlMutation := `
	mutation   addOrganizationNutritionKYC($input:OrganizationNutritionInput!){
		addOrganizationNutritionKYC(input:$input) {    
			organizationTypeName
			KRAPIN
			KRAPINUploadID		
			practiceLicenseID
			registrationNumber
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
			name: "valid mutation request",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"input": map[string]interface{}{
							"organizationTypeName": "LIMITED_COMPANY",
							"KRAPIN":               "KRA-123456789",
							"KRAPINUploadID":       "KRA-UPLOAD-123456789",
							"practiceLicenseID":    "PRACL",
							"registrationNumber":   "10222",
						},
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
	_, err = RemoveTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}

func TestSetupAsExperimentParticipant(t *testing.T) {
	// create a user and their profile
	phoneNumber := interserviceclient.TestUserPhoneNumber
	user, err := CreateTestUserByPhone(t, phoneNumber)
	if err != nil {
		t.Errorf("failed to create a user by phone %v", err)
		return
	}

	idToken := user.Auth.IDToken
	headers, err := CreatedUserGraphQLHeaders(idToken)
	if err != nil {
		t.Errorf("error in getting headers: %w", err)
		return
	}

	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")

	graphqlMutation := `
	mutation setupAsExperimentParticipant($participate:Boolean!){
		setupAsExperimentParticipant(participate:$participate)
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
			name: "valid mutation request",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutation,
					"variables": map[string]interface{}{
						"participate": true,
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
	_, err = RemoveTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
	if err != nil {
		t.Errorf("unable to remove test user: %s", err)
	}
}

func TestAddOrganizationPractitionerKyc(t *testing.T) {
	ctx := context.Background()
	phoneNumber := interserviceclient.TestUserPhoneNumber
	user, err := CreateTestUserByPhone(t, phoneNumber)
	log.Printf("the user is %v", user)
	if err != nil {
		t.Errorf("failed to create a user by phone %v", err)
		return
	}
	idToken := user.Auth.IDToken
	headers, err := CreatedUserGraphQLHeaders(idToken)
	if err != nil {
		t.Errorf("error in getting headers: %w", err)
		return
	}

	authToken, err := firebasetools.ValidateBearerToken(ctx, *idToken)
	if err != nil {
		t.Errorf("invalid token: %w", err)
		return
	}
	authenticatedContext := context.WithValue(ctx, firebasetools.AuthTokenContextKey, authToken)

	err = setPrimaryEmailAddress(authenticatedContext, t, testEmail)
	if err != nil {
		t.Errorf("failed to set primary email address: %v", err)
		return
	}
	dateOfBirth2 := scalarutils.Date{
		Day:   12,
		Year:  1995,
		Month: 10,
	}
	firstName2 := "makmende"
	lastName2 := "juha"

	completeUserDetails := profileutils.BioData{
		DateOfBirth: &dateOfBirth2,
		FirstName:   &firstName2,
		LastName:    &lastName2,
	}
	partnerName := "nutrition"
	partnerType := profileutils.PartnerTypePractitioner

	_, err = addPartnerType(authenticatedContext, t, &partnerName, partnerType)
	if err != nil {
		t.Errorf("failed to add partnerType: %v", err)
		return
	}
	_, err = setUpSupplier(authenticatedContext, t, profileutils.AccountTypeOrganisation)
	if err != nil {
		t.Errorf("failed to setup supplier: %v", err)
		return
	}
	err = updateBioData(authenticatedContext, t, completeUserDetails)
	if err != nil {
		t.Errorf("failed to update biodata: %v", err)
		return
	}
	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")
	graphqlMutationPayload := `mutation AddOrganizationPractitionerKyc(
		$input: OrganizationPractitionerInput!
	  ) {
		addOrganizationPractitionerKYC(input: $input) {
		  organizationTypeName
		  KRAPIN
		  KRAPINUploadID		  
		  certificateOfIncorporation
		  certificateOfInCorporationUploadID
		  directorIdentifications {
			identificationDocType
			identificationDocNumber
			identificationDocNumberUploadID
		  }
		  organizationCertificate
		  registrationNumber
		  practiceLicenseUploadID
		  practiceServices
		  cadre
		}
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
			name: "Happy Case - Successfully Add organization practitioner kyc",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutationPayload,
					"variables": map[string]interface{}{
						"input": map[string]interface{}{
							"directorIdentifications": []map[string]interface{}{
								{
									"identificationDocType":           "NATIONALID",
									"identificationDocNumber":         "12345678",
									"identificationDocNumberUploadID": "12345678",
								},
							},
							"organizationTypeName":               "LIMITED_COMPANY",
							"KRAPIN":                             "12345678",
							"KRAPINUploadID":                     "12345678",
							"certificateOfIncorporation":         "12345",
							"certificateOfInCorporationUploadID": "12345",
							"organizationCertificate":            "12345",
							"registrationNumber":                 "REG-123",
							"practiceLicenseUploadID":            "UPLOAD-123456",
							"practiceLicenseID":                  "1289",
							"practiceServices":                   []string{"OUTPATIENT_SERVICES"},
							"cadre":                              "DOCTOR",
						},
					},
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "Sad Case - Use invalid input - missing KRA",
			args: args{
				query: map[string]interface{}{
					"query": graphqlMutationPayload,
					"variables": map[string]interface{}{
						"input": map[string]interface{}{
							"directorIdentifications": []map[string]interface{}{
								{
									"identificationDocType":           "NATIONALID",
									"identificationDocNumber":         "12345678",
									"identificationDocNumberUploadID": "12345678",
								},
							},
							"organizationTypeName":               "LIMITED_COMPANY",
							"certificateOfIncorporation":         12345,
							"certificateOfInCorporationUploadID": 12345,
							"organizationCertificate":            12345,
							"registrationNumber":                 "REG-123",
							"practiceLicenseUploadID":            "UPLOAD-123456",
							"practiceLicenseID":                  "1289",
							"practiceServices":                   []string{"OUTPATIENT_SERVICES"},
							"cadre":                              "DOCTOR",
						},
					},
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
			if tt.wantStatus != resp.StatusCode {
				b, _ := httputil.DumpResponse(resp, true)
				t.Errorf("Bad status response returned; %v ", string(b))
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
