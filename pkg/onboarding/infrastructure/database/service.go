package database

import (
	"context"
	"log"

	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/database/fb"
	"github.com/savannahghi/profileutils"
)

// Repository interface that provide access to all persistent storage operations
type Repository interface {
	UserProfileRepository

	SupplierRepository

	CustomerRepository

	RolesRepository

	// creates a user profile of using the provided phone number and uid
	CreateUserProfile(
		ctx context.Context,
		phoneNumber, uid string,
	) (*profileutils.UserProfile, error)

	// creates a new user profile that is pre-filled using the provided phone number
	CreateDetailedUserProfile(
		ctx context.Context,
		phoneNumber string,
		profile profileutils.UserProfile,
	) (*profileutils.UserProfile, error)

	// fetches a user profile by uid
	GetUserProfileByUID(
		ctx context.Context,
		uid string,
		suspended bool,
	) (*profileutils.UserProfile, error)

	// fetches a user profile by id. returns the unsuspend profile
	GetUserProfileByID(
		ctx context.Context,
		id string,
		suspended bool,
	) (*profileutils.UserProfile, error)

	// fetches a user profile by phone number
	GetUserProfileByPhoneNumber(
		ctx context.Context,
		phoneNumber string,
		suspended bool,
	) (*profileutils.UserProfile, error)

	// fetches a user profile by primary phone number
	GetUserProfileByPrimaryPhoneNumber(
		ctx context.Context,
		phoneNumber string,
		suspend bool,
	) (*profileutils.UserProfile, error)

	// checks if a specific phone number has already been registered to another user
	CheckIfPhoneNumberExists(ctx context.Context, phone string) (bool, error)

	// checks if a specific email has already been registered to another user
	CheckIfEmailExists(ctx context.Context, phone string) (bool, error)

	// checks if a specific username has already been registered to another user
	CheckIfUsernameExists(ctx context.Context, phone string) (bool, error)

	GenerateAuthCredentialsForAnonymousUser(
		ctx context.Context,
	) (*profileutils.AuthCredentialResponse, error)

	GenerateAuthCredentials(
		ctx context.Context,
		phone string,
		profile *profileutils.UserProfile,
	) (*profileutils.AuthCredentialResponse, error)

	FetchAdminUsers(ctx context.Context) ([]*profileutils.UserProfile, error)

	FetchAllUsers(ctx context.Context, callbackURL string)

	// removes user completely. This should be used only under testing environment
	PurgeUserByPhoneNumber(ctx context.Context, phone string) error

	HardResetSecondaryPhoneNumbers(
		ctx context.Context,
		profile *profileutils.UserProfile,
		newSecondaryPhones []string,
	) error

	HardResetSecondaryEmailAddress(
		ctx context.Context,
		profile *profileutils.UserProfile,
		newSecondaryEmails []string,
	) error

	// PINs
	GetPINByProfileID(
		ctx context.Context,
		ProfileID string,
	) (*domain.PIN, error)

	// User Pin methods
	SavePIN(ctx context.Context, pin *domain.PIN) (bool, error)
	UpdatePIN(ctx context.Context, id string, pin *domain.PIN) (bool, error)

	ExchangeRefreshTokenForIDToken(
		ctx context.Context,
		token string,
	) (*profileutils.AuthCredentialResponse, error)

	GetOrCreatePhoneNumberUser(ctx context.Context, phone string) (*dto.CreatedUserResponse, error)

	AddUserAsExperimentParticipant(
		ctx context.Context,
		profile *profileutils.UserProfile,
	) (bool, error)

	RemoveUserAsExperimentParticipant(
		ctx context.Context,
		profile *profileutils.UserProfile,
	) (bool, error)

	CheckIfExperimentParticipant(ctx context.Context, profileID string) (bool, error)

	GetUserCommunicationsSettings(
		ctx context.Context,
		profileID string,
	) (*profileutils.UserCommunicationsSetting, error)

	SetUserCommunicationsSettings(ctx context.Context, profileID string,
		allowWhatsApp *bool, allowTextSms *bool, allowPush *bool, allowEmail *bool) (*profileutils.UserCommunicationsSetting, error)
}

// SupplierRepository  defines signatures that relate to suppliers
type SupplierRepository interface {
	StageProfileNudge(ctx context.Context, nudge *feedlib.Nudge) error
	CheckIfAdmin(profile *profileutils.UserProfile) bool
}

// CustomerRepository  defines signatures that relate to customers
type CustomerRepository interface {
	// GetUserProfileByPhoneOrEmail gets usser profile by phone or email
	GetUserProfileByPhoneOrEmail(ctx context.Context, payload *dto.RetrieveUserProfileInput) (*profileutils.UserProfile, error)

	// UpdateUserProfileEmail updates user profile's email
	UpdateUserProfileEmail(ctx context.Context, phone string, email string) error
}

// UserProfileRepository interface that provide access to all persistent storage operations for user profile
type UserProfileRepository interface {
	UpdateUserName(ctx context.Context, id string, userName string) error
	UpdatePrimaryPhoneNumber(ctx context.Context, id string, phoneNumber string) error
	UpdatePrimaryEmailAddress(ctx context.Context, id string, emailAddress string) error
	UpdateSecondaryPhoneNumbers(ctx context.Context, id string, phoneNumbers []string) error
	UpdateSecondaryEmailAddresses(ctx context.Context, id string, emailAddresses []string) error
	UpdateVerifiedIdentifiers(
		ctx context.Context,
		id string,
		identifiers []profileutils.VerifiedIdentifier,
	) error
	UpdateVerifiedUIDS(ctx context.Context, id string, uids []string) error
	UpdateSuspended(ctx context.Context, id string, status bool) error
	UpdatePhotoUploadID(ctx context.Context, id string, uploadID string) error
	UpdatePushTokens(ctx context.Context, id string, pushToken []string) error
	UpdatePermissions(ctx context.Context, id string, perms []profileutils.PermissionType) error
	UpdateRole(ctx context.Context, id string, role profileutils.RoleType) error
	UpdateUserRoleIDs(ctx context.Context, id string, roleIDs []string) error
	UpdateBioData(ctx context.Context, id string, data profileutils.BioData) error
	UpdateAddresses(
		ctx context.Context,
		id string,
		address profileutils.Address,
		addressType enumutils.AddressType,
	) error
	UpdateFavNavActions(ctx context.Context, id string, favActions []string) error
	ListUserProfiles(
		ctx context.Context,
		role profileutils.RoleType,
	) ([]*profileutils.UserProfile, error)
}

// RolesRepository interface that provide access to all persistent storage operations for roles
type RolesRepository interface {
	CreateRole(
		ctx context.Context,
		profileID string,
		input dto.RoleInput,
	) (*profileutils.Role, error)

	GetAllRoles(ctx context.Context) (*[]profileutils.Role, error)

	GetRoleByID(ctx context.Context, roleID string) (*profileutils.Role, error)

	GetRoleByName(ctx context.Context, roleName string) (*profileutils.Role, error)

	GetRolesByIDs(ctx context.Context, roleIDs []string) (*[]profileutils.Role, error)

	CheckIfRoleNameExists(ctx context.Context, name string) (bool, error)

	UpdateRoleDetails(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error)

	DeleteRole(ctx context.Context, roleID string) (bool, error)

	CheckIfUserHasPermission(
		ctx context.Context,
		UID string,
		requiredPermission profileutils.Permission,
	) (bool, error)

	// GetUserProfilesByRole retrieves userprofiles with a particular role
	GetUserProfilesByRoleID(ctx context.Context, role string) ([]*profileutils.UserProfile, error)

	SaveRoleRevocation(ctx context.Context, userID string, revocation dto.RoleRevocationInput) error
}

// DbService is an implementation of the database repository
// It is implementation agnostic i.e logic should be handled using
// the preferred database
type DbService struct {
	firestore *fb.Repository
}

// NewDbService creates a new database service
func NewDbService() *DbService {
	ctx := context.Background()
	fc := &firebasetools.FirebaseClient{}
	firebaseApp, err := fc.InitFirebase()
	if err != nil {
		return nil
	}
	fbc, err := firebaseApp.Auth(ctx)
	if err != nil {
		log.Panicf("can't initialize Firebase auth when setting up profile service: %s", err)
	}
	fsc, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("unable to initialize Firestore: %s", err)
	}
	firestoreExtension := fb.NewFirestoreClientExtension(fsc)

	firestore := fb.NewFirebaseRepository(firestoreExtension, fbc)
	return &DbService{
		firestore: firestore,
	}
}

// CheckPreconditions ensures correct initialization
func (d DbService) CheckPreconditions() {
	if d.firestore == nil {
		log.Panicf("nil firestore service in database service")
	}
}

// StageProfileNudge stages nudges published from this service.
func (d DbService) StageProfileNudge(ctx context.Context, nudge *feedlib.Nudge) error {
	return d.firestore.StageProfileNudge(ctx, nudge)
}

// CheckIfAdmin checks if a user has admin permissions
func (d DbService) CheckIfAdmin(profile *profileutils.UserProfile) bool {
	return d.firestore.CheckIfAdmin(profile)
}

// UpdateUserName updates the username of a profile that matches the id
// this method should be called after asserting the username is unique and not associated with another userProfile
func (d DbService) UpdateUserName(ctx context.Context, id string, userName string) error {
	return d.firestore.UpdateUserName(ctx, id, userName)
}

// UpdatePrimaryPhoneNumber append a new primary phone number to the user profile
// this method should be called after asserting the phone number is unique and not associated with another userProfile
func (d DbService) UpdatePrimaryPhoneNumber(ctx context.Context, id string, phoneNumber string) error {
	return d.firestore.UpdatePrimaryPhoneNumber(ctx, id, phoneNumber)
}

// UpdatePrimaryEmailAddress the primary email addresses of the profile that matches the id
// this method should be called after asserting the emailAddress is unique and not associated with another userProfile
func (d DbService) UpdatePrimaryEmailAddress(ctx context.Context, id string, emailAddress string) error {
	return d.firestore.UpdatePrimaryEmailAddress(ctx, id, emailAddress)
}

// UpdateSecondaryPhoneNumbers updates the secondary phone numbers of the profile that matches the id
// this method should be called after asserting the phone numbers are unique and not associated with another userProfile
func (d DbService) UpdateSecondaryPhoneNumbers(ctx context.Context, id string, phoneNumbers []string) error {
	return d.firestore.UpdateSecondaryPhoneNumbers(ctx, id, phoneNumbers)
}

// UpdateSecondaryEmailAddresses the secondary email addresses of the profile that matches the id
// this method should be called after asserting the emailAddresses  as unique and not associated with another userProfile
func (d DbService) UpdateSecondaryEmailAddresses(ctx context.Context, id string, emailAddresses []string) error {
	return d.firestore.UpdateSecondaryEmailAddresses(ctx, id, emailAddresses)
}

// UpdateVerifiedIdentifiers adds a UID to a user profile during login if it does not exist
func (d DbService) UpdateVerifiedIdentifiers(
	ctx context.Context,
	id string,
	identifiers []profileutils.VerifiedIdentifier,
) error {
	return d.firestore.UpdateVerifiedIdentifiers(ctx, id, identifiers)
}

// UpdateVerifiedUIDS adds a UID to a user profile during login if it does not exist
func (d DbService) UpdateVerifiedUIDS(ctx context.Context, id string, uids []string) error {
	return d.firestore.UpdateVerifiedUIDS(ctx, id, uids)
}

// UpdateSuspended updates the suspend attribute of the profile that matches the id
func (d DbService) UpdateSuspended(ctx context.Context, id string, status bool) error {
	return d.firestore.UpdateSuspended(ctx, id, status)
}

// UpdatePhotoUploadID updates the photoUploadID attribute of the profile that matches the id
func (d DbService) UpdatePhotoUploadID(ctx context.Context, id string, uploadID string) error {
	return d.firestore.UpdatePhotoUploadID(ctx, id, uploadID)
}

// UpdatePushTokens updates the pushTokens attribute of the profile that matches the id. This function does a hard reset instead of prior
// matching
func (d DbService) UpdatePushTokens(ctx context.Context, id string, pushToken []string) error {
	return d.firestore.UpdatePushTokens(ctx, id, pushToken)
}

// UpdatePermissions update the permissions of the user profile
func (d DbService) UpdatePermissions(ctx context.Context, id string, perms []profileutils.PermissionType) error {
	return d.firestore.UpdatePermissions(ctx, id, perms)
}

// UpdateRole update the permissions of the user profile
func (d DbService) UpdateRole(ctx context.Context, id string, role profileutils.RoleType) error {
	return d.firestore.UpdateRole(ctx, id, role)
}

// UpdateUserRoleIDs updates the roles for a user
func (d DbService) UpdateUserRoleIDs(ctx context.Context, id string, roleIDs []string) error {
	return d.firestore.UpdateUserRoleIDs(ctx, id, roleIDs)
}

// UpdateBioData updates the biodate of the profile that matches the id
func (d DbService) UpdateBioData(ctx context.Context, id string, data profileutils.BioData) error {
	return d.firestore.UpdateBioData(ctx, id, data)
}

// UpdateAddresses persists a user's home or work address information to the database
func (d DbService) UpdateAddresses(
	ctx context.Context,
	id string,
	address profileutils.Address,
	addressType enumutils.AddressType,
) error {
	return d.firestore.UpdateAddresses(ctx, id, address, addressType)
}

// UpdateFavNavActions update the permissions of the user profile
func (d DbService) UpdateFavNavActions(ctx context.Context, id string, favActions []string) error {
	return d.firestore.UpdateFavNavActions(ctx, id, favActions)
}

// ListUserProfiles fetches all users with the specified role from the database
func (d DbService) ListUserProfiles(
	ctx context.Context,
	role profileutils.RoleType,
) ([]*profileutils.UserProfile, error) {
	return d.firestore.ListUserProfiles(ctx, role)
}

// CreateRole creates a new role and persists it to the database
func (d DbService) CreateRole(
	ctx context.Context,
	profileID string,
	input dto.RoleInput,
) (*profileutils.Role, error) {
	return d.firestore.CreateRole(ctx, profileID, input)
}

// GetAllRoles returns a list of all created roles
func (d DbService) GetAllRoles(ctx context.Context) (*[]profileutils.Role, error) {
	return d.firestore.GetAllRoles(ctx)
}

// GetRoleByID gets role with matching id
func (d DbService) GetRoleByID(ctx context.Context, roleID string) (*profileutils.Role, error) {
	return d.firestore.GetRoleByID(ctx, roleID)
}

// GetRoleByName retrieves a role using it's name
func (d DbService) GetRoleByName(ctx context.Context, roleName string) (*profileutils.Role, error) {
	return d.firestore.GetRoleByName(ctx, roleName)
}

// GetRolesByIDs gets all roles matching provided roleIDs if specified otherwise all roles
func (d DbService) GetRolesByIDs(ctx context.Context, roleIDs []string) (*[]profileutils.Role, error) {
	return d.firestore.GetRolesByIDs(ctx, roleIDs)
}

// CheckIfRoleNameExists checks if a role with a similar name exists
// Ensures unique name for each role during creation
func (d DbService) CheckIfRoleNameExists(ctx context.Context, name string) (bool, error) {
	return d.firestore.CheckIfRoleNameExists(ctx, name)
}

// UpdateRoleDetails  updates the details of a role
func (d DbService) UpdateRoleDetails(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error) {
	return d.firestore.UpdateRoleDetails(ctx, profileID, role)
}

// DeleteRole removes a role permanently from the database
func (d DbService) DeleteRole(ctx context.Context, roleID string) (bool, error) {
	return d.firestore.DeleteRole(ctx, roleID)
}

// CheckIfUserHasPermission checks if a user has the required permission
func (d DbService) CheckIfUserHasPermission(
	ctx context.Context,
	UID string,
	requiredPermission profileutils.Permission,
) (bool, error) {
	return d.firestore.CheckIfUserHasPermission(ctx, UID, requiredPermission)
}

// GetUserProfilesByRoleID returns a list of user profiles with the role ID
// i.e users assigned a particular role
func (d DbService) GetUserProfilesByRoleID(ctx context.Context, role string) ([]*profileutils.UserProfile, error) {
	return d.firestore.GetUserProfilesByRoleID(ctx, role)
}

// CreateUserProfile creates a user profile of using the provided phone number and uid
func (d DbService) CreateUserProfile(
	ctx context.Context,
	phoneNumber, uid string,
) (*profileutils.UserProfile, error) {
	return d.firestore.CreateUserProfile(ctx, phoneNumber, uid)
}

// CreateDetailedUserProfile creates a new user profile that is pre-filled using the provided phone number
func (d DbService) CreateDetailedUserProfile(
	ctx context.Context,
	phoneNumber string,
	profile profileutils.UserProfile,
) (*profileutils.UserProfile, error) {
	return d.firestore.CreateDetailedUserProfile(ctx, phoneNumber, profile)
}

// GetUserProfileByUID fetches a user profile by uid
func (d DbService) GetUserProfileByUID(
	ctx context.Context,
	uid string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	return d.firestore.GetUserProfileByUID(ctx, uid, suspended)
}

// GetUserProfileByID fetches a user profile by id. returns the unsuspend profile
func (d DbService) GetUserProfileByID(
	ctx context.Context,
	id string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	return d.firestore.GetUserProfileByID(ctx, id, suspended)
}

// GetUserProfileByPhoneNumber fetches a user profile by phone number
func (d DbService) GetUserProfileByPhoneNumber(
	ctx context.Context,
	phoneNumber string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	return d.firestore.GetUserProfileByPhoneNumber(ctx, phoneNumber, suspended)
}

// GetUserProfileByPrimaryPhoneNumber fetches a user profile by primary phone number
func (d DbService) GetUserProfileByPrimaryPhoneNumber(
	ctx context.Context,
	phoneNumber string,
	suspend bool,
) (*profileutils.UserProfile, error) {
	return d.firestore.GetUserProfileByPrimaryPhoneNumber(ctx, phoneNumber, suspend)
}

// CheckIfPhoneNumberExists checks if a specific phone number has already been registered to another user
func (d DbService) CheckIfPhoneNumberExists(ctx context.Context, phone string) (bool, error) {
	return d.firestore.CheckIfPhoneNumberExists(ctx, phone)
}

// CheckIfEmailExists checks if a specific email has already been registered to another user
func (d DbService) CheckIfEmailExists(ctx context.Context, email string) (bool, error) {
	return d.firestore.CheckIfEmailExists(ctx, email)
}

// CheckIfUsernameExists checks if a specific username has already been registered to another user
func (d DbService) CheckIfUsernameExists(ctx context.Context, phone string) (bool, error) {
	username := phone
	return d.firestore.CheckIfUsernameExists(ctx, username)
}

// GenerateAuthCredentialsForAnonymousUser ...
func (d DbService) GenerateAuthCredentialsForAnonymousUser(
	ctx context.Context,
) (*profileutils.AuthCredentialResponse, error) {
	return d.firestore.GenerateAuthCredentialsForAnonymousUser(ctx)
}

// GenerateAuthCredentials ...
func (d DbService) GenerateAuthCredentials(
	ctx context.Context,
	phone string,
	profile *profileutils.UserProfile,
) (*profileutils.AuthCredentialResponse, error) {
	return d.firestore.GenerateAuthCredentials(ctx, phone, profile)
}

// FetchAdminUsers ...
func (d DbService) FetchAdminUsers(ctx context.Context) ([]*profileutils.UserProfile, error) {
	return d.firestore.FetchAdminUsers(ctx)
}

func (d DbService) FetchAllUsers(ctx context.Context, callbackURL string) {
	d.firestore.FetchAllUsers(ctx, callbackURL)
}

// PurgeUserByPhoneNumber removes user completely. This should be used only under testing environment
func (d DbService) PurgeUserByPhoneNumber(ctx context.Context, phone string) error {
	return d.firestore.PurgeUserByPhoneNumber(ctx, phone)
}

// HardResetSecondaryPhoneNumbers ...
func (d DbService) HardResetSecondaryPhoneNumbers(
	ctx context.Context,
	profile *profileutils.UserProfile,
	newSecondaryPhones []string,
) error {
	return d.firestore.HardResetSecondaryPhoneNumbers(ctx, profile, newSecondaryPhones)
}

// HardResetSecondaryEmailAddress ...
func (d DbService) HardResetSecondaryEmailAddress(
	ctx context.Context,
	profile *profileutils.UserProfile,
	newSecondaryEmails []string,
) error {
	return d.firestore.HardResetSecondaryEmailAddress(ctx, profile, newSecondaryEmails)
}

// GetPINByProfileID ...
func (d DbService) GetPINByProfileID(
	ctx context.Context,
	ProfileID string,
) (*domain.PIN, error) {
	return d.firestore.GetPINByProfileID(ctx, ProfileID)
}

// SavePIN  User Pin methods
func (d DbService) SavePIN(ctx context.Context, pin *domain.PIN) (bool, error) {
	return d.firestore.SavePIN(ctx, pin)
}

// UpdatePIN ...
func (d DbService) UpdatePIN(ctx context.Context, id string, pin *domain.PIN) (bool, error) {
	return d.firestore.UpdatePIN(ctx, id, pin)
}

// ExchangeRefreshTokenForIDToken ...
func (d DbService) ExchangeRefreshTokenForIDToken(
	ctx context.Context,
	token string,
) (*profileutils.AuthCredentialResponse, error) {
	return d.firestore.ExchangeRefreshTokenForIDToken(ctx, token)
}

// GetOrCreatePhoneNumberUser ...
func (d DbService) GetOrCreatePhoneNumberUser(ctx context.Context, phone string) (*dto.CreatedUserResponse, error) {
	return d.firestore.GetOrCreatePhoneNumberUser(ctx, phone)
}

// AddUserAsExperimentParticipant ...
func (d DbService) AddUserAsExperimentParticipant(
	ctx context.Context,
	profile *profileutils.UserProfile,
) (bool, error) {
	return d.firestore.AddUserAsExperimentParticipant(ctx, profile)
}

// RemoveUserAsExperimentParticipant ...
func (d DbService) RemoveUserAsExperimentParticipant(
	ctx context.Context,
	profile *profileutils.UserProfile,
) (bool, error) {
	return d.firestore.RemoveUserAsExperimentParticipant(ctx, profile)
}

// CheckIfExperimentParticipant ...
func (d DbService) CheckIfExperimentParticipant(ctx context.Context, profileID string) (bool, error) {
	return d.firestore.CheckIfExperimentParticipant(ctx, profileID)
}

// GetUserCommunicationsSettings ...
func (d DbService) GetUserCommunicationsSettings(
	ctx context.Context,
	profileID string,
) (*profileutils.UserCommunicationsSetting, error) {
	return d.firestore.GetUserCommunicationsSettings(ctx, profileID)
}

// SetUserCommunicationsSettings ...
func (d DbService) SetUserCommunicationsSettings(ctx context.Context, profileID string,
	allowWhatsApp *bool, allowTextSms *bool, allowPush *bool, allowEmail *bool) (*profileutils.UserCommunicationsSetting, error) {
	return d.firestore.SetUserCommunicationsSettings(ctx, profileID, allowWhatsApp, allowTextSms, allowPush, allowEmail)
}

// SaveRoleRevocation records a log for a role revocation
//
// userId is the ID of the user removing a role from a user
func (d DbService) SaveRoleRevocation(ctx context.Context, userID string, revocation dto.RoleRevocationInput) error {
	return d.firestore.SaveRoleRevocation(ctx, userID, revocation)
}

// GetUserProfileByPhoneOrEmail gets usser profile by phone or email
func (d DbService) GetUserProfileByPhoneOrEmail(ctx context.Context, payload *dto.RetrieveUserProfileInput) (*profileutils.UserProfile, error) {
	return d.firestore.GetUserProfileByPhoneOrEmail(ctx, payload)
}

// UpdateUserProfileEmail updates user profile's email
func (d DbService) UpdateUserProfileEmail(ctx context.Context, phone string, email string) error {
	return d.firestore.UpdateUserProfileEmail(ctx, phone, email)
}
