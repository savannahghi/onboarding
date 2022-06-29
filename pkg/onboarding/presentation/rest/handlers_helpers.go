package rest

import (
	"context"
	"fmt"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/serverutils"
	"go.opentelemetry.io/otel/trace"
)

func decodePhoneNumberPayload(
	w http.ResponseWriter,
	r *http.Request,
	span trace.Span,
) (*dto.PhoneNumberPayload, error) {
	payload := &dto.PhoneNumberPayload{}
	serverutils.DecodeJSONToTargetStruct(w, r, payload)

	span.AddEvent("decode json payload to struct")

	if payload.PhoneNumber == nil {
		return nil, fmt.Errorf(
			"expected a phone number to be given but it was not supplied",
		)
	}

	return payload, nil
}

func decodeOTPPayload(
	w http.ResponseWriter,
	r *http.Request,
	span trace.Span,
) (*dto.OtpPayload, error) {
	payload := &dto.OtpPayload{}
	serverutils.DecodeJSONToTargetStruct(w, r, payload)

	span.AddEvent("decode json payload to struct")

	if payload.PhoneNumber == nil {
		return nil, fmt.Errorf(
			"expected a phone number to be given but it was not supplied",
		)
	}

	return payload, nil
}

func addUIDToContext(ctx context.Context, uid string) context.Context {
	return context.WithValue(
		context.Background(),
		firebasetools.AuthTokenContextKey,
		&auth.Token{UID: uid},
	)
}
