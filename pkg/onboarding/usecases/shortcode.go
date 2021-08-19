package usecases

import (
	"context"
	"fmt"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/utils"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/crm"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/engagement"
	pubsubmessaging "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/pubsub"
	"github.com/savannahghi/onboarding/pkg/onboarding/repository"
	log "github.com/sirupsen/logrus"
	"gitlab.slade360emr.com/go/commontools/crm/pkg/domain"
)

const (
	// The right copy of replies will be availed later
	bewellUserReplyMessage = "We have received your request and one of our representatives will reach out to you. Thank you"
	newSenderReplyMessage  = "We have received your request and one of our representatives will reach out to you. Thanks you.\n\nDid you know you can now view your medical cover and benefits on Be.Well. To get started, Download Now https://bwl.mobi/1bvf"
)

// SMSUsecase represent the logic involved in receiving an SMS
type SMSUsecase interface {
	CreateSMSData(ctx context.Context, input *dto.AfricasTalkingMessage) error
}

//SMSImpl represents usecase implemention object
type SMSImpl struct {
	onboardingRepository repository.OnboardingRepository
	baseExt              extension.BaseExtension
	engagement           engagement.ServiceEngagement
	pubSub               pubsubmessaging.ServicePubSub
	hubspotCRM           crm.ServiceCrm
}

// NewSMSUsecase returns a new SMS usecase
func NewSMSUsecase(
	r repository.OnboardingRepository,
	ext extension.BaseExtension,
	engage engagement.ServiceEngagement,
	ps pubsubmessaging.ServicePubSub,
	crm crm.ServiceCrm,

) SMSUsecase {
	return &SMSImpl{
		onboardingRepository: r,
		baseExt:              ext,
		engagement:           engage,
		pubSub:               ps,
		hubspotCRM:           crm,
	}
}

// CreateSMSData saves and creates the SMS data of the message received
func (s *SMSImpl) CreateSMSData(ctx context.Context, input *dto.AfricasTalkingMessage) error {
	ctx, span := tracer.Start(ctx, "CreateSMSData")
	defer span.End()

	validatedInput, err := utils.ValidateAficasTalkingSMSData(input)
	if err != nil {
		utils.RecordSpanError(span, err)
		return err
	}

	// Checking if the sender has a profile
	profile, err := s.onboardingRepository.GetUserProfileByPhoneNumber(ctx, validatedInput.From, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		//should not panic when the user profile is not found
		log.Errorf("an error occurred: %v", err)
	}

	if profile != nil {
		// Replying to the customer without the download link and notifying the admin
		to := validatedInput.From
		message := bewellUserReplyMessage

		supportEmailPayload := &dto.EmailNotificationPayload{
			SubjectTitle: validatedInput.Text,
			EmailBody:    validatedInput.Text,
			PrimaryPhone: validatedInput.From,
			BeWellUser:   domain.GeneralOptionTypeYes,
			Time:         validatedInput.Date,
		}

		return s.replyAndNotifyAdmin(ctx, to, message, validatedInput, supportEmailPayload)
	}

	//Reply back with the download link for non-bewell user(new user)
	to := validatedInput.From
	message := newSenderReplyMessage

	supportEmailPayload := &dto.EmailNotificationPayload{
		SubjectTitle: validatedInput.Text,
		EmailBody:    validatedInput.Text,
		PrimaryPhone: validatedInput.From,
		BeWellUser:   domain.GeneralOptionTypeNo,
		Time:         validatedInput.Date,
	}

	err = s.replyAndNotifyAdmin(ctx, to, message, validatedInput, supportEmailPayload)
	if err != nil {
		return err
	}

	//Create CRM contact for the new user
	contact := &domain.CRMContact{
		Properties: domain.ContactProperties{
			Phone:                 validatedInput.From,
			Gender:                string(domain.GeneralOptionTypeNotGiven),
			FirstChannelOfContact: domain.ChannelOfContactShortcode,
			BeWellEnrolled:        domain.GeneralOptionTypeNo,
			OptOut:                domain.GeneralOptionTypeNo,
		},
	}

	err = s.pubSub.NotifyCreateContact(ctx, *contact)
	if err != nil {
		utils.RecordSpanError(span, err)
		log.Printf("failed to publish to crm.contact.create topic: %v", err)
	}

	return nil
}

func (s *SMSImpl) replyAndNotifyAdmin(
	ctx context.Context,
	to string,
	message string,
	validatedInput *dto.AfricasTalkingMessage,
	supportEmailPayload *dto.EmailNotificationPayload) error {
	ctx, span := tracer.Start(ctx, "CreateSMSData")
	defer span.End()

	err := s.engagement.SendSMS(ctx, []string{to}, message)
	if err != nil {
		return fmt.Errorf("an error occurred while sending SMS: %v", err)
	}

	err = s.engagement.NotifyAdmins(ctx, *supportEmailPayload)
	if err != nil {
		return fmt.Errorf("an error occurred while notifying admins: %v", err)
	}

	err = s.onboardingRepository.PersistIncomingSMSData(ctx, validatedInput)
	if err != nil {
		utils.RecordSpanError(span, err)

		return err
	}

	return nil
}
