package profile

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"gitlab.slade360emr.com/go/base"
)

func TestValidateEmail(t *testing.T) {
	fc := &base.FirebaseClient{}
	firebaseApp, err := fc.InitFirebase()
	assert.Nil(t, err)

	ctx := base.GetAuthenticatedContext(t)
	firestoreClient, err := firebaseApp.Firestore(ctx)
	assert.Nil(t, err)

	validOtpCode := rand.Int()
	validOtpData := map[string]interface{}{
		"authorizationCode": strconv.Itoa(validOtpCode),
		"isValid":           true,
		"message":           "Testing email OTP message",
		"timestamp":         time.Now(),
		"email":             "ngure.nyaga@healthcloud.co.ke",
	}
	_, err = base.SaveDataToFirestore(firestoreClient, base.SuffixCollection(base.OTPCollectionName), validOtpData)
	assert.Nil(t, err)

	invalidOtpCode := rand.Int()
	invalidOtpData := map[string]interface{}{
		"authorizationCode": strconv.Itoa(invalidOtpCode),
		"isValid":           false,
		"message":           "testing OTP message",
		"email":             "ngure.nyaga@healthcloud.co.ke",
		"timestamp":         time.Now(),
	}
	_, err = base.SaveDataToFirestore(firestoreClient, base.SuffixCollection(base.OTPCollectionName), invalidOtpData)
	assert.Nil(t, err)

	type args struct {
		email            string
		verificationCode string
		firestoreClient  *firestore.Client
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "invalid email",
			args: args{
				email: "not a valid email",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "valid email",
			args: args{
				email:            "ngure.nyaga@healthcloud.co.ke",
				verificationCode: strconv.Itoa(validOtpCode),
				firestoreClient:  firestoreClient,
			},
			want:    "ngure.nyaga@healthcloud.co.ke",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateEmail(tt.args.email, tt.args.verificationCode, tt.args.firestoreClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateMsisdn(t *testing.T) {
	goodData := &PinRecovery{
		MSISDN: "+254712789456",
	}
	goodDataJSONBytes, err := json.Marshal(goodData)
	assert.Nil(t, err)
	assert.NotNil(t, goodDataJSONBytes)

	validRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	validRequest.Body = ioutil.NopCloser(bytes.NewReader(goodDataJSONBytes))

	emptyDataRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	emptyDataRequest.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))

	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    *PinRecovery
		wantErr bool
	}{
		{
			name: "valid data",
			args: args{
				w: httptest.NewRecorder(),
				r: validRequest,
			},
			want: &PinRecovery{
				MSISDN: "+254712789456",
			},
			wantErr: false,
		},
		{
			name: "invalid data",
			args: args{
				w: httptest.NewRecorder(),
				r: emptyDataRequest,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateMsisdn(tt.args.w, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMsisdn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateMsisdn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateUpdatePinPayload(t *testing.T) {
	goodData := &PinRecovery{
		MSISDN: "+254712789456",
		PIN:    "1234",
		OTP:    "123456",
	}
	goodDataJSONBytes, err := json.Marshal(goodData)
	assert.Nil(t, err)
	assert.NotNil(t, goodDataJSONBytes)

	validRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	validRequest.Body = ioutil.NopCloser(bytes.NewReader(goodDataJSONBytes))

	emptyDataRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	emptyDataRequest.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))

	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    *PinRecovery
		wantErr bool
	}{
		{
			name: "valid data",
			args: args{
				w: httptest.NewRecorder(),
				r: validRequest,
			},
			want: &PinRecovery{
				MSISDN: "+254712789456",
				PIN:    "1234",
				OTP:    "123456",
			},
			wantErr: false,
		},
		{
			name: "invalid data",
			args: args{
				w: httptest.NewRecorder(),
				r: emptyDataRequest,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateUpdatePinPayload(tt.args.w, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdatePinPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateUpdatePinPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}