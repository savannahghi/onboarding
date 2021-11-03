package utils_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidateSignUpInput(t *testing.T) {
	phone := interserviceclient.TestUserPhoneNumber
	pin := interserviceclient.TestUserPin
	flavour := feedlib.FlavourConsumer
	otp := "12345"

	alphanumericPhone := "+254-not-valid-123"
	badPhone := "+254712"
	shortPin := "123"
	longPin := "1234567"
	alphabeticalPin := "abcd"

	type args struct {
		input *dto.SignUpInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success: return a valid output",
			args: args{
				input: &dto.SignUpInput{
					PhoneNumber: &phone,
					PIN:         &pin,
					Flavour:     flavour,
					OTP:         &otp,
				},
			},
			wantErr: false,
		},
		{
			name: "failure: bad phone number provided",
			args: args{
				input: &dto.SignUpInput{
					PhoneNumber: &badPhone,
					PIN:         &pin,
					Flavour:     flavour,
					OTP:         &otp,
				},
			},
			wantErr: true,
		},
		{
			name: "failure: alphanumeric phone number provided",
			args: args{
				input: &dto.SignUpInput{
					PhoneNumber: &alphanumericPhone,
					PIN:         &pin,
					Flavour:     flavour,
					OTP:         &otp,
				},
			},
			wantErr: true,
		},
		{
			name: "failure: short pin number provided",
			args: args{
				input: &dto.SignUpInput{
					PhoneNumber: &phone,
					PIN:         &shortPin,
					Flavour:     flavour,
					OTP:         &otp,
				},
			},
			wantErr: true,
		},
		{
			name: "failure: long pin number provided",
			args: args{
				input: &dto.SignUpInput{
					PhoneNumber: &phone,
					PIN:         &longPin,
					Flavour:     flavour,
					OTP:         &otp,
				},
			},
			wantErr: true,
		},
		{
			name: "failure: alphabetical pin number provided",
			args: args{
				input: &dto.SignUpInput{
					PhoneNumber: &phone,
					PIN:         &alphabeticalPin,
					Flavour:     flavour,
					OTP:         &otp,
				},
			},
			wantErr: true,
		},
		{
			name: "failure: bad flavour provided",
			args: args{
				input: &dto.SignUpInput{
					PhoneNumber: &phone,
					PIN:         &pin,
					Flavour:     "not-a-flavour",
					OTP:         &otp,
				},
			},
			wantErr: true,
		},
		{
			name: "failure: no OTP provided",
			args: args{
				input: &dto.SignUpInput{
					PhoneNumber: &phone,
					PIN:         &pin,
					Flavour:     flavour,
					OTP:         nil,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validInput, err := utils.ValidateSignUpInput(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSignUpInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && validInput != nil {
				t.Errorf("expected a nil valid input since an error :%v occurred", err)
			}

			if err == nil && validInput == nil {
				t.Errorf("expected a valid input %v since no error occurred", validInput)
			}
		})
	}
}

func TestValidateUID(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid",
			args: map[string]interface{}{
				"uid": uuid.New().String(),
			},
			wantErr: false,
		},
		{
			name: "invalid",
			args: map[string]interface{}{
				"uuid": uuid.New().String(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.args)
			if err != nil {
				t.Errorf("failed to marshal body: %v", err)
				return
			}
			// Create a request to pass to our handler.
			req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewBuffer(body))
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}
			rw := httptest.NewRecorder()
			resp, err := utils.ValidateUID(rw, req)
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Nil(t, resp)
			}
			if !tt.wantErr {
				assert.Nil(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.UID)
			}

		})
	}
}
