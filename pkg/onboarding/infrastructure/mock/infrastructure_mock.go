package mock

import (
	"context"

	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/profileutils"
)

// FakeInfrastructure mocks the infrastructure interface
type FakeInfrastructure struct {
	// StageProfileNudge stages nudges published from this service.
	StageProfileNudgeFn func(ctx context.Context, nudge *feedlib.Nudge) error

	// CreateRole creates a new role and persists it to the database
	CreateRoleFn func(
		ctx context.Context,
		profileID string,
		input dto.RoleInput,
	) (*profileutils.Role, error)

	// GetAllRoles returns a list of all created roles
	GetAllRolesFn func(ctx context.Context) (*[]profileutils.Role, error)

	// GetRoleByID gets role with matching id
	GetRoleByIDFn func(ctx context.Context, roleID string) (*profileutils.Role, error)

	// GetRoleByName retrieves a role using it's name
	GetRoleByNameFn func(ctx context.Context, roleName string) (*profileutils.Role, error)

	// GetRolesByIDs gets all roles matching provided roleIDs if specified otherwise all roles
	GetRolesByIDsFn func(ctx context.Context, roleIDs []string) (*[]profileutils.Role, error)

	// CheckIfRoleNameExists checks if a role with a similar name exists
	// Ensures unique name for each role during creation
	CheckIfRoleNameExistsFn func(ctx context.Context, name string) (bool, error)

	// UpdateRoleDetails  updates the details of a role
	UpdateRoleDetailsFn func(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error)
	// DeleteRole removes a role permanently from the database
	DeleteRoleFn func(ctx context.Context, roleID string) (bool, error)

	//CheckIfUserHasPermission checks if a user has the required permission
	CheckIfUserHasPermissionFn func(
		ctx context.Context,
		UID string,
		requiredPermission profileutils.Permission,
	) (bool, error)

	// GetUserProfilesByRoleID returns a list of user profiles with the role ID
	// f.e users assigned a particular role
	GetUserProfilesByRoleIDFn func(ctx context.Context, role string) ([]*profileutils.UserProfile, error)

	// SaveRoleRevocation records a log for a role revocation
	//
	// userId is the ID of the user removing a role from a user
	SaveRoleRevocationFn func(ctx context.Context, userID string, revocation dto.RoleRevocationInput) error

	// CheckIfAdmin checks if a user has admin permissions
	CheckIfAdminFn func(profile *profileutils.UserProfile) bool

	// UpdateUserName updates the username of a profile that matches the id
	// this method should be called after asserting the username is unique and not associated with another userProfile
	UpdateUserNameFn func(ctx context.Context, id string, userName string) error

	// UpdatePrimaryPhoneNumber append a new primary phone number to the user profile
	// this method should be called after asserting the phone number is unique and not associated with another userProfile
	UpdatePrimaryPhoneNumberFn func(ctx context.Context, id string, phoneNumber string) error

	// UpdatePrimaryEmailAddress the primary email addresses of the profile that matches the id
	// this method should be called after asserting the emailAddress is unique and not associated with another userProfile
	UpdatePrimaryEmailAddressFn func(ctx context.Context, id string, emailAddress string) error

	// UpdateSecondaryPhoneNumbers updates the secondary phone numbers of the profile that matches the id
	// this method should be called after asserting the phone numbers are unique and not associated with another userProfile
	UpdateSecondaryPhoneNumbersFn func(ctx context.Context, id string, phoneNumbers []string) error

	// UpdateSecondaryEmailAddresses the secondary email addresses of the profile that matches the id
	// this method should be called after asserting the emailAddresses  as unique and not associated with another userProfile
	UpdateSecondaryEmailAddressesFn func(ctx context.Context, id string, emailAddresses []string) error

	// UpdateVerifiedIdentifiers adds a UID to a user profile during login if it does not exist
	UpdateVerifiedIdentifiersFn func(
		ctx context.Context,
		id string,
		identifiers []profileutils.VerifiedIdentifier,
	) error

	// UpdateVerifiedUIDS adds a UID to a user profile during login if it does not exist
	UpdateVerifiedUIDSFn func(ctx context.Context, id string, uids []string) error

	// UpdateSuspended updates the suspend attribute of the profile that matches the id
	UpdateSuspendedFn func(ctx context.Context, id string, status bool) error

	// UpdatePhotoUploadID updates the photoUploadID attribute of the profile that matches the id
	UpdatePhotoUploadIDFn func(ctx context.Context, id string, uploadID string) error

	// UpdatePushTokens updates the pushTokens attribute of the profile that matches the id. This function does a hard reset instead of prior
	// matching
	UpdatePushTokensFn func(ctx context.Context, id string, pushToken []string) error

	// UpdatePermissions update the permissions of the user profile
	UpdatePermissionsFn func(ctx context.Context, id string, perms []profileutils.PermissionType) error

	// UpdateRole update the permissions of the user profile
	UpdateRoleFn func(ctx context.Context, id string, role profileutils.RoleType) error

	// UpdateUserRoleIDs updates the roles for a user
	UpdateUserRoleIDsFn func(ctx context.Context, id string, roleIDs []string) error

	// UpdateBioData updates the biodate of the profile that matches the id
	UpdateBioDataFn func(ctx context.Context, id string, data profileutils.BioData) error

	// UpdateAddresses persists a user's home or work address information to the database
	UpdateAddressesFn func(
		ctx context.Context,
		id string,
		address profileutils.Address,
		addressType enumutils.AddressType,
	) error

	// UpdateFavNavActions update the permissions of the user profile
	UpdateFavNavActionsFn func(ctx context.Context, id string, favActions []string) error

	// ListUserProfiles fetches all users with the specified role from the database
	ListUserProfilesFn func(
		ctx context.Context,
		role profileutils.RoleType,
	) ([]*profileutils.UserProfile, error)

	// GetUserProfileByPhoneOrEmail gets usser profile by phone or email
	GetUserProfileByPhoneOrEmailFn func(ctx context.Context, payload *dto.RetrieveUserProfileInput) (*profileutils.UserProfile, error)

	// UpdateUserProfileEmail updates user profile's email
	UpdateUserProfileEmailFn func(ctx context.Context, phone string, email string) error

	// CreateUserProfile creates a user profile of using the provided phone number and uid
	CreateUserProfileFn func(
		ctx context.Context,
		phoneNumber, uid string,
	) (*profileutils.UserProfile, error)

	// CreateDetailedUserProfile creates a new user profile that is pre-filled using the provided phone number
	CreateDetailedUserProfileFn func(
		ctx context.Context,
		phoneNumber string,
		profile profileutils.UserProfile,
	) (*profileutils.UserProfile, error)

	// GetUserProfileByUID fetches a user profile by uid
	GetUserProfileByUIDFn func(
		ctx context.Context,
		uid string,
		suspended bool,
	) (*profileutils.UserProfile, error)

	// GetUserProfileByID fetches a user profile by id. returns the unsuspend profile
	GetUserProfileByIDFn func(
		ctx context.Context,
		id string,
		suspended bool,
	) (*profileutils.UserProfile, error)

	// GetUserProfileByPhoneNumber fetches a user profile by phone number
	GetUserProfileByPhoneNumberFn func(
		ctx context.Context,
		phoneNumber string,
		suspended bool,
	) (*profileutils.UserProfile, error)

	// GetUserProfileByPrimaryPhoneNumber fetches a user profile by primary phone number
	GetUserProfileByPrimaryPhoneNumberFn func(
		ctx context.Context,
		phoneNumber string,
		suspend bool,
	) (*profileutils.UserProfile, error)

	// CheckIfPhoneNumberExists checks if a specific phone number has already been registered to another user
	CheckIfPhoneNumberExistsFn func(ctx context.Context, phone string) (bool, error)

	// CheckIfEmailExists checks if a specific email has already been registered to another user
	CheckIfEmailExistsFn func(ctx context.Context, email string) (bool, error)

	// CheckIfUsernameExists checks if a specific username has already been registered to another user
	CheckIfUsernameExistsFn func(ctx context.Context, username string) (bool, error)

	// GenerateAuthCredentialsForAnonymousUser ...
	GenerateAuthCredentialsForAnonymousUserFn func(
		ctx context.Context,
	) (*profileutils.AuthCredentialResponse, error)

	// GenerateAuthCredentials ...
	GenerateAuthCredentialsFn func(
		ctx context.Context,
		phone string,
		profile *profileutils.UserProfile,
	) (*profileutils.AuthCredentialResponse, error)

	// FetchAdminUsers ...
	FetchAdminUsersFn func(ctx context.Context) ([]*profileutils.UserProfile, error)

	// PurgeUserByPhoneNumber removes user completely. This should be used only under testing environment
	PurgeUserByPhoneNumberFn func(ctx context.Context, phone string) error

	// HardResetSecondaryPhoneNumbers ...
	HardResetSecondaryPhoneNumbersFn func(
		ctx context.Context,
		profile *profileutils.UserProfile,
		newSecondaryPhones []string,
	) error

	// HardResetSecondaryEmailAddress ...
	HardResetSecondaryEmailAddressFn func(
		ctx context.Context,
		profile *profileutils.UserProfile,
		newSecondaryEmails []string,
	) error

	// GetPINByProfileID ...
	GetPINByProfileIDFn func(ctx context.Context, ProfileID string) (*domain.PIN, error)

	// RecordPostVisitSurvey records the  post visit survey
	RecordPostVisitSurveyFn func(ctx context.Context, input dto.PostVisitSurveyInput, UID string) error

	// SavePIN  User Pin methods
	SavePINFn func(ctx context.Context, pin *domain.PIN) (bool, error)

	// UpdatePIN ...
	UpdatePINFn func(ctx context.Context, id string, pin *domain.PIN) (bool, error)

	// ExchangeRefreshTokenForIDToken ...
	ExchangeRefreshTokenForIDTokenFn func(ctx context.Context, token string) (*profileutils.AuthCredentialResponse, error)

	// GetOrCreatePhoneNumberUser ...
	GetOrCreatePhoneNumberUserFn func(ctx context.Context, phone string) (*dto.CreatedUserResponse, error)

	// AddUserAsExperimentParticipant ...
	AddUserAsExperimentParticipantFn func(ctx context.Context, profile *profileutils.UserProfile) (bool, error)

	// RemoveUserAsExperimentParticipant ...
	RemoveUserAsExperimentParticipantFn func(ctx context.Context, profile *profileutils.UserProfile) (bool, error)

	// CheckIfExperimentParticipant ...
	CheckIfExperimentParticipantFn func(ctx context.Context, profileID string) (bool, error)

	// GetUserCommunicationsSettings ...
	GetUserCommunicationsSettingsFn func(ctx context.Context, profileID string) (*profileutils.UserCommunicationsSetting, error)

	// SetUserCommunicationsSettings ...
	SetUserCommunicationsSettingsFn func(ctx context.Context, profileID string,
		allowWhatsApp *bool, allowTextSms *bool, allowPush *bool, allowEmail *bool) (*profileutils.UserCommunicationsSetting, error)

	// ResolveDefaultNudgeByTitle ...
	ResolveDefaultNudgeByTitleFn func(ctx context.Context, UID string, flavour feedlib.Flavour, nudgeTitle string) error

	// SendMail ...
	SendMailFn func(ctx context.Context, email string, message string, subject string) error

	// GenerateAndSendOTP ...
	GenerateAndSendOTPFn func(ctx context.Context, phone string, appID *string) (*profileutils.OtpResponse, error)

	// SendRetryOTP ...
	SendRetryOTPFn func(ctx context.Context, msisdn string, retryStep int, appID *string) (*profileutils.OtpResponse, error)

	// VerifyOTP ...
	VerifyOTPFn func(ctx context.Context, phone, OTP string) (bool, error)

	// VerifyEmailOTP ...
	VerifyEmailOTPFn func(ctx context.Context, email, OTP string) (bool, error)

	// SendSMS ...
	SendSMSFn func(ctx context.Context, phoneNumbers []string, message string) error

	// AddEngagementPubsubNameSpace creates a namespaced topic that resembles the one in
	// engagement service, which is prepended with the word "engagement". This solves the problem
	// where namespaced topics from "onboarding" are different from the ones in engagement.
	// This fix allows for uniformity of topic names between the engagement and onboarding services.
	AddEngagementPubsubNameSpaceFn func(topic string) string

	// AddPubSubNamespace creates a namespaced topic name
	AddPubSubNamespaceFn func(topicName string) string

	// TopicIDs returns the known (registered) topic IDs
	TopicIDsFn func() []string

	// PublishToPubsub sends a message to a specifeid Topic
	PublishToPubsubFn func(ctx context.Context, topicID string, payload []byte) error

	// EnsureTopicsExist creates the topic(s) in the suppplied list if they do not
	// already exist.
	EnsureTopicsExistFn func(ctx context.Context, topicIDs []string) error

	// EnsureSubscriptionsExist ensures that the subscriptions named in the supplied
	// topic:subscription map exist. If any does not exist, it is created.
	EnsureSubscriptionsExistFn func(ctx context.Context) error

	// SubscriptionIDs returns a map of topic IDs to subscription IDs
	SubscriptionIDsFn func() map[string]string
}

// StageProfileNudge stages nudges published from this service.
func (f FakeInfrastructure) StageProfileNudge(ctx context.Context, nudge *feedlib.Nudge) error {
	return f.StageProfileNudgeFn(ctx, nudge)
}

// CreateRole creates a new role and persists it to the database
func (f FakeInfrastructure) CreateRole(
	ctx context.Context,
	profileID string,
	input dto.RoleInput,
) (*profileutils.Role, error) {
	return f.CreateRoleFn(ctx, profileID, input)
}

// GetAllRoles returns a list of all created roles
func (f FakeInfrastructure) GetAllRoles(ctx context.Context) (*[]profileutils.Role, error) {
	return f.GetAllRolesFn(ctx)
}

// GetRoleByID gets role with matching id
func (f FakeInfrastructure) GetRoleByID(ctx context.Context, roleID string) (*profileutils.Role, error) {
	return f.GetRoleByIDFn(ctx, roleID)
}

// GetRoleByName retrieves a role using it's name
func (f FakeInfrastructure) GetRoleByName(ctx context.Context, roleName string) (*profileutils.Role, error) {
	return f.GetRoleByNameFn(ctx, roleName)
}

// GetRolesByIDs gets all roles matching provided roleIDs if specified otherwise all roles
func (f FakeInfrastructure) GetRolesByIDs(ctx context.Context, roleIDs []string) (*[]profileutils.Role, error) {
	return f.GetRolesByIDsFn(ctx, roleIDs)
}

// CheckIfRoleNameExists checks if a role with a similar name exists
// Ensures unique name for each role during creation
func (f FakeInfrastructure) CheckIfRoleNameExists(ctx context.Context, name string) (bool, error) {
	return f.CheckIfRoleNameExistsFn(ctx, name)
}

// UpdateRoleDetails  updates the details of a role
func (f FakeInfrastructure) UpdateRoleDetails(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error) {
	return f.UpdateRoleDetailsFn(ctx, profileID, role)
}

// DeleteRole removes a role permanently from the database
func (f FakeInfrastructure) DeleteRole(ctx context.Context, roleID string) (bool, error) {
	return f.DeleteRoleFn(ctx, roleID)
}

//CheckIfUserHasPermission checks if a user has the required permission
func (f FakeInfrastructure) CheckIfUserHasPermission(
	ctx context.Context,
	UID string,
	requiredPermission profileutils.Permission,
) (bool, error) {
	return f.CheckIfUserHasPermissionFn(ctx, UID, requiredPermission)
}

// GetUserProfilesByRoleID returns a list of user profiles with the role ID
// f.e users assigned a particular role
func (f FakeInfrastructure) GetUserProfilesByRoleID(ctx context.Context, role string) ([]*profileutils.UserProfile, error) {
	return f.GetUserProfilesByRoleIDFn(ctx, role)
}

// SaveRoleRevocation records a log for a role revocation
//
// userId is the ID of the user removing a role from a user
func (f FakeInfrastructure) SaveRoleRevocation(ctx context.Context, userID string, revocation dto.RoleRevocationInput) error {
	return f.SaveRoleRevocationFn(ctx, userID, revocation)
}

// CheckIfAdmin checks if a user has admin permissions
func (f FakeInfrastructure) CheckIfAdmin(profile *profileutils.UserProfile) bool {
	return f.CheckIfAdminFn(profile)
}

// UpdateUserName updates the username of a profile that matches the id
// this method should be called after asserting the username is unique and not associated with another userProfile
func (f FakeInfrastructure) UpdateUserName(ctx context.Context, id string, userName string) error {
	return f.UpdateUserNameFn(ctx, id, userName)
}

// UpdatePrimaryPhoneNumber append a new primary phone number to the user profile
// this method should be called after asserting the phone number is unique and not associated with another userProfile
func (f FakeInfrastructure) UpdatePrimaryPhoneNumber(ctx context.Context, id string, phoneNumber string) error {
	return f.UpdatePrimaryPhoneNumberFn(ctx, id, phoneNumber)
}

// UpdatePrimaryEmailAddress the primary email addresses of the profile that matches the id
// this method should be called after asserting the emailAddress is unique and not associated with another userProfile
func (f FakeInfrastructure) UpdatePrimaryEmailAddress(ctx context.Context, id string, emailAddress string) error {
	return f.UpdatePrimaryEmailAddressFn(ctx, id, emailAddress)
}

// UpdateSecondaryPhoneNumbers updates the secondary phone numbers of the profile that matches the id
// this method should be called after asserting the phone numbers are unique and not associated with another userProfile
func (f FakeInfrastructure) UpdateSecondaryPhoneNumbers(ctx context.Context, id string, phoneNumbers []string) error {
	return f.UpdateSecondaryPhoneNumbersFn(ctx, id, phoneNumbers)
}

// UpdateSecondaryEmailAddresses the secondary email addresses of the profile that matches the id
// this method should be called after asserting the emailAddresses  as unique and not associated with another userProfile
func (f FakeInfrastructure) UpdateSecondaryEmailAddresses(ctx context.Context, id string, emailAddresses []string) error {
	return f.UpdateSecondaryEmailAddressesFn(ctx, id, emailAddresses)
}

// UpdateVerifiedIdentifiers adds a UID to a user profile during login if it does not exist
func (f FakeInfrastructure) UpdateVerifiedIdentifiers(
	ctx context.Context,
	id string,
	identifiers []profileutils.VerifiedIdentifier,
) error {
	return f.UpdateVerifiedIdentifiersFn(ctx, id, identifiers)
}

// UpdateVerifiedUIDS adds a UID to a user profile during login if it does not exist
func (f FakeInfrastructure) UpdateVerifiedUIDS(ctx context.Context, id string, uids []string) error {
	return f.UpdateVerifiedUIDSFn(ctx, id, uids)
}

// UpdateSuspended updates the suspend attribute of the profile that matches the id
func (f FakeInfrastructure) UpdateSuspended(ctx context.Context, id string, status bool) error {
	return f.UpdateSuspendedFn(ctx, id, status)
}

// UpdatePhotoUploadID updates the photoUploadID attribute of the profile that matches the id
func (f FakeInfrastructure) UpdatePhotoUploadID(ctx context.Context, id string, uploadID string) error {
	return f.UpdatePhotoUploadIDFn(ctx, id, uploadID)
}

// UpdatePushTokens updates the pushTokens attribute of the profile that matches the id. This function does a hard reset instead of prior
// matching
func (f FakeInfrastructure) UpdatePushTokens(ctx context.Context, id string, pushToken []string) error {
	return f.UpdatePushTokensFn(ctx, id, pushToken)
}

// UpdatePermissions update the permissions of the user profile
func (f FakeInfrastructure) UpdatePermissions(ctx context.Context, id string, perms []profileutils.PermissionType) error {
	return f.UpdatePermissionsFn(ctx, id, perms)
}

// UpdateRole update the permissions of the user profile
func (f FakeInfrastructure) UpdateRole(ctx context.Context, id string, role profileutils.RoleType) error {
	return f.UpdateRoleFn(ctx, id, role)
}

// UpdateUserRoleIDs updates the roles for a user
func (f FakeInfrastructure) UpdateUserRoleIDs(ctx context.Context, id string, roleIDs []string) error {
	return f.UpdateUserRoleIDsFn(ctx, id, roleIDs)
}

// UpdateBioData updates the biodate of the profile that matches the id
func (f FakeInfrastructure) UpdateBioData(ctx context.Context, id string, data profileutils.BioData) error {
	return f.UpdateBioDataFn(ctx, id, data)
}

// UpdateAddresses persists a user's home or work address information to the database
func (f FakeInfrastructure) UpdateAddresses(
	ctx context.Context,
	id string,
	address profileutils.Address,
	addressType enumutils.AddressType,
) error {
	return f.UpdateAddressesFn(ctx, id, address, addressType)
}

// UpdateFavNavActions update the permissions of the user profile
func (f FakeInfrastructure) UpdateFavNavActions(ctx context.Context, id string, favActions []string) error {
	return f.UpdateFavNavActionsFn(ctx, id, favActions)
}

// ListUserProfiles fetches all users with the specified role from the database
func (f FakeInfrastructure) ListUserProfiles(
	ctx context.Context,
	role profileutils.RoleType,
) ([]*profileutils.UserProfile, error) {
	return f.ListUserProfilesFn(ctx, role)
}

// GetUserProfileByPhoneOrEmail gets usser profile by phone or email
func (f FakeInfrastructure) GetUserProfileByPhoneOrEmail(ctx context.Context, payload *dto.RetrieveUserProfileInput) (*profileutils.UserProfile, error) {
	return f.GetUserProfileByPhoneOrEmailFn(ctx, payload)
}

// UpdateUserProfileEmail updates user profile's email
func (f FakeInfrastructure) UpdateUserProfileEmail(ctx context.Context, phone string, email string) error {
	return f.UpdateUserProfileEmailFn(ctx, phone, email)
}

// CreateUserProfile creates a user profile of using the provided phone number and uid
func (f FakeInfrastructure) CreateUserProfile(
	ctx context.Context,
	phoneNumber, uid string,
) (*profileutils.UserProfile, error) {
	return f.CreateUserProfileFn(ctx, phoneNumber, uid)
}

// CreateDetailedUserProfile creates a new user profile that is pre-filled using the provided phone number
func (f FakeInfrastructure) CreateDetailedUserProfile(
	ctx context.Context,
	phoneNumber string,
	profile profileutils.UserProfile,
) (*profileutils.UserProfile, error) {
	return f.CreateDetailedUserProfileFn(ctx, phoneNumber, profile)
}

// GetUserProfileByUID fetches a user profile by uid
func (f FakeInfrastructure) GetUserProfileByUID(
	ctx context.Context,
	uid string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	return f.GetUserProfileByUIDFn(ctx, uid, suspended)
}

// GetUserProfileByID fetches a user profile by id. returns the unsuspend profile
func (f FakeInfrastructure) GetUserProfileByID(
	ctx context.Context,
	id string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	return f.GetUserProfileByIDFn(ctx, id, suspended)
}

// GetUserProfileByPhoneNumber fetches a user profile by phone number
func (f FakeInfrastructure) GetUserProfileByPhoneNumber(
	ctx context.Context,
	phoneNumber string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	return f.GetUserProfileByPhoneNumberFn(ctx, phoneNumber, suspended)
}

// GetUserProfileByPrimaryPhoneNumber fetches a user profile by primary phone number
func (f FakeInfrastructure) GetUserProfileByPrimaryPhoneNumber(
	ctx context.Context,
	phoneNumber string,
	suspend bool,
) (*profileutils.UserProfile, error) {
	return f.GetUserProfileByPrimaryPhoneNumberFn(ctx, phoneNumber, suspend)
}

// CheckIfPhoneNumberExists checks if a specific phone number has already been registered to another user
func (f FakeInfrastructure) CheckIfPhoneNumberExists(ctx context.Context, phone string) (bool, error) {
	return f.CheckIfPhoneNumberExistsFn(ctx, phone)
}

// CheckIfEmailExists checks if a specific email has already been registered to another user
func (f FakeInfrastructure) CheckIfEmailExists(ctx context.Context, email string) (bool, error) {
	return f.CheckIfEmailExistsFn(ctx, email)
}

// CheckIfUsernameExists checks if a specific username has already been registered to another user
func (f FakeInfrastructure) CheckIfUsernameExists(ctx context.Context, username string) (bool, error) {
	return f.CheckIfUsernameExistsFn(ctx, username)
}

// GenerateAuthCredentialsForAnonymousUser ...
func (f FakeInfrastructure) GenerateAuthCredentialsForAnonymousUser(
	ctx context.Context,
) (*profileutils.AuthCredentialResponse, error) {
	return f.GenerateAuthCredentialsForAnonymousUserFn(ctx)
}

// GenerateAuthCredentials ...
func (f FakeInfrastructure) GenerateAuthCredentials(
	ctx context.Context,
	phone string,
	profile *profileutils.UserProfile,
) (*profileutils.AuthCredentialResponse, error) {
	return f.GenerateAuthCredentialsFn(ctx, phone, profile)
}

// FetchAdminUsers ...
func (f FakeInfrastructure) FetchAdminUsers(ctx context.Context) ([]*profileutils.UserProfile, error) {
	return f.FetchAdminUsersFn(ctx)
}

// PurgeUserByPhoneNumber removes user completely. This should be used only under testing environment
func (f FakeInfrastructure) PurgeUserByPhoneNumber(ctx context.Context, phone string) error {
	return f.PurgeUserByPhoneNumberFn(ctx, phone)
}

// HardResetSecondaryPhoneNumbers ...
func (f FakeInfrastructure) HardResetSecondaryPhoneNumbers(
	ctx context.Context,
	profile *profileutils.UserProfile,
	newSecondaryPhones []string,
) error {
	return f.HardResetSecondaryPhoneNumbersFn(ctx, profile, newSecondaryPhones)
}

// HardResetSecondaryEmailAddress ...
func (f FakeInfrastructure) HardResetSecondaryEmailAddress(
	ctx context.Context,
	profile *profileutils.UserProfile,
	newSecondaryEmails []string,
) error {
	return f.HardResetSecondaryEmailAddressFn(ctx, profile, newSecondaryEmails)
}

// GetPINByProfileID ...
func (f FakeInfrastructure) GetPINByProfileID(
	ctx context.Context,
	ProfileID string,
) (*domain.PIN, error) {
	return f.GetPINByProfileIDFn(ctx, ProfileID)
}

// RecordPostVisitSurvey records the  post visit survey
func (f FakeInfrastructure) RecordPostVisitSurvey(
	ctx context.Context,
	input dto.PostVisitSurveyInput,
	UID string,
) error {
	return f.RecordPostVisitSurveyFn(ctx, input, UID)
}

// SavePIN  User Pin methods
func (f FakeInfrastructure) SavePIN(ctx context.Context, pin *domain.PIN) (bool, error) {
	return f.SavePINFn(ctx, pin)
}

// UpdatePIN ...
func (f FakeInfrastructure) UpdatePIN(ctx context.Context, id string, pin *domain.PIN) (bool, error) {
	return f.UpdatePINFn(ctx, id, pin)
}

// ExchangeRefreshTokenForIDToken ...
func (f FakeInfrastructure) ExchangeRefreshTokenForIDToken(
	ctx context.Context,
	token string,
) (*profileutils.AuthCredentialResponse, error) {
	return f.ExchangeRefreshTokenForIDTokenFn(ctx, token)
}

// GetOrCreatePhoneNumberUser ...
func (f FakeInfrastructure) GetOrCreatePhoneNumberUser(ctx context.Context, phone string) (*dto.CreatedUserResponse, error) {
	return f.GetOrCreatePhoneNumberUserFn(ctx, phone)
}

// AddUserAsExperimentParticipant ...
func (f FakeInfrastructure) AddUserAsExperimentParticipant(
	ctx context.Context,
	profile *profileutils.UserProfile,
) (bool, error) {
	return f.AddUserAsExperimentParticipantFn(ctx, profile)
}

// RemoveUserAsExperimentParticipant ...
func (f FakeInfrastructure) RemoveUserAsExperimentParticipant(
	ctx context.Context,
	profile *profileutils.UserProfile,
) (bool, error) {
	return f.RemoveUserAsExperimentParticipantFn(ctx, profile)
}

// CheckIfExperimentParticipant ...
func (f FakeInfrastructure) CheckIfExperimentParticipant(ctx context.Context, profileID string) (bool, error) {
	return f.CheckIfExperimentParticipantFn(ctx, profileID)
}

// GetUserCommunicationsSettings ...
func (f FakeInfrastructure) GetUserCommunicationsSettings(
	ctx context.Context,
	profileID string,
) (*profileutils.UserCommunicationsSetting, error) {
	return f.GetUserCommunicationsSettingsFn(ctx, profileID)
}

// SetUserCommunicationsSettings ...
func (f FakeInfrastructure) SetUserCommunicationsSettings(ctx context.Context, profileID string,
	allowWhatsApp *bool, allowTextSms *bool, allowPush *bool, allowEmail *bool) (*profileutils.UserCommunicationsSetting, error) {
	return f.SetUserCommunicationsSettingsFn(ctx, profileID, allowWhatsApp, allowTextSms, allowPush, allowEmail)
}

// ResolveDefaultNudgeByTitle ...
func (f FakeInfrastructure) ResolveDefaultNudgeByTitle(ctx context.Context, UID string, flavour feedlib.Flavour,
	nudgeTitle string) error {
	return f.ResolveDefaultNudgeByTitleFn(ctx, UID, flavour, nudgeTitle)
}

// SendMail ...
func (f FakeInfrastructure) SendMail(ctx context.Context, email string, message string, subject string) error {
	return f.SendMailFn(ctx, email, message, subject)
}

// GenerateAndSendOTP ...
func (f FakeInfrastructure) GenerateAndSendOTP(ctx context.Context, phone string, appID *string) (*profileutils.OtpResponse, error) {
	return f.GenerateAndSendOTPFn(ctx, phone, appID)
}

// SendRetryOTP ...
func (f FakeInfrastructure) SendRetryOTP(ctx context.Context, msisdn string, retryStep int, appID *string) (*profileutils.OtpResponse, error) {
	return f.SendRetryOTPFn(ctx, msisdn, retryStep, appID)
}

// VerifyOTP ...
func (f FakeInfrastructure) VerifyOTP(ctx context.Context, phone, OTP string) (bool, error) {
	return f.VerifyOTPFn(ctx, phone, OTP)
}

// VerifyEmailOTP ...
func (f FakeInfrastructure) VerifyEmailOTP(ctx context.Context, email, OTP string) (bool, error) {
	return f.VerifyEmailOTPFn(ctx, email, OTP)
}

// SendSMS ...
func (f FakeInfrastructure) SendSMS(ctx context.Context, phoneNumbers []string, message string) error {
	return f.SendSMSFn(ctx, phoneNumbers, message)
}

// AddEngagementPubsubNameSpace creates a namespaced topic that resembles the one in
// engagement service, which is prepended with the word "engagement". This solves the problem
// where namespaced topics from "onboarding" are different from the ones in engagement.
// This fix allows for uniformity of topic names between the engagement and onboarding services.
func (f FakeInfrastructure) AddEngagementPubsubNameSpace(
	topic string,
) string {
	return f.AddEngagementPubsubNameSpaceFn(topic)
}

// AddPubSubNamespace creates a namespaced topic name
func (f FakeInfrastructure) AddPubSubNamespace(topicName string) string {
	return f.AddPubSubNamespaceFn(topicName)
}

// TopicIDs returns the known (registered) topic IDs
func (f FakeInfrastructure) TopicIDs() []string {
	return f.TopicIDsFn()
}

// PublishToPubsub sends a message to a specifeid Topic
func (f FakeInfrastructure) PublishToPubsub(ctx context.Context, topicID string, payload []byte) error {
	return f.PublishToPubsubFn(ctx, topicID, payload)
}

// EnsureTopicsExist creates the topic(s) in the suppplied list if they do not
// already exist.
func (f FakeInfrastructure) EnsureTopicsExist(
	ctx context.Context,
	topicIDs []string,
) error {
	return f.EnsureTopicsExistFn(
		ctx,
		topicIDs,
	)
}

// EnsureSubscriptionsExist ensures that the subscriptions named in the supplied
// topic:subscription map exist. If any does not exist, it is created.
func (f FakeInfrastructure) EnsureSubscriptionsExist(
	ctx context.Context,
) error {
	return f.EnsureSubscriptionsExistFn(ctx)
}

// SubscriptionIDs returns a map of topic IDs to subscription IDs
func (f FakeInfrastructure) SubscriptionIDs() map[string]string {
	return f.SubscriptionIDsFn()
}
