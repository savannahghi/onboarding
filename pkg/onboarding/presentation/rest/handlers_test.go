package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/profileutils"
	erp "gitlab.slade360emr.com/go/commontools/accounting/pkg/usecases"
	erpMock "gitlab.slade360emr.com/go/commontools/accounting/pkg/usecases/mock"
	crmDomain "gitlab.slade360emr.com/go/commontools/crm/pkg/domain"
	"gitlab.slade360emr.com/go/commontools/crm/pkg/infrastructure/services/hubspot"

	extMock "github.com/savannahghi/onboarding/pkg/onboarding/application/extension/mock"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/chargemaster"
	chargemasterMock "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/chargemaster/mock"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/edi"
	ediMock "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/edi/mock"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/engagement"
	engagementMock "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/engagement/mock"

	crmExt "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/crm"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/messaging"
	messagingMock "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/messaging/mock"
	pubsubmessaging "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/pubsub"
	pubsubmessagingMock "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/pubsub/mock"
	"github.com/savannahghi/onboarding/pkg/onboarding/presentation/interactor"
	"github.com/savannahghi/onboarding/pkg/onboarding/presentation/rest"
	"github.com/savannahghi/onboarding/pkg/onboarding/repository"
	mockRepo "github.com/savannahghi/onboarding/pkg/onboarding/repository/mock"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases"
	adminSrv "github.com/savannahghi/onboarding/pkg/onboarding/usecases/admin"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases/ussd"
	hubspotRepo "gitlab.slade360emr.com/go/commontools/crm/pkg/infrastructure/database/fs"
	hubspotUsecases "gitlab.slade360emr.com/go/commontools/crm/pkg/usecases"
)

var fakeRepo mockRepo.FakeOnboardingRepository
var fakeEngagementSvs engagementMock.FakeServiceEngagement
var fakeBaseExt extMock.FakeBaseExtensionImpl
var fakePinExt extMock.PINExtensionImpl
var serverUrl = "http://localhost:5000"
var fakePubSub pubsubmessagingMock.FakeServicePubSub
var fakeEDISvc ediMock.FakeServiceEDI

// InitializeFakeOnboardingInteractor represents a fakeonboarding interactor
func InitializeFakeOnboardingInteractor() (*interactor.Interactor, error) {
	var r repository.OnboardingRepository = &fakeRepo
	var erpSvc erp.AccountingUsecase = &erpMock.FakeServiceCommonTools{}
	var chargemasterSvc chargemaster.ServiceChargeMaster = &chargemasterMock.FakeServiceChargeMaster{}
	var engagementSvc engagement.ServiceEngagement = &fakeEngagementSvs
	var messagingSvc messaging.ServiceMessaging = &messagingMock.FakeServiceMessaging{}
	var ext extension.BaseExtension = &fakeBaseExt
	var pinExt extension.PINExtension = &fakePinExt
	var ps pubsubmessaging.ServicePubSub = &fakePubSub
	var ediSvc edi.ServiceEdi = &fakeEDISvc

	// hubspot usecases
	hubspotService := hubspot.NewHubSpotService()
	hubspotfr, err := hubspotRepo.NewHubSpotFirebaseRepository(context.Background(), hubspotService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize hubspot crm repository: %w", err)
	}
	hubspotUsecases := hubspotUsecases.NewHubSpotUsecases(hubspotfr)
	crmExt := crmExt.NewCrmService(hubspotUsecases)
	profile := usecases.NewProfileUseCase(r, ext, engagementSvc, ps, crmExt)
	login := usecases.NewLoginUseCases(r, profile, ext, pinExt)
	survey := usecases.NewSurveyUseCases(r, ext)
	supplier := usecases.NewSupplierUseCases(
		r, profile, erpSvc, chargemasterSvc, engagementSvc, messagingSvc, ext, ps,
	)
	userpin := usecases.NewUserPinUseCase(r, profile, ext, pinExt, engagementSvc)
	su := usecases.NewSignUpUseCases(r, profile, userpin, supplier, ext, engagementSvc, ps, ediSvc)
	nhif := usecases.NewNHIFUseCases(r, profile, ext, engagementSvc)
	sms := usecases.NewSMSUsecase(r, ext)
	admin := usecases.NewAdminUseCases(r, engagementSvc, ext, userpin)
	agent := usecases.NewAgentUseCases(r, engagementSvc, ext, userpin)

	aitUssd := ussd.NewUssdUsecases(r, ext, profile, userpin, su, pinExt, ps, crmExt)
	adminSrv := adminSrv.NewService(ext)

	role := usecases.NewRoleUseCases(r, ext)

	i, err := interactor.NewOnboardingInteractor(
		r, profile, su, supplier, login,
		survey, userpin, erpSvc, chargemasterSvc,
		engagementSvc, messagingSvc, nhif, ps, sms,
		aitUssd, agent, admin, ediSvc, adminSrv, crmExt,
		role,
	)
	if err != nil {
		return nil, fmt.Errorf("can't instantiate service : %w", err)
	}
	return i, nil

}

func composeValidPhonePayload(t *testing.T, phone string) *bytes.Buffer {
	phoneNumber := struct {
		PhoneNumber string
	}{
		PhoneNumber: phone,
	}
	bs, err := json.Marshal(phoneNumber)
	if err != nil {
		t.Errorf("unable to marshal phoneNumber to JSON: %s", err)
	}
	return bytes.NewBuffer(bs)
}

func composeValidRolePayload(t *testing.T, phone string, role profileutils.RoleType) *bytes.Buffer {
	payload := &dto.RolePayload{
		PhoneNumber: &phone,
		Role:        &role,
	}
	bs, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("unable to marshal token string to JSON: %s", err)
	}
	return bytes.NewBuffer(bs)
}

func composeSignupPayload(
	t *testing.T,
	phone, pin, otp string,
	flavour feedlib.Flavour,
) *bytes.Buffer {
	payload := dto.SignUpInput{
		PhoneNumber: &phone,
		PIN:         &pin,
		Flavour:     flavour,
		OTP:         &otp,
	}

	bs, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	return bytes.NewBuffer(bs)
}

func composeChangePinPayload(t *testing.T, phone, pin, otp string) *bytes.Buffer {
	payload := domain.ChangePINRequest{
		PhoneNumber: phone,
		PIN:         pin,
		OTP:         otp,
	}

	bs, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	return bytes.NewBuffer(bs)
}

func composeRefreshTokenPayload(t *testing.T, token *string) *bytes.Buffer {
	refreshToken := &dto.RefreshTokenPayload{RefreshToken: token}
	bs, err := json.Marshal(refreshToken)
	if err != nil {
		t.Errorf("unable to marshal token string to JSON: %s", err)
	}
	return bytes.NewBuffer(bs)
}

func composeUIDPayload(t *testing.T, uid *string) *bytes.Buffer {
	uidPayload := &dto.UIDPayload{UID: uid}
	bs, err := json.Marshal(uidPayload)
	if err != nil {
		t.Errorf("unable to marshal token string to JSON: %s", err)
	}
	return bytes.NewBuffer(bs)
}

func composePushTokenPayload(t *testing.T, UID, token string) *bytes.Buffer {
	payload := &dto.PushTokenPayload{
		PushToken: token,
		UID:       UID,
	}
	bs, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("unable to marshal token string to JSON: %s", err)
	}
	return bytes.NewBuffer(bs)
}

func composeLoginPayload(t *testing.T, phone, pin string, flavour feedlib.Flavour) *bytes.Buffer {
	payload := dto.LoginPayload{
		PhoneNumber: &phone,
		PIN:         &pin,
		Flavour:     flavour,
	}

	bs, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("unable to marshal test item to JSON: %s", err)
	}
	return bytes.NewBuffer(bs)
}

func composeSendRetryOTPPayload(t *testing.T, phone string, retryStep int) *bytes.Buffer {
	payload := dto.SendRetryOTPPayload{
		Phone:     &phone,
		RetryStep: &retryStep,
	}

	bs, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("unable to marshal payload to JSON: %s", err)
		return nil
	}
	return bytes.NewBuffer(bs)
}

func composeCoversUpdatePayload(t *testing.T, payload *dto.UpdateCoversPayload) *bytes.Buffer {

	bs, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("unable to marshal payload to JSON: %s", err)
		return nil
	}
	return bytes.NewBuffer(bs)
}

func composeSetPrimaryPhoneNumberPayload(t *testing.T, phone, otp string) *bytes.Buffer {
	payload := dto.SetPrimaryPhoneNumberPayload{
		PhoneNumber: &phone,
		OTP:         &otp,
	}
	bs, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("unable to marshal payload to JSON: %s", err)
		return nil
	}
	return bytes.NewBuffer(bs)
}

func composeSMSMessageDataPayload(
	t *testing.T,
	payload *dto.AfricasTalkingMessage,
) *strings.Reader {
	data := url.Values{}
	data.Set("date", payload.Date)
	data.Set("from", payload.From)
	data.Set("id", payload.ID)
	data.Set("linkId", payload.LinkID)
	data.Set("text", payload.Text)
	data.Set("to", payload.To)

	smspayload := strings.NewReader(data.Encode())
	return smspayload
}

func composeUssdPayload(t *testing.T, payload *dto.SessionDetails) *strings.Reader {
	data := url.Values{}
	data.Set("phoneNumber", *payload.PhoneNumber)
	data.Set("sessionId", payload.SessionID)
	data.Set("text", payload.Text)
	data.Set("level", strconv.Itoa(payload.Level))

	ussdPayload := strings.NewReader(data.Encode())
	return ussdPayload
}

func TestHandlersInterfacesImpl_VerifySignUpPhoneNumber(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)
	// valid:_successfully_verifies_a_phone_number
	payload := composeValidPhonePayload(t, interserviceclient.TestUserPhoneNumber)

	// payload 2
	payload2 := composeValidPhonePayload(t, interserviceclient.TestUserPhoneNumberWithPin)

	// payload 3
	payload3 := composeValidPhonePayload(t, "0700100200")

	// payload 4
	payload4 := composeValidPhonePayload(t, "0700600300")

	// payload 5
	payload5 := composeValidPhonePayload(t, "*")

	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "invalid:_phone_number_is_empty",
			args: args{
				url:        fmt.Sprintf("%s/verify_phone", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload5,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "valid:_successfully_verifies_a_phone_number",
			args: args{
				url:        fmt.Sprintf("%s/verify_phone", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:_user_phone_already_exists",
			args: args{
				url:        fmt.Sprintf("%s/verify_phone", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload2,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_unable_to_check_if_phone_exists",
			args: args{
				url:        fmt.Sprintf("%s/verify_phone", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload3,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_sending_and_generation_of_OTP_fails",
			args: args{
				url:        fmt.Sprintf("%s/verify_phone", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload4,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			// Create a request to pass to our handler.
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			response := httptest.NewRecorder()
			// we mock the required methods for a valid case
			fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
				phone := "+254721123123"
				return &phone, nil
			}
			if tt.name == "invalid:_phone_number_is_empty" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("empty phone number")
				}
			}
			if tt.name == "valid:_successfully_verifies_a_phone_number" {
				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return false, nil
				}
				fakeEngagementSvs.GenerateAndSendOTPFn = func(ctx context.Context, phone string) (*profileutils.OtpResponse, error) {
					return &profileutils.OtpResponse{OTP: "1234"}, nil
				}
			}
			// we mock `CheckPhoneExists` to return true
			// we dont need to mock `GenerateAndSendOTP` because we won't get there
			if tt.name == "invalid:_user_phone_already_exists" {
				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return true, nil
				}
			}
			// we mock `CheckPhoneExists` to return error,
			if tt.name == "invalid:_unable_to_check_if_phone_exists" {
				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return false, fmt.Errorf("unable to check if phone exists")
				}
			}
			// we mock `GenerateAndSendOTP` to return error,
			if tt.name == "invalid:_sending_and_generation_of_OTP_fails" {
				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return false, nil
				}
				fakeEngagementSvs.GenerateAndSendOTPFn = func(ctx context.Context, phone string) (*profileutils.OtpResponse, error) {
					return nil, fmt.Errorf("unable generate and send otp")
				}
			}
			// Our handlers satisfy http.Handler, so we can call its ServeHTTP method
			// directly and pass in our Request and ResponseRecorder.
			svr := h.VerifySignUpPhoneNumber()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
			if err != nil {
				t.Errorf("can't read response body: %v", err)
				return
			}
			if dataResponse == nil {
				t.Errorf("nil response body data")
				return
			}
			if !tt.wantErr {
				data := map[string]interface{}{}
				err = json.Unmarshal(dataResponse, &data)
				if err != nil {
					t.Errorf("bad data returned")
					return
				}
				if !tt.wantErr {
					_, ok := data["error"]
					if ok {
						t.Errorf("error not expected")
						return
					}
				}
			}

		})

	}
}

func TestHandlersInterfacesImpl_CreateUserWithPhoneNumber(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)

	// payload
	pin := "2030"
	flavour := feedlib.FlavourPro
	otp := "1234"
	phoneNumber := interserviceclient.TestUserPhoneNumber
	payload := composeSignupPayload(t, phoneNumber, pin, otp, flavour)

	// payload 2
	pin2 := "1000"
	flavour2 := feedlib.FlavourConsumer
	otp2 := "9000"
	phoneNumber2 := "+254720125456"
	payload2 := composeSignupPayload(t, phoneNumber2, pin2, otp2, flavour2)

	// payload 3
	pin3 := "2000"
	flavour3 := feedlib.FlavourConsumer
	otp3 := "3000"
	phoneNumber3 := "+254721100200"
	payload3 := composeSignupPayload(t, phoneNumber3, pin3, otp3, flavour3)

	// payload6
	pin6 := "0000"
	flavour6 := feedlib.FlavourConsumer
	otp6 := "9520"
	phoneNumber6 := "+254721410589"
	payload6 := composeSignupPayload(t, phoneNumber6, pin6, otp6, flavour6)

	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid:_successfully_create_user_by_phone",
			args: args{
				url:        fmt.Sprintf("%s/create_user_by_phone", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name: "invalid_unable_to_verify_otp",
			args: args{
				url:        fmt.Sprintf("%s/create_user_by_phone", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload2,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid_verify_otp_returns_false",
			args: args{
				url:        fmt.Sprintf("%s/create_user_by_phone", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload3,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid_get_or_create_phone_returns_error",
			args: args{
				url:        fmt.Sprintf("%s/create_user_by_phone", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload6,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request to pass to our handler.
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			response := httptest.NewRecorder()
			// we mock the required methods for a valid case
			if tt.name == "valid:_successfully_create_user_by_phone" {
				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return false, nil
				}
				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}
				fakeRepo.GetOrCreatePhoneNumberUserFn = func(ctx context.Context, phone string) (*dto.CreatedUserResponse, error) {
					return &dto.CreatedUserResponse{
						UID:         "1106f10f-bea6-4fa3-bdba-16b1e39bd318",
						DisplayName: "kalulu juha",
						Email:       "juha@gmail.com",
						PhoneNumber: "0756232452",
						PhotoURL:    "",
						ProviderID:  "google.com",
					}, nil
				}
				fakeRepo.CreateUserProfileFn = func(ctx context.Context, phoneNumber, uid string) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "123",
						VerifiedIdentifiers: []profileutils.VerifiedIdentifier{
							{
								UID:           "125",
								LoginProvider: "Phone",
							},
						},
						PrimaryPhone: &phoneNumber,
					}, nil
				}
				fakeRepo.GenerateAuthCredentialsFn = func(ctx context.Context, phone string, profile *profileutils.UserProfile) (*profileutils.AuthCredentialResponse, error) {
					return &profileutils.AuthCredentialResponse{
						UID:          "5550",
						RefreshToken: "55550",
					}, nil
				}
				fakePinExt.EncryptPINFn = func(rawPwd string, options *extension.Options) (string, string) {
					return "salt", "passw"
				}
				fakePubSub.NotifyCoverLinkingFn = func(ctx context.Context, data dto.LinkCoverPubSubMessage) error {
					return nil
				}
				fakeRepo.CreateEmptySupplierProfileFn = func(ctx context.Context, profileID string) (*profileutils.Supplier, error) {
					return &profileutils.Supplier{
						ID:         "5550",
						SupplierID: "5555",
					}, nil
				}
				fakeRepo.CreateEmptyCustomerProfileFn = func(ctx context.Context, profileID string) (*profileutils.Customer, error) {
					return &profileutils.Customer{
						ID:         "0000",
						CustomerID: "22222",
					}, nil
				}
				// should return a profile with an ID
				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "123",
						VerifiedIdentifiers: []profileutils.VerifiedIdentifier{
							{
								UID:           "125",
								LoginProvider: "Phone",
							},
						},
						PrimaryPhone: &phoneNumber,
					}, nil
				}
				// should return true
				fakeRepo.SavePINFn = func(ctx context.Context, pin *domain.PIN) (bool, error) {
					return true, nil
				}

				fakeRepo.SetUserCommunicationsSettingsFn = func(ctx context.Context, profileID string,
					allowWhatsApp *bool, allowTextSms *bool, allowPush *bool, allowEmail *bool) (*profileutils.UserCommunicationsSetting, error) {
					return &profileutils.UserCommunicationsSetting{
						ID:            "111",
						ProfileID:     "profile-id",
						AllowWhatsApp: true,
						AllowEmail:    true,
						AllowTextSMS:  true,
						AllowPush:     true,
					}, nil
				}

				fakePubSub.NotifyCreateContactFn = func(ctx context.Context, contact crmDomain.CRMContact) error {
					return nil
				}

				fakeRepo.GetUserCommunicationsSettingsFn = func(ctx context.Context, profileID string) (*profileutils.UserCommunicationsSetting, error) {
					return &profileutils.UserCommunicationsSetting{
						ID:            "111",
						ProfileID:     "profile-id",
						AllowWhatsApp: true,
						AllowEmail:    true,
						AllowTextSMS:  true,
						AllowPush:     true,
					}, nil
				}

				fakePubSub.TopicIDsFn = func() []string {
					return []string{uuid.New().String()}
				}

				fakePubSub.AddPubSubNamespaceFn = func(topicName string) string {
					return uuid.New().String()
				}

				fakePubSub.PublishToPubsubFn = func(ctx context.Context, topicID string, payload []byte) error {
					return nil
				}

			}

			// mock VerifyOTP to return an error
			if tt.name == "invalid_unable_to_verify_otp" {
				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return false, fmt.Errorf("unable to verify otp")
				}
			}

			// mock VerifyOTP to return an false
			if tt.name == "invalid_verify_otp_returns_false" {
				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return false, nil
				}
			}

			// mock `GetOrCreatePhoneNumberUser` to return an error
			if tt.name == "invalid_get_or_create_phone_returns_error" {
				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}
				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return false, nil
				}
				fakeRepo.GetOrCreatePhoneNumberUserFn = func(ctx context.Context, phone string) (*dto.CreatedUserResponse, error) {
					return nil, fmt.Errorf("unable to create user")
				}
			}

			// Our handlers satisfy http.Handler, so we can call its ServeHTTP method
			// directly and pass in our Request and ResponseRecorder.
			svr := h.CreateUserWithPhoneNumber()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
			if err != nil {
				t.Errorf("can't read response body: %v", err)
				return
			}
			if dataResponse == nil {
				t.Errorf("nil response body data")
				return
			}
			if !tt.wantErr {
				data := map[string]interface{}{}
				err = json.Unmarshal(dataResponse, &data)
				if err != nil {
					t.Errorf("bad data returned")
					return
				}
			}
		})
	}
}

func TestHandlersInterfacesImpl_UserRecoveryPhoneNumbers(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)
	// payload 1
	payload := composeValidPhonePayload(t, interserviceclient.TestUserPhoneNumber)

	// payload 2
	payload2 := composeValidPhonePayload(t, "0710100595")

	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid:_successfully_get_a_recovery_phone",
			args: args{
				url:        fmt.Sprintf("%s/user_recovery_phonenumbers", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:_unable_to_get_profile",
			args: args{
				url:        fmt.Sprintf("%s/user_recovery_phonenumbers", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload2,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request to pass to our handler.
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			response := httptest.NewRecorder()
			// we mock the required methods for a valid case
			if tt.name == "valid:_successfully_get_a_recovery_phone" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}
				fakeRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "123",
						PrimaryPhone: &phoneNumber,
						SecondaryPhoneNumbers: []string{
							"0721521456", "0721856741",
						},
					}, nil
				}
			}

			// we set GetUserProfileByPhoneNumber to return an error
			if tt.name == "invalid:_unable_to_get_profile" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}
				fakeRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to retrieve profile")
				}
			}

			// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
			// directly and pass in our Request and ResponseRecorder.
			svr := h.UserRecoveryPhoneNumbers()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
			if err != nil {
				t.Errorf("can't read response body: %v", err)
				return
			}
			if dataResponse == nil {
				t.Errorf("nil response body data")
				return
			}
			if !tt.wantErr {
				data := map[string]interface{}{}
				err = json.Unmarshal(dataResponse, &data)
				if err != nil {
					t.Errorf("bad data returned")
					return
				}
				if !tt.wantErr {
					_, ok := data["error"]
					if ok {
						t.Errorf("error not expected")
						return
					}
				}
			}
		})
	}
}

func TestHandlersInterfacesImpl_RequestPINReset(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)
	// payload successfully_request_pin_reset
	payload := composeValidPhonePayload(t, interserviceclient.TestUserPhoneNumber)
	// _phone_number_invalid
	payload1 := composeValidPhonePayload(t, "")
	//invalid:_inable_to_get_primary_phone
	payload2 := composeValidPhonePayload(t, "0725123456")
	//invalid:check_has_pin_failed
	payload3 := composeValidPhonePayload(t, "0700100400")

	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid:successfully_request_pin_reset",
			args: args{
				url:        fmt.Sprintf("%s/request_pin_reset", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:_phone_number_invalid",
			args: args{
				url:        fmt.Sprintf("%s/request_pin_reset", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload1,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_inable_to_get_primary_phone",
			args: args{
				url:        fmt.Sprintf("%s/request_pin_reset", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload2,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:check_has_pin_failed",
			args: args{
				url:        fmt.Sprintf("%s/request_pin_reset", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload3,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:otp_generation_fails",
			args: args{
				url:        fmt.Sprintf("%s/request_pin_reset", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload3,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request to pass to our handler.
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			response := httptest.NewRecorder()
			// we mock the required methods for a valid case
			fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
				phone := "+254721123123"
				return &phone, nil
			}
			if tt.name == "invalid:_phone_number_invalid" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("invalid phone number")
				}
			}

			if tt.name == "valid:successfully_request_pin_reset" {
				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "123",
						PrimaryPhone: &phoneNumber,
					}, nil
				}
				fakeRepo.GetPINByProfileIDFn = func(ctx context.Context, profileID string) (*domain.PIN, error) {
					return &domain.PIN{ID: "123", ProfileID: "456"}, nil
				}
				fakeEngagementSvs.GenerateAndSendOTPFn = func(ctx context.Context, phone string) (*profileutils.OtpResponse, error) {
					return &profileutils.OtpResponse{OTP: "1234"}, nil
				}
			}

			if tt.name == "invalid:otp_generation_fails" {
				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "123",
						PrimaryPhone: &phoneNumber,
					}, nil
				}
				fakeRepo.GetPINByProfileIDFn = func(ctx context.Context, profileID string) (*domain.PIN, error) {
					return &domain.PIN{ID: "123", ProfileID: "456"}, nil
				}
				fakeEngagementSvs.GenerateAndSendOTPFn = func(ctx context.Context, phone string) (*profileutils.OtpResponse, error) {
					return nil, fmt.Errorf("unable to generate otp")
				}
			}

			if tt.name == "invalid:_inable_to_get_primary_phone" {
				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to fetch profile")
				}
			}

			if tt.name == "invalid:check_has_pin_failed" {
				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "123",
						PrimaryPhone: &phoneNumber,
					}, nil
				}
				fakeRepo.GetPINByProfileIDFn = func(ctx context.Context, profileID string) (*domain.PIN, error) {
					return nil, fmt.Errorf("unable to retrieve pin")
				}
			}

			// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
			// directly and pass in our Request and ResponseRecorder.
			svr := h.RequestPINReset()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
			if err != nil {
				t.Errorf("can't read response body: %v", err)
				return
			}
			if dataResponse == nil {
				t.Errorf("nil response body data")
				return
			}
			if !tt.wantErr {
				data := map[string]interface{}{}
				err = json.Unmarshal(dataResponse, &data)
				if err != nil {
					t.Errorf("bad data returned")
					return
				}
				if !tt.wantErr {
					_, ok := data["error"]
					if ok {
						t.Errorf("error not expected")
						return
					}
				}
			}

		})
	}
}

func TestHandlersInterfacesImpl_ResetPin(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)
	// payload
	phone := "0712456784"
	pin := "1897"
	otp := "000548"
	payload := composeChangePinPayload(t, phone, pin, otp)
	// payload2
	phone1 := "0710472196"
	pin1 := "02222"
	otp1 := "0002358"
	payload1 := composeChangePinPayload(t, phone1, pin1, otp1)
	// payload3
	phone2 := ""
	pin2 := ""
	otp2 := "6666"
	payload2 := composeChangePinPayload(t, phone2, pin2, otp2)

	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "invalid:empty_payload",
			args: args{
				url:        fmt.Sprintf("%s/reset_pin", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload2,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "valid:successfully_reset_pin",
			args: args{
				url:        fmt.Sprintf("%s/reset_pin", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name: "invalid:unable_to_update_pin",
			args: args{
				url:        fmt.Sprintf("%s/reset_pin", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload1,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request to pass to our handler.
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			response := httptest.NewRecorder()
			// we mock the required methods for a valid case
			if tt.name == "valid:successfully_reset_pin" {
				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "123",
						PrimaryPhone: &phoneNumber,
					}, nil
				}
				fakeRepo.GetPINByProfileIDFn = func(ctx context.Context, profileID string) (*domain.PIN, error) {
					return &domain.PIN{ID: "123", ProfileID: "456"}, nil
				}
				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}
				fakeRepo.UpdatePINFn = func(ctx context.Context, id string, pin *domain.PIN) (bool, error) {
					return true, nil
				}
			}

			// we set `UpdatePIN` to return an error
			if tt.name == "invalid:unable_to_update_pin" {
				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "123",
						PrimaryPhone: &phoneNumber,
					}, nil
				}
				fakeRepo.GetPINByProfileIDFn = func(ctx context.Context, profileID string) (*domain.PIN, error) {
					return &domain.PIN{ID: "123", ProfileID: "456"}, nil
				}
				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}
				fakeRepo.UpdatePINFn = func(ctx context.Context, id string, pin *domain.PIN) (bool, error) {
					return false, fmt.Errorf("unable to update pin")
				}
			}
			svr := h.ResetPin()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func TestHandlersInterfacesImpl_RefreshToken(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)

	token := "10c17f3b-a9a9-431c-ad0a-94c684eccd85"
	payload := composeRefreshTokenPayload(t, &token)

	token1 := "b5c52b32-7dd5-4dd5-9ddb-44cac9701d6c"
	payload1 := composeRefreshTokenPayload(t, &token1)

	token2 := "*"
	payload2 := composeRefreshTokenPayload(t, &token2)

	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid:_successfully_refresh_token",
			args: args{
				url:        fmt.Sprintf("%s/refresh_token", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:_refresh_token_fails",
			args: args{
				url:        fmt.Sprintf("%s/refresh_token", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload1,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_refresh_token_with_invalid_payload",
			args: args{
				url:        fmt.Sprintf("%s/refresh_token", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload2,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request to pass to our handler.
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			response := httptest.NewRecorder()
			if tt.name == "valid:_successfully_refresh_token" {
				fakeRepo.ExchangeRefreshTokenForIDTokenFn = func(ctx context.Context, token string) (*profileutils.AuthCredentialResponse, error) {
					return &profileutils.AuthCredentialResponse{
						UID:          "5550",
						RefreshToken: "55550",
					}, nil
				}
			}

			if tt.name == "invalid:_refresh_token_fails" {
				fakeRepo.ExchangeRefreshTokenForIDTokenFn = func(ctx context.Context, token string) (*profileutils.AuthCredentialResponse, error) {
					return nil, fmt.Errorf("unable to refresh token")
				}
			}

			svr := h.RefreshToken()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func TestHandlersInterfacesImpl_GetUserProfileByUID(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)

	uid := "db963177-21b2-489f-83e6-3521bf5db516"
	payload := composeUIDPayload(t, &uid)

	uid1 := "584799be-97c5-4aa4-8b0f-094990bd55b1"
	payload1 := composeUIDPayload(t, &uid1)

	uid2 := "*"
	payload2 := composeUIDPayload(t, &uid2)

	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid:_successfully_get_profile_by_uid",
			args: args{
				url:        fmt.Sprintf("%s/user_profile", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:_unable_to_get_profile_by_uid",
			args: args{
				url:        fmt.Sprintf("%s/user_profile", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload1,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_get_profile_by_uid_with_invalid_payload",
			args: args{
				url:        fmt.Sprintf("%s/user_profile", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload2,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request to pass to our handler.
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			response := httptest.NewRecorder()
			if tt.name == "valid:_successfully_get_profile_by_uid" {
				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7-5b64-4c2f-91bd-42b3af315a4e",
					}, nil
				}
			}

			if tt.name == "invalid:_unable_to_get_profile_by_uid" {
				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get profile")
				}
			}

			svr := h.GetUserProfileByUID()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func TestHandlersInterfacesImpl_SendOTP(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)

	payload := composeValidPhonePayload(t, interserviceclient.TestUserPhoneNumber)

	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid:_successfully_send_otp",
			args: args{
				url:        fmt.Sprintf("%s/send_otp", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:_unable_to_send_otp",
			args: args{
				url:        fmt.Sprintf("%s/send_otp", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request to pass to our handler.
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			response := httptest.NewRecorder()
			if tt.name == "valid:_successfully_send_otp" {
				fakeEngagementSvs.GenerateAndSendOTPFn = func(ctx context.Context, phone string) (*profileutils.OtpResponse, error) {
					return &profileutils.OtpResponse{OTP: "1234"}, nil
				}
			}

			if tt.name == "invalid:_unable_to_send_otp" {
				fakeEngagementSvs.GenerateAndSendOTPFn = func(ctx context.Context, phone string) (*profileutils.OtpResponse, error) {
					return nil, fmt.Errorf("unable to send otp")
				}
			}

			svr := h.SendOTP()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func TestHandlersInterfacesImpl_LoginByPhone(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)
	// payload
	phone := "0712456784"
	pin := "1897"
	flavour := feedlib.FlavourPro
	payload := composeLoginPayload(t, phone, pin, flavour)

	// payload1 : invalid:_get_userprofile_by_primary_phone_fails
	phone1 := "0708598520"
	pin1 := "1800"
	flavour1 := feedlib.FlavourConsumer
	payload1 := composeLoginPayload(t, phone1, pin1, flavour1)

	// payload2 : invalid:_get_pinbyprofileid_fails
	phone2 := "0708590000"
	pin2 := "1000"
	flavour2 := feedlib.FlavourConsumer
	payload2 := composeLoginPayload(t, phone2, pin2, flavour2)

	// payload4 invalid:_pin_mismatch
	phone4 := "0702960230"
	pin4 := "1023"
	flavour4 := feedlib.FlavourConsumer
	payload4 := composeLoginPayload(t, phone4, pin4, flavour4)

	// payload5 invalid:_generate_auth_credentials_fails
	phone5 := "0705222888"
	pin5 := "1093"
	flavour5 := feedlib.FlavourConsumer
	payload5 := composeLoginPayload(t, phone5, pin5, flavour5)

	// payload7 invalid:_invalid_flavour_used
	phone7 := "0712456784"
	pin7 := "1897"
	payload7 := composeLoginPayload(t, phone7, pin7, "invalidFlavour")
	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid:_successfully_login_by_phone",
			args: args{
				url:        fmt.Sprintf("%s/login_by_phone", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:_get_userprofile_by_primary_phone_fails",
			args: args{
				url:        fmt.Sprintf("%s/login_by_phone", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload1,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_get_pinbyprofileid_fails",
			args: args{
				url:        fmt.Sprintf("%s/login_by_phone", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload2,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_pin_mismatch",
			args: args{
				url:        fmt.Sprintf("%s/login_by_phone", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload4,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_generate_auth_credentials_fails",
			args: args{
				url:        fmt.Sprintf("%s/login_by_phone", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload5,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_invalid_flavour_used",
			args: args{
				url:        fmt.Sprintf("%s/login_by_phone", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload7,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request to pass to our handler.
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			response := httptest.NewRecorder()
			// we mock the required methods for a valid case
			if tt.name == "valid:_successfully_login_by_phone" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}
				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "123",
						VerifiedIdentifiers: []profileutils.VerifiedIdentifier{
							{
								UID:           "125",
								LoginProvider: "Phone",
							},
						},
						PrimaryPhone: &phoneNumber,
					}, nil
				}
				fakeRepo.GetPINByProfileIDFn = func(ctx context.Context, profileID string) (*domain.PIN, error) {
					return &domain.PIN{ID: "123", ProfileID: "456"}, nil
				}
				fakePinExt.ComparePINFn = func(rawPwd string, salt string, encodedPwd string, options *extension.Options) bool {
					return true
				}
				fakeRepo.GenerateAuthCredentialsFn = func(ctx context.Context, phone string, profile *profileutils.UserProfile) (*profileutils.AuthCredentialResponse, error) {
					return &profileutils.AuthCredentialResponse{
						UID: "5550",
						// IDToken:      "555",
						RefreshToken: "55550",
					}, nil
				}
				fakeRepo.GetCustomerOrSupplierProfileByProfileIDFn = func(ctx context.Context, flavour feedlib.Flavour, profileID string) (*profileutils.Customer, *profileutils.Supplier, error) {
					return &profileutils.Customer{
							ID: "5550",
						}, &profileutils.Supplier{
							ID: "5550",
						}, nil
				}
				fakeRepo.GetUserCommunicationsSettingsFn = func(ctx context.Context, profileID string) (*profileutils.UserCommunicationsSetting, error) {
					return &profileutils.UserCommunicationsSetting{
						ID:            "111",
						ProfileID:     "profile-id",
						AllowWhatsApp: true,
						AllowEmail:    true,
						AllowTextSMS:  true,
						AllowPush:     true,
					}, nil
				}
			}

			if tt.name == "invalid:_get_userprofile_by_primary_phone_fails" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}
				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get user profile")
				}
			}

			if tt.name == "invalid:_get_pinbyprofileid_fails" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}

				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "123",
					}, nil
				}
				fakeRepo.GetPINByProfileIDFn = func(ctx context.Context, profileID string) (*domain.PIN, error) {
					return nil, fmt.Errorf("unable to get pin by profileID")
				}

			}

			if tt.name == "invalid:_pin_mismatch" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}
				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "123",
						VerifiedIdentifiers: []profileutils.VerifiedIdentifier{
							{
								UID:           "125",
								LoginProvider: "Phone",
							},
						},
						PrimaryPhone: &phoneNumber,
					}, nil
				}
				fakeRepo.GetPINByProfileIDFn = func(ctx context.Context, profileID string) (*domain.PIN, error) {
					return &domain.PIN{ID: "123", ProfileID: "456"}, nil
				}
				fakePinExt.ComparePINFn = func(rawPwd string, salt string, encodedPwd string, options *extension.Options) bool {
					return false
				}
			}

			if tt.name == "invalid:_generate_auth_credentials_fails" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721123123"
					return &phone, nil
				}
				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "123",
						VerifiedIdentifiers: []profileutils.VerifiedIdentifier{
							{
								UID:           "125",
								LoginProvider: "Phone",
							},
						},
						PrimaryPhone: &phoneNumber,
					}, nil
				}
				fakeRepo.GetPINByProfileIDFn = func(ctx context.Context, profileID string) (*domain.PIN, error) {
					return &domain.PIN{ID: "123", ProfileID: "456"}, nil
				}
				fakePinExt.ComparePINFn = func(rawPwd string, salt string, encodedPwd string, options *extension.Options) bool {
					return true
				}
				fakeRepo.GenerateAuthCredentialsFn = func(ctx context.Context, phone string, profile *profileutils.UserProfile) (*profileutils.AuthCredentialResponse, error) {
					return nil, fmt.Errorf("unable to generate auth credentials")
				}
			}

			if tt.name == "invalid:_invalid_flavour_used" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("invalid flavour defined")
				}
			}

			svr := h.LoginByPhone()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
			if err != nil {
				t.Errorf("can't read response body: %v", err)
				return
			}
			if dataResponse == nil {
				t.Errorf("nil response body data")
				return
			}

			if !tt.wantErr {
				data := map[string]interface{}{}
				err = json.Unmarshal(dataResponse, &data)
				if err != nil {
					t.Errorf("bad data returned")
					return
				}
				if !tt.wantErr {
					_, ok := data["error"]
					if ok {
						t.Errorf("error not expected")
						return
					}
				}
			}
		})
	}
}

func TestHandlersInterfacesImpl_SendRetryOTP(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)

	// valid payload
	validPayload := composeSendRetryOTPPayload(t, interserviceclient.TestUserPhoneNumber, 1)

	invalidPayload := composeSendRetryOTPPayload(t, "", 2)
	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid:_successfully_send_retry_otp",
			args: args{
				url:        fmt.Sprintf("%s/send_retry_otp", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:_unable_to_send_otp",
			args: args{
				url:        fmt.Sprintf("%s/send_retry_otp", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_unable_to_send_otp_due_to_missing_msisdn",
			args: args{
				url:        fmt.Sprintf("%s/send_retry_otp", serverUrl),
				httpMethod: http.MethodPost,
				body:       invalidPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create a new request: %v", err)
				return
			}

			response := httptest.NewRecorder()

			if tt.name == "valid:_successfully_send_retry_otp" {
				fakeEngagementSvs.SendRetryOTPFn = func(ctx context.Context, msisdn string, retryStep int) (*profileutils.OtpResponse, error) {
					return &profileutils.OtpResponse{
						OTP: "123456",
					}, nil
				}
			}

			if tt.name == "invalid:_unable_to_send_otp" {
				fakeEngagementSvs.SendRetryOTPFn = func(ctx context.Context, msisdn string, retryStep int) (*profileutils.OtpResponse, error) {
					return nil, fmt.Errorf("unable to send OTP")
				}
			}

			if tt.name == "invalid:_unable_to_send_otp_due_to_missing_msisdn" {
				fakeEngagementSvs.SendRetryOTPFn = func(ctx context.Context, msisdn string, retryStep int) (*profileutils.OtpResponse, error) {
					return nil, fmt.Errorf("unable to send OTP")
				}
			}

			svr := h.SendRetryOTP()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func TestHandlersInterfacesImpl_LoginAnonymous(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)

	validPayload := composeLoginPayload(t, "", "", feedlib.FlavourConsumer)
	invalidPayload := composeLoginPayload(t, "", "", " ")

	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid:_successfully_login_as_anonymous",
			args: args{
				url:        fmt.Sprintf("%s/login_anonymous", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:_invalid_flavour_defined",
			args: args{
				url:        fmt.Sprintf("%s/login_anonymous", serverUrl),
				httpMethod: http.MethodPost,
				body:       invalidPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_missing_flavour",
			args: args{
				url:        fmt.Sprintf("%s/login_anonymous", serverUrl),
				httpMethod: http.MethodPost,
				body:       invalidPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			if tt.name == "valid:_successfully_login_as_anonymous" {
				fakeRepo.GenerateAuthCredentialsForAnonymousUserFn = func(ctx context.Context) (*profileutils.AuthCredentialResponse, error) {
					return &profileutils.AuthCredentialResponse{
						UID:          "6660",
						RefreshToken: "6660",
					}, nil
				}
			}

			if tt.name == "invalid:_invalid_flavour_defined" {
				fakeRepo.GenerateAuthCredentialsForAnonymousUserFn = func(ctx context.Context) (*profileutils.AuthCredentialResponse, error) {
					return nil, fmt.Errorf("an invalid `flavour` defined")
				}
			}

			if tt.name == "invalid:_missing_flavour" {
				fakeRepo.GenerateAuthCredentialsForAnonymousUserFn = func(ctx context.Context) (*profileutils.AuthCredentialResponse, error) {
					return nil, fmt.Errorf("expected `flavour` to be defined")
				}
			}

			response := httptest.NewRecorder()

			svr := h.LoginAnonymous()

			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func TestHandlersInterfacesImpl_UpdateCovers(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)

	invalidUID := " "
	uid := "5cf354a2-1d3e-400d-8716-7e2aead29f2c"
	payerName := "Payer Name"
	memberName := "Member Name"
	memberNumber := "5678"
	payerSladeCode := 1234
	beneficiaryID := 15689
	effectivePolicyNumber := "14582"

	validFromString := "2021-01-01T00:00:00+03:00"
	validFrom, err := time.Parse(time.RFC3339, validFromString)
	if err != nil {
		t.Errorf("failed parse date string: %v", err)
		return
	}

	validToString := "2022-01-01T00:00:00+03:00"
	validTo, err := time.Parse(time.RFC3339, validToString)
	if err != nil {
		t.Errorf("failed parse date string: %v", err)
		return
	}

	updateCoversPayloadValid := &dto.UpdateCoversPayload{
		UID:                   &uid,
		PayerName:             &payerName,
		PayerSladeCode:        &payerSladeCode,
		MemberName:            &memberName,
		MemberNumber:          &memberNumber,
		BeneficiaryID:         &beneficiaryID,
		EffectivePolicyNumber: &effectivePolicyNumber,
		ValidFrom:             &validFrom,
		ValidTo:               &validTo,
	}

	updateCoversPayloadInValid := &dto.UpdateCoversPayload{
		UID:                   &invalidUID,
		PayerName:             &payerName,
		PayerSladeCode:        &payerSladeCode,
		MemberName:            &memberName,
		MemberNumber:          &memberNumber,
		BeneficiaryID:         &beneficiaryID,
		EffectivePolicyNumber: &effectivePolicyNumber,
		ValidFrom:             &validFrom,
		ValidTo:               &validTo,
	}

	validPayload := composeCoversUpdatePayload(t, updateCoversPayloadValid)
	inValidPayload := composeCoversUpdatePayload(t, updateCoversPayloadInValid)

	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid:_Successfully_update_covers",
			args: args{
				url:        fmt.Sprintf("%s/update_covers", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},

		{
			name: "invalid:_update_covers_fails",
			args: args{
				url:        fmt.Sprintf("%s/update_covers", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_get_user_profile_by_UID_fails",
			args: args{
				url:        fmt.Sprintf("%s/update_covers", serverUrl),
				httpMethod: http.MethodPost,
				body:       inValidPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			response := httptest.NewRecorder()

			if tt.name == "valid:_Successfully_update_covers" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "8716-7e2aead29f2c", nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7",
					}, nil
				}

				fakeRepo.UpdateCoversFn = func(ctx context.Context, id string, covers []profileutils.Cover) error {
					return nil
				}
			}

			if tt.name == "invalid:_get_user_profile_by_UID_fails" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "", fmt.Errorf("failed to get logged in user UID")
				}
			}

			if tt.name == "invalid:_update_covers_fails" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "5cf354a2-1d3e-400d-8716-7e2aead29f2c", nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "f4f39af7",
					}, nil
				}

				fakeRepo.UpdateCoversFn = func(ctx context.Context, id string, covers []profileutils.Cover) error {
					return fmt.Errorf("unable to update covers")
				}
			}

			svr := h.UpdateCovers()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func TestHandlersInterfacesImpl_FindSupplierByUID(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)

	uid := "5cf354a2-1d3e-400d-8716-7e2aead29f2c"
	payload := composeUIDPayload(t, &uid)

	uid1 := "98cbf5e8-162b-4b8a-a618-f6fff3c36ef9"
	payload1 := composeUIDPayload(t, &uid1)

	uid2 := "53298383-eb8a-4a3e-8428-cf76e7af644e"
	payload2 := composeUIDPayload(t, &uid2)

	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid:_successfully_get_supplier_by_uid",
			args: args{
				url:        fmt.Sprintf("%s/supplier", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:_fail_to_get_supplier_by_uid",
			args: args{
				url:        fmt.Sprintf("%s/supplier", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload1,
			},
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name: "invalid:_unable_to_get_user_profile_by_uid",
			args: args{
				url:        fmt.Sprintf("%s/supplier", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload2,
			},
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("failed to create a new request: %s", err)
				return
			}

			response := httptest.NewRecorder()

			if tt.name == "valid:_successfully_get_supplier_by_uid" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{
						UID:         "12233",
						Email:       "test@example.com",
						PhoneNumber: "0721568526",
					}, nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "AD-FSO798",
					}, nil
				}
				fakeRepo.GetSupplierProfileByProfileIDFn = func(ctx context.Context, profileID string) (*profileutils.Supplier, error) {
					return &profileutils.Supplier{
						ProfileID: &profileID,
					}, nil
				}
			}

			if tt.name == "invalid:_fail_to_get_supplier_by_uid" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "FSO798-AD3", nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "AD-FSO798",
					}, nil
				}
				fakeRepo.GetSupplierProfileByProfileIDFn = func(ctx context.Context, profileID string) (*profileutils.Supplier, error) {
					return nil, fmt.Errorf("failed to get the supplier profile")
				}
			}

			if tt.name == "invalid:_unable_to_get_user_profile_by_uid" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "FSO798-AD3", nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get profile")
				}
			}

			svr := h.FindSupplierByUID()
			svr.ServeHTTP(response, req)
			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func TestHandlersInterfacesImpl_RemoveUserByPhoneNumber(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)

	primaryPhone := "+254711445566"
	validPayload := composeValidPhonePayload(t, primaryPhone)
	validPayload1 := composeValidPhonePayload(t, "+254777882200")
	validPayload2 := composeValidPhonePayload(t, "+")

	invalidPayload := composeValidPhonePayload(t, " ")
	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid:_Successfully_remove_user_by_phone",
			args: args{
				url:        fmt.Sprintf("%s/remove_user", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:_unable_to_remove_user_by_phone",
			args: args{
				url:        fmt.Sprintf("%s/remove_user", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload1,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},

		{
			name: "invalid:_empty_phonenumber",
			args: args{
				url:        fmt.Sprintf("%s/remove_user", serverUrl),
				httpMethod: http.MethodPost,
				body:       invalidPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_unable_to_check_if_phone_exists",
			args: args{
				url:        fmt.Sprintf("%s/remove_user", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload2,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			response := httptest.NewRecorder()

			if tt.name == "valid:_Successfully_remove_user_by_phone" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721026491"
					return &phone, nil
				}

				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return true, nil
				}

				fakeRepo.PurgeUserByPhoneNumberFn = func(ctx context.Context, phone string) error {
					return nil
				}

			}

			if tt.name == "invalid:_unable_to_remove_user_by_phone" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721026491"
					return &phone, nil
				}

				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return true, nil
				}

				fakeRepo.PurgeUserByPhoneNumberFn = func(ctx context.Context, phone string) error {
					return fmt.Errorf("unable to purge user by phonenumber")
				}

			}

			if tt.name == "invalid:_empty_phonenumber" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("empty phone number")
				}
			}

			if tt.name == "invalid:_unable_to_check_if_phone_exists" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "0788554422"
					return &phone, nil
				}

				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return false, fmt.Errorf("the phone does not exist")
				}

			}

			svr := h.RemoveUserByPhoneNumber()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func TestHandlersInterfacesImpl_SetPrimaryPhoneNumber(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)

	primaryPhone := "+254701567839"
	otp := "890087"
	validPayload := composeSetPrimaryPhoneNumberPayload(t, primaryPhone, otp)

	primaryPhone1 := "+254765738293"
	otp1 := "345678"
	payload1 := composeSetPrimaryPhoneNumberPayload(t, primaryPhone1, otp1)

	primaryPhone2 := " "
	otp2 := " "
	payload2 := composeSetPrimaryPhoneNumberPayload(t, primaryPhone2, otp2)
	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid:_successfully_set_a_primary_phonenumber",
			args: args{
				url:        fmt.Sprintf("%s/set_primary_phonenumber", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:_fail_to_set_a_primary_phonenumber",
			args: args{
				url:        fmt.Sprintf("%s/set_primary_phonenumber", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload1,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_phonenumber_and_otp_missing",
			args: args{
				url:        fmt.Sprintf("%s/set_primary_phonenumber", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload2,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			if tt.name == "valid:_successfully_set_a_primary_phonenumber" {

				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254799774466"
					return &phone, nil
				}

				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}

				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254799774466"
					return &phone, nil
				}

				fakeRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "123",
						PrimaryPhone: &phoneNumber,
						SecondaryPhoneNumbers: []string{
							"0721521456", "0721856741",
						},
					}, nil
				}

				fakeRepo.UpdatePrimaryPhoneNumberFn = func(ctx context.Context, id string, phoneNumber string) error {
					return nil
				}

				fakeRepo.UpdateSecondaryPhoneNumbersFn = func(ctx context.Context, id string, phoneNumbers []string) error {
					return nil
				}
			}

			if tt.name == "invalid:_fail_to_set_a_primary_phonenumber" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254799774466"
					return &phone, nil
				}

				fakeEngagementSvs.VerifyOTPFn = func(ctx context.Context, phone, OTP string) (bool, error) {
					return true, nil
				}

				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254799774466"
					return &phone, nil
				}

				fakeRepo.GetUserProfileByPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:           "123",
						PrimaryPhone: &phoneNumber,
						SecondaryPhoneNumbers: []string{
							"0721521456", "0721856741",
						},
					}, nil
				}

				fakeRepo.UpdatePrimaryPhoneNumberFn = func(ctx context.Context, id string, phoneNumber string) error {
					return fmt.Errorf("failed to set a primary phone number")
				}
			}

			if tt.name == "invalid:_phonenumber_and_otp_missing" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("empty phone number provided")
				}
			}

			response := httptest.NewRecorder()

			svr := h.SetPrimaryPhoneNumber()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func TestHandlersInterfacesImpl_RegisterPushToken(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)
	uid := "5cf354a2-1d3e-400d-8716-7e2aead29f2c"
	token := "10c17f3b-a9a9-431c-ad0a-94c684eccd85"
	payload := composePushTokenPayload(t, token, uid)

	token1 := ""
	uid1 := "5cf354a2-1d3e-400d-8716-7e2aead29f2c"
	invalidPayload := composePushTokenPayload(t, token1, uid1)

	type args struct {
		url        string
		httpMethod string
		body       io.Reader
	}
	tests := []struct {
		name       string
		args       args
		want       http.HandlerFunc
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid:_successfully_register_push_token",
			args: args{
				url:        fmt.Sprintf("%s/register_push_token", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:_unsuccessfully_register_push_token",
			args: args{
				url:        fmt.Sprintf("%s/register_push_token", serverUrl),
				httpMethod: http.MethodPost,
				body:       invalidPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			if tt.name == "valid:_successfully_register_push_token" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "400d-8716-7e2aead29f2c", nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:         "f4f39af7--91bd-42b3af315a4e",
						PushTokens: []string{"token12", "token23", "token34"},
					}, nil
				}

				fakeRepo.UpdatePushTokensFn = func(ctx context.Context, id string, pushToken []string) error {
					return nil
				}
			}
			if tt.name == "invalid:_unsuccessfully_register_push_token" {
				fakeBaseExt.GetLoggedInUserUIDFn = func(ctx context.Context) (string, error) {
					return "400d-8716-7e2aead29f2c", nil
				}

				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:         "f4f39af7--91bd-42b3af315a4e",
						PushTokens: []string{"token12", "token23", "token34"},
					}, nil
				}

				fakeRepo.UpdatePushTokensFn = func(ctx context.Context, id string, pushToken []string) error {
					return fmt.Errorf("failed to register push tokens")
				}
			}
			response := httptest.NewRecorder()

			svr := h.RegisterPushToken()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func TestHandlersInterfacesImpl_AddAdminPermsToUser(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)

	primaryPhone := "+254711445566"
	validPayload := composeValidPhonePayload(t, primaryPhone)
	validPayload1 := composeValidPhonePayload(t, "+254777882200")
	validPayload2 := composeValidPhonePayload(t, "+")

	invalidPayload := composeValidPhonePayload(t, " ")

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
			name: "valid:_Successfully_update_user_permissions",
			args: args{
				url:        fmt.Sprintf("%s/update_user_permissions", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:_update_user_permissions",
			args: args{
				url:        fmt.Sprintf("%s/update_user_permissions", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload1,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},

		{
			name: "invalid:_empty_phonenumber",
			args: args{
				url:        fmt.Sprintf("%s/update_user_permissions", serverUrl),
				httpMethod: http.MethodPost,
				body:       invalidPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_unable_to_check_if_phone_exists",
			args: args{
				url:        fmt.Sprintf("%s/update_user_permissions", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload2,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			response := httptest.NewRecorder()

			if tt.name == "valid:_Successfully_update_user_permissions" {
				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return true, nil
				}

				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721026491"
					return &phone, nil
				}

				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}

				fakeRepo.UpdatePermissionsFn = func(ctx context.Context, id string, perms []profileutils.PermissionType) error {
					return nil
				}

			}

			if tt.name == "invalid:_update_user_permissions" {
				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return true, nil
				}

				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721026491"
					return &phone, nil
				}

				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}

				fakeRepo.UpdatePermissionsFn = func(ctx context.Context, id string, perms []profileutils.PermissionType) error {
					return fmt.Errorf("unable to update user permissions")
				}

			}

			if tt.name == "invalid:_empty_phonenumber" {
				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return true, nil
				}
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("empty phone number")
				}
			}

			if tt.name == "invalid:_unable_to_check_if_phone_exists" {
				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return false, fmt.Errorf("the phone does not exist")
				}

			}

			svr := h.AddAdminPermsToUser()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func TestHandlersInterfacesImpl_RemoveAdminPermsToUser(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)

	primaryPhone := "+254711445566"
	validPayload := composeValidPhonePayload(t, primaryPhone)
	validPayload1 := composeValidPhonePayload(t, "+254777882200")
	validPayload2 := composeValidPhonePayload(t, "+")

	invalidPayload := composeValidPhonePayload(t, " ")

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
			name: "valid:_Successfully_remove_admin_permissions",
			args: args{
				url:        fmt.Sprintf("%s/remove_admin_permissions", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:_remove_admin_permissions",
			args: args{
				url:        fmt.Sprintf("%s/remove_admin_permissions", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload1,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},

		{
			name: "invalid:_empty_phonenumber",
			args: args{
				url:        fmt.Sprintf("%s/remove_admin_permissions", serverUrl),
				httpMethod: http.MethodPost,
				body:       invalidPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:_unable_to_check_if_phone_exists",
			args: args{
				url:        fmt.Sprintf("%s/remove_admin_permissions", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload2,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			response := httptest.NewRecorder()

			if tt.name == "valid:_Successfully_remove_admin_permissions" {

				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return true, nil
				}

				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721026491"
					return &phone, nil
				}

				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}

				fakeRepo.UpdatePermissionsFn = func(ctx context.Context, id string, perms []profileutils.PermissionType) error {
					return nil
				}

			}

			if tt.name == "invalid:_remove_admin_permissions" {

				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return true, nil
				}

				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721026491"
					return &phone, nil
				}

				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}

				fakeRepo.UpdatePermissionsFn = func(ctx context.Context, id string, perms []profileutils.PermissionType) error {
					return fmt.Errorf("unable to update user permissions")
				}

			}

			if tt.name == "invalid:_empty_phonenumber" {

				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return true, nil
				}

				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("empty phone number")
				}
			}

			if tt.name == "invalid:_unable_to_check_if_phone_exists" {

				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return true, nil
				}

				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return false, fmt.Errorf("the phone does not exist")
				}

			}

			svr := h.RemoveAdminPermsToUser()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func TestHandlersInterfacesImpl_AddRoleToUser(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)

	validPhone := "+254711445566"
	invalidPhone := "+254777882200"
	validRole := profileutils.RoleTypeEmployee
	var invalidRole profileutils.RoleType = "STANGER"
	payload := composeValidRolePayload(t, validPhone, validRole)
	payload1 := composeValidRolePayload(t, invalidPhone, validRole)
	payload2 := composeValidRolePayload(t, validPhone, invalidRole)

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
			name: "valid:successful_added_user_role",
			args: args{
				url:        fmt.Sprintf("%s/add_user_role", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:failed_to_find_user",
			args: args{
				url:        fmt.Sprintf("%s/add_user_role", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload1,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid:failed_invalid_role",
			args: args{
				url:        fmt.Sprintf("%s/add_user_role", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload2,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			response := httptest.NewRecorder()

			if tt.name == "valid:successful_added_user_role" {

				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721026491"
					return &phone, nil
				}
				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}
				fakeRepo.UpdateRoleFn = func(ctx context.Context, id string, role profileutils.RoleType) error {
					return nil
				}
			}

			if tt.name == "invalid:failed_to_find_user" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("Invalid phone number provided")
				}
			}

			if tt.name == "invalid:failed_invalid_role" {

				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721026491"
					return &phone, nil
				}
				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}
				fakeRepo.UpdateRoleFn = func(ctx context.Context, id string, role profileutils.RoleType) error {
					return fmt.Errorf("Invalid role provided")
				}
			}

			serverResponse := h.AddRoleToUser()
			serverResponse.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func TestHandlersInterfacesImpl_RemoveRoleToUser(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}

	h := rest.NewHandlersInterfaces(i)

	validPhone := "+254711445566"
	payload := composeValidPhonePayload(t, validPhone)
	payload1 := composeValidPhonePayload(t, validPhone)

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
			name: "valid:successful_removed_user_role",
			args: args{
				url:        fmt.Sprintf("%s/remove_user_role", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid:failed_to_find_user",
			args: args{
				url:        fmt.Sprintf("%s/remove_user_role", serverUrl),
				httpMethod: http.MethodPost,
				body:       payload1,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}

			response := httptest.NewRecorder()

			if tt.name == "valid:successful_removed_user_role" {

				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721026491"
					return &phone, nil
				}
				fakeRepo.GetUserProfileByPrimaryPhoneNumberFn = func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: uuid.New().String(),
					}, nil
				}
				fakeRepo.UpdateRoleFn = func(ctx context.Context, id string, role profileutils.RoleType) error {
					return nil
				}
			}

			if tt.name == "invalid:failed_to_find_user" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("Invalid phone number provided")
				}
			}

			serverResponse := h.RemoveRoleToUser()
			serverResponse.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func composeSMSMessageDataJSONPayload(
	t *testing.T,
	payload *dto.AfricasTalkingMessage,
) *bytes.Buffer {

	bs, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("unable to marshal payload to JSON: %s", err)
		return nil
	}
	return bytes.NewBuffer(bs)
}

func TestHandlersInterfacesImpl_IncomingATSMS(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}
	h := rest.NewHandlersInterfaces(i)

	invalidLinkId := ""
	validLinkId := uuid.New().String()
	text := "Test Covers"
	to := "3601"
	id := "60119"
	from := "+254705385894"
	date := "2021-05-17T13:20:04.490Z"

	validSMSData := &dto.AfricasTalkingMessage{
		LinkID: validLinkId,
		Text:   text,
		To:     to,
		ID:     id,
		Date:   date,
		From:   from,
	}

	invalidSMSData := &dto.AfricasTalkingMessage{
		LinkID: invalidLinkId,
		Text:   text,
		To:     to,
		ID:     id,
		Date:   date,
		From:   from,
	}

	validPayload := composeSMSMessageDataPayload(t, validSMSData)
	invalidPayload := composeSMSMessageDataPayload(t, invalidSMSData)
	invalidJSONPayload := composeSMSMessageDataJSONPayload(t, validSMSData)
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
			name: "VALID_CASE:Valid_incoming_sms",
			args: args{
				url:        fmt.Sprintf("%s/incoming_ait_sms", serverUrl),
				httpMethod: http.MethodPost,
				body:       validPayload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "INVALID_CASE:Nil_incoming_sms_JSON",
			args: args{
				url:        fmt.Sprintf("%s/incoming_ait_sms", serverUrl),
				httpMethod: http.MethodPost,
				body:       invalidJSONPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "INVALID_CASE:Invalid_incoming_sms",
			args: args{
				url:        fmt.Sprintf("%s/incoming_ait_sms", serverUrl),
				httpMethod: http.MethodPost,
				body:       invalidPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			response := httptest.NewRecorder()

			if tt.name == "VALID_CASE:Valid_incoming_sms" {
				fakeRepo.PersistIncomingSMSDataFn = func(ctx context.Context, input *dto.AfricasTalkingMessage) error {
					return nil
				}
			}

			if tt.name == "INVALID_CASE:Nil_incoming_sms_JSON" {
				fakeRepo.PersistIncomingSMSDataFn = func(ctx context.Context, input *dto.AfricasTalkingMessage) error {
					return fmt.Errorf("invalid sms")
				}
			}

			if tt.name == "INVALID_CASE:Invalid_incoming_sms" {
				fakeRepo.PersistIncomingSMSDataFn = func(ctx context.Context, input *dto.AfricasTalkingMessage) error {
					return fmt.Errorf("invalid sms")
				}
			}

			svr := h.IncomingATSMS()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
			if err != nil {
				t.Errorf("can't read response body: %v", err)
			}
			if dataResponse == nil {
				t.Errorf("nil response body data")
				return
			}
		})
	}
}

func TestHandlersInterfacesImpl_USSDHandler(t *testing.T) {

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}
	h := rest.NewHandlersInterfaces(i)

	USSDPhoneNumber := "+254711445566"
	invalidUSSDPhoneNumber := ""
	sessionId := "123456778"
	invalidSessionId := ""
	text := "1"
	level := 0

	validPayload := &dto.SessionDetails{
		SessionID:   sessionId,
		PhoneNumber: &USSDPhoneNumber,
		Level:       level,
		Text:        text,
	}

	invalidPayload := &dto.SessionDetails{
		SessionID:   invalidSessionId,
		PhoneNumber: &invalidUSSDPhoneNumber,
		Level:       level,
		Text:        text,
	}

	validUSSDPayload := composeUssdPayload(t, validPayload)
	invalidUSSDPayload := composeUssdPayload(t, invalidPayload)

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
			name: "valid:_successful_USSD",
			args: args{
				url:        fmt.Sprintf("%s/ait_ussd", serverUrl),
				httpMethod: http.MethodPost,
				body:       validUSSDPayload,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "Invalid:_unsuccessful_USSD",
			args: args{
				url:        fmt.Sprintf("%s/ait_ussd", serverUrl),
				httpMethod: http.MethodPost,
				body:       invalidUSSDPayload,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "Invalid:_nil_body",
			args: args{
				url:        fmt.Sprintf("%s/ait_ussd", serverUrl),
				httpMethod: http.MethodPost,
				body:       nil,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			response := httptest.NewRecorder()
			if tt.name == "valid:_successful_USSD" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					phone := "+254721026491"
					return &phone, nil
				}

				fakeRepo.AddAITSessionDetailsFn = func(ctx context.Context, input *dto.SessionDetails) (*domain.USSDLeadDetails, error) {
					return &domain.USSDLeadDetails{
						ID:          uuid.New().String(),
						SessionID:   input.SessionID,
						PhoneNumber: *input.PhoneNumber,
						Level:       input.Level,
					}, nil
				}

				fakeRepo.UpdateSessionLevelFn = func(ctx context.Context, sessionID string, level int) (*domain.USSDLeadDetails, error) {
					return nil, nil
				}

				fakeRepo.GetAITSessionDetailsFn = func(ctx context.Context, sessionID string) (*domain.USSDLeadDetails, error) {
					return &domain.USSDLeadDetails{
						Level: 2,
					}, nil
				}

				fakeRepo.CheckIfPhoneNumberExistsFn = func(ctx context.Context, phone string) (bool, error) {
					return true, nil
				}
				fakeRepo.UpdateSessionPINFn = func(ctx context.Context, sessionID, pin string) (*domain.USSDLeadDetails, error) {
					return nil, nil
				}
			}
			if tt.name == "Invalid:_unsuccessful_USSD" {
				fakeBaseExt.NormalizeMSISDNFn = func(msisdn string) (*string, error) {
					return nil, fmt.Errorf("empty phone number")
				}
			}
			svr := h.IncomingUSSDHandler()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}

			dataResponse, err := ioutil.ReadAll(response.Body)
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
}

func TestHandlers_PollServices(t *testing.T) {
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to initialize onboarding interactor: %v", err)
		return
	}
	h := rest.NewHandlersInterfaces(i)

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
			name: "return handler func",
			args: args{
				url:        fmt.Sprintf("%s/poll_services", serverUrl),
				httpMethod: http.MethodGet,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request to pass to our handler.
			req, err := http.NewRequest(tt.args.httpMethod, tt.args.url, tt.args.body)
			if err != nil {
				t.Errorf("can't create new request: %v", err)
				return
			}
			// We create a ResponseRecorder to record the response.
			response := httptest.NewRecorder()

			/*
				Some Mockery
			*/

			// call its ServeHTTP method and pass in our Request and ResponseRecorder.
			svr := h.PollServices()
			svr.ServeHTTP(response, req)

			if tt.wantStatus != response.Code {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
				return
			}
		})
	}
}
