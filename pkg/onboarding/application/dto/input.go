package dto

import (
	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/profileutils"
	"github.com/savannahghi/scalarutils"
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

// SignUpInput represents the user information required to create a new account
type SignUpInput struct {
	PhoneNumber *string         `json:"phoneNumber"`
	PIN         *string         `json:"pin"`
	Flavour     feedlib.Flavour `json:"flavour"`
	OTP         *string         `json:"otp"`
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

// PushTokenPayload represents user device push token
type PushTokenPayload struct {
	PushToken string `json:"pushTokens"`
	UID       string `json:"uid"`
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

// AssignRolePayload is the payload used to assign a role to a user
type AssignRolePayload struct {
	UserID string `json:"userID"`
	RoleID string `json:"roleID"`
}

// DeleteRolePayload is the payload used to delete a role
type DeleteRolePayload struct {
	Name string `json:"name"`

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

// RegisterUserInput is the data required to creat a new user.
// this data can be used by cross service requests
type RegisterUserInput struct {
	UID            *string           `json:"createdByID,omitempty"`
	FirstName      *string           `json:"firstName,omitempty"`
	LastName       *string           `json:"lastName,omitempty"`
	Gender         *enumutils.Gender `json:"gender,omitempty"`
	PhoneNumber    *string           `json:"phoneNumber,omitempty"`
	Email          *string           `json:"email,omitempty"`
	DateOfBirth    *scalarutils.Date `json:"dateOfBirth,omitempty"`
	RoleIDs        []string          `json:"roleIDs,omitempty"`
	WelcomeMessage *string           `json:"welcomeMessage,omitempty"`
}
