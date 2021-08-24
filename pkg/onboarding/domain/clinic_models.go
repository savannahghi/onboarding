package domain

import (
	"fmt"
	"io"
	"strconv"

	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/scalarutils"
)

// NameInput is used to input patient names.
type NameInput struct {
	FirstName  string  `json:"firstName"`
	LastName   string  `json:"lastName"`
	OtherNames *string `json:"otherNames"`
}

// IdentificationDocument is used to input e.g National ID or passport document
// numbers at patient registration.
type IdentificationDocument struct {
	DocumentType     IDDocumentType         `json:"documentType"`
	DocumentNumber   string                 `json:"documentNumber"`
	Title            *string                `json:"title,omitempty"`
	ImageContentType *enumutils.ContentType `json:"imageContentType,omitempty"`
	ImageBase64      *string                `json:"imageBase64,omitempty"`
}

// PhoneNumberInput is used to input phone numbers.
type PhoneNumberInput struct {
	Msisdn             string `json:"msisdn"`
	VerificationCode   string `json:"verificationCode"`
	IsUssd             bool   `json:"isUSSD"`
	CommunicationOptIn bool   `json:"communicationOptIn"`
}

// PhotoInput is used to upload patient photos.
type PhotoInput struct {
	PhotoContentType enumutils.ContentType `json:"photoContentType"`
	PhotoBase64data  string                `json:"photoBase64data"`
	PhotoFilename    string                `json:"photoFilename"`
}

// EmailInput is used to register patient emails.
type EmailInput struct {
	Email              string `json:"email"`
	CommunicationOptIn bool   `json:"communicationOptIn"`
}

// PhysicalAddress is used to record a precise physical address.
type PhysicalAddress struct {
	MapsCode        string `json:"mapsCode"`
	PhysicalAddress string `json:"physicalAddress"`
}

// PostalAddress is used to record patient's postal addresses
type PostalAddress struct {
	PostalAddress string `json:"postalAddress"`
	PostalCode    string `json:"postalCode"`
}

// SimplePatientRegistrationInput provides a simplified API to support registration
// of patients.
type SimplePatientRegistrationInput struct {
	ID                      string                    `json:"id"`
	Names                   []*NameInput              `json:"names"`
	IdentificationDocuments []*IdentificationDocument `json:"identificationDocuments"`
	BirthDate               scalarutils.Date          `json:"birthDate"`
	PhoneNumbers            []*PhoneNumberInput       `json:"phoneNumbers"`
	Photos                  []*PhotoInput             `json:"photos"`
	Emails                  []*EmailInput             `json:"emails"`
	PhysicalAddresses       []*PhysicalAddress        `json:"physicalAddresses"`
	PostalAddresses         []*PostalAddress          `json:"postalAddresses"`
	Gender                  string                    `json:"gender"`
	Active                  bool                      `json:"active"`
	MaritalStatus           MaritalStatus             `json:"maritalStatus"`
	Languages               []enumutils.Language      `json:"languages"`
	ReplicateUSSD           bool                      `json:"replicate_ussd,omitempty"`
}

// IDDocumentType is an internal code system for identification document types.
type IDDocumentType string

// ID type constants
const (
	// IDDocumentTypeNationalID ...
	IDDocumentTypeNationalID IDDocumentType = "national_id"
	// IDDocumentTypePassport ...
	IDDocumentTypePassport IDDocumentType = "passport"
	// IDDocumentTypeAlienID ...
	IDDocumentTypeAlienID IDDocumentType = "alien_id"
)

// AllIDDocumentType is a list of known ID types
var AllIDDocumentType = []IDDocumentType{
	IDDocumentTypeNationalID,
	IDDocumentTypePassport,
	IDDocumentTypeAlienID,
}

// IsValid checks that the ID type is valid
func (e IDDocumentType) IsValid() bool {
	switch e {
	case IDDocumentTypeNationalID, IDDocumentTypePassport, IDDocumentTypeAlienID:
		return true
	}
	return false
}

// String ...
func (e IDDocumentType) String() string {
	return string(e)
}

// UnmarshalGQL translates the input value to an ID type
func (e *IDDocumentType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = IDDocumentType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid IDDocumentType", str)
	}
	return nil
}

// MarshalGQL writes the enum value to the supplied writer
func (e IDDocumentType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))

}

// MaritalStatus is used to code individuals' marital statuses.
//
// See: https://www.hl7.org/fhir/valueset-marital-status.html
type MaritalStatus string

// known marital statuses
const (
	// MaritalStatusA ...
	MaritalStatusA MaritalStatus = "A"
	// MaritalStatusD ...
	MaritalStatusD MaritalStatus = "D"
	// MaritalStatusI ...
	MaritalStatusI MaritalStatus = "I"
	// MaritalStatusL ...
	MaritalStatusL MaritalStatus = "L"
	// MaritalStatusM ...
	MaritalStatusM MaritalStatus = "M"
	// MaritalStatusP ...
	MaritalStatusP MaritalStatus = "P"
	// MaritalStatusS ...
	MaritalStatusS MaritalStatus = "S"
	// MaritalStatusT ...
	MaritalStatusT MaritalStatus = "T"
	// MaritalStatusU ...
	MaritalStatusU MaritalStatus = "U"
	// MaritalStatusW ...
	MaritalStatusW MaritalStatus = "W"
	// MaritalStatusUnk ...
	MaritalStatusUnk MaritalStatus = "UNK"
)

// AllMaritalStatus is a list of known marital statuses
var AllMaritalStatus = []MaritalStatus{
	MaritalStatusA,
	MaritalStatusD,
	MaritalStatusI,
	MaritalStatusL,
	MaritalStatusM,
	MaritalStatusP,
	MaritalStatusS,
	MaritalStatusT,
	MaritalStatusU,
	MaritalStatusW,
	MaritalStatusUnk,
}

// IsValid checks that the marital status is valid
func (e MaritalStatus) IsValid() bool {
	switch e {
	case MaritalStatusA, MaritalStatusD, MaritalStatusI, MaritalStatusL, MaritalStatusM, MaritalStatusP, MaritalStatusS, MaritalStatusT, MaritalStatusU, MaritalStatusW, MaritalStatusUnk:
		return true
	}
	return false
}

// String ...
func (e MaritalStatus) String() string {
	return string(e)
}

// UnmarshalGQL turns the supplied input into a marital status enum value
func (e *MaritalStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = MaritalStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid MaritalStatus", str)
	}
	return nil
}

// MarshalGQL writes the enum value to the supplied writer
func (e MaritalStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
