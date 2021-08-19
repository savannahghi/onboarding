package usecases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"gitlab.slade360emr.com/go/commontools/crm/pkg/domain"
)

func TestSMSImpl_CreateSMSData_integration(t *testing.T) {
	ctx := context.Background()

	s, err := InitializeTestService(ctx)
	if err != nil {
		t.Errorf("unable to initialize the test service")
		return
	}

	validLinkId := uuid.New().String()
	text := "Test Covers"
	to := "3601"
	id := "60119"
	from := interserviceclient.TestUserPhoneNumber
	date := "2021-05-17T13:20:04.490Z"

	validPayload := &dto.AfricasTalkingMessage{
		LinkID: validLinkId,
		Text:   text,
		To:     to,
		ID:     id,
		Date:   date,
		From:   from,
	}

	invalidPayload := &dto.AfricasTalkingMessage{
		LinkID: "",
		Text:   text,
		To:     to,
		ID:     id,
		Date:   "",
		From:   from,
	}

	_ = s.Engagement.SendSMS(ctx, []string{from}, text)

	supportNotification := &dto.EmailNotificationPayload{
		EmailBody: text,
	}
	_ = s.Engagement.NotifyAdmins(ctx, *supportNotification)

	contact := &domain.CRMContact{
		Properties: domain.ContactProperties{
			Phone:                 validPayload.From,
			Gender:                string(domain.GeneralOptionTypeNotGiven),
			FirstChannelOfContact: domain.ChannelOfContactShortcode,
			BeWellEnrolled:        domain.GeneralOptionTypeNo,
			OptOut:                domain.GeneralOptionTypeNo,
		},
	}

	_ = s.PubSub.NotifyCreateContact(ctx, *contact)

	type args struct {
		ctx   context.Context
		input dto.AfricasTalkingMessage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy :) successfully persist sms data",
			args: args{
				ctx:   ctx,
				input: *validPayload,
			},
			wantErr: false,
		},
		{
			name: "Sad :( fail to persist sms data with empty sms data",
			args: args{
				ctx:   ctx,
				input: dto.AfricasTalkingMessage{},
			},
			wantErr: true,
		},
		{
			name: "Sad :( fail to persist sms data with invalid sms data",
			args: args{
				ctx:   ctx,
				input: *invalidPayload,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.SMS.CreateSMSData(tt.args.ctx, &tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SMSImpl.CreateSMSData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
