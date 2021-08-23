package ussd

import (
	"context"
	"time"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/utils"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	hubspotDomain "gitlab.slade360emr.com/go/commontools/crm/pkg/domain"
)

const (
	//MarketingInput indicates users who want to opt out or opt in  to be send marketing sms(messages)
	MarketingInput = "1"

	// OptOutText ...
	OptOutText = "1. Opt out from marketing messages\r\n"

	// OptInText ...
	OptInText = "1. Opt in to marketing messages\r\n"
)

// WelcomeMenu represents  the default welcome submenu
func (u *Impl) WelcomeMenu(text string) string {
	resp := "CON Welcome to Be.Well\r\n"
	resp += text
	resp += "2. Change PIN"
	return resp
}

// ResetPinMenu ...
func (u *Impl) ResetPinMenu(text string) string {
	resp := "CON Your PIN was reset successfully.\r\n"
	resp += text
	resp += "2. Change PIN"
	return resp
}

// HandleHomeMenu represents the default home menu
func (u *Impl) HandleHomeMenu(ctx context.Context, level int, session *domain.USSDLeadDetails, userResponse string) string {
	ctx, span := tracer.Start(ctx, "HandleHomeMenu")
	defer span.End()

	OptedOut, err := u.crm.IsOptedOut(ctx, session.PhoneNumber)
	if err != nil {
		utils.RecordSpanError(span, err)
		return "END Something went wrong. Please try again."
	}

	if userResponse == EmptyInput || userResponse == GoBackHomeInput {
		if !OptedOut {
			return u.WelcomeMenu(OptOutText)
		}
		return u.WelcomeMenu(OptInText)

	} else if userResponse == MarketingInput {

		if !OptedOut {
			resp := u.OptOutOrOptIn(ctx, session, hubspotDomain.GeneralOptionTypeYes)
			return resp
		}

		resp := u.OptOutOrOptIn(ctx, session, hubspotDomain.GeneralOptionTypeNo)
		return resp
	} else if userResponse == ChangePINInput {
		err := u.UpdateSessionLevel(ctx, ChangeUserPINState, session.SessionID)
		if err != nil {
			utils.RecordSpanError(span, err)
			return "END Something went wrong. Please try again"
		}
		return u.HandleChangePIN(ctx, session, userResponse)

	} else {
		if !OptedOut {
			resp := "CON Invalid choice. Try again.\r\n"
			resp += "1. Opt out from marketing messages\r\n"
			resp += "2. Change PIN"
			return resp
		}
		resp := "CON Invalid choice. Try again.\r\n"
		resp += "1. Opt in from marketing messages\r\n"
		resp += "2. Change PIN"
		return resp
	}
}

//OptOutOrOptIn ...
func (u *Impl) OptOutOrOptIn(ctx context.Context, session *domain.USSDLeadDetails, text hubspotDomain.GeneralOptionType) string {
	ctx, span := tracer.Start(ctx, "OptOutOrOptIn")
	defer span.End()
	time := time.Now()

	_, err := u.crm.OptOutOrOptIn(ctx, session.PhoneNumber, text)
	if err != nil {
		utils.RecordSpanError(span, err)
		return "END Something went wrong. Please try again."
	}

	if _, err := u.onboardingRepository.SaveUSSDEvent(ctx, &dto.USSDEvent{
		SessionID:         session.SessionID,
		PhoneNumber:       session.PhoneNumber,
		USSDEventDateTime: &time,
		USSDEventName:     USSDOptOut,
	}); err != nil {
		return "END Something went wrong. Please try again."
	}

	if text == "YES" {
		resp := "CON We have successfully opted you\r\n"
		resp += "out of marketing messages\r\n"
		resp += "0. Go back home"
		return resp
	}
	resp := "CON We have successfully opted you\r\n"
	resp += "in to marketing messages\r\n"
	resp += "0. Go back home"
	return resp

}
