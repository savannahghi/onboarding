package infrastructure

import (
	"context"

	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/database"
	"github.com/savannahghi/profileutils"
)

// Infrastructure defines the contract provided by the infrastructure layer
// It's a combination of interactions with external services/dependencies
type Infrastructure interface {
	database.Repository
}

// Interactor is an implementation of the infrastructure interface
// It combines each individual service implementation
type Interactor struct {
	database *database.DbService
}

// NewInfrastructureInteractor initializes a new infrastructure interactor
func NewInfrastructureInteractor() *Interactor {
	db := database.NewDbService()
	return &Interactor{
		database: db,
	}
}

// CheckPreconditions ensures correct initialization
func (i Interactor) CheckPreconditions() {}

// StageProfileNudge stages nudges published from this service.
func (i Interactor) StageProfileNudge(ctx context.Context, nudge *feedlib.Nudge) error {
	return i.database.StageProfileNudge(ctx, nudge)
}

// CreateRole creates a new role and persists it to the database
func (i Interactor) CreateRole(
	ctx context.Context,
	profileID string,
	input dto.RoleInput,
) (*profileutils.Role, error) {
	return i.database.CreateRole(ctx, profileID, input)
}

// GetAllRoles returns a list of all created roles
func (i Interactor) GetAllRoles(ctx context.Context) (*[]profileutils.Role, error) {
	return i.database.GetAllRoles(ctx)
}

// GetRoleByID gets role with matching id
func (i Interactor) GetRoleByID(ctx context.Context, roleID string) (*profileutils.Role, error) {
	return i.database.GetRoleByID(ctx, roleID)
}

// GetRoleByName retrieves a role using it's name
func (i Interactor) GetRoleByName(ctx context.Context, roleName string) (*profileutils.Role, error) {
	return i.database.GetRoleByName(ctx, roleName)
}

// GetRolesByIDs gets all roles matching provided roleIDs if specified otherwise all roles
func (i Interactor) GetRolesByIDs(ctx context.Context, roleIDs []string) (*[]profileutils.Role, error) {
	return i.database.GetRolesByIDs(ctx, roleIDs)
}

// CheckIfRoleNameExists checks if a role with a similar name exists
// Ensures unique name for each role during creation
func (i Interactor) CheckIfRoleNameExists(ctx context.Context, name string) (bool, error) {
	return i.database.CheckIfRoleNameExists(ctx, name)
}

// UpdateRoleDetails  updates the details of a role
func (i Interactor) UpdateRoleDetails(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error) {
	return i.database.UpdateRoleDetails(ctx, profileID, role)
}

// DeleteRole removes a role permanently from the database
func (i Interactor) DeleteRole(ctx context.Context, roleID string) (bool, error) {
	return i.database.DeleteRole(ctx, roleID)
}

//CheckIfUserHasPermission checks if a user has the required permission
func (i Interactor) CheckIfUserHasPermission(
	ctx context.Context,
	UID string,
	requiredPermission profileutils.Permission,
) (bool, error) {
	return i.database.CheckIfUserHasPermission(ctx, UID, requiredPermission)
}

// GetUserProfilesByRoleID returns a list of user profiles with the role ID
// i.e users assigned a particular role
func (i Interactor) GetUserProfilesByRoleID(ctx context.Context, role string) ([]*profileutils.UserProfile, error) {
	return i.database.GetUserProfilesByRoleID(ctx, role)
}

// SaveRoleRevocation records a log for a role revocation
//
// userId is the ID of the user removing a role from a user
func (i Interactor) SaveRoleRevocation(ctx context.Context, userID string, revocation dto.RoleRevocationInput) error {
	return i.database.SaveRoleRevocation(ctx, userID, revocation)
}

// CheckIfAdmin checks if a user has admin permissions
func (i Interactor) CheckIfAdmin(profile *profileutils.UserProfile) bool {
	return i.database.CheckIfAdmin(profile)
}

// UpdateUserName updates the username of a profile that matches the id
// this method should be called after asserting the username is unique and not associated with another userProfile
func (i Interactor) UpdateUserName(ctx context.Context, id string, userName string) error {
	return i.database.UpdateUserName(ctx, id, userName)
}

// UpdatePrimaryPhoneNumber append a new primary phone number to the user profile
// this method should be called after asserting the phone number is unique and not associated with another userProfile
func (i Interactor) UpdatePrimaryPhoneNumber(ctx context.Context, id string, phoneNumber string) error {
	return i.database.UpdatePrimaryPhoneNumber(ctx, id, phoneNumber)
}

// UpdatePrimaryEmailAddress the primary email addresses of the profile that matches the id
// this method should be called after asserting the emailAddress is unique and not associated with another userProfile
func (i Interactor) UpdatePrimaryEmailAddress(ctx context.Context, id string, emailAddress string) error {
	return i.database.UpdatePrimaryEmailAddress(ctx, id, emailAddress)
}

// UpdateSecondaryPhoneNumbers updates the secondary phone numbers of the profile that matches the id
// this method should be called after asserting the phone numbers are unique and not associated with another userProfile
func (i Interactor) UpdateSecondaryPhoneNumbers(ctx context.Context, id string, phoneNumbers []string) error {
	return i.database.UpdateSecondaryPhoneNumbers(ctx, id, phoneNumbers)
}

// UpdateSecondaryEmailAddresses the secondary email addresses of the profile that matches the id
// this method should be called after asserting the emailAddresses  as unique and not associated with another userProfile
func (i Interactor) UpdateSecondaryEmailAddresses(ctx context.Context, id string, emailAddresses []string) error {
	return i.database.UpdateSecondaryEmailAddresses(ctx, id, emailAddresses)
}

// UpdateVerifiedIdentifiers adds a UID to a user profile during login if it does not exist
func (i Interactor) UpdateVerifiedIdentifiers(
	ctx context.Context,
	id string,
	identifiers []profileutils.VerifiedIdentifier,
) error {
	return i.database.UpdateVerifiedIdentifiers(ctx, id, identifiers)
}

// UpdateVerifiedUIDS adds a UID to a user profile during login if it does not exist
func (i Interactor) UpdateVerifiedUIDS(ctx context.Context, id string, uids []string) error {
	return i.database.UpdateVerifiedUIDS(ctx, id, uids)
}

// UpdateSuspended updates the suspend attribute of the profile that matches the id
func (i Interactor) UpdateSuspended(ctx context.Context, id string, status bool) error {
	return i.database.UpdateSuspended(ctx, id, status)
}

// UpdatePhotoUploadID updates the photoUploadID attribute of the profile that matches the id
func (i Interactor) UpdatePhotoUploadID(ctx context.Context, id string, uploadID string) error {
	return i.database.UpdatePhotoUploadID(ctx, id, uploadID)
}

// UpdateCovers updates the covers attribute of the profile that matches the id
func (i Interactor) UpdateCovers(ctx context.Context, id string, covers []profileutils.Cover) error {
	return i.database.UpdateCovers(ctx, id, covers)
}

// UpdatePushTokens updates the pushTokens attribute of the profile that matches the id. This function does a hard reset instead of prior
// matching
func (i Interactor) UpdatePushTokens(ctx context.Context, id string, pushToken []string) error {
	return i.database.UpdatePushTokens(ctx, id, pushToken)
}

// UpdatePermissions update the permissions of the user profile
func (i Interactor) UpdatePermissions(ctx context.Context, id string, perms []profileutils.PermissionType) error {
	return i.database.UpdatePermissions(ctx, id, perms)
}

// UpdateRole update the permissions of the user profile
func (i Interactor) UpdateRole(ctx context.Context, id string, role profileutils.RoleType) error {
	return i.database.UpdateRole(ctx, id, role)
}

// UpdateUserRoleIDs updates the roles for a user
func (i Interactor) UpdateUserRoleIDs(ctx context.Context, id string, roleIDs []string) error {
	return i.database.UpdateUserRoleIDs(ctx, id, roleIDs)
}

// UpdateBioData updates the biodate of the profile that matches the id
func (i Interactor) UpdateBioData(ctx context.Context, id string, data profileutils.BioData) error {
	return i.database.UpdateBioData(ctx, id, data)
}

// UpdateAddresses persists a user's home or work address information to the database
func (i Interactor) UpdateAddresses(
	ctx context.Context,
	id string,
	address profileutils.Address,
	addressType enumutils.AddressType,
) error {
	return i.database.UpdateAddresses(ctx, id, address, addressType)
}

// UpdateFavNavActions update the permissions of the user profile
func (i Interactor) UpdateFavNavActions(ctx context.Context, id string, favActions []string) error {
	return i.database.UpdateFavNavActions(ctx, id, favActions)
}

// ListUserProfiles fetches all users with the specified role from the database
func (i Interactor) ListUserProfiles(
	ctx context.Context,
	role profileutils.RoleType,
) ([]*profileutils.UserProfile, error) {
	return i.database.ListUserProfiles(ctx, role)
}

// GetUserProfileByPhoneOrEmail gets usser profile by phone or email
func (i Interactor) GetUserProfileByPhoneOrEmail(ctx context.Context, payload *dto.RetrieveUserProfileInput) (*profileutils.UserProfile, error) {
	return i.database.GetUserProfileByPhoneOrEmail(ctx, payload)
}

// UpdateUserProfileEmail updates user profile's email
func (i Interactor) UpdateUserProfileEmail(ctx context.Context, phone string, email string) error {
	return i.database.UpdateUserProfileEmail(ctx, phone, email)
}

// CreateUserProfile creates a user profile of using the provided phone number and uid
func (i Interactor) CreateUserProfile(
	ctx context.Context,
	phoneNumber, uid string,
) (*profileutils.UserProfile, error) {
	return i.database.CreateUserProfile(ctx, phoneNumber, uid)
}

// CreateDetailedUserProfile creates a new user profile that is pre-filled using the provided phone number
func (i Interactor) CreateDetailedUserProfile(
	ctx context.Context,
	phoneNumber string,
	profile profileutils.UserProfile,
) (*profileutils.UserProfile, error) {
	return i.database.CreateDetailedUserProfile(ctx, phoneNumber, profile)
}

// GetUserProfileByUID fetches a user profile by uid
func (i Interactor) GetUserProfileByUID(
	ctx context.Context,
	uid string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	return i.database.GetUserProfileByUID(ctx, uid, suspended)
}

// GetUserProfileByID fetches a user profile by id. returns the unsuspend profile
func (i Interactor) GetUserProfileByID(
	ctx context.Context,
	id string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	return i.database.GetUserProfileByID(ctx, id, suspended)
}

// GetUserProfileByPhoneNumber fetches a user profile by phone number
func (i Interactor) GetUserProfileByPhoneNumber(
	ctx context.Context,
	phoneNumber string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	return i.database.GetUserProfileByPhoneNumber(ctx, phoneNumber, suspended)
}

// GetUserProfileByPrimaryPhoneNumber fetches a user profile by primary phone number
func (i Interactor) GetUserProfileByPrimaryPhoneNumber(
	ctx context.Context,
	phoneNumber string,
	suspend bool,
) (*profileutils.UserProfile, error) {
	return i.database.GetUserProfileByPrimaryPhoneNumber(ctx, phoneNumber, suspend)
}

// CheckIfPhoneNumberExists checks if a specific phone number has already been registered to another user
func (i Interactor) CheckIfPhoneNumberExists(ctx context.Context, phone string) (bool, error) {
	return i.database.CheckIfPhoneNumberExists(ctx, phone)
}

// CheckIfEmailExists checks if a specific email has already been registered to another user
func (i Interactor) CheckIfEmailExists(ctx context.Context, email string) (bool, error) {
	return i.database.CheckIfEmailExists(ctx, email)
}

// CheckIfUsernameExists checks if a specific username has already been registered to another user
func (i Interactor) CheckIfUsernameExists(ctx context.Context, username string) (bool, error) {
	return i.database.CheckIfUsernameExists(ctx, username)
}

// GenerateAuthCredentialsForAnonymousUser ...
func (i Interactor) GenerateAuthCredentialsForAnonymousUser(
	ctx context.Context,
) (*profileutils.AuthCredentialResponse, error) {
	return i.database.GenerateAuthCredentialsForAnonymousUser(ctx)
}

// GenerateAuthCredentials ...
func (i Interactor) GenerateAuthCredentials(
	ctx context.Context,
	phone string,
	profile *profileutils.UserProfile,
) (*profileutils.AuthCredentialResponse, error) {
	return i.database.GenerateAuthCredentials(ctx, phone, profile)
}

// FetchAdminUsers ...
func (i Interactor) FetchAdminUsers(ctx context.Context) ([]*profileutils.UserProfile, error) {
	return i.database.FetchAdminUsers(ctx)
}

// PurgeUserByPhoneNumber removes user completely. This should be used only under testing environment
func (i Interactor) PurgeUserByPhoneNumber(ctx context.Context, phone string) error {
	return i.database.PurgeUserByPhoneNumber(ctx, phone)
}

// HardResetSecondaryPhoneNumbers ...
func (i Interactor) HardResetSecondaryPhoneNumbers(
	ctx context.Context,
	profile *profileutils.UserProfile,
	newSecondaryPhones []string,
) error {
	return i.database.HardResetSecondaryPhoneNumbers(ctx, profile, newSecondaryPhones)
}

// HardResetSecondaryEmailAddress ...
func (i Interactor) HardResetSecondaryEmailAddress(
	ctx context.Context,
	profile *profileutils.UserProfile,
	newSecondaryEmails []string,
) error {
	return i.database.HardResetSecondaryEmailAddress(ctx, profile, newSecondaryEmails)
}

// GetPINByProfileID ...
func (i Interactor) GetPINByProfileID(
	ctx context.Context,
	ProfileID string,
) (*domain.PIN, error) {
	return i.database.GetPINByProfileID(ctx, ProfileID)
}

// RecordPostVisitSurvey records the  post visit survey
func (i Interactor) RecordPostVisitSurvey(
	ctx context.Context,
	input dto.PostVisitSurveyInput,
	UID string,
) error {
	return i.database.RecordPostVisitSurvey(ctx, input, UID)
}

// SavePIN  User Pin methods
func (i Interactor) SavePIN(ctx context.Context, pin *domain.PIN) (bool, error) {
	return i.database.SavePIN(ctx, pin)
}

// UpdatePIN ...
func (i Interactor) UpdatePIN(ctx context.Context, id string, pin *domain.PIN) (bool, error) {
	return i.database.UpdatePIN(ctx, id, pin)
}

// ExchangeRefreshTokenForIDToken ...
func (i Interactor) ExchangeRefreshTokenForIDToken(
	ctx context.Context,
	token string,
) (*profileutils.AuthCredentialResponse, error) {
	return i.database.ExchangeRefreshTokenForIDToken(ctx, token)
}

// GetOrCreatePhoneNumberUser ...
func (i Interactor) GetOrCreatePhoneNumberUser(ctx context.Context, phone string) (*dto.CreatedUserResponse, error) {
	return i.database.GetOrCreatePhoneNumberUser(ctx, phone)
}

// AddUserAsExperimentParticipant ...
func (i Interactor) AddUserAsExperimentParticipant(
	ctx context.Context,
	profile *profileutils.UserProfile,
) (bool, error) {
	return i.database.AddUserAsExperimentParticipant(ctx, profile)
}

// RemoveUserAsExperimentParticipant ...
func (i Interactor) RemoveUserAsExperimentParticipant(
	ctx context.Context,
	profile *profileutils.UserProfile,
) (bool, error) {
	return i.database.RemoveUserAsExperimentParticipant(ctx, profile)
}

// CheckIfExperimentParticipant ...
func (i Interactor) CheckIfExperimentParticipant(ctx context.Context, profileID string) (bool, error) {
	return i.database.CheckIfExperimentParticipant(ctx, profileID)
}

// GetUserCommunicationsSettings ...
func (i Interactor) GetUserCommunicationsSettings(
	ctx context.Context,
	profileID string,
) (*profileutils.UserCommunicationsSetting, error) {
	return i.database.GetUserCommunicationsSettings(ctx, profileID)
}

// SetUserCommunicationsSettings ...
func (i Interactor) SetUserCommunicationsSettings(ctx context.Context, profileID string,
	allowWhatsApp *bool, allowTextSms *bool, allowPush *bool, allowEmail *bool) (*profileutils.UserCommunicationsSetting, error) {
	return i.database.SetUserCommunicationsSettings(ctx, profileID, allowWhatsApp, allowTextSms, allowPush, allowEmail)
}
