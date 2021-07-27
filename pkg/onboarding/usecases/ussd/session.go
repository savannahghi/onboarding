package ussd

import (
	"context"

	"github.com/savannahghi/converterandformatter"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/exceptions"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/utils"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/profileutils"
)

// UpdateSessionLevel updates user current level of interaction with USSD
func (u *Impl) UpdateSessionLevel(ctx context.Context, level int, sessionID string) error {
	ctx, span := tracer.Start(ctx, "UpdateSessionLevel")
	defer span.End()

	validSessionID, err := utils.CheckEmptyString(sessionID)
	if err != nil {
		utils.RecordSpanError(span, err)
		return err
	}

	_, err = u.onboardingRepository.UpdateSessionLevel(ctx, *validSessionID, level)
	if err != nil {
		utils.RecordSpanError(span, err)
		return err
	}
	return nil

}

// UpdateSessionPIN updates user current session PIN
func (u *Impl) UpdateSessionPIN(ctx context.Context, pin string, sessionID string) (*domain.USSDLeadDetails, error) {
	ctx, span := tracer.Start(ctx, "UpdateSessionLevel")
	defer span.End()

	validSessionID, err := utils.CheckEmptyString(sessionID)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	ussdLead, err := u.onboardingRepository.UpdateSessionPIN(ctx, *validSessionID, pin)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}
	return ussdLead, nil

}

// GetUserProfileByPrimaryPhoneNumber ...
func (u *Impl) GetUserProfileByPrimaryPhoneNumber(ctx context.Context, phoneNumber string, suspend bool) (*profileutils.UserProfile, error) {
	profile, err := u.onboardingRepository.GetUserProfileByPrimaryPhoneNumber(ctx, phoneNumber, false)
	if err != nil {
		return nil, err
	}
	return profile, err
}

// StageCRMPayload ...
func (u *Impl) StageCRMPayload(ctx context.Context, payload *dto.ContactLeadInput) error {
	err := u.onboardingRepository.StageCRMPayload(ctx, payload)
	if err != nil {
		return err
	}
	return nil
}

//AddAITSessionDetails persists USSD details
func (u *Impl) AddAITSessionDetails(ctx context.Context, input *dto.SessionDetails) (*domain.USSDLeadDetails, error) {
	ctx, span := tracer.Start(ctx, "AddAITSessionDetails")
	defer span.End()

	phone, err := converterandformatter.NormalizeMSISDN(*input.PhoneNumber)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.NormalizeMSISDNError(err)
	}
	sessionDetails := &dto.SessionDetails{
		PhoneNumber: phone,
		SessionID:   input.SessionID,
		Level:       input.Level,
	}
	result, err := u.onboardingRepository.AddAITSessionDetails(ctx, sessionDetails)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}
	return result, nil
}

// GetOrCreateSessionState is used to set or return a user session
func (u *Impl) GetOrCreateSessionState(ctx context.Context, payload *dto.SessionDetails) (*domain.USSDLeadDetails, error) {
	ctx, span := tracer.Start(ctx, "GetOrCreateSessionState")
	defer span.End()

	sessionDetails, err := u.onboardingRepository.GetAITSessionDetails(ctx, payload.SessionID)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}
	if sessionDetails == nil {
		payload.Level = 0
		sessionDetails, err = u.AddAITSessionDetails(ctx, payload)
		if err != nil {
			utils.RecordSpanError(span, err)
			return nil, err
		}
	}
	return sessionDetails, nil
}

// GetOrCreatePhoneNumberUser ...
func (u *Impl) GetOrCreatePhoneNumberUser(ctx context.Context, phone string) (*dto.CreatedUserResponse, error) {
	ctx, span := tracer.Start(ctx, "GetOrCreatePhoneNumberUser")
	defer span.End()

	user, err := u.onboardingRepository.GetOrCreatePhoneNumberUser(ctx, phone)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	return user, nil
}

// CreateUserProfile ...
func (u *Impl) CreateUserProfile(ctx context.Context, phoneNumber string, uid string) (*profileutils.UserProfile, error) {
	ctx, span := tracer.Start(ctx, "GetOrCreatePhoneNumberUser")
	defer span.End()

	userProfile, err := u.onboardingRepository.CreateUserProfile(ctx, phoneNumber, uid)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	return userProfile, nil
}

// CreateEmptyCustomerProfile ...
func (u *Impl) CreateEmptyCustomerProfile(ctx context.Context, profileID string) (*profileutils.Customer, error) {
	ctx, span := tracer.Start(ctx, "GetOrCreatePhoneNumberUser")
	defer span.End()

	customer, err := u.onboardingRepository.CreateEmptyCustomerProfile(ctx, profileID)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	return customer, nil
}

// UpdateBioData ...
func (u *Impl) UpdateBioData(ctx context.Context, id string, data profileutils.BioData) error {
	ctx, span := tracer.Start(ctx, "GetOrCreatePhoneNumberUser")
	defer span.End()

	validID, err := utils.CheckEmptyString(id)
	if err != nil {
		return err
	}

	err = u.onboardingRepository.UpdateBioData(ctx, *validID, data)
	if err != nil {
		utils.RecordSpanError(span, err)
		return err
	}

	return nil
}

// RemoveUserByPhoneNumber ...
func (u *Impl) RemoveUserByPhoneNumber(ctx context.Context, phone string) error {
	ctx, span := tracer.Start(ctx, "PurgeUserByPhoneNumber")
	defer span.End()
	err := u.signUp.RemoveUserByPhoneNumber(ctx, phone)
	if err != nil {
		utils.RecordSpanError(span, err)
		return err
	}
	return nil
}

// SetUserPIN ...
func (u *Impl) SetUserPIN(ctx context.Context, pin string, profileID string) (bool, error) {
	ctx, span := tracer.Start(ctx, "SetUserPIN")
	defer span.End()
	setPIN, err := u.pinUsecase.SetUserPIN(ctx, pin, profileID)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, err
	}
	return setPIN, nil
}

// SaveUSSDEvent saves USSD event in the database
func (u *Impl) SaveUSSDEvent(ctx context.Context, input *dto.USSDEvent) (*dto.USSDEvent, error) {
	return u.onboardingRepository.SaveUSSDEvent(ctx, input)
}
