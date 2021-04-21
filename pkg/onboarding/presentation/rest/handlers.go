package rest

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gorilla/mux"
	"gitlab.slade360emr.com/go/profile/pkg/onboarding/application/resources"
	"gitlab.slade360emr.com/go/profile/pkg/onboarding/application/utils"
	"gitlab.slade360emr.com/go/profile/pkg/onboarding/presentation/interactor"

	"gitlab.slade360emr.com/go/base"
)

// HandlersInterfaces represents all the REST API logic
type HandlersInterfaces interface {
	VerifySignUpPhoneNumber(ctx context.Context) http.HandlerFunc
	CreateUserWithPhoneNumber(ctx context.Context) http.HandlerFunc
	UserRecoveryPhoneNumbers(ctx context.Context) http.HandlerFunc
	SetPrimaryPhoneNumber(ctx context.Context) http.HandlerFunc
	LoginByPhone(ctx context.Context) http.HandlerFunc
	LoginAnonymous(ctx context.Context) http.HandlerFunc
	RequestPINReset(ctx context.Context) http.HandlerFunc
	ResetPin(ctx context.Context) http.HandlerFunc
	SendOTP(ctx context.Context) http.HandlerFunc
	SendRetryOTP(ctx context.Context) http.HandlerFunc
	RefreshToken(ctx context.Context) http.HandlerFunc
	FindSupplierByUID(ctx context.Context) http.HandlerFunc
	RemoveUserByPhoneNumber(ctx context.Context) http.HandlerFunc
	GetUserProfileByUID(ctx context.Context) http.HandlerFunc
	UpdateCovers(ctx context.Context) http.HandlerFunc
	ProfileAttributes(ctx context.Context) http.HandlerFunc
	RegisterPushToken(ctx context.Context) http.HandlerFunc
	AddAdminPermsToUser(ctx context.Context) http.HandlerFunc
}

// HandlersInterfacesImpl represents the usecase implementation object
type HandlersInterfacesImpl struct {
	interactor *interactor.Interactor
}

// NewHandlersInterfaces initializes a new rest handlers usecase
func NewHandlersInterfaces(i *interactor.Interactor) HandlersInterfaces {
	return &HandlersInterfacesImpl{i}
}

// VerifySignUpPhoneNumber is an unauthenticated endpoint that does a
// check on the supplied phone number asserting whether the phone is associated with
// a user profile. It check both the PRIMARY PHONE and SECONDARY PHONE NUMBER.
// If the phone number does not exist, it sends the OTP to the phone number
func (h *HandlersInterfacesImpl) VerifySignUpPhoneNumber(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := &resources.PhoneNumberPayload{}
		base.DecodeJSONToTargetStruct(w, r, p)
		if p.PhoneNumber == nil {
			err := fmt.Errorf("expected `phoneNumber` to be defined")
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}

		otpResp, err := h.interactor.Signup.VerifyPhoneNumber(ctx, *p.PhoneNumber)
		if err != nil {
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}
		base.WriteJSONResponse(w, otpResp, http.StatusOK)
	}
}

// CreateUserWithPhoneNumber is an unauthenticated endpoint that is called to create
func (h *HandlersInterfacesImpl) CreateUserWithPhoneNumber(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := &resources.SignUpInput{}
		base.DecodeJSONToTargetStruct(w, r, p)

		response, err := h.interactor.Signup.CreateUserByPhone(ctx, p)
		if err != nil {
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}

		base.WriteJSONResponse(w, response, http.StatusCreated)
	}
}

// UserRecoveryPhoneNumbers fetches the phone numbers associated with a profile for the purpose of account recovery.
// The returned phone numbers slice should be masked. E.G +254700***123
func (h *HandlersInterfacesImpl) UserRecoveryPhoneNumbers(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		p := &resources.PhoneNumberPayload{}
		base.DecodeJSONToTargetStruct(w, r, p)
		if p.PhoneNumber == nil {
			err := fmt.Errorf("expected `phoneNumber` to be defined")
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		response, err := h.interactor.Signup.GetUserRecoveryPhoneNumbers(
			ctx,
			*p.PhoneNumber,
		)

		if err != nil {
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}
		base.WriteJSONResponse(w, response, http.StatusOK)
	}
}

// SetPrimaryPhoneNumber sets the provided phone number as the primary phone of the profile associated with it
func (h *HandlersInterfacesImpl) SetPrimaryPhoneNumber(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		p := &resources.SetPrimaryPhoneNumberPayload{}
		base.DecodeJSONToTargetStruct(w, r, p)
		if p.PhoneNumber == nil || p.OTP == nil {
			err := fmt.Errorf("expected `phoneNumber` and `otp` to be defined")
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		response, err := h.interactor.Signup.SetPhoneAsPrimary(
			ctx,
			*p.PhoneNumber,
			*p.OTP,
		)

		if err != nil {
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}
		base.WriteJSONResponse(w, resources.NewOKResp(response), http.StatusOK)
	}
}

// LoginByPhone is an unauthenticated endpoint that:
// Collects a phonenumber and pin from the user and checks if the phonenumber
// is an existing PRIMARY PHONENUMBER. If it does then it fetches the PIN that
// belongs to the profile and returns auth credentials to allow the user to login
func (h *HandlersInterfacesImpl) LoginByPhone(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := &resources.LoginPayload{}
		base.DecodeJSONToTargetStruct(w, r, p)
		if p.PhoneNumber == nil || p.PIN == nil {
			err := fmt.Errorf("expected `phoneNumber`, `pin` to be defined")
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		if !p.Flavour.IsValid() {
			err := fmt.Errorf("an invalid `flavour` defined")
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		response, err := h.interactor.Login.LoginByPhone(
			ctx,
			*p.PhoneNumber,
			*p.PIN,
			p.Flavour,
		)
		if err != nil {
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}

		base.WriteJSONResponse(w, response, http.StatusOK)
	}
}

// LoginAnonymous is an unauthenticated endpoint that returns only auth credentials for anonymous users
func (h *HandlersInterfacesImpl) LoginAnonymous(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := &resources.LoginPayload{}
		base.DecodeJSONToTargetStruct(w, r, p)
		if p.Flavour.String() == "" {
			err := fmt.Errorf("expected `flavour` to be defined")
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		if !p.Flavour.IsValid() || p.Flavour != base.FlavourConsumer {
			err := fmt.Errorf("an invalid `flavour` defined")
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		response, err := h.interactor.Login.LoginAsAnonymous(ctx)
		if err != nil {
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}

		base.WriteJSONResponse(w, response, http.StatusOK)
	}
}

// RequestPINReset is an unauthenticated request that takes in a phone number
// sends an otp to an msisdn that requests a PIN reset request during login
func (h *HandlersInterfacesImpl) RequestPINReset(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := &resources.PhoneNumberPayload{}
		base.DecodeJSONToTargetStruct(w, r, p)
		if p.PhoneNumber == nil {
			err := fmt.Errorf("expected `phoneNumber` to be defined")
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		otpResp, err := h.interactor.UserPIN.RequestPINReset(ctx, *p.PhoneNumber)
		if err != nil {
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}
		base.WriteJSONResponse(w, otpResp, http.StatusOK)
	}
}

// ResetPin used to change/update a user's PIN
func (h *HandlersInterfacesImpl) ResetPin(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pin := &resources.ChangePINRequest{}
		base.DecodeJSONToTargetStruct(w, r, pin)
		if pin.PhoneNumber == "" || pin.PIN == "" || pin.OTP == "" {
			err := fmt.Errorf(
				"expected `phoneNumber`, `PIN` to be defined, `OTP` to be defined")
			base.WriteJSONResponse(
				w,
				base.CustomError{
					Err:     err,
					Message: err.Error(),
				},
				http.StatusBadRequest,
			)
			return
		}

		response, err := h.interactor.UserPIN.ResetUserPIN(
			ctx,
			pin.PhoneNumber,
			pin.PIN,
			pin.OTP,
		)
		if err != nil {
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}

		base.WriteJSONResponse(w, response, http.StatusCreated)
	}
}

// SendOTP is an unauthenticated request that takes in a phone number
// and generates an OTP and sends a valid OTP to the phone number. This API will mostly be used
// during account recovery workflow
func (h *HandlersInterfacesImpl) SendOTP(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := &resources.PhoneNumberPayload{}
		base.DecodeJSONToTargetStruct(w, r, payload)
		if payload.PhoneNumber == nil {
			err := fmt.Errorf("expected `phoneNumber` to be defined")
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		response, err := h.interactor.Otp.GenerateAndSendOTP(
			ctx,
			*payload.PhoneNumber,
		)
		if err != nil {
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}

		base.WriteJSONResponse(w, response, http.StatusOK)
	}
}

// SendRetryOTP is an unauthenticated request that takes in a phone number
// and a retry step (1 for sending an OTP via WhatsApp and 2 for Twilio Messages)
// and generates and sends a valid OTP to the phone number
func (h *HandlersInterfacesImpl) SendRetryOTP(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		retryPayload := &resources.SendRetryOTPPayload{}
		base.DecodeJSONToTargetStruct(w, r, retryPayload)
		if retryPayload.Phone == nil || retryPayload.RetryStep == nil {
			err := fmt.Errorf("expected `phoneNumber`, `retryStep` to be defined")
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		response, err := h.interactor.Otp.SendRetryOTP(
			ctx,
			*retryPayload.Phone,
			*retryPayload.RetryStep,
		)
		if err != nil {
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}

		base.WriteJSONResponse(w, response, http.StatusOK)
	}
}

// RefreshToken is an unauthenticated endpoint that
// takes a custom Firebase refresh token and tries to fetch
// an ID token and returns auth credentials if successful
// Otherwise, an error is returned
func (h *HandlersInterfacesImpl) RefreshToken(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := &resources.RefreshTokenPayload{}
		base.DecodeJSONToTargetStruct(w, r, p)
		if p.RefreshToken == nil {
			err := fmt.Errorf("expected `refreshToken` to be defined")
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		response, err := h.interactor.Login.RefreshToken(*p.RefreshToken)
		if err != nil {
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}

		base.WriteJSONResponse(w, response, http.StatusOK)
	}
}

// FindSupplierByUID fetch supplier profile via REST
func (h *HandlersInterfacesImpl) FindSupplierByUID(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := utils.ValidateUID(w, r)
		if err != nil {
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		if s.UID == nil {
			err := fmt.Errorf("expected `uid` to be defined")
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		var supplier *base.Supplier
		authCred := &auth.Token{UID: *s.UID}
		newContext := context.WithValue(ctx, base.AuthTokenContextKey, authCred)
		supplier, err = h.interactor.Supplier.FindSupplierByUID(newContext)
		log.Printf("the supplier is %v", supplier)
		log.Printf("the err is %v", err)
		if supplier == nil || err != nil {
			err := fmt.Errorf("supplier profile not found")
			base.WriteJSONResponse(w, err, http.StatusNotFound)
			return
		}

		base.WriteJSONResponse(w, supplier, http.StatusOK)
	}
}

// RemoveUserByPhoneNumber is an unauthenticated endpoint that removes a user
// whose phone number, either PRIMARY PHONE NUMBER or SECONDARY PHONE NUMBERS,matches the provided
// phone number in the request. This endpoint will ONLY be available under testing environment
func (h *HandlersInterfacesImpl) RemoveUserByPhoneNumber(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		p := &resources.PhoneNumberPayload{}
		base.DecodeJSONToTargetStruct(w, r, p)
		if p.PhoneNumber == nil {
			err := fmt.Errorf("expected `phoneNumber` to be defined")
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}
		v, err := h.interactor.Onboarding.CheckPhoneExists(ctx, *p.PhoneNumber)
		if err != nil {
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		if v {
			if err := h.interactor.Signup.RemoveUserByPhoneNumber(ctx, *p.PhoneNumber); err != nil {
				base.WriteJSONResponse(w, base.CustomError{
					Err:     err,
					Message: err.Error(),
				}, http.StatusBadRequest)
				return
			}
			base.WriteJSONResponse(w, resources.OKResp{Status: "OK"}, http.StatusOK)
			return
		}
		err = fmt.Errorf("`phoneNumber` does not exist and not associated with any user ")
		base.WriteJSONResponse(w, base.CustomError{
			Err:     err,
			Message: err.Error(),
		}, http.StatusBadRequest)
	}
}

// GetUserProfileByUID fetches and returns a user profile via REST ISC
func (h *HandlersInterfacesImpl) GetUserProfileByUID(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := &resources.UIDPayload{}
		base.DecodeJSONToTargetStruct(w, r, p)
		if p.UID == nil {
			err := fmt.Errorf("expected `UID` to be defined")
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}

		profile, err := h.interactor.Onboarding.GetUserProfileByUID(ctx, *p.UID)
		if err != nil {
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}

		base.WriteJSONResponse(w, profile, http.StatusOK)
	}
}

// RegisterPushToken adds a new push token in the users profile if the push token does not exist
// via REST ISC
func (h *HandlersInterfacesImpl) RegisterPushToken(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := &resources.PushTokenPayload{}
		base.DecodeJSONToTargetStruct(w, r, t)
		if t.PushToken == "" || t.UID == "" {
			err := fmt.Errorf("expected `PushToken` or `UID` to be defined")
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}
		authCred := &auth.Token{UID: t.UID}
		newContext := context.WithValue(ctx, base.AuthTokenContextKey, authCred)
		profile, err := h.interactor.Signup.RegisterPushToken(newContext, t.PushToken)
		if err != nil {
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}

		base.WriteJSONResponse(w, profile, http.StatusOK)
	}
}

// UpdateCovers is an unauthenticated ISC endpoint that updates the cover of
// a given user given their UID and cover details
func (h *HandlersInterfacesImpl) UpdateCovers(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := &resources.UpdateCoversPayload{}
		base.DecodeJSONToTargetStruct(w, r, p)
		if p.UID == nil {
			err := fmt.Errorf("expected `UID` to be defined")
			base.ReportErr(w, err, http.StatusBadRequest)
			return
		}

		if p.BeneficiaryID == nil {
			err := fmt.Errorf("expected `BeneficiaryID` to be defined")
			base.ReportErr(w, err, http.StatusBadRequest)
			return
		}

		if p.EffectivePolicyNumber == nil {
			err := fmt.Errorf("expected `EffectivePolicyNumber` to be defined")
			base.ReportErr(w, err, http.StatusBadRequest)
			return
		}

		if p.ValidFrom == nil {
			err := fmt.Errorf("expected `ValidFrom` to be defined")
			base.ReportErr(w, err, http.StatusBadRequest)
			return
		}

		if p.ValidTo == nil {
			err := fmt.Errorf("expected `ValidTo` to be defined")
			base.ReportErr(w, err, http.StatusBadRequest)
			return
		}

		auth := &auth.Token{UID: *p.UID}
		newContext := context.WithValue(ctx, base.AuthTokenContextKey, auth)
		cover := base.Cover{
			PayerName:             *p.PayerName,
			MemberNumber:          *p.MemberNumber,
			MemberName:            *p.MemberName,
			PayerSladeCode:        *p.PayerSladeCode,
			BeneficiaryID:         *p.BeneficiaryID,
			EffectivePolicyNumber: *p.EffectivePolicyNumber,
			ValidFrom:             *p.ValidFrom,
			ValidTo:               *p.ValidTo,
		}
		var covers []base.Cover
		covers = append(covers, cover)
		err := h.interactor.Onboarding.UpdateCovers(newContext, covers)
		if err != nil {
			base.ReportErr(w, err, http.StatusBadRequest)
			return
		}

		base.WriteJSONResponse(
			w,
			resources.OKResp{
				Status: "Covers successfully updated",
			},
			http.StatusOK,
		)
	}
}

// ProfileAttributes retreives confirmed user profile attributes.
// These attributes include a user's verified phone numbers, verfied emails
// and verified FCM push tokens
func (h *HandlersInterfacesImpl) ProfileAttributes(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		attribute, found := vars["attribute"]
		if !found {
			err := fmt.Errorf("request does not have a path var named `%s`",
				attribute,
			)
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}

		p := &resources.UIDsPayload{}
		base.DecodeJSONToTargetStruct(w, r, p)
		if len(p.UIDs) == 0 {
			err := fmt.Errorf("expected a `UID` to be defined")
			base.WriteJSONResponse(w, err, http.StatusBadRequest)
			return
		}

		output, err := h.interactor.Onboarding.ProfileAttributes(
			ctx,
			p.UIDs,
			attribute,
		)
		if err != nil {
			base.ReportErr(w, err, http.StatusBadRequest)
			return
		}

		base.WriteJSONResponse(w, output, http.StatusOK)
	}
}

// AddAdminPermsToUser is authenticated endpoint that adds admin permissions to a
// whose phone number, either PRIMARY PHONE NUMBER or SECONDARY PHONE NUMBERS,matches
// the provided phone number in the request.
func (h *HandlersInterfacesImpl) AddAdminPermsToUser(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		p := &resources.PhoneNumberPayload{}
		base.DecodeJSONToTargetStruct(w, r, p)
		if p.PhoneNumber == nil {
			err := fmt.Errorf("expected `phoneNumber` to be defined")
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}
		v, err := h.interactor.Onboarding.CheckPhoneExists(ctx, *p.PhoneNumber)
		if err != nil {
			base.WriteJSONResponse(w, base.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		if v {
			if err := h.interactor.Onboarding.AddAdminPermsToUser(ctx, *p.PhoneNumber); err != nil {
				base.WriteJSONResponse(w, base.CustomError{
					Err:     err,
					Message: err.Error(),
				}, http.StatusBadRequest)
				return
			}
			base.WriteJSONResponse(w, resources.OKResp{Status: "OK"}, http.StatusOK)
			return
		}
		err = fmt.Errorf("`phoneNumber` does not exist and not associated with any user ")
		base.WriteJSONResponse(w, base.CustomError{
			Err:     err,
			Message: err.Error(),
		}, http.StatusBadRequest)
	}
}
