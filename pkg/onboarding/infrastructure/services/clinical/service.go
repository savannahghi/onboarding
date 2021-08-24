package clinical

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
)

// internal apis definitions
const (
	registerPatient = "internal/register-patient"

	// addNextOfKin = "internal/add-next_of_kin"

	// addNHIF = "internal/add-nhif"
)

// ServiceClinical defines the business logic required to interact with Clinical
type ServiceClinical interface {
	RegisterPatient(
		ctx context.Context,
		payload domain.SimplePatientRegistrationInput,
	) (*domain.SimplePatientRegistrationInput, error)

	CheckIfPatientExists(
		ctx context.Context,
		phoneNumber string,
	) (*domain.SimplePatientRegistrationInput, bool, error)
}

// ServiceClinicalImpl represents clinical usecases
type ServiceClinicalImpl struct {
	iscExt extension.ISCClientExtension
}

// NewClinicalService returns a new instance of clinical implementations
func NewClinicalService(
	serviceClient extension.ISCClientExtension,
) ServiceClinical {
	return &ServiceClinicalImpl{
		iscExt: serviceClient,
	}
}

// CheckIfPatientExists registers a patient and returns patientID of the registered patient
func (e *ServiceClinicalImpl) CheckIfPatientExists(
	ctx context.Context,
	phoneNumber string,
) (*domain.SimplePatientRegistrationInput, bool, error) {
	payload := struct {
		PhoneNumber string `json:"phoneNumber,omitempty"`
	}{
		PhoneNumber: phoneNumber,
	}
	resp, err := e.iscExt.MakeRequest(
		ctx,
		http.MethodPost,
		registerPatient,
		payload,
	)
	if err != nil {
		return nil, false, fmt.Errorf("failed to make patient search request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("unable to get data, with status code %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, false, err
	}

	output := dto.SearchPatientResult{}

	err = json.Unmarshal(data, &output)
	if err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal patient data: %v", err)
	}

	return &output.Patient, output.Found, nil
}

// RegisterPatient registers a patient and returns patientID of the registered patient
func (e *ServiceClinicalImpl) RegisterPatient(
	ctx context.Context,
	payload domain.SimplePatientRegistrationInput,
) (*domain.SimplePatientRegistrationInput, error) {

	//check if patient is already registered
	patient, found, err := e.CheckIfPatientExists(ctx, payload.PhoneNumbers[0].Msisdn)

	if err != nil {
		return nil, fmt.Errorf("failed unable to check if patient exists: %w", err)
	}

	//patient already exists
	if found {
		return patient, nil
	}

	resp, err := e.iscExt.MakeRequest(
		ctx,
		http.MethodPost,
		registerPatient,
		payload,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make patient registration request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to get data, with status code %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	patient = &domain.SimplePatientRegistrationInput{}
	err = json.Unmarshal(data, &patient)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal patient data: %v", err)
	}

	return patient, nil
}
