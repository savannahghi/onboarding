package usecases

import (
	"context"

	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases/admin"
	"github.com/savannahghi/profileutils"
)

// Usecases is an interface that combines of all usescases
type Usecases interface {
	LoginUseCases
	ProfileUseCase
	RoleUseCase
	SignUpUseCases
	SurveyUseCases
	UserPINUseCases
	admin.Usecase
}

// Interactor is an implementation of the usecases interface
type Interactor struct {
	login    LoginUseCases
	profile  ProfileUseCase
	roles    RoleUseCase
	signup   SignUpUseCases
	surveys  SurveyUseCases
	pins     UserPINUseCases
	services admin.Usecase
}

// NewUsecasesInteractor initializes a new usecases interactor
func NewUsecasesInteractor(infrastructure infrastructure.Infrastructure, baseExtension extension.BaseExtension, pinsExtension extension.PINExtension) Usecases {

	profile := NewProfileUseCase(infrastructure, baseExtension)
	login := NewLoginUseCases(infrastructure, profile, baseExtension, pinsExtension)
	roles := NewRoleUseCases(infrastructure, baseExtension)
	pins := NewUserPinUseCase(infrastructure, profile, baseExtension, pinsExtension)
	signup := NewSignUpUseCases(infrastructure, profile, pins, baseExtension)
	surveys := NewSurveyUseCases(infrastructure, baseExtension)
	services := admin.NewService(baseExtension)

	impl := &Interactor{
		login:    login,
		profile:  profile,
		roles:    roles,
		signup:   signup,
		surveys:  surveys,
		pins:     pins,
		services: services,
	}

	return impl
}

// LoginByPhone ...
func (i *Interactor) LoginByPhone(
	ctx context.Context,
	phone string,
	PIN string,
	flavour feedlib.Flavour,
) (*profileutils.UserResponse, error) {
	return i.login.LoginByPhone(ctx, phone, PIN, flavour)
}

// RefreshToken ...
func (i *Interactor) RefreshToken(ctx context.Context, token string) (*profileutils.AuthCredentialResponse, error) {
	return i.login.RefreshToken(ctx, token)
}

// LoginAsAnonymous ...
func (i *Interactor) LoginAsAnonymous(ctx context.Context) (*profileutils.AuthCredentialResponse, error) {
	return i.login.LoginAsAnonymous(ctx)
}

// ResumeWithPin ...
func (i *Interactor) ResumeWithPin(ctx context.Context, pin string) (bool, error) {
	return i.login.ResumeWithPin(ctx, pin)
}

// UserProfile ...
func (i *Interactor) UserProfile(ctx context.Context) (*profileutils.UserProfile, error) {
	return i.profile.UserProfile(ctx)
}

// GetProfileByID ...
func (i *Interactor) GetProfileByID(ctx context.Context, id *string) (*profileutils.UserProfile, error) {
	return i.profile.GetProfileByID(ctx, id)
}

// UpdateUserName ...
func (i *Interactor) UpdateUserName(ctx context.Context, userName string) error {
	return i.profile.UpdateUserName(ctx, userName)
}

// UpdatePrimaryPhoneNumber ...
func (i *Interactor) UpdatePrimaryPhoneNumber(ctx context.Context, phoneNumber string, useContext bool) error {
	return i.profile.UpdatePrimaryPhoneNumber(ctx, phoneNumber, useContext)
}

// UpdatePrimaryEmailAddress ...
func (i *Interactor) UpdatePrimaryEmailAddress(ctx context.Context, emailAddress string) error {
	return i.profile.UpdatePrimaryEmailAddress(ctx, emailAddress)
}

// UpdateSecondaryPhoneNumbers ...
func (i *Interactor) UpdateSecondaryPhoneNumbers(ctx context.Context, phoneNumbers []string) error {
	return i.profile.UpdateSecondaryPhoneNumbers(ctx, phoneNumbers)
}

// UpdateSecondaryEmailAddresses ...
func (i *Interactor) UpdateSecondaryEmailAddresses(ctx context.Context, emailAddresses []string) error {
	return i.profile.UpdateSecondaryEmailAddresses(ctx, emailAddresses)
}

// UpdateVerifiedIdentifiers ...
func (i *Interactor) UpdateVerifiedIdentifiers(
	ctx context.Context,
	identifiers []profileutils.VerifiedIdentifier,
) error {
	return i.profile.UpdateVerifiedIdentifiers(
		ctx,
		identifiers,
	)
}

// UpdateVerifiedUIDS ...
func (i *Interactor) UpdateVerifiedUIDS(ctx context.Context, uids []string) error {
	return i.profile.UpdateVerifiedUIDS(ctx, uids)
}

// UpdateSuspended ...
func (i *Interactor) UpdateSuspended(ctx context.Context, status bool, phoneNumber string, useContext bool) error {
	return i.profile.UpdateSuspended(ctx, status, phoneNumber, useContext)
}

// UpdatePhotoUploadID ...
func (i *Interactor) UpdatePhotoUploadID(ctx context.Context, uploadID string) error {
	return i.profile.UpdatePhotoUploadID(ctx, uploadID)
}

// UpdateCovers ...
func (i *Interactor) UpdateCovers(ctx context.Context, covers []profileutils.Cover) error {
	return i.profile.UpdateCovers(ctx, covers)
}

// UpdatePushTokens ...
func (i *Interactor) UpdatePushTokens(ctx context.Context, pushToken string, retire bool) error {
	return i.profile.UpdatePushTokens(ctx, pushToken, retire)
}

// UpdatePermissions ...
func (i *Interactor) UpdatePermissions(ctx context.Context, perms []profileutils.PermissionType) error {
	return i.profile.UpdatePermissions(ctx, perms)
}

// AddAdminPermsToUser ...
func (i *Interactor) AddAdminPermsToUser(ctx context.Context, phone string) error {
	return i.profile.AddAdminPermsToUser(ctx, phone)
}

// RemoveAdminPermsToUser ...
func (i *Interactor) RemoveAdminPermsToUser(ctx context.Context, phone string) error {
	return i.profile.RemoveAdminPermsToUser(ctx, phone)
}

// AddRoleToUser ...
func (i *Interactor) AddRoleToUser(ctx context.Context, phone string, role profileutils.RoleType) error {
	return i.profile.AddRoleToUser(ctx, phone, role)
}

// RemoveRoleToUser ...
func (i *Interactor) RemoveRoleToUser(ctx context.Context, phone string) error {
	return i.profile.RemoveRoleToUser(ctx, phone)
}

// UpdateBioData ...
func (i *Interactor) UpdateBioData(ctx context.Context, data profileutils.BioData) error {
	return i.profile.UpdateBioData(ctx, data)
}

// GetUserProfileByUID ...
func (i *Interactor) GetUserProfileByUID(
	ctx context.Context,
	UID string,
) (*profileutils.UserProfile, error) {
	return i.profile.GetUserProfileByUID(
		ctx,
		UID,
	)
}

// GetUserProfileByPhoneOrEmail ...
func (i *Interactor) GetUserProfileByPhoneOrEmail(
	ctx context.Context,
	payload *dto.RetrieveUserProfileInput,
) (*profileutils.UserProfile, error) {
	return i.profile.GetUserProfileByPhoneOrEmail(
		ctx,
		payload,
	)
}

// MaskPhoneNumbers ...
func (i *Interactor) MaskPhoneNumbers(phones []string) []string {
	return i.profile.MaskPhoneNumbers(phones)
}

// SetPrimaryPhoneNumber ...
func (i *Interactor) SetPrimaryPhoneNumber(
	ctx context.Context,
	phoneNumber string,
	otp string,
	useContext bool,
) error {
	return i.profile.SetPrimaryPhoneNumber(
		ctx,
		phoneNumber,
		otp,
		useContext,
	)
}

// SetPrimaryEmailAddress ...
func (i *Interactor) SetPrimaryEmailAddress(
	ctx context.Context,
	emailAddress string,
	otp string,
) error {
	return i.profile.SetPrimaryEmailAddress(
		ctx,
		emailAddress,
		otp,
	)
}

// CheckPhoneExists ...
func (i *Interactor) CheckPhoneExists(ctx context.Context, phone string) (bool, error) {
	return i.profile.CheckPhoneExists(ctx, phone)
}

// CheckEmailExists ...
func (i *Interactor) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	return i.profile.CheckEmailExists(ctx, email)
}

// RetireSecondaryPhoneNumbers ...
func (i *Interactor) RetireSecondaryPhoneNumbers(ctx context.Context, toRemovePhoneNumbers []string) (bool, error) {
	return i.profile.RetireSecondaryPhoneNumbers(ctx, toRemovePhoneNumbers)
}

// RetireSecondaryEmailAddress ...
func (i *Interactor) RetireSecondaryEmailAddress(ctx context.Context, toRemoveEmails []string) (bool, error) {
	return i.profile.RetireSecondaryEmailAddress(ctx, toRemoveEmails)
}

// GetUserProfileAttributes ...
func (i *Interactor) GetUserProfileAttributes(
	ctx context.Context,
	UIDs []string,
	attribute string,
) (map[string][]string, error) {
	return i.profile.GetUserProfileAttributes(
		ctx,
		UIDs,
		attribute,
	)
}

// ConfirmedEmailAddresses ...
func (i *Interactor) ConfirmedEmailAddresses(
	ctx context.Context,
	UIDs []string,
) (map[string][]string, error) {
	return i.profile.ConfirmedEmailAddresses(
		ctx,
		UIDs,
	)
}

// ConfirmedPhoneNumbers ...
func (i *Interactor) ConfirmedPhoneNumbers(
	ctx context.Context,
	UIDs []string,
) (map[string][]string, error) {
	return i.profile.ConfirmedPhoneNumbers(
		ctx,
		UIDs,
	)
}

// ValidFCMTokens ...
func (i *Interactor) ValidFCMTokens(
	ctx context.Context,
	UIDs []string,
) (map[string][]string, error) {
	return i.profile.ValidFCMTokens(
		ctx,
		UIDs,
	)
}

// ProfileAttributes ...
func (i *Interactor) ProfileAttributes(
	ctx context.Context,
	UIDs []string,
	attribute string,
) (map[string][]string, error) {
	return i.profile.ProfileAttributes(
		ctx,
		UIDs,
		attribute,
	)
}

// SetupAsExperimentParticipant ...
func (i *Interactor) SetupAsExperimentParticipant(ctx context.Context, participate *bool) (bool, error) {
	return i.profile.SetupAsExperimentParticipant(ctx, participate)
}

// AddAddress ...
func (i *Interactor) AddAddress(
	ctx context.Context,
	input dto.UserAddressInput,
	addressType enumutils.AddressType,
) (*profileutils.Address, error) {
	return i.profile.AddAddress(
		ctx,
		input,
		addressType,
	)
}

// GetAddresses ...
func (i *Interactor) GetAddresses(ctx context.Context) (*domain.UserAddresses, error) {
	return i.profile.GetAddresses(ctx)
}

// GetUserCommunicationsSettings ...
func (i *Interactor) GetUserCommunicationsSettings(
	ctx context.Context,
) (*profileutils.UserCommunicationsSetting, error) {
	return i.profile.GetUserCommunicationsSettings(ctx)
}

// SetUserCommunicationsSettings ...
func (i *Interactor) SetUserCommunicationsSettings(
	ctx context.Context,
	allowWhatsApp *bool,
	allowTextSms *bool,
	allowPush *bool,
	allowEmail *bool,
) (*profileutils.UserCommunicationsSetting, error) {
	return i.profile.SetUserCommunicationsSettings(
		ctx,
		allowWhatsApp,
		allowTextSms,
		allowPush,
		allowEmail,
	)
}

// GetNavigationActions ...
func (i *Interactor) GetNavigationActions(ctx context.Context) (*dto.GroupedNavigationActions, error) {
	return i.profile.GetNavigationActions(ctx)
}

// SaveFavoriteNavActions ...
func (i *Interactor) SaveFavoriteNavActions(ctx context.Context, title string) (bool, error) {
	return i.profile.SaveFavoriteNavActions(ctx, title)
}

// DeleteFavoriteNavActions ...
func (i *Interactor) DeleteFavoriteNavActions(ctx context.Context, title string) (bool, error) {
	return i.profile.DeleteFavoriteNavActions(ctx, title)
}

// RefreshNavigationActions ...
func (i *Interactor) RefreshNavigationActions(ctx context.Context) (*profileutils.NavigationActions, error) {
	return i.profile.RefreshNavigationActions(ctx)
}

// SwitchUserFlaggedFeatures ...
func (i *Interactor) SwitchUserFlaggedFeatures(ctx context.Context, phoneNumber string) (*dto.OKResp, error) {
	return i.profile.SwitchUserFlaggedFeatures(ctx, phoneNumber)
}

// FindUserByPhone ...
func (i *Interactor) FindUserByPhone(ctx context.Context, phoneNumber string) (*profileutils.UserProfile, error) {
	return i.profile.FindUserByPhone(ctx, phoneNumber)
}

// CreateRole ...
func (i *Interactor) CreateRole(ctx context.Context, input dto.RoleInput) (*dto.RoleOutput, error) {
	return i.roles.CreateRole(ctx, input)
}

// DeleteRole ...
func (i *Interactor) DeleteRole(ctx context.Context, roleID string) (bool, error) {
	return i.roles.DeleteRole(ctx, roleID)
}

// GetAllRoles ...
func (i *Interactor) GetAllRoles(ctx context.Context) ([]*dto.RoleOutput, error) {
	return i.roles.GetAllRoles(ctx)
}

// FindRoleByName ...
func (i *Interactor) FindRoleByName(ctx context.Context, roleName *string) ([]*dto.RoleOutput, error) {
	return i.roles.FindRoleByName(ctx, roleName)
}

// GetAllPermissions ...
func (i *Interactor) GetAllPermissions(ctx context.Context) ([]*profileutils.Permission, error) {
	return i.roles.GetAllPermissions(ctx)
}

// GetRoleByName ...
func (i *Interactor) GetRoleByName(ctx context.Context, name string) (*dto.RoleOutput, error) {
	return i.roles.GetRoleByName(ctx, name)
}

// AddPermissionsToRole ...
func (i *Interactor) AddPermissionsToRole(
	ctx context.Context,
	input dto.RolePermissionInput,
) (*dto.RoleOutput, error) {
	return i.roles.AddPermissionsToRole(ctx, input)
}

// RevokeRolePermission ...
func (i *Interactor) RevokeRolePermission(
	ctx context.Context,
	input dto.RolePermissionInput,
) (*dto.RoleOutput, error) {
	return i.roles.RevokeRolePermission(ctx, input)
}

// UpdateRolePermissions ...
func (i *Interactor) UpdateRolePermissions(
	ctx context.Context,
	input dto.RolePermissionInput,
) (*dto.RoleOutput, error) {
	return i.roles.UpdateRolePermissions(ctx, input)
}

// AssignRole ...
func (i *Interactor) AssignRole(ctx context.Context, userID string, roleID string) (bool, error) {
	return i.roles.AssignRole(ctx, userID, roleID)
}

// AssignMultipleRoles ...
func (i *Interactor) AssignMultipleRoles(ctx context.Context, userID string, roleIDs []string) (bool, error) {
	return i.roles.AssignMultipleRoles(ctx, userID, roleIDs)
}

// RevokeRole ...
func (i *Interactor) RevokeRole(ctx context.Context, userID, roleID, reason string) (bool, error) {
	return i.roles.RevokeRole(ctx, userID, roleID, reason)
}

// ActivateRole ...
func (i *Interactor) ActivateRole(ctx context.Context, roleID string) (*dto.RoleOutput, error) {
	return i.roles.ActivateRole(ctx, roleID)
}

// DeactivateRole ...
func (i *Interactor) DeactivateRole(ctx context.Context, roleID string) (*dto.RoleOutput, error) {
	return i.roles.DeactivateRole(ctx, roleID)
}

// CheckPermission ...
func (i *Interactor) CheckPermission(
	ctx context.Context,
	uid string,
	permission profileutils.Permission,
) (bool, error) {
	return i.roles.CheckPermission(ctx, uid, permission)
}

// CreateUnauthorizedRole ...
func (i *Interactor) CreateUnauthorizedRole(ctx context.Context, input dto.RoleInput) (*dto.RoleOutput, error) {
	return i.roles.CreateUnauthorizedRole(ctx, input)
}

// UnauthorizedDeleteRole ...
func (i *Interactor) UnauthorizedDeleteRole(ctx context.Context, roleID string) (bool, error) {
	return i.roles.UnauthorizedDeleteRole(ctx, roleID)
}

// GetRolesByIDs ...
func (i *Interactor) GetRolesByIDs(ctx context.Context, roleIDs []string) ([]*dto.RoleOutput, error) {
	return i.roles.GetRolesByIDs(ctx, roleIDs)
}

// SetUserPIN ...
func (i *Interactor) SetUserPIN(ctx context.Context, pin string, profileID string) (bool, error) {
	return i.pins.SetUserPIN(ctx, pin, profileID)
}

// SetUserTempPIN ...
func (i *Interactor) SetUserTempPIN(ctx context.Context, profileID string) (string, error) {
	return i.pins.SetUserTempPIN(ctx, profileID)
}

// ResetUserPIN ...
func (i *Interactor) ResetUserPIN(
	ctx context.Context,
	phone string,
	PIN string,
	OTP string,
) (bool, error) {
	return i.pins.ResetUserPIN(ctx, phone, PIN, OTP)
}

// ChangeUserPIN ...
func (i *Interactor) ChangeUserPIN(ctx context.Context, phone string, pin string) (bool, error) {
	return i.pins.ChangeUserPIN(ctx, phone, pin)
}

// RequestPINReset ...
func (i *Interactor) RequestPINReset(ctx context.Context, phone string, appID *string) (*profileutils.OtpResponse, error) {
	return i.pins.RequestPINReset(ctx, phone, appID)
}

// CheckHasPIN ...
func (i *Interactor) CheckHasPIN(ctx context.Context, profileID string) (bool, error) {
	return i.pins.CheckHasPIN(ctx, profileID)
}

// VerifyPhoneNumber ...
func (i *Interactor) VerifyPhoneNumber(ctx context.Context, phone string, appID *string) (*profileutils.OtpResponse, error) {
	return i.signup.VerifyPhoneNumber(ctx, phone, appID)
}

// CreateUserByPhone ...
func (i *Interactor) CreateUserByPhone(ctx context.Context, input *dto.SignUpInput) (*profileutils.UserResponse, error) {
	return i.signup.CreateUserByPhone(ctx, input)
}

// UpdateUserProfile ...
func (i *Interactor) UpdateUserProfile(
	ctx context.Context,
	input *dto.UserProfileInput,
) (*profileutils.UserProfile, error) {
	return i.signup.UpdateUserProfile(ctx, input)
}

// RegisterPushToken ...
func (i *Interactor) RegisterPushToken(ctx context.Context, token string) (bool, error) {
	return i.signup.RegisterPushToken(ctx, token)
}

// CompleteSignup ...
func (i *Interactor) CompleteSignup(ctx context.Context, flavour feedlib.Flavour) (bool, error) {
	return i.signup.CompleteSignup(ctx, flavour)
}

// RetirePushToken ...
func (i *Interactor) RetirePushToken(ctx context.Context, token string) (bool, error) {
	return i.signup.RetirePushToken(ctx, token)
}

// GetUserRecoveryPhoneNumbers ...
func (i *Interactor) GetUserRecoveryPhoneNumbers(
	ctx context.Context,
	phoneNumber string,
) (*dto.AccountRecoveryPhonesResponse, error) {
	return i.signup.GetUserRecoveryPhoneNumbers(ctx, phoneNumber)
}

// SetPhoneAsPrimary ...
func (i *Interactor) SetPhoneAsPrimary(ctx context.Context, phone, otp string) (bool, error) {
	return i.signup.SetPhoneAsPrimary(ctx, phone, otp)
}

// RemoveUserByPhoneNumber ...
func (i *Interactor) RemoveUserByPhoneNumber(ctx context.Context, phone string) error {
	return i.signup.RemoveUserByPhoneNumber(ctx, phone)
}

// RecordPostVisitSurvey ...
func (i *Interactor) RecordPostVisitSurvey(ctx context.Context, input dto.PostVisitSurveyInput) (bool, error) {
	return i.surveys.RecordPostVisitSurvey(ctx, input)
}

// RegisterMicroservice ...
func (i *Interactor) RegisterMicroservice(
	ctx context.Context,
	input domain.Microservice,
) (*domain.Microservice, error) {
	return i.services.RegisterMicroservice(ctx, input)
}

// ListMicroservices ...
func (i *Interactor) ListMicroservices(ctx context.Context) ([]*domain.Microservice, error) {
	return i.services.ListMicroservices(ctx)
}

// DeregisterMicroservice ...
func (i *Interactor) DeregisterMicroservice(ctx context.Context, id string) (bool, error) {
	return i.services.DeregisterMicroservice(ctx, id)
}

// DeregisterAllMicroservices ...
func (i *Interactor) DeregisterAllMicroservices(ctx context.Context) (bool, error) {
	return i.services.DeregisterAllMicroservices(ctx)
}

// FindMicroserviceByID ...
func (i *Interactor) FindMicroserviceByID(ctx context.Context, id string) (*domain.Microservice, error) {
	return i.services.FindMicroserviceByID(ctx, id)
}

// PollMicroservicesStatus ...
func (i *Interactor) PollMicroservicesStatus(ctx context.Context) ([]*domain.MicroserviceStatus, error) {
	return i.services.PollMicroservicesStatus(ctx)
}
