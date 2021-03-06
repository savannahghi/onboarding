package mock

import (
	"context"

	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/profileutils"
)

// FakeOnboardingRepository is a mock onboarding repository.
type FakeOnboardingRepository struct {
	AddRoleToUserfn func(ctx context.Context, phone string, role profileutils.RoleType) error

	StageProfileNudgeFn func(ctx context.Context, nudge *feedlib.Nudge) error

	CreateUserProfileFn func(ctx context.Context, phoneNumber, uid string) (*profileutils.UserProfile, error)

	CreateDetailedUserProfileFn func(ctx context.Context, phoneNumber string, profile profileutils.UserProfile) (*profileutils.UserProfile, error)

	// fetches a user profile by uid
	GetUserProfileByUIDFn func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error)

	// fetches user profile by email
	GetUserProfileByPhoneOrEmailFn func(ctx context.Context, payload *dto.RetrieveUserProfileInput) (*profileutils.UserProfile, error)

	// fetches a user profile by id
	GetUserProfileByIDFn func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error)

	// fetches a user profile by phone number
	GetUserProfileByPhoneNumberFn func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error)

	// fetches a user profile by primary phone number
	GetUserProfileByPrimaryPhoneNumberFn func(ctx context.Context, phoneNumber string, suspended bool) (*profileutils.UserProfile, error)

	// checks if a specific phone number has already been registered to another user
	CheckIfPhoneNumberExistsFn func(ctx context.Context, phone string) (bool, error)

	CheckIfEmailExistsFn func(ctx context.Context, email string) (bool, error)

	// checks if a specific username has already been registered to another user
	CheckIfUsernameExistsFn func(ctx context.Context, phone string) (bool, error)

	GenerateAuthCredentialsForAnonymousUserFn func(ctx context.Context) (*profileutils.AuthCredentialResponse, error)

	GenerateAuthCredentialsFn func(ctx context.Context, phone string, profile *profileutils.UserProfile) (*profileutils.AuthCredentialResponse, error)

	FetchAdminUsersFn func(ctx context.Context) ([]*profileutils.UserProfile, error)
	CheckIfAdminFn    func(profile *profileutils.UserProfile) bool

	FetchAllUsersFn func(ctx context.Context, callbackURL string)

	// removes user completely. This should be used only under testing environment
	PurgeUserByPhoneNumberFn func(ctx context.Context, phone string) error

	HardResetSecondaryPhoneNumbersFn func(ctx context.Context, profile *profileutils.UserProfile, phoneNumbers []string) error

	HardResetSecondaryEmailAddressFn func(ctx context.Context, profile *profileutils.UserProfile, newSecondaryEmails []string) error

	// PINs
	GetPINByProfileIDFn func(ctx context.Context, ProfileID string) (*domain.PIN, error)

	// Record post visit survey
	RecordPostVisitSurveyFn func(ctx context.Context, input dto.PostVisitSurveyInput, UID string) error

	// User Pin methods
	SavePINFn   func(ctx context.Context, pin *domain.PIN) (bool, error)
	UpdatePINFn func(ctx context.Context, id string, pin *domain.PIN) (bool, error)

	ExchangeRefreshTokenForIDTokenFn func(
		ctx context.Context,
		token string,
	) (*profileutils.AuthCredentialResponse, error)

	GetOrCreatePhoneNumberUserFn func(
		ctx context.Context,
		phone string,
	) (*dto.CreatedUserResponse, error)

	GetUserProfileAttributesFn func(
		ctx context.Context,
		UIDs []string,
		attribute string,
	) (map[string][]string, error)

	CheckIfExperimentParticipantFn func(ctx context.Context, profileID string) (bool, error)

	AddUserAsExperimentParticipantFn func(ctx context.Context, profile *profileutils.UserProfile) (bool, error)

	RemoveUserAsExperimentParticipantFn func(ctx context.Context, profile *profileutils.UserProfile) (bool, error)

	GetUserCommunicationsSettingsFn func(ctx context.Context, profileID string) (*profileutils.UserCommunicationsSetting, error)

	SetUserCommunicationsSettingsFn func(ctx context.Context, profileID string,
		allowWhatsApp *bool, allowTextSms *bool, allowPush *bool, allowEmail *bool) (*profileutils.UserCommunicationsSetting, error)

	// Userprofile
	UpdateUserNameFn                func(ctx context.Context, id string, phoneNumber string) error
	UpdatePrimaryPhoneNumberFn      func(ctx context.Context, id string, phoneNumber string) error
	UpdatePrimaryEmailAddressFn     func(ctx context.Context, id string, emailAddress string) error
	UpdateSecondaryPhoneNumbersFn   func(ctx context.Context, id string, phoneNumbers []string) error
	UpdateSecondaryEmailAddressesFn func(ctx context.Context, id string, emailAddresses []string) error
	UpdateUserRoleIDsFn             func(ctx context.Context, id string, roleIDs []string) error
	UpdateSuspendedFn               func(ctx context.Context, id string, status bool) error
	UpdatePhotoUploadIDFn           func(ctx context.Context, id string, uploadID string) error
	UpdatePushTokensFn              func(ctx context.Context, id string, pushToken []string) error
	UpdatePermissionsFn             func(ctx context.Context, id string, perms []profileutils.PermissionType) error
	UpdateRoleFn                    func(ctx context.Context, id string, role profileutils.RoleType) error
	UpdateBioDataFn                 func(ctx context.Context, id string, data profileutils.BioData) error
	UpdateVerifiedIdentifiersFn     func(ctx context.Context, id string, identifiers []profileutils.VerifiedIdentifier) error
	UpdateVerifiedUIDSFn            func(ctx context.Context, id string, uids []string) error
	UpdateAddressesFn               func(ctx context.Context, id string, address profileutils.Address, addressType enumutils.AddressType) error
	ListUserProfilesFn              func(ctx context.Context, role profileutils.RoleType) ([]*profileutils.UserProfile, error)
	UpdateOptOutFn                  func(ctx context.Context, option string, phoneNumber string) error
	UpdateFavNavActionsFn           func(ctx context.Context, id string, favActions []string) error //roles
	CreateRoleFn                    func(ctx context.Context, profileID string, role dto.RoleInput) (*profileutils.Role, error)
	GetAllRolesFn                   func(ctx context.Context) (*[]profileutils.Role, error)
	UpdateRoleDetailsFn             func(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error)
	GetRolesByIDsFn                 func(ctx context.Context, roleIDs []string) (*[]profileutils.Role, error)
	GetRoleByIDFn                   func(ctx context.Context, roleID string) (*profileutils.Role, error)
	GetRoleByNameFn                 func(ctx context.Context, roleName string) (*profileutils.Role, error)
	CheckIfRoleNameExistsFn         func(ctx context.Context, name string) (bool, error)
	DeleteRoleFn                    func(ctx context.Context, roleID string) (bool, error)
	CheckIfUserHasPermissionFn      func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error)
	UpdateUserProfileEmailFn        func(ctx context.Context, phone string, email string) error
	GetUserProfilesByRoleIDFn       func(ctx context.Context, role string) ([]*profileutils.UserProfile, error)
	SaveRoleRevocationFn            func(ctx context.Context, userID string, revocation dto.RoleRevocationInput) error
}

// CheckIfAdmin ...
func (f *FakeOnboardingRepository) CheckIfAdmin(profile *profileutils.UserProfile) bool {
	return f.CheckIfAdminFn(profile)
}

// AddRoleToUser ...
func (f *FakeOnboardingRepository) AddRoleToUser(
	ctx context.Context,
	phone string,
	role profileutils.RoleType,
) error {
	return f.AddRoleToUserfn(ctx, phone, role)
}

// StageProfileNudge ...
func (f *FakeOnboardingRepository) StageProfileNudge(
	ctx context.Context,
	nudge *feedlib.Nudge,
) error {
	return f.StageProfileNudgeFn(ctx, nudge)
}

// CreateUserProfile ...
func (f *FakeOnboardingRepository) CreateUserProfile(
	ctx context.Context,
	phoneNumber, uid string,
) (*profileutils.UserProfile, error) {
	return f.CreateUserProfileFn(ctx, phoneNumber, uid)
}

// GetUserProfileByUID fetches a user profile by uid
func (f *FakeOnboardingRepository) GetUserProfileByUID(
	ctx context.Context,
	uid string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	return f.GetUserProfileByUIDFn(ctx, uid, suspended)
}

// GetUserProfileByPhoneOrEmail fetches user profile by email or phone
func (f *FakeOnboardingRepository) GetUserProfileByPhoneOrEmail(ctx context.Context, payload *dto.RetrieveUserProfileInput) (*profileutils.UserProfile, error) {
	return f.GetUserProfileByPhoneOrEmailFn(ctx, payload)

}

// GetUserProfileByID fetches a user profile by id
func (f *FakeOnboardingRepository) GetUserProfileByID(
	ctx context.Context,
	id string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	return f.GetUserProfileByIDFn(ctx, id, suspended)
}

// GetUserProfileByPhoneNumber fetches a user profile by phone number
func (f *FakeOnboardingRepository) GetUserProfileByPhoneNumber(
	ctx context.Context,
	phoneNumber string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	return f.GetUserProfileByPhoneNumberFn(ctx, phoneNumber, suspended)
}

// GetUserProfileByPrimaryPhoneNumber fetches a user profile by primary phone number
func (f *FakeOnboardingRepository) GetUserProfileByPrimaryPhoneNumber(
	ctx context.Context,
	phoneNumber string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	return f.GetUserProfileByPrimaryPhoneNumberFn(ctx, phoneNumber, suspended)
}

// CheckIfPhoneNumberExists checks if a specific phone number has already been registered to another user
func (f *FakeOnboardingRepository) CheckIfPhoneNumberExists(
	ctx context.Context,
	phone string,
) (bool, error) {
	return f.CheckIfPhoneNumberExistsFn(ctx, phone)
}

// CheckIfEmailExists ...
func (f *FakeOnboardingRepository) CheckIfEmailExists(
	ctx context.Context,
	email string,
) (bool, error) {
	return f.CheckIfEmailExistsFn(ctx, email)
}

// CheckIfUsernameExists checks if a specific username has already been registered to another user
func (f *FakeOnboardingRepository) CheckIfUsernameExists(
	ctx context.Context,
	phone string,
) (bool, error) {
	return f.CheckIfUsernameExistsFn(ctx, phone)
}

// GenerateAuthCredentialsForAnonymousUser ...
func (f *FakeOnboardingRepository) GenerateAuthCredentialsForAnonymousUser(
	ctx context.Context,
) (*profileutils.AuthCredentialResponse, error) {
	return f.GenerateAuthCredentialsForAnonymousUserFn(ctx)
}

// GenerateAuthCredentials ...
func (f *FakeOnboardingRepository) GenerateAuthCredentials(
	ctx context.Context,
	phone string,
	profile *profileutils.UserProfile,
) (*profileutils.AuthCredentialResponse, error) {
	return f.GenerateAuthCredentialsFn(ctx, phone, profile)
}

// FetchAdminUsers ...
func (f *FakeOnboardingRepository) FetchAdminUsers(
	ctx context.Context,
) ([]*profileutils.UserProfile, error) {
	return f.FetchAdminUsersFn(ctx)
}

// PurgeUserByPhoneNumber removes user completely. This should be used only under testing environment
func (f *FakeOnboardingRepository) PurgeUserByPhoneNumber(ctx context.Context, phone string) error {
	return f.PurgeUserByPhoneNumberFn(ctx, phone)
}

// GetPINByProfileID PINs
func (f *FakeOnboardingRepository) GetPINByProfileID(
	ctx context.Context,
	ProfileID string,
) (*domain.PIN, error) {
	return f.GetPINByProfileIDFn(ctx, ProfileID)
}

//RecordPostVisitSurvey Record post visit survey
func (f *FakeOnboardingRepository) RecordPostVisitSurvey(
	ctx context.Context,
	input dto.PostVisitSurveyInput,
	UID string,
) error {
	return f.RecordPostVisitSurveyFn(ctx, input, UID)
}

//SavePIN  User Pin methods
func (f *FakeOnboardingRepository) SavePIN(ctx context.Context, pin *domain.PIN) (bool, error) {
	return f.SavePINFn(ctx, pin)
}

// UpdatePIN ...
func (f *FakeOnboardingRepository) UpdatePIN(
	ctx context.Context,
	id string,
	pin *domain.PIN,
) (bool, error) {
	return f.UpdatePINFn(ctx, id, pin)
}

// ExchangeRefreshTokenForIDToken ...
func (f *FakeOnboardingRepository) ExchangeRefreshTokenForIDToken(
	ctx context.Context,
	token string,
) (*profileutils.AuthCredentialResponse, error) {
	return f.ExchangeRefreshTokenForIDTokenFn(ctx, token)
}

// UpdateUserName ...
func (f *FakeOnboardingRepository) UpdateUserName(
	ctx context.Context,
	id string,
	phoneNumber string,
) error {
	return f.UpdateUserNameFn(ctx, id, phoneNumber)
}

// UpdatePrimaryPhoneNumber ...
func (f *FakeOnboardingRepository) UpdatePrimaryPhoneNumber(
	ctx context.Context,
	id string,
	phoneNumber string,
) error {
	return f.UpdatePrimaryPhoneNumberFn(ctx, id, phoneNumber)
}

// UpdatePrimaryEmailAddress ...
func (f *FakeOnboardingRepository) UpdatePrimaryEmailAddress(
	ctx context.Context,
	id string,
	emailAddress string,
) error {
	return f.UpdatePrimaryEmailAddressFn(ctx, id, emailAddress)
}

// UpdateSecondaryPhoneNumbers ...
func (f *FakeOnboardingRepository) UpdateSecondaryPhoneNumbers(
	ctx context.Context,
	id string,
	phoneNumbers []string,
) error {
	return f.UpdateSecondaryPhoneNumbersFn(ctx, id, phoneNumbers)
}

// UpdateSecondaryEmailAddresses ...
func (f *FakeOnboardingRepository) UpdateSecondaryEmailAddresses(
	ctx context.Context,
	id string,
	emailAddresses []string,
) error {
	return f.UpdateSecondaryEmailAddressesFn(ctx, id, emailAddresses)
}

// UpdateSuspended ...
func (f *FakeOnboardingRepository) UpdateSuspended(
	ctx context.Context,
	id string,
	status bool,
) error {
	return f.UpdateSuspendedFn(ctx, id, status)
}

// UpdatePhotoUploadID ...
func (f *FakeOnboardingRepository) UpdatePhotoUploadID(
	ctx context.Context,
	id string,
	uploadID string,
) error {
	return f.UpdatePhotoUploadIDFn(ctx, id, uploadID)
}

// UpdatePushTokens ...
func (f *FakeOnboardingRepository) UpdatePushTokens(
	ctx context.Context,
	id string,
	pushToken []string,
) error {
	return f.UpdatePushTokensFn(ctx, id, pushToken)
}

// UpdatePermissions ...
func (f *FakeOnboardingRepository) UpdatePermissions(
	ctx context.Context,
	id string,
	perms []profileutils.PermissionType,
) error {
	return f.UpdatePermissionsFn(ctx, id, perms)
}

// UpdateRole ...
func (f *FakeOnboardingRepository) UpdateRole(
	ctx context.Context,
	id string,
	role profileutils.RoleType,
) error {
	return f.UpdateRoleFn(ctx, id, role)
}

// UpdateBioData ...
func (f *FakeOnboardingRepository) UpdateBioData(
	ctx context.Context,
	id string,
	data profileutils.BioData,
) error {
	return f.UpdateBioDataFn(ctx, id, data)
}

// UpdateVerifiedIdentifiers ...
func (f *FakeOnboardingRepository) UpdateVerifiedIdentifiers(
	ctx context.Context,
	id string,
	identifiers []profileutils.VerifiedIdentifier,
) error {
	return f.UpdateVerifiedIdentifiersFn(ctx, id, identifiers)
}

// UpdateVerifiedUIDS ...
func (f *FakeOnboardingRepository) UpdateVerifiedUIDS(
	ctx context.Context,
	id string,
	uids []string,
) error {
	return f.UpdateVerifiedUIDSFn(ctx, id, uids)
}

// GetOrCreatePhoneNumberUser ...
func (f *FakeOnboardingRepository) GetOrCreatePhoneNumberUser(ctx context.Context,
	phone string,
) (*dto.CreatedUserResponse, error) {
	return f.GetOrCreatePhoneNumberUserFn(ctx, phone)
}

// HardResetSecondaryPhoneNumbers ...
func (f *FakeOnboardingRepository) HardResetSecondaryPhoneNumbers(
	ctx context.Context,
	profile *profileutils.UserProfile,
	phoneNumbers []string,
) error {
	return f.HardResetSecondaryPhoneNumbersFn(ctx, profile, phoneNumbers)
}

// HardResetSecondaryEmailAddress ...
func (f *FakeOnboardingRepository) HardResetSecondaryEmailAddress(
	ctx context.Context,
	profile *profileutils.UserProfile,
	newSecondaryEmails []string,
) error {
	return f.HardResetSecondaryEmailAddressFn(ctx, profile, newSecondaryEmails)
}

// GetUserProfileAttributes ...
func (f *FakeOnboardingRepository) GetUserProfileAttributes(
	ctx context.Context,
	UIDs []string,
	attribute string,
) (map[string][]string, error) {
	return f.GetUserProfileAttributesFn(
		ctx,
		UIDs,
		attribute,
	)
}

// CheckIfExperimentParticipant ...
func (f *FakeOnboardingRepository) CheckIfExperimentParticipant(
	ctx context.Context,
	profileID string,
) (bool, error) {
	return f.CheckIfExperimentParticipantFn(ctx, profileID)
}

// AddUserAsExperimentParticipant ...
func (f *FakeOnboardingRepository) AddUserAsExperimentParticipant(
	ctx context.Context,
	profile *profileutils.UserProfile,
) (bool, error) {
	return f.AddUserAsExperimentParticipantFn(ctx, profile)
}

// RemoveUserAsExperimentParticipant ...
func (f *FakeOnboardingRepository) RemoveUserAsExperimentParticipant(
	ctx context.Context,
	profile *profileutils.UserProfile,
) (bool, error) {
	return f.RemoveUserAsExperimentParticipantFn(ctx, profile)
}

// UpdateAddresses ...
func (f *FakeOnboardingRepository) UpdateAddresses(
	ctx context.Context,
	id string,
	address profileutils.Address,
	addressType enumutils.AddressType,
) error {
	return f.UpdateAddressesFn(ctx, id, address, addressType)
}

// GetUserCommunicationsSettings ...
func (f *FakeOnboardingRepository) GetUserCommunicationsSettings(
	ctx context.Context,
	profileID string,
) (*profileutils.UserCommunicationsSetting, error) {
	return f.GetUserCommunicationsSettingsFn(ctx, profileID)
}

// SetUserCommunicationsSettings ...
func (f *FakeOnboardingRepository) SetUserCommunicationsSettings(
	ctx context.Context,
	profileID string,
	allowWhatsApp *bool,
	allowTextSms *bool,
	allowPush *bool,
	allowEmail *bool,
) (*profileutils.UserCommunicationsSetting, error) {
	return f.SetUserCommunicationsSettingsFn(
		ctx,
		profileID,
		allowWhatsApp,
		allowTextSms,
		allowPush,
		allowEmail,
	)
}

// ListUserProfiles ...
func (f *FakeOnboardingRepository) ListUserProfiles(
	ctx context.Context,
	role profileutils.RoleType,
) ([]*profileutils.UserProfile, error) {
	return f.ListUserProfilesFn(ctx, role)
}

// CreateDetailedUserProfile ...
func (f *FakeOnboardingRepository) CreateDetailedUserProfile(
	ctx context.Context,
	phoneNumber string,
	profile profileutils.UserProfile,
) (*profileutils.UserProfile, error) {
	return f.CreateDetailedUserProfileFn(ctx, phoneNumber, profile)
}

// UpdateFavNavActions ...
func (f *FakeOnboardingRepository) UpdateFavNavActions(
	ctx context.Context,
	id string,
	favActions []string,
) error {
	return f.UpdateFavNavActionsFn(ctx, id, favActions)
}

//CreateRole ...
func (f *FakeOnboardingRepository) CreateRole(
	ctx context.Context,
	profileID string,
	input dto.RoleInput,
) (*profileutils.Role, error) {
	return f.CreateRoleFn(ctx, profileID, input)
}

//UpdateRoleDetails ...
func (f *FakeOnboardingRepository) UpdateRoleDetails(
	ctx context.Context,
	profileID string,
	role profileutils.Role,
) (*profileutils.Role, error) {
	return f.UpdateRoleDetailsFn(ctx, profileID, role)
}

//GetRoleByID ...
func (f *FakeOnboardingRepository) GetRoleByID(
	ctx context.Context,
	roleID string,
) (*profileutils.Role, error) {
	return f.GetRoleByIDFn(ctx, roleID)
}

//GetAllRoles ...
func (f *FakeOnboardingRepository) GetAllRoles(
	ctx context.Context,
) (*[]profileutils.Role, error) {
	return f.GetAllRolesFn(ctx)
}

// GetRolesByIDs ...
func (f *FakeOnboardingRepository) GetRolesByIDs(
	ctx context.Context,
	roleIDs []string,
) (*[]profileutils.Role, error) {
	return f.GetRolesByIDsFn(ctx, roleIDs)
}

// CheckIfRoleNameExists ...
func (f *FakeOnboardingRepository) CheckIfRoleNameExists(
	ctx context.Context,
	name string,
) (bool, error) {
	return f.CheckIfRoleNameExistsFn(ctx, name)
}

// CheckIfUserHasPermission ...
func (f *FakeOnboardingRepository) CheckIfUserHasPermission(
	ctx context.Context,
	UID string,
	requiredPermission profileutils.Permission,
) (bool, error) {
	return f.CheckIfUserHasPermissionFn(ctx, UID, requiredPermission)
}

// UpdateUserRoleIDs ...
func (f *FakeOnboardingRepository) UpdateUserRoleIDs(
	ctx context.Context,
	id string,
	roleIDs []string,
) error {
	return f.UpdateUserRoleIDsFn(ctx, id, roleIDs)
}

// DeleteRole ...
func (f *FakeOnboardingRepository) DeleteRole(ctx context.Context, roleID string) (bool, error) {
	return f.DeleteRoleFn(ctx, roleID)
}

// UpdateUserProfileEmail ...
func (f *FakeOnboardingRepository) UpdateUserProfileEmail(ctx context.Context, phone string, email string) error {
	return f.UpdateUserProfileEmailFn(ctx, phone, email)
}

// GetRoleByName ...
func (f *FakeOnboardingRepository) GetRoleByName(ctx context.Context, roleName string) (*profileutils.Role, error) {
	return f.GetRoleByNameFn(ctx, roleName)
}

// GetUserProfilesByRoleID ...
func (f *FakeOnboardingRepository) GetUserProfilesByRoleID(ctx context.Context, role string) ([]*profileutils.UserProfile, error) {
	return f.GetUserProfilesByRoleIDFn(ctx, role)
}

// SaveRoleRevocation ...
func (f *FakeOnboardingRepository) SaveRoleRevocation(ctx context.Context, userID string, revocation dto.RoleRevocationInput) error {
	return f.SaveRoleRevocationFn(ctx, userID, revocation)
}

// FetchAllUsers ...
func (f *FakeOnboardingRepository) FetchAllUsers(ctx context.Context, callbackURL string) {
	f.FetchAllUsersFn(ctx, callbackURL)
}
