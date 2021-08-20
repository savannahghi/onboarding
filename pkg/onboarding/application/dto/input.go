package dto

import (
	"net/url"
	"time"

	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/profileutils"
	"github.com/savannahghi/scalarutils"
	dm "gitlab.slade360emr.com/go/commontools/accounting/pkg/domain"
	CRMDomain "gitlab.slade360emr.com/go/commontools/crm/pkg/domain"
)

// UserProfileInput is used to create or update a user's profile.
type UserProfileInput struct {
	PhotoUploadID *string           `json:"photoUploadID"`
	DateOfBirth   *scalarutils.Date `json:"dateOfBirth,omitempty"`
	Gender        *enumutils.Gender `json:"gender,omitempty"`
	FirstName     *string           `json:"lastName"`
	LastName      *string           `json:"firstName"`
}

// PostVisitSurveyInput is used to send the results of post-visit surveys to the
// server.
type PostVisitSurveyInput struct {
	LikelyToRecommend int    `json:"likelyToRecommend" firestore:"likelyToRecommend"`
	Criticism         string `json:"criticism"         firestore:"criticism"`
	Suggestions       string `json:"suggestions"       firestore:"suggestions"`
}

// BusinessPartnerFilterInput is used to supply filter parameters for organizatiom filter inputs
type BusinessPartnerFilterInput struct {
	Search    *string `json:"search"`
	Name      *string `json:"name"`
	SladeCode *string `json:"slade_code"`
}

// ToURLValues transforms the filter input to `url.Values`
func (i *BusinessPartnerFilterInput) ToURLValues() (values url.Values) {
	vals := url.Values{}
	if i.Search != nil {
		vals.Add("search", *i.Search)
	}
	if i.Name != nil {
		vals.Add("name", *i.Name)
	}
	if i.SladeCode != nil {
		vals.Add("slade_code", *i.SladeCode)
	}
	return vals
}

// BusinessPartnerSortInput is used to supply sort input for organization list queries
type BusinessPartnerSortInput struct {
	Name      *enumutils.SortOrder `json:"name"`
	SladeCode *enumutils.SortOrder `json:"slade_code"`
}

// ToURLValues transforms the filter input to `url.Values`
func (i *BusinessPartnerSortInput) ToURLValues() (values url.Values) {
	vals := url.Values{}
	if i.Name != nil {
		if *i.Name == enumutils.SortOrderAsc {
			vals.Add("order_by", "name")
		} else {
			vals.Add("order_by", "-name")
		}
	}
	if i.SladeCode != nil {
		if *i.Name == enumutils.SortOrderAsc {
			vals.Add("slade_code", "number")
		} else {
			vals.Add("slade_code", "-number")
		}
	}
	return vals
}

// BranchSortInput is used to supply sorting input for location list queries
type BranchSortInput struct {
	Name      *enumutils.SortOrder `json:"name"`
	SladeCode *enumutils.SortOrder `json:"slade_code"`
}

// ToURLValues transforms the sort input to `url.Values`
func (i *BranchSortInput) ToURLValues() (values url.Values) {
	vals := url.Values{}
	if i.Name != nil {
		if *i.Name == enumutils.SortOrderAsc {
			vals.Add("order_by", "name")
		} else {
			vals.Add("order_by", "-name")
		}
	}
	if i.SladeCode != nil {
		if *i.SladeCode == enumutils.SortOrderAsc {
			vals.Add("slade_code", "number")
		} else {
			vals.Add("slade_code", "-number")
		}
	}
	return vals
}

// SignUpInput represents the user information required to create a new account
type SignUpInput struct {
	PhoneNumber *string         `json:"phoneNumber"`
	PIN         *string         `json:"pin"`
	Flavour     feedlib.Flavour `json:"flavour"`
	OTP         *string         `json:"otp"`
}

// BranchEdge is used to serialize GraphQL Relay edges for locations
type BranchEdge struct {
	Cursor *string        `json:"cursor"`
	Node   *domain.Branch `json:"node"`
}

// BranchConnection is used tu serialize GraphQL Relay connections for locations
type BranchConnection struct {
	Edges    []*BranchEdge           `json:"edges"`
	PageInfo *firebasetools.PageInfo `json:"pageInfo"`
}

// BranchFilterInput is used to supply filter parameters for locatioon list queries
type BranchFilterInput struct {
	Search               *string `json:"search"`
	SladeCode            *string `json:"sladeCode"`
	ParentOrganizationID *string `json:"parentOrganizationID"`
}

// ToURLValues transforms the filter input to `url.Values`
func (i *BranchFilterInput) ToURLValues() url.Values {
	vals := url.Values{}
	if i.Search != nil {
		vals.Add("search", *i.Search)
	}
	if i.SladeCode != nil {
		vals.Add("slade_code", *i.SladeCode)
	}
	if i.ParentOrganizationID != nil {
		vals.Add("parent", *i.ParentOrganizationID)
	}
	return vals
}

// PhoneNumberPayload used when verifying a phone number.
type PhoneNumberPayload struct {
	PhoneNumber *string `json:"phoneNumber"`
}

// SetPrimaryPhoneNumberPayload used when veriying and setting a user's primary phone number via REST
type SetPrimaryPhoneNumberPayload struct {
	PhoneNumber *string `json:"phoneNumber"`
	OTP         *string `json:"otp"`
}

// ChangePINRequest payload to set or change PIN information
type ChangePINRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	PIN         string `json:"pin"`
	OTP         string `json:"otp"`
}

// LoginPayload used when calling the REST API to log a user in
type LoginPayload struct {
	PhoneNumber *string         `json:"phoneNumber"`
	PIN         *string         `json:"pin"`
	Flavour     feedlib.Flavour `json:"flavour"`
}

// SendRetryOTPPayload is used when calling the REST API to resend an otp
type SendRetryOTPPayload struct {
	Phone     *string `json:"phoneNumber"`
	RetryStep *int    `json:"retryStep"`
	AppID     *string `json:"appId"`
}

// RefreshTokenExchangePayload is marshalled into JSON
// and sent to the Firebase Auth REST API when exchanging a
// refresh token for an ID token that can be used to make API calls
type RefreshTokenExchangePayload struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenPayload is used when calling the REST API to
// exchange a Refresh Token for new ID Token
type RefreshTokenPayload struct {
	RefreshToken *string `json:"refreshToken"`
}

// UIDPayload is the user ID used in some inter-service requests
type UIDPayload struct {
	UID *string `json:"uid"`
}

// UpdateCoversPayload is used to make a REST
// request to update a user's covers in their user profile
type UpdateCoversPayload struct {
	UID                   *string    `json:"uid"`
	PayerName             *string    `json:"payerName"`
	MemberName            *string    `json:"memberName"`
	MemberNumber          *string    `json:"memberNumber"`
	PayerSladeCode        *int       `json:"payerSladeCode"`
	BeneficiaryID         *int       `json:"beneficiaryID"`
	EffectivePolicyNumber *string    `json:"effectivePolicyNumber"`
	ValidFrom             *time.Time `json:"validFrom"`
	ValidTo               *time.Time `json:"validTo"`
}

// UIDsPayload is an input of a slice of users' UIDs used
// for ISC requests to retrieve contact details of the users
type UIDsPayload struct {
	UIDs []string `json:"uids"`
}

// UserAddressInput represents a user's geo location input
type UserAddressInput struct {
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	Locality         *string `json:"locality"`
	Name             *string `json:"name"`
	PlaceID          *string `json:"placeID"`
	FormattedAddress *string `json:"formattedAddress"`
}

// NHIFDetailsInput represents a user's thin NHIF input details
type NHIFDetailsInput struct {
	MembershipNumber          string                          `json:"membershipNumber"`
	Employment                domain.EmploymentType           `json:"employmentType"`
	IDDocType                 enumutils.IdentificationDocType `json:"IDDocType"`
	IDNumber                  string                          `json:"IDNumber"`
	IdentificationCardPhotoID string                          `json:"identificationCardPhotoID"`
	NHIFCardPhotoID           string                          `json:"nhifCardPhotoID"`
}

// PushTokenPayload represents user device push token
type PushTokenPayload struct {
	PushToken string `json:"pushTokens"`
	UID       string `json:"uid"`
}

// CustomerPubSubMessage is an `onboarding` PubSub message struct
type CustomerPubSubMessage struct {
	CustomerPayload CustomerPayload `json:"customerPayload"`
	UID             string          `json:"uid"`
}

// CustomerPayload is the customer data used to create a customer
// business partner in the ERP
type CustomerPayload struct {
	Active       bool                     `json:"active"`
	PartnerName  string                   `json:"partner_name"`
	Country      string                   `json:"country"`
	Currency     string                   `json:"currency"`
	IsCustomer   bool                     `json:"is_customer"`
	CustomerType profileutils.PartnerType `json:"customer_type"`
}

// SupplierPubSubMessage is an `onboarding` PubSub message struct
type SupplierPubSubMessage struct {
	SupplierPayload SupplierPayload `json:"supplierPayload"`
	UID             string          `json:"uid"`
}

// SupplierPayload is the supplier data used to create a supplier
// business partner in the ERP
type SupplierPayload struct {
	Active       bool                     `json:"active"`
	PartnerName  string                   `json:"partner_name"`
	Country      string                   `json:"country"`
	Currency     string                   `json:"currency"`
	IsSupplier   bool                     `json:"is_supplier"`
	SupplierType profileutils.PartnerType `json:"supplier_type"`
}

// EmailNotificationPayload is the email payload used to send email
// supplier and admins for KYC requests
type EmailNotificationPayload struct {
	SupplierName string                      `json:"supplier_name"`
	PartnerType  string                      `json:"partner_type"`
	AccountType  string                      `json:"account_type"`
	SubjectTitle string                      `json:"subject_title"`
	EmailBody    string                      `json:"email_body"`
	EmailAddress string                      `json:"email_address"`
	PrimaryPhone string                      `json:"primary_phone"`
	BeWellUser   CRMDomain.GeneralOptionType `json:"bewell_user"`
	Time         string                      `json:"sending_time"`
}

// UserProfilePayload is used to update a user's profile.
// This payload is used for REST endpoints
type UserProfilePayload struct {
	UID           *string           `json:"uid"`
	PhotoUploadID *string           `json:"photoUploadID"`
	DateOfBirth   *scalarutils.Date `json:"dateOfBirth,omitempty"`
	Gender        *enumutils.Gender `json:"gender,omitempty"`
	FirstName     *string           `json:"lastName"`
	LastName      *string           `json:"firstName"`
}

// PermissionInput input required to create a permission
type PermissionInput struct {
	Action   string
	Resource string
}

// RolePayload used when adding roles to a user
type RolePayload struct {
	PhoneNumber *string                `json:"phoneNumber"`
	Role        *profileutils.RoleType `json:"role"`
}

// RegisterAgentInput provides the data payload required to create an Agent
type RegisterAgentInput struct {
	FirstName   string           `json:"lastName"`
	LastName    string           `json:"firstName"`
	Gender      enumutils.Gender `json:"gender"`
	PhoneNumber string           `json:"phoneNumber"`
	Email       string           `json:"email"`
	DateOfBirth scalarutils.Date `json:"dateOfBirth"`
	// ID of the Role being assigned to the new agent
	RoleIDs []string `json:"roleIDs"`
}

// RegisterAdminInput provides the data payload required to create an Admin
type RegisterAdminInput struct {
	FirstName   string           `json:"lastName"`
	LastName    string           `json:"firstName"`
	Gender      enumutils.Gender `json:"gender"`
	PhoneNumber string           `json:"phoneNumber"`
	Email       string           `json:"email"`
	DateOfBirth scalarutils.Date `json:"dateOfBirth"`
	// ID of the Role being assigned to the new employee
	RoleIDs []string `json:"roleIDs"`
}

// ContactLeadInput ...
type ContactLeadInput struct {
	ContactType    string                      `json:"contact_type,omitempty"`
	ContactValue   string                      `json:"contact_value,omitempty"`
	FirstName      string                      `json:"first_name,omitempty"`
	LastName       string                      `json:"last_name,omitempty"`
	DateOfBirth    scalarutils.Date            `json:"date_of_birth,omitempty"`
	IsSync         bool                        `json:"isSync"                    firestore:"IsSync"`
	TimeSync       *time.Time                  `json:"timeSync"                  firestore:"TimeSync"`
	OptOut         CRMDomain.GeneralOptionType `json:"opt_out,omitempty"`
	WantCover      bool                        `json:"wantCover"                 firestore:"wantCover"`
	ContactChannel string                      `json:"contact_channel,omitempty"`
	IsRegistered   bool                        `json:"is_registered,omitempty"`
}

// AgentFilterInput is used to supply filter parameters for agent filter inputs
type AgentFilterInput struct {
	PhoneNumber string `json:"phoneNumber"`
}

// CoverInput is used to add covers
type CoverInput struct {
	PayerSladeCode int      `json:"payerSladeCode"`
	MemberNumber   string   `json:"memberNumber"`
	UID            string   `json:"uid"`
	PushToken      []string `json:"pushToken"`
}

// LinkCoverPubSubMessage is a `cover linking` pub sub message struct
type LinkCoverPubSubMessage struct {
	PhoneNumber string   `json:"phoneNumber"`
	UID         string   `json:"uid"`
	PushToken   []string `json:"pushToken"`
}

//CustomerPubSubMessagePayload is an `onboarding` PubSub message struct for commontools
type CustomerPubSubMessagePayload struct {
	CustomerPayload dm.CustomerPayload `json:"customerPayload"`
	UID             string             `json:"uid"`
}

//SupplierPubSubMessagePayload is an `onboarding` PubSub message struct for commontools
type SupplierPubSubMessagePayload struct {
	SupplierPayload dm.SupplierPayload `json:"supplierPayload"`
	UID             string             `json:"uid"`
}

// USSDEvent records any USSD event(e.g. entering firstname, lastname etc.) that happens for every session and the time
type USSDEvent struct {
	SessionID         string     `firestore:"sessionID"`
	PhoneNumber       string     `firestore:"phoneNumber"`
	USSDEventDateTime *time.Time `firestore:"ussdEventDateTime"`
	Level             int        `firestore:"level"`
	USSDEventName     string     `firestore:"ussdEventName"`
}

// CoverLinkingEvent is a cover linking struct for cover linking events(started or completed)
type CoverLinkingEvent struct {
	ID                    string     `firestore:"id"`
	CoverLinkingEventTime *time.Time `firestore:"coverLinkingEventTime"`
	CoverStatus           string     `firestore:"coverStatus"`
	MemberNumber          string     `firestore:"memberNumber"`
	PhoneNumber           string     `firestore:"phoneNumber"`
}

// AssignRolePayload is the payload used to assign a role to a user
type AssignRolePayload struct {
	UserID string `json:"userID"`
	RoleID string `json:"roleID"`
}

// DeleteRolePayload is the payload used to delete a role
type DeleteRolePayload struct {
	Name   string `json:"name"`
	RoleID string `json:"roleID"`
}

// RoleInput represents the information required when creating a role
type RoleInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Scopes      []string `json:"scopes"`
}

// RolePermissionInput input required to create a permission
type RolePermissionInput struct {
	RoleID string   `json:"roleID"`
	Scopes []string `json:"scopes"`
}

// OtpPayload used when sending OTP messages
type OtpPayload struct {
	PhoneNumber *string `json:"phoneNumber"`
	AppID       *string `json:"appId"`
}

// RetrieveUserProfileInput will be used to fetch a user profile by either email address or phone
type RetrieveUserProfileInput struct {
	Email       *string `json:"email" firestore:"emailAddress"`
	PhoneNumber *string `json:"phone" firestore:"phoneNumber"`
}

//ProfileSuspensionInput is the input required to suspend/unsuspend a PRO account
type ProfileSuspensionInput struct {
	ID      string   `json:"id"`
	RoleIDs []string `json:"roleIDs"`
	Reason  string   `json:"reason"`
}

// EDICoverLinkingPubSubMessage holds the data required to add a cover to the profile
// of EDI members who received a message with the bewell link an went ahead to
// download the app
type EDICoverLinkingPubSubMessage struct {
	PayerSladeCode int    `json:"payersladecode"`
	MemberNumber   string `json:"membernumber"`
	PhoneNumber    string `json:"phonenumber"`
}

// CheckPermissionPayload is the payload used when checking if a user is authorized
type CheckPermissionPayload struct {
	UID        *string                  `json:"uid"`
	Permission *profileutils.Permission `json:"permission"`
}

// RoleRevocationInput is the input when revoking a user's role
type RoleRevocationInput struct {
	ProfileID string
	RoleID    string
	Reason    string
}
