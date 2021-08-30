package infrastructure

import (
	"context"

	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
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
