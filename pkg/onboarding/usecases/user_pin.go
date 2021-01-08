package usecases

import (
	"context"
	"fmt"

	"gitlab.slade360emr.com/go/profile/pkg/onboarding/application/resources"

	"github.com/google/uuid"
	"gitlab.slade360emr.com/go/base"
	"gitlab.slade360emr.com/go/profile/pkg/onboarding/application/exceptions"
	"gitlab.slade360emr.com/go/profile/pkg/onboarding/application/utils"
	"gitlab.slade360emr.com/go/profile/pkg/onboarding/domain"
	"gitlab.slade360emr.com/go/profile/pkg/onboarding/infrastructure/services/otp"
	"gitlab.slade360emr.com/go/profile/pkg/onboarding/repository"
)

// UserPINUseCases represents all the business logic that touch on user PIN Management
type UserPINUseCases interface {
	SetUserPIN(ctx context.Context, pin string, phone string) (bool, error)
	ResetUserPIN(
		ctx context.Context,
		phone string,
		PIN string,
		OTP string,
	) (*resources.PINOutput, error)
	ChangeUserPIN(ctx context.Context, phone string, pin string) (*resources.PINOutput, error)
	RequestPINReset(ctx context.Context, phone string) (*resources.OtpResponse, error)
}

// UserPinUseCaseImpl represents usecase implementation object
type UserPinUseCaseImpl struct {
	onboardingRepository repository.OnboardingRepository
	otpUseCases          otp.ServiceOTP
	profileUseCases      ProfileUseCase
}

// NewUserPinUseCase returns a new UserPin usecase
func NewUserPinUseCase(r repository.OnboardingRepository, otp otp.ServiceOTP, p ProfileUseCase) UserPINUseCases {
	return &UserPinUseCaseImpl{
		onboardingRepository: r,
		otpUseCases:          otp,
		profileUseCases:      p,
	}
}

// SetUserPIN receives phone number and pin from phonenumber sign up
func (u *UserPinUseCaseImpl) SetUserPIN(ctx context.Context, pin string, phone string) (bool, error) {
	phoneNumber, err := base.NormalizeMSISDN(phone)
	if err != nil {
		return false, exceptions.NormalizeMSISDNError(err)
	}

	if err := utils.ValidatePINLength(pin); err != nil {
		return false, err
	}

	if err = utils.ValidatePINDigits(pin); err != nil {
		return false, err
	}

	pr, err := u.onboardingRepository.GetUserProfileByPrimaryPhoneNumber(ctx, *phoneNumber)
	if err != nil {
		return false, exceptions.ProfileNotFoundError(err)
	}
	// EncryptPIN the PIN
	salt, encryptedPin := utils.EncryptPIN(pin, nil)

	pinPayload := &domain.PIN{
		ID:        uuid.New().String(),
		ProfileID: pr.ID,
		PINNumber: encryptedPin,
		Salt:      salt,
	}
	if _, err := u.onboardingRepository.SavePIN(ctx, pinPayload); err != nil {
		return false, fmt.Errorf("unable to save user PIN: %v", err)
	}

	return true, nil
}

// RequestPINReset sends a request given an existing user's phone number,
// sends an otp to the phone number that is then used in the process of
// updating their old PIN to a new one
func (u *UserPinUseCaseImpl) RequestPINReset(ctx context.Context, phone string) (*resources.OtpResponse, error) {
	phoneNumber, err := base.NormalizeMSISDN(phone)
	if err != nil {
		return nil, exceptions.NormalizeMSISDNError(err)
	}

	pr, err := u.onboardingRepository.GetUserProfileByPrimaryPhoneNumber(ctx, *phoneNumber)
	if err != nil {
		return nil, exceptions.ProfileNotFoundError(err)
	}

	exists, err := u.CheckHasPIN(ctx, pr.ID)
	if err != nil {
		return nil, exceptions.CheckUserPINError(err)
	}
	if !exists {
		return nil, exceptions.ExistingPINError(err)
	}
	fmt.Println("Tumefika apa")
	// generate and send otp to the phone number
	otpResp, err := u.otpUseCases.GenerateAndSendOTP(ctx, phone)
	if err != nil {
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
) (*resources.PINOutput, error) {
	phoneNumber, err := base.NormalizeMSISDN(phone)
	if err != nil {
		return nil, exceptions.NormalizeMSISDNError(err)
	}

	verified, err := u.otpUseCases.VerifyOTP(ctx, phone, OTP)
	if err != nil {
		return nil, exceptions.VerifyOTPError(err)
	}

	if !verified {
		return nil, exceptions.VerifyOTPError(nil)
	}

	profile, err := u.onboardingRepository.GetUserProfileByPrimaryPhoneNumber(ctx, *phoneNumber)
	if err != nil {
		return nil, exceptions.ProfileNotFoundError(err)
	}

	exists, err := u.CheckHasPIN(ctx, profile.ID)
	if !exists {
		return nil, exceptions.ExistingPINError(err)
	}

	// EncryptPIN the PIN
	salt, encryptedPin := utils.EncryptPIN(PIN, nil)
	if err != nil {
		return nil, exceptions.EncryptPINError(err)
	}

	pinPayload := &domain.PIN{
		ID:        uuid.New().String(),
		ProfileID: profile.ID,
		PINNumber: encryptedPin,
		Salt:      salt,
	}
	createdPin, err := u.onboardingRepository.UpdatePIN(ctx, profile.ID, pinPayload)
	if err != nil {
		return nil, exceptions.EncryptPINError(err)
	}
	return &resources.PINOutput{
		ProfileID: createdPin.ProfileID,
		PINNumber: createdPin.PINNumber,
	}, nil
}

// ChangeUserPIN updates authenticated user's pin with the newly supplied pin
func (u *UserPinUseCaseImpl) ChangeUserPIN(ctx context.Context, phone string, pin string) (*resources.PINOutput, error) {
	phoneNumber, err := base.NormalizeMSISDN(phone)
	if err != nil {
		return nil, exceptions.NormalizeMSISDNError(err)
	}

	profile, err := u.onboardingRepository.GetUserProfileByPrimaryPhoneNumber(ctx, *phoneNumber)
	if err != nil {
		return nil, exceptions.ProfileNotFoundError(err)
	}

	exists, err := u.CheckHasPIN(ctx, profile.ID)
	if !exists {
		return nil, exceptions.ExistingPINError(err)
	}

	// EncryptPIN the PIN
	salt, encryptedPin := utils.EncryptPIN(pin, nil)
	if err != nil {
		return nil, exceptions.EncryptPINError(err)
	}

	pinPayload := &domain.PIN{
		ID:        uuid.New().String(),
		ProfileID: profile.ID,
		PINNumber: encryptedPin,
		Salt:      salt,
	}
	createdPin, err := u.onboardingRepository.UpdatePIN(ctx, profile.ID, pinPayload)
	if err != nil {
		return nil, exceptions.EncryptPINError(err)
	}
	return &resources.PINOutput{
		ProfileID: createdPin.ProfileID,
		PINNumber: createdPin.PINNumber,
	}, nil
}

// CheckHasPIN given a phone number checks if the phonenumber is present in our collections
// which essentially means that the number has an already existing PIN
func (u *UserPinUseCaseImpl) CheckHasPIN(ctx context.Context, profileID string) (bool, error) {

	pinData, err := u.onboardingRepository.GetPINByProfileID(ctx, profileID)
	if err != nil {
		return false, err
	}

	if pinData == nil {
		return false, fmt.Errorf("%v", base.PINNotFound)
	}

	return true, nil
}
