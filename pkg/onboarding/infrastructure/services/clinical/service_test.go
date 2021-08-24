package clinical

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	extMock "github.com/savannahghi/onboarding/pkg/onboarding/application/extension/mock"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/clinical/mock"
)

var fakeISCExt extMock.ISCClientExtension
var clinicalClient extension.ISCClientExtension = &fakeISCExt
var fakeClinic mock.FakeServiceClinical

func TestServiceClinicalImpl_RegisterPatient(t *testing.T) {
	ctx := context.Background()
	s := &ServiceClinicalImpl{
		iscExt: clinicalClient,
	}

	ID := uuid.New().String()
	phoneNumber := interserviceclient.TestUserPhoneNumber
	input := domain.SimplePatientRegistrationInput{
		PhoneNumbers: []*domain.PhoneNumberInput{
			{
				Msisdn: phoneNumber,
			},
		},
	}

	json := fmt.Sprintf(`{"ID": "%s"}`, ID)

	type args struct {
		ctx     context.Context
		payload domain.SimplePatientRegistrationInput
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
		wantErr bool
	}{
		{
			name: "sad: unable to make request",
			args: args{
				ctx:     ctx,
				payload: input,
			},
			wantNil: true,
			wantErr: true,
		},
		{
			name: "sad: failed with a status code",
			args: args{
				ctx:     ctx,
				payload: input,
			},
			wantNil: true,
			wantErr: true,
		},
		{
			name: "happy: patient already registered",
			args: args{
				ctx:     ctx,
				payload: input,
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "happy: registered patient",
			args: args{
				ctx:     ctx,
				payload: input,
			},
			wantNil: false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "sad: unable to make request" {
				fakeClinic.CheckIfPatientExistsFn = func(ctx context.Context, phoneNumber string) (*domain.SimplePatientRegistrationInput, bool, error) {
					return &domain.SimplePatientRegistrationInput{}, true, nil
				}
				fakeISCExt.MakeRequestFn = func(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
					return nil, fmt.Errorf("error unable to make request")
				}
			}
			if tt.name == "sad: failed with a status code" {
				fakeClinic.CheckIfPatientExistsFn = func(ctx context.Context, phoneNumber string) (*domain.SimplePatientRegistrationInput, bool, error) {
					return &domain.SimplePatientRegistrationInput{}, true, nil
				}
				fakeISCExt.MakeRequestFn = func(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
					return nil, fmt.Errorf("error unable to make request")
				}
				fakeISCExt.MakeRequestFn = func(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusBadRequest}, nil
				}
			}
			if tt.name == "happy: patient already registered" {
				fakeClinic.CheckIfPatientExistsFn = func(ctx context.Context, phoneNumber string) (*domain.SimplePatientRegistrationInput, bool, error) {
					return &domain.SimplePatientRegistrationInput{}, true, nil
				}
				fakeISCExt.MakeRequestFn = func(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
					return nil, fmt.Errorf("error unable to make request")
				}
				fakeISCExt.MakeRequestFn = func(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
					r := io.NopCloser(bytes.NewReader([]byte(json)))
					return &http.Response{
						Body:       r,
						StatusCode: http.StatusOK}, nil
				}
			}
			if tt.name == "happy: registered patient" {
				fakeClinic.CheckIfPatientExistsFn = func(ctx context.Context, phoneNumber string) (*domain.SimplePatientRegistrationInput, bool, error) {
					return &domain.SimplePatientRegistrationInput{}, true, nil
				}
				fakeISCExt.MakeRequestFn = func(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
					return nil, fmt.Errorf("error unable to make request")
				}
				fakeISCExt.MakeRequestFn = func(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
					r := io.NopCloser(bytes.NewReader([]byte(json)))
					return &http.Response{
						Body:       r,
						StatusCode: http.StatusOK}, nil
				}
			}
			got, err := s.RegisterPatient(tt.args.ctx, tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServiceClinicalImpl.RegisterPatient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && tt.wantNil {
				t.Errorf("ServiceClinicalImpl.RegisterPatient() = %v, want %v", got, tt.wantNil)
			}
		})
	}
}

func TestServiceClinicalImpl_CheckIfPatientExists(t *testing.T) {
	ctx := context.Background()
	s := &ServiceClinicalImpl{
		iscExt: clinicalClient,
	}

	phone := interserviceclient.TestUserPhoneNumber

	ID := uuid.New().String()

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
			name: "sad: unable to make request",
			args: args{
				ctx:         ctx,
				phoneNumber: phone,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "sad: failed with a status code",
			args: args{
				ctx:         ctx,
				phoneNumber: phone,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "sad: patient not found",
			args: args{
				ctx:         ctx,
				phoneNumber: phone,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "happy: got patient",
			args: args{
				ctx:         ctx,
				phoneNumber: phone,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "sad: unable to make request" {
				fakeISCExt.MakeRequestFn = func(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
					return nil, fmt.Errorf("error unable to make request")
				}
			}
			if tt.name == "sad: failed with a status code" {
				fakeISCExt.MakeRequestFn = func(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusBadRequest}, nil
				}
			}
			if tt.name == "sad: patient not found" {
				fakeISCExt.MakeRequestFn = func(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
					json := `{"found": false}`
					r := io.NopCloser(bytes.NewReader([]byte(json)))
					return &http.Response{
						Body:       r,
						StatusCode: http.StatusOK}, nil
				}
			}
			if tt.name == "happy: got patient" {
				fakeISCExt.MakeRequestFn = func(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
					json := fmt.Sprintf(`{"patient": {"ID": "%s"}, "found": true}`, ID)
					r := io.NopCloser(bytes.NewReader([]byte(json)))
					return &http.Response{
						Body:       r,
						StatusCode: http.StatusOK}, nil
				}
			}

			_, got, err := s.CheckIfPatientExists(tt.args.ctx, tt.args.phoneNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServiceClinicalImpl.CheckIfPatientExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServiceClinicalImpl.CheckIfPatientExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
