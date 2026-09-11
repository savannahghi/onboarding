package usecases

import (
	"context"
	"fmt"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/exceptions"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/utils"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure"
	"github.com/savannahghi/profileutils"
)

// PermissionUseCase represent the business logic required for management of permissions
type PermissionUseCase interface {
	CreatePermission(
		ctx context.Context,
		input dto.PermissionInput,
	) (*domain.RolePermission, error)
	GetPermissions(ctx context.Context) (*[]domain.RolePermission, error)
	DeletePermission(
		ctx context.Context,
		permissionScope string,
	) (bool, error)
}

// PermissionUseCase  represents usecase implementation object
type PermissionUseCaseImpl struct {
	infrastructure infrastructure.Infrastructure
	baseExt        extension.BaseExtension
}

// NewPermissionUseCases returns a new onboarding usecase
func NewPermissionUseCases(
	infrastructure infrastructure.Infrastructure,
	ext extension.BaseExtension,
) PermissionUseCase {
	return &PermissionUseCaseImpl{
		infrastructure: infrastructure,
		baseExt:        ext,
	}
}

// CreatePermission creates a new Role
func (r *PermissionUseCaseImpl) CreatePermission(
	ctx context.Context,
	input dto.PermissionInput,
) (*domain.RolePermission, error) {
	ctx, span := tracer.Start(ctx, "CreatePermission")
	defer span.End()

	user, err := r.baseExt.GetLoggedInUser(ctx)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	// Check logged in user has the right permissions
	allowed, err := r.infrastructure.Database.CheckIfUserHasPermission(ctx, user.UID, profileutils.CanCreateRole)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}
	if !allowed {
		return nil, exceptions.RoleNotValid(
			fmt.Errorf("error: logged in user does not have permissions to create role"),
		)
	}

	userProfile, err := r.infrastructure.Database.GetUserProfileByUID(ctx, user.UID, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	permission, err := r.infrastructure.Database.CreatePermission(ctx, userProfile.ID, input)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	return permission, nil
}

// GetPermissions returns a list of all created permissions
func (r *PermissionUseCaseImpl) GetPermissions(ctx context.Context) (*[]domain.RolePermission, error) {
	ctx, span := tracer.Start(ctx, "GetPermissions")
	defer span.End()

	user, err := r.baseExt.GetLoggedInUser(ctx)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	// Check logged in user has the right permissions
	allowed, err := r.infrastructure.Database.CheckIfUserHasPermission(ctx, user.UID, profileutils.CanViewRole)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}
	if !allowed {
		return nil, exceptions.RoleNotValid(
			fmt.Errorf("error: logged in user does not have permissions to list roles"),
		)
	}

	permissions, err := r.infrastructure.Database.GetAllPermissions(ctx)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	return permissions, nil
}

// DeletePermission removes a permission permanently from the database
func (r *PermissionUseCaseImpl) DeletePermission(
	ctx context.Context,
	permissionScope string,
) (bool, error) {
	ctx, span := tracer.Start(ctx, "DeletePermission")
	defer span.End()

	user, err := r.baseExt.GetLoggedInUser(ctx)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, err
	}

	// Check logged in user has the right permissions
	allowed, err := r.infrastructure.Database.CheckIfUserHasPermission(ctx, user.UID, profileutils.CanEditRole)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, err
	}
	if !allowed {
		return false, exceptions.RoleNotValid(
			fmt.Errorf("logged in user does not have permission to delete a permission"),
		)
	}

	return r.infrastructure.Database.DeletePermission(ctx, permissionScope, user.UID)
}
