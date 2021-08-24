package mock

import (
	"context"

	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
)

// FakeServiceClinical is a `Clinical` service mock
type FakeServiceClinical struct {
	RegisterPatientFn func(
		ctx context.Context,
		payload domain.SimplePatientRegistrationInput,
	) (*domain.SimplePatientRegistrationInput, error)

	CheckIfPatientExistsFn func(
		ctx context.Context,
		phoneNumber string,
	) (*domain.SimplePatientRegistrationInput, bool, error)
}

// RegisterPatient ...
func (f *FakeServiceClinical) RegisterPatient(
	ctx context.Context,
	payload domain.SimplePatientRegistrationInput,
) (*domain.SimplePatientRegistrationInput, error) {
	return f.RegisterPatientFn(ctx, payload)
}

//CheckIfPatientExists ...
func (f *FakeServiceClinical) CheckIfPatientExists(
	ctx context.Context,
	phoneNumber string,
) (*domain.SimplePatientRegistrationInput, bool, error) {
	return f.CheckIfPatientExistsFn(ctx, phoneNumber)
}
