package repository

import (
	"context"

	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/profileutils"

	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
)

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

// OnboardingRepository interface that provide access to all persistent storage operations
type OnboardingRepository interface {
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

	// Record post visit survey
	RecordPostVisitSurvey(
		ctx context.Context,
		input dto.PostVisitSurveyInput,
		UID string,
	) error

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

//RolesRepository interface that provide access to all persistent storage operations for roles
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
