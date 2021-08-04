package tests

// func TestResumeWithPin(t *testing.T) {
// 	// create a user and their profile
// 	phoneNumber := interserviceclient.TestUserPhoneNumber
// 	user, err := CreateTestUserByPhone(t, phoneNumber)
// 	if err != nil {
// 		t.Errorf("failed to create a user by phone %v", err)
// 		return
// 	}

// 	idToken := user.Auth.IDToken
// 	headers, err := CreatedUserGraphQLHeaders(idToken)
// 	if err != nil {
// 		t.Errorf("error in getting headers: %w", err)
// 		return
// 	}

// 	graphQLURL := fmt.Sprintf("%s/%s", baseURL, "graphql")

// 	graphqlMutation := `
//     query resumeWithPin($pin:String!){
// 		resumeWithPIN(pin:$pin)
// 	}`

// 	type args struct {
// 		query map[string]interface{}
// 	}

// 	tests := []struct {
// 		name       string
// 		args       args
// 		wantStatus int
// 		wantErr    bool
// 	}{
// 		{
// 			name: "resume with pin successfully",
// 			args: args{
// 				query: map[string]interface{}{
// 					"query": graphqlMutation,
// 					"variables": map[string]interface{}{
// 						"pin": "2030",
// 					},
// 				},
// 			},
// 			wantStatus: http.StatusOK,
// 			wantErr:    false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {

// 			body, err := mapToJSONReader(tt.args.query)
// 			if err != nil {
// 				t.Errorf("unable to get GQL JSON io Reader: %s", err)
// 				return
// 			}

// 			r, err := http.NewRequest(
// 				http.MethodPost,
// 				graphQLURL,
// 				body,
// 			)

// 			if err != nil {
// 				t.Errorf("unable to compose request: %s", err)
// 				return
// 			}

// 			if r == nil {
// 				t.Errorf("nil request")
// 				return
// 			}

// 			for k, v := range headers {
// 				r.Header.Add(k, v)
// 			}
// 			client := http.Client{
// 				Timeout: time.Second * testHTTPClientTimeout,
// 			}
// 			resp, err := client.Do(r)
// 			if err != nil {
// 				t.Errorf("request error: %s", err)
// 				return
// 			}

// 			dataResponse, err := ioutil.ReadAll(resp.Body)
// 			if err != nil {
// 				t.Errorf("can't read request body: %s", err)
// 				return
// 			}
// 			if dataResponse == nil {
// 				t.Errorf("nil response data")
// 				return
// 			}

// 			data := map[string]interface{}{}
// 			err = json.Unmarshal(dataResponse, &data)
// 			if err != nil {
// 				t.Errorf("bad data returned")
// 				return
// 			}

// 			if tt.wantErr {
// 				_, ok := data["errors"]
// 				if !ok {
// 					t.Errorf("expected an error")
// 					return
// 				}
// 			}

// 			if !tt.wantErr {
// 				_, ok := data["errors"]
// 				if ok {
// 					t.Errorf("error not expected got error: %w", data["errors"])
// 					return
// 				}
// 			}
// 			if tt.wantStatus != resp.StatusCode {
// 				b, _ := httputil.DumpResponse(resp, true)
// 				t.Errorf("Bad status response returned; %v ", string(b))
// 				return
// 			}
// 		})
// 	}

// 	// perform tear down; remove user
// 	_, err = RemoveTestUserByPhone(t, interserviceclient.TestUserPhoneNumber)
// 	if err != nil {
// 		t.Errorf("unable to remove test user: %s", err)
// 	}
// }
