package domain

import (
	"fmt"
	"io"
	"log"
	"strconv"
)

// PractitionerCadre is a list of health worker cadres.
type PractitionerCadre string

// practitioner cadre constants
const (
	PractitionerCadreDoctor          PractitionerCadre = "DOCTOR"
	PractitionerCadreClinicalOfficer PractitionerCadre = "CLINICAL_OFFICER"
	PractitionerCadreNurse           PractitionerCadre = "NURSE"
)

// AllPractitionerCadre is the set of known valid practitioner cadres
var AllPractitionerCadre = []PractitionerCadre{
	PractitionerCadreDoctor,
	PractitionerCadreClinicalOfficer,
	PractitionerCadreNurse,
}

// IsValid returns true if a practitioner cadre is valid
func (e PractitionerCadre) IsValid() bool {
	switch e {
	case PractitionerCadreDoctor, PractitionerCadreClinicalOfficer, PractitionerCadreNurse:
		return true
	}
	return false
}

func (e PractitionerCadre) String() string {
	return string(e)
}

// UnmarshalGQL converts the supplied value to a practitioner cadre
func (e *PractitionerCadre) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PractitionerCadre(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PractitionerCadre", str)
	}
	return nil
}

// MarshalGQL writes the practitioner cadre to the supplied writer
func (e PractitionerCadre) MarshalGQL(w io.Writer) {
	_, err := fmt.Fprint(w, strconv.Quote(e.String()))
	if err != nil {
		log.Printf("%v\n", err)
	}
}

// FivePointRating is used to implement
type FivePointRating string

// known ratings
const (
	FivePointRatingPoor           FivePointRating = "POOR"
	FivePointRatingUnsatisfactory FivePointRating = "UNSATISFACTORY"
	FivePointRatingAverage        FivePointRating = "AVERAGE"
	FivePointRatingSatisfactory   FivePointRating = "SATISFACTORY"
	FivePointRatingExcellent      FivePointRating = "EXCELLENT"
)

// AllFivePointRating is a list of all known ratings
var AllFivePointRating = []FivePointRating{
	FivePointRatingPoor,
	FivePointRatingUnsatisfactory,
	FivePointRatingAverage,
	FivePointRatingSatisfactory,
	FivePointRatingExcellent,
}

// IsValid returns true for valid ratings
func (e FivePointRating) IsValid() bool {
	switch e {
	case FivePointRatingPoor, FivePointRatingUnsatisfactory, FivePointRatingAverage, FivePointRatingSatisfactory, FivePointRatingExcellent:
		return true
	}
	return false
}

func (e FivePointRating) String() string {
	return string(e)
}

// UnmarshalGQL converts the input, if valid, into a rating value
func (e *FivePointRating) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = FivePointRating(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid FivePointRating", str)
	}
	return nil
}

// MarshalGQL converts the rating into a valid JSON string
func (e FivePointRating) MarshalGQL(w io.Writer) {
	_, err := fmt.Fprint(w, strconv.Quote(e.String()))
	if err != nil {
		log.Printf("%v\n", err)
	}
}

// PractitionerService defines the various services practitioners offer
type PractitionerService string

// PractitionerServiceOutpatientServices is a constant of all known practitioner service
const (
	PractitionerServiceOutpatientServices PractitionerService = "OUTPATIENT_SERVICES"
	PractitionerServiceInpatientServices  PractitionerService = "INPATIENT_SERVICES"
	PractitionerServicePharmacy           PractitionerService = "PHARMACY"
	PractitionerServiceMaternity          PractitionerService = "MATERNITY"
	PractitionerServiceLabServices        PractitionerService = "LAB_SERVICES"
	PractitionerServiceOther              PractitionerService = "OTHER"
)

//AllPractitionerService is a list of all known practitioner service
var AllPractitionerService = []PractitionerService{
	PractitionerServiceOutpatientServices,
	PractitionerServiceInpatientServices,
	PractitionerServicePharmacy,
	PractitionerServiceMaternity,
	PractitionerServiceLabServices,
	PractitionerServiceOther,
}

// IsValid returns true for valid practitioner service
func (e PractitionerService) IsValid() bool {
	switch e {
	case PractitionerServiceOutpatientServices, PractitionerServiceInpatientServices, PractitionerServicePharmacy, PractitionerServiceMaternity, PractitionerServiceLabServices, PractitionerServiceOther:
		return true
	}
	return false
}

func (e PractitionerService) String() string {
	return string(e)
}

// UnmarshalGQL converts the input, if valid, into a practitioner service value
func (e *PractitionerService) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PractitionerService(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PractitionerService", str)
	}
	return nil
}

// MarshalGQL converts the practitioner service into a valid JSON string
func (e PractitionerService) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

// BeneficiaryRelationship defines the various relationships with beneficiaries
type BeneficiaryRelationship string

// BeneficiaryRelationshipSpouse is a constant of beneficiary spouse relationship
const (
	BeneficiaryRelationshipSpouse BeneficiaryRelationship = "SPOUSE"
	BeneficiaryRelationshipChild  BeneficiaryRelationship = "CHILD"
)

//AllBeneficiaryRelationship is a list of all known beneficiary relationships
var AllBeneficiaryRelationship = []BeneficiaryRelationship{
	BeneficiaryRelationshipSpouse,
	BeneficiaryRelationshipChild,
}

// IsValid returns true for valid beneficiary relationship
func (e BeneficiaryRelationship) IsValid() bool {
	switch e {
	case BeneficiaryRelationshipSpouse, BeneficiaryRelationshipChild:
		return true
	}
	return false
}

func (e BeneficiaryRelationship) String() string {
	return string(e)
}

// UnmarshalGQL converts the input, if valid, into a beneficiary relationship value
func (e *BeneficiaryRelationship) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = BeneficiaryRelationship(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid BeneficiaryRelationship", str)
	}
	return nil
}

// MarshalGQL converts the beneficiary relationship into a valid JSON string
func (e BeneficiaryRelationship) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

// EmploymentType ...
type EmploymentType string

// EmploymentTypeEmployed ..
const (
	EmploymentTypeEmployed     EmploymentType = "EMPLOYED"
	EmploymentTypeSelfEmployed EmploymentType = "SELF_EMPLOYED"
)

// AllEmploymentType ..
var AllEmploymentType = []EmploymentType{
	EmploymentTypeEmployed,
	EmploymentTypeSelfEmployed,
}

// IsValid ..
func (e EmploymentType) IsValid() bool {
	switch e {
	case EmploymentTypeEmployed, EmploymentTypeSelfEmployed:
		return true
	}
	return false
}

func (e EmploymentType) String() string {
	return string(e)
}

// UnmarshalGQL ..
func (e *EmploymentType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = EmploymentType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid EmploymentType", str)
	}
	return nil
}

// MarshalGQL ..
func (e EmploymentType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
