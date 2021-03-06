package usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/savannahghi/errorcodeutil"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/exceptions"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/utils"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure"
	"github.com/savannahghi/profileutils"
)

// UserPINUseCases represents all the business logic that touch on user PIN Management
type UserPINUseCases interface {
	SetUserPIN(ctx context.Context, pin string, profileID string) (bool, error)
	// SetUserTempPIN is used to set a temporary PIN for a created user.
	SetUserTempPIN(ctx context.Context, profileID string) (string, error)
	ResetUserPIN(
		ctx context.Context,
		phone string,
		PIN string,
		OTP string,
	) (bool, error)
	ChangeUserPIN(ctx context.Context, phone string, pin string) (bool, error)
	RequestPINReset(ctx context.Context, phone string, appID *string) (*profileutils.OtpResponse, error)
	CheckHasPIN(ctx context.Context, profileID string) (bool, error)
}

// UserPinUseCaseImpl represents usecase implementation object
type UserPinUseCaseImpl struct {
	infrastructure  infrastructure.Infrastructure
	profileUseCases ProfileUseCase
	baseExt         extension.BaseExtension
	pinExt          extension.PINExtension
}

// NewUserPinUseCase returns a new UserPin usecase
func NewUserPinUseCase(
	infrastructure infrastructure.Infrastructure,
	p ProfileUseCase,
	ext extension.BaseExtension,
	pin extension.PINExtension,
) UserPINUseCases {
	return &UserPinUseCaseImpl{
		infrastructure:  infrastructure,
		profileUseCases: p,
		baseExt:         ext,
		pinExt:          pin,
	}
}

// SetUserPIN receives phone number and pin from phonenumber sign up
func (u *UserPinUseCaseImpl) SetUserPIN(
	ctx context.Context,
	pin string,
	profileID string,
) (bool, error) {
	ctx, span := tracer.Start(ctx, "SetUserPIN")
	defer span.End()

	if err := extension.ValidatePINLength(pin); err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.ValidatePINLengthError(err)
	}

	if err := extension.ValidatePINDigits(pin); err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.ValidatePINDigitsError(err)
	}

	// EncryptPIN the PIN
	salt, encryptedPin := u.pinExt.EncryptPIN(pin, nil)

	pinPayload := &domain.PIN{
		ID:        uuid.New().String(),
		ProfileID: profileID,
		PINNumber: encryptedPin,
		Salt:      salt,
	}
	if _, err := u.infrastructure.Database.SavePIN(ctx, pinPayload); err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.SaveUserPinError(err)
	}

	return true, nil
}

// RequestPINReset sends a request given an existing user's phone number,
// sends an otp to the phone number that is then used in the process of
// updating their old PIN to a new one
func (u *UserPinUseCaseImpl) RequestPINReset(
	ctx context.Context,
	phone string,
	appID *string,
) (*profileutils.OtpResponse, error) {
	ctx, span := tracer.Start(ctx, "RequestPINReset")
	defer span.End()

	phoneNumber, err := u.baseExt.NormalizeMSISDN(phone)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.NormalizeMSISDNError(err)
	}

	pr, err := u.infrastructure.Database.GetUserProfileByPrimaryPhoneNumber(ctx, *phoneNumber, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return nil, err
	}

	exists, err := u.CheckHasPIN(ctx, pr.ID)
	if !exists {
		return nil, exceptions.ExistingPINError(err)
	}
	// generate and send otp to the phone number
	otpResp, err := u.infrastructure.Engagement.GenerateAndSendOTP(ctx, phone, appID)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.GenerateAndSendOTPError(err)
	}

	return otpResp, nil
}

// ResetUserPIN resets a user's PIN with the newly supplied PIN
func (u *UserPinUseCaseImpl) ResetUserPIN(
	ctx context.Context,
	phone string,
	PIN string,
	OTP string,
) (bool, error) {
	ctx, span := tracer.Start(ctx, "ResetUserPIN")
	defer span.End()

	phoneNumber, err := u.baseExt.NormalizeMSISDN(phone)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.NormalizeMSISDNError(err)
	}

	verified, err := u.infrastructure.Engagement.VerifyOTP(ctx, phone, OTP)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.VerifyOTPError(err)
	}

	if !verified {
		return false, exceptions.VerifyOTPError(nil)
	}

	profile, err := u.infrastructure.Database.GetUserProfileByPrimaryPhoneNumber(
		ctx,
		*phoneNumber,
		false,
	)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return false, err
	}

	_, err = u.CheckHasPIN(ctx, profile.ID)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.EncryptPINError(err)
	}

	// EncryptPIN the PIN
	salt, encryptedPin := u.pinExt.EncryptPIN(PIN, nil)

	pinPayload := &domain.PIN{
		ID:        uuid.New().String(),
		ProfileID: profile.ID,
		PINNumber: encryptedPin,
		Salt:      salt,
	}
	_, err = u.infrastructure.Database.UpdatePIN(ctx, profile.ID, pinPayload)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.InternalServerError(err)
	}
	return true, nil
}

// ChangeUserPIN updates authenticated user's pin with the newly supplied pin
func (u *UserPinUseCaseImpl) ChangeUserPIN(
	ctx context.Context,
	phone string,
	pin string,
) (bool, error) {
	ctx, span := tracer.Start(ctx, "ChangeUserPIN")
	defer span.End()

	phoneNumber, err := u.baseExt.NormalizeMSISDN(phone)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.NormalizeMSISDNError(err)
	}

	profile, err := u.infrastructure.Database.GetUserProfileByPrimaryPhoneNumber(
		ctx,
		*phoneNumber,
		false,
	)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return false, err
	}

	_, err = u.CheckHasPIN(ctx, profile.ID)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.EncryptPINError(err)
	}

	salt, encryptedPin := u.pinExt.EncryptPIN(pin, nil)

	pinPayload := &domain.PIN{
		ID:        uuid.New().String(),
		ProfileID: profile.ID,
		PINNumber: encryptedPin,
		Salt:      salt,
	}
	_, err = u.infrastructure.Database.UpdatePIN(ctx, profile.ID, pinPayload)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.InternalServerError(err)
	}
	return true, nil
}

// CheckHasPIN given a phone number checks if the phonenumber is present in our collections
// which essentially means that the number has an already existing PIN
func (u *UserPinUseCaseImpl) CheckHasPIN(ctx context.Context, profileID string) (bool, error) {
	ctx, span := tracer.Start(ctx, "CheckHasPIN")
	defer span.End()

	pinData, err := u.infrastructure.Database.GetPINByProfileID(ctx, profileID)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, err
	}

	if pinData == nil {
		return false, fmt.Errorf("%v", errorcodeutil.PINNotFound)
	}

	return true, nil
}

// SetUserTempPIN generates a random one time pin.
// The pin acts as a temporary PIN and should be changed by the user.
func (u *UserPinUseCaseImpl) SetUserTempPIN(ctx context.Context, profileID string) (string, error) {
	ctx, span := tracer.Start(ctx, "SetUserTempPIN")
	defer span.End()

	pin, err := u.pinExt.GenerateTempPIN(ctx)
	if err != nil {
		utils.RecordSpanError(span, err)
		return "", exceptions.GeneratePinError(err)
	}

	// Encrypt the PIN
	salt, encryptedPin := u.pinExt.EncryptPIN(pin, nil)

	pinPayload := &domain.PIN{
		ID:        uuid.New().String(),
		ProfileID: profileID,
		PINNumber: encryptedPin,
		Salt:      salt,
		IsOTP:     true,
	}
	if _, err := u.infrastructure.Database.SavePIN(ctx, pinPayload); err != nil {
		utils.RecordSpanError(span, err)
		return "", exceptions.SaveUserPinError(err)
	}

	return pin, nil
}
