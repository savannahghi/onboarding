package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"time"

	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/presentation/graph/generated"
	"github.com/savannahghi/profileutils"
	"github.com/savannahghi/serverutils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *mutationResolver) CompleteSignup(ctx context.Context, flavour feedlib.Flavour) (bool, error) {
	startTime := time.Now()

	completeSignup, err := r.usecases.CompleteSignup(ctx, flavour)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "completeSignup", err)

	return completeSignup, err
}

func (r *mutationResolver) UpdateUserProfile(ctx context.Context, input dto.UserProfileInput) (*profileutils.UserProfile, error) {
	startTime := time.Now()

	updateUserProfile, err := r.usecases.UpdateUserProfile(ctx, &input)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "updateUserProfile", err)

	return updateUserProfile, err
}

func (r *mutationResolver) UpdateUserPin(ctx context.Context, phone string, pin string) (bool, error) {
	startTime := time.Now()

	updateUserPIN, err := r.usecases.ChangeUserPIN(ctx, phone, pin)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "updateUserPIN", err)

	return updateUserPIN, err
}

func (r *mutationResolver) SetPrimaryPhoneNumber(ctx context.Context, phone string, otp string) (bool, error) {
	startTime := time.Now()

	err := r.usecases.SetPrimaryPhoneNumber(ctx, phone, otp, true)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "setPrimaryPhoneNumber", err)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *mutationResolver) SetPrimaryEmailAddress(ctx context.Context, email string, otp string) (bool, error) {
	startTime := time.Now()

	err := r.usecases.SetPrimaryEmailAddress(ctx, email, otp)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "setPrimaryEmailAddress", err)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *mutationResolver) AddSecondaryPhoneNumber(ctx context.Context, phone []string) (bool, error) {
	startTime := time.Now()

	err := r.usecases.UpdateSecondaryPhoneNumbers(ctx, phone)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "addSecondaryPhoneNumber", err)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *mutationResolver) RetireSecondaryPhoneNumbers(ctx context.Context, phones []string) (bool, error) {
	startTime := time.Now()

	retireSecondaryPhoneNumbers, err := r.usecases.RetireSecondaryPhoneNumbers(
		ctx,
		phones,
	)

	defer serverutils.RecordGraphqlResolverMetrics(
		ctx,
		startTime,
		"retireSecondaryPhoneNumbers",
		err,
	)

	return retireSecondaryPhoneNumbers, err
}

func (r *mutationResolver) AddSecondaryEmailAddress(ctx context.Context, email []string) (bool, error) {
	startTime := time.Now()

	err := r.usecases.UpdateSecondaryEmailAddresses(ctx, email)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "addSecondaryEmailAddress", err)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *mutationResolver) RetireSecondaryEmailAddresses(ctx context.Context, emails []string) (bool, error) {
	startTime := time.Now()

	retireSecondaryEmailAddresses, err := r.usecases.RetireSecondaryEmailAddress(
		ctx,
		emails,
	)

	defer serverutils.RecordGraphqlResolverMetrics(
		ctx,
		startTime,
		"retireSecondaryEmailAddresses",
		err,
	)

	return retireSecondaryEmailAddresses, err
}

func (r *mutationResolver) UpdateUserName(ctx context.Context, username string) (bool, error) {
	startTime := time.Now()

	err := r.usecases.UpdateUserName(ctx, username)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "updateUserName", err)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *mutationResolver) RegisterPushToken(ctx context.Context, token string) (bool, error) {
	startTime := time.Now()

	registerPushToken, err := r.usecases.RegisterPushToken(ctx, token)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "registerPushToken", err)

	return registerPushToken, err
}

func (r *mutationResolver) RecordPostVisitSurvey(ctx context.Context, input dto.PostVisitSurveyInput) (bool, error) {
	startTime := time.Now()

	recordPostVisitSurvey, err := r.usecases.RecordPostVisitSurvey(ctx, input)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "recordPostVisitSurvey", err)

	return recordPostVisitSurvey, err
}

func (r *mutationResolver) SetupAsExperimentParticipant(ctx context.Context, participate *bool) (bool, error) {
	startTime := time.Now()

	setupAsExperimentParticipant, err := r.usecases.SetupAsExperimentParticipant(
		ctx,
		participate,
	)

	defer serverutils.RecordGraphqlResolverMetrics(
		ctx,
		startTime,
		"setupAsExperimentParticipant",
		err,
	)

	return setupAsExperimentParticipant, err
}

func (r *mutationResolver) AddAddress(ctx context.Context, input dto.UserAddressInput, addressType enumutils.AddressType) (*profileutils.Address, error) {
	startTime := time.Now()

	addAddress, err := r.usecases.AddAddress(
		ctx,
		input,
		addressType,
	)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "addAddress", err)

	return addAddress, err
}

func (r *mutationResolver) SetUserCommunicationsSettings(ctx context.Context, allowWhatsApp *bool, allowTextSms *bool, allowPush *bool, allowEmail *bool) (*profileutils.UserCommunicationsSetting, error) {
	startTime := time.Now()

	setUserCommunicationsSettings, err := r.usecases.SetUserCommunicationsSettings(
		ctx,
		allowWhatsApp,
		allowTextSms,
		allowPush,
		allowEmail,
	)

	defer serverutils.RecordGraphqlResolverMetrics(
		ctx,
		startTime,
		"setUserCommunicationsSettings",
		err,
	)

	return setUserCommunicationsSettings, err
}

func (r *mutationResolver) SaveFavoriteNavAction(ctx context.Context, title string) (bool, error) {
	startTime := time.Now()

	success, err := r.usecases.SaveFavoriteNavActions(ctx, title)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "saveFavoriteNavAction", err)

	return success, err
}

func (r *mutationResolver) DeleteFavoriteNavAction(ctx context.Context, title string) (bool, error) {
	startTime := time.Now()

	success, err := r.usecases.DeleteFavoriteNavActions(ctx, title)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "deleteFavoriteNavAction", err)

	return success, err
}

func (r *mutationResolver) RegisterMicroservice(ctx context.Context, input domain.Microservice) (*domain.Microservice, error) {
	startTime := time.Now()

	service, err := r.usecases.RegisterMicroservice(ctx, input)
	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "registerMicroservice", err)

	return service, err
}

func (r *mutationResolver) DeregisterMicroservice(ctx context.Context, id string) (bool, error) {
	startTime := time.Now()

	status, err := r.usecases.DeregisterMicroservice(ctx, id)
	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "deregisterMicroservice", err)

	return status, err
}

func (r *mutationResolver) DeregisterAllMicroservices(ctx context.Context) (bool, error) {
	startTime := time.Now()

	status, err := r.usecases.DeregisterAllMicroservices(ctx)
	defer serverutils.RecordGraphqlResolverMetrics(
		ctx,
		startTime,
		"deregisterAllMicroservices",
		err,
	)

	return status, err
}

func (r *mutationResolver) CreateRole(ctx context.Context, input dto.RoleInput) (*dto.RoleOutput, error) {
	startTime := time.Now()

	role, err := r.usecases.CreateRole(ctx, input)
	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "createRole", err)

	return role, err
}

func (r *mutationResolver) DeleteRole(ctx context.Context, roleID string) (bool, error) {
	startTime := time.Now()

	success, err := r.usecases.DeleteRole(ctx, roleID)
	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "deleteRole", err)

	return success, err
}

func (r *mutationResolver) AddPermissionsToRole(ctx context.Context, input dto.RolePermissionInput) (*dto.RoleOutput, error) {
	startTime := time.Now()

	role, err := r.usecases.AddPermissionsToRole(ctx, input)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "addPermissionsToRole", err)

	return role, err
}

func (r *mutationResolver) RevokeRolePermission(ctx context.Context, input dto.RolePermissionInput) (*dto.RoleOutput, error) {
	startTime := time.Now()

	role, err := r.usecases.RevokeRolePermission(ctx, input)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "revokeRolePermission", err)

	return role, err
}

func (r *mutationResolver) UpdateRolePermissions(ctx context.Context, input dto.RolePermissionInput) (*dto.RoleOutput, error) {
	startTime := time.Now()

	role, err := r.usecases.UpdateRolePermissions(ctx, input)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "updateRolePermissions", err)

	return role, err
}

func (r *mutationResolver) AssignRole(ctx context.Context, userID string, roleID string) (bool, error) {
	startTime := time.Now()

	status, err := r.usecases.AssignRole(ctx, userID, roleID)
	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "assignRole", err)

	return status, err
}

func (r *mutationResolver) AssignMultipleRoles(ctx context.Context, userID string, roleIDs []string) (bool, error) {
	startTime := time.Now()

	status, err := r.usecases.AssignMultipleRoles(ctx, userID, roleIDs)
	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "assignMultipleRoles", err)

	return status, err
}

func (r *mutationResolver) RevokeRole(ctx context.Context, userID string, roleID string, reason string) (bool, error) {
	startTime := time.Now()

	status, err := r.usecases.RevokeRole(ctx, userID, roleID, reason)
	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "revokeRole", err)

	return status, err
}

func (r *mutationResolver) ActivateRole(ctx context.Context, roleID string) (*dto.RoleOutput, error) {
	startTime := time.Now()

	role, err := r.usecases.ActivateRole(ctx, roleID)
	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "activateRole", err)

	return role, err
}

func (r *mutationResolver) DeactivateRole(ctx context.Context, roleID string) (*dto.RoleOutput, error) {
	startTime := time.Now()

	role, err := r.usecases.DeactivateRole(ctx, roleID)
	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "deactivateRole", err)

	return role, err
}

func (r *queryResolver) DummyQuery(ctx context.Context) (*bool, error) {
	dummy := true
	return &dummy, nil
}

func (r *queryResolver) UserProfile(ctx context.Context) (*profileutils.UserProfile, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("resolver.name", "userProfile"),
	)

	startTime := time.Now()

	userProfile, err := r.usecases.UserProfile(ctx)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "userProfile", err)

	return userProfile, err
}

func (r *queryResolver) ResumeWithPin(ctx context.Context, pin string) (bool, error) {
	startTime := time.Now()

	resumeWithPin, err := r.usecases.ResumeWithPin(ctx, pin)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "resumeWithPin", err)

	return resumeWithPin, err
}

func (r *queryResolver) GetAddresses(ctx context.Context) (*domain.UserAddresses, error) {
	startTime := time.Now()

	addresses, err := r.usecases.GetAddresses(ctx)

	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "getAddresses", err)

	return addresses, err
}

func (r *queryResolver) GetUserCommunicationsSettings(ctx context.Context) (*profileutils.UserCommunicationsSetting, error) {
	startTime := time.Now()

	userCommunicationsSettings, err := r.usecases.GetUserCommunicationsSettings(ctx)

	defer serverutils.RecordGraphqlResolverMetrics(
		ctx,
		startTime,
		"getUserCommunicationsSettings",
		err,
	)

	return userCommunicationsSettings, err
}

func (r *queryResolver) FetchUserNavigationActions(ctx context.Context) (*profileutils.NavigationActions, error) {
	startTime := time.Now()

	navactions, err := r.usecases.RefreshNavigationActions(ctx)

	defer serverutils.RecordGraphqlResolverMetrics(
		ctx,
		startTime,
		"fetchUserNavigationActions",
		err,
	)

	return navactions, err
}

func (r *queryResolver) ListMicroservices(ctx context.Context) ([]*domain.Microservice, error) {
	startTime := time.Now()

	services, err := r.usecases.ListMicroservices(ctx)
	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "listMicroservices", err)

	return services, err
}

func (r *queryResolver) GetAllRoles(ctx context.Context) ([]*dto.RoleOutput, error) {
	startTime := time.Now()

	roles, err := r.usecases.GetAllRoles(ctx)
	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "getAllRoles", err)

	return roles, err
}

func (r *queryResolver) FindRoleByName(ctx context.Context, roleName *string) ([]*dto.RoleOutput, error) {
	startTime := time.Now()

	roles, err := r.usecases.FindRoleByName(ctx, roleName)
	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "findRoleByName", err)

	return roles, err
}

func (r *queryResolver) GetAllPermissions(ctx context.Context) ([]*profileutils.Permission, error) {
	startTime := time.Now()

	permissions, err := r.usecases.GetAllPermissions(ctx)
	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "getAllPermissions", err)

	return permissions, err
}

func (r *queryResolver) FindUserByPhone(ctx context.Context, phoneNumber string) (*profileutils.UserProfile, error) {
	startTime := time.Now()

	profile, err := r.usecases.FindUserByPhone(ctx, phoneNumber)
	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "findUserByPhone", err)

	return profile, err
}

func (r *queryResolver) GetNavigationActions(ctx context.Context) (*dto.GroupedNavigationActions, error) {
	startTime := time.Now()

	navActions, err := r.usecases.GetNavigationActions(ctx)
	defer serverutils.RecordGraphqlResolverMetrics(ctx, startTime, "getNavigationActions", err)

	return navActions, err
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
