package usecases

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/utils"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/crm"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/engagement"
	pubsubmessaging "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/pubsub"
	"github.com/savannahghi/onboarding/pkg/onboarding/repository"
	hubspotDomain "gitlab.slade360emr.com/go/commontools/crm/pkg/domain"
)

const (
	// The right copy of reply will be availed later
	shortcodeReplyMessage = "We have received your request and one of our representatives will reach out to you. Thanks you.\n\nDid you know you can now view your medical cover and benefits on Be.Well. To get started, Download Now from https://bwl.mobi/1bvf"
)

// SMSUsecase represent the logic involved in processing SMSs from shortcode
type SMSUsecase interface {
	ProcessShortCodeSMS(ctx context.Context, input *dto.AfricasTalkingMessage) error
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

// ProcessShortCodeSMS saves incoming shortcode messages, replies to the sender, creates hubspot contact and engagement
func (s *SMSImpl) ProcessShortCodeSMS(ctx context.Context, input *dto.AfricasTalkingMessage) error {
	ctx, span := tracer.Start(ctx, "ProcessShortCodeSMS")
	defer span.End()

	validatedInput, err := utils.ValidateAficasTalkingSMSData(input)
	if err != nil {
		utils.RecordSpanError(span, err)
		return err
	}

	// Checking if the shortcode SMS sender has a profile
	profile, err := s.onboardingRepository.GetUserProfileByPhoneNumber(ctx, validatedInput.From, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		//should not panic when the user profile is not found but proceed to send a reply sms and send an email to support

		err = s.createUserContactInCRM(ctx, validatedInput.From)
		if err != nil {
			return err
		}

		// replying an notifying support
		to := validatedInput.From
		message := shortcodeReplyMessage

		supportEmailPayload := &dto.EmailNotificationPayload{
			SubjectTitle: validatedInput.Text,
			EmailBody:    validatedInput.Text,
			PrimaryPhone: validatedInput.From,
			BeWellUser:   hubspotDomain.GeneralOptionTypeNo,
			Time:         validatedInput.Date,
		}

		return s.replyAndNotify(ctx, to, message, validatedInput, supportEmailPayload)

	}

	// an existing user
	if profile != nil {
		to := validatedInput.From
		message := shortcodeReplyMessage

		supportEmailPayload := &dto.EmailNotificationPayload{
			SubjectTitle: validatedInput.Text,
			EmailBody:    validatedInput.Text,
			PrimaryPhone: validatedInput.From,
			BeWellUser:   hubspotDomain.GeneralOptionTypeYes,
			Time:         validatedInput.Date,
		}

		err = s.saveShortCodeMessageAsCRMEngagement(ctx, validatedInput)
		if err != nil {
			return fmt.Errorf("an error occurred while saving user message as an engagement: %v", err)
		}

		return s.replyAndNotify(ctx, to, message, validatedInput, supportEmailPayload)
	}

	return nil
}

// replyAndNotify replies with an sms to the sender and notifies support
func (s *SMSImpl) replyAndNotify(
	ctx context.Context,
	to string,
	message string,
	validatedInput *dto.AfricasTalkingMessage,
	supportEmailPayload *dto.EmailNotificationPayload) error {
	ctx, span := tracer.Start(ctx, "replyAndNotify")
	defer span.End()

	err := s.onboardingRepository.PersistIncomingSMSData(ctx, validatedInput)
	if err != nil {
		utils.RecordSpanError(span, err)

		return err
	}

	err = s.engagement.SendSMS(ctx, []string{to}, message)
	if err != nil {
		return fmt.Errorf("an error occurred while sending SMS: %v", err)
	}

	err = s.engagement.NotifySupportTeam(ctx, *supportEmailPayload)
	if err != nil {
		return fmt.Errorf("an error occurred while notifying support: %v", err)
	}

	return nil
}

// saveShortCodeMessageAsCRMEngagement create and saves an engagement in the CRM
func (s *SMSImpl) saveShortCodeMessageAsCRMEngagement(ctx context.Context, payload *dto.AfricasTalkingMessage) error {
	if payload == nil {
		return fmt.Errorf("nil africastalking payload")
	}

	contact, err := s.hubspotCRM.GetContactByPhone(ctx, payload.From)
	if err != nil {
		return fmt.Errorf("failed to get contacts by phone: %w", err)
	}

	if contact.Properties.FirstChannelOfContact == "" {
		contact.Properties.FirstChannelOfContact = hubspotDomain.ChannelOfContact("SHORTCODE")
		_, err := s.hubspotCRM.UpdateHubSpotContact(ctx, contact)
		if err != nil {
			return fmt.Errorf("failed to update contact with First channel of contact as WHATSAPP: %w", err)
		}
	}

	searchContactResp, err := s.hubspotCRM.SearchContactByPhone(payload.From)
	if err != nil {
		return fmt.Errorf("unable to search contact by phone: %v", err)
	}

	contactID, err := strconv.Atoi(searchContactResp.Results[0].ContactID)
	if err != nil {
		return fmt.Errorf("unable to convert contact ID to int")
	}

	engagementData := &hubspotDomain.EngagementData{
		Engagement: hubspotDomain.Engagement{
			Active:    true,
			Type:      hubspotDomain.EngagementTypeNOTE,
			Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
			OwnerID:   0,
		},
		Associations: hubspotDomain.Associations{
			ContactIDs: []int{contactID},
		},
		Metadata: map[string]interface{}{
			"body": payload.Text,
		},
	}

	engagement, err := s.hubspotCRM.CreateHubspotEngagement(ctx, engagementData)
	if err != nil {
		return fmt.Errorf("unable to create an engagement: %v", err)
	}
	if engagement == nil {
		return fmt.Errorf("nil engagement returned")
	}

	return nil
}

func (s *SMSImpl) createUserContactInCRM(ctx context.Context, phoneNumber string) error {
	// check if contact is in CRM
	contactResponse, err := s.hubspotCRM.SearchContactByPhone(phoneNumber)
	if err != nil {
		return err
	}

	if len(contactResponse.Results) == 0 {
		contact := hubspotDomain.CRMContact{
			Properties: hubspotDomain.ContactProperties{
				Phone:                 phoneNumber,
				FirstChannelOfContact: hubspotDomain.ChannelOfContactShortcode,
				BeWellEnrolled:        hubspotDomain.GeneralOptionTypeNo,
				OptOut:                hubspotDomain.GeneralOptionTypeNo,
			},
		}

		err := s.pubSub.NotifyCreateContact(ctx, contact)
		if err != nil {
			return fmt.Errorf("error occurred creating a CRMContact:%w", err)
		}
	}
	return nil
}
