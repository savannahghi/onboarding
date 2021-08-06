package mock

import (
	"context"

	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/profileutils"
)

// FakeServiceEngagement is an `engagement` service mock .
type FakeServiceEngagement struct {
	ResolveDefaultNudgeByTitleFn func(ctx context.Context, UID string, flavour feedlib.Flavour, nudgeTitle string) error
	SendMailFn                   func(ctx context.Context, email string, message string, subject string) error
	NotifySupplierOnSuspensionFn func(ctx context.Context, input dto.EmailNotificationPayload) error
	GenerateAndSendOTPFn         func(
		ctx context.Context,
		phone string,
	) (*profileutils.OtpResponse, error)

	SendRetryOTPFn func(
		ctx context.Context,
		msisdn string,
		retryStep int,
	) (*profileutils.OtpResponse, error)

	VerifyOTPFn func(ctx context.Context, phone, OTP string) (bool, error)

	VerifyEmailOTPFn func(ctx context.Context, email, OTP string) (bool, error)

	SendSMSFn func(ctx context.Context, phoneNumbers []string, message string) error
}

// ResolveDefaultNudgeByTitle ...
func (f *FakeServiceEngagement) ResolveDefaultNudgeByTitle(
	ctx context.Context,
	UID string,
	flavour feedlib.Flavour,
	nudgeTitle string,
) error {
	return f.ResolveDefaultNudgeByTitleFn(
		ctx,
		UID,
		flavour,
		nudgeTitle,
	)
}

// SendMail ...
func (f *FakeServiceEngagement) SendMail(
	ctx context.Context,
	email string,
	message string,
	subject string,
) error {
	return f.SendMailFn(ctx, email, message, subject)
}

// GenerateAndSendOTP ...
func (f *FakeServiceEngagement) GenerateAndSendOTP(
	ctx context.Context,
	phone string,
) (*profileutils.OtpResponse, error) {
	return f.GenerateAndSendOTPFn(ctx, phone)
}

// SendRetryOTP ...
func (f *FakeServiceEngagement) SendRetryOTP(
	ctx context.Context,
	msisdn string,
	retryStep int,
) (*profileutils.OtpResponse, error) {
	return f.SendRetryOTPFn(ctx, msisdn, retryStep)
}

// VerifyOTP ...
func (f *FakeServiceEngagement) VerifyOTP(ctx context.Context, phone, OTP string) (bool, error) {
	return f.VerifyOTPFn(ctx, phone, OTP)
}

// VerifyEmailOTP ...
func (f *FakeServiceEngagement) VerifyEmailOTP(ctx context.Context, email, OTP string) (bool, error) {
	return f.VerifyEmailOTPFn(ctx, email, OTP)
}

// NotifySupplierOnSuspension ...
func (f *FakeServiceEngagement) NotifySupplierOnSuspension(ctx context.Context, input dto.EmailNotificationPayload) error {
	return f.NotifySupplierOnSuspensionFn(ctx, input)
}

// SendSMS ...
func (f *FakeServiceEngagement) SendSMS(ctx context.Context, phoneNumbers []string, message string) error {
	return f.SendSMSFn(ctx, phoneNumbers, message)
}
