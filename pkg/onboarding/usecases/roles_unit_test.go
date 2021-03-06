package usecases_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/profileutils"
)

func TestRoleUseCaseImpl_CreateRole(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()

	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	input := dto.RoleInput{
		Name: "Agents",
	}

	allPerms, err := profileutils.AllPermissions(ctx)
	if err != nil {
		t.Error("error did not get all permissions")
		return
	}

	perms := []profileutils.Permission{}
	for _, perm := range allPerms {
		if perm.Scope == "role.edit" {
			perm.Allowed = true
		}
		perms = append(perms, perm)
	}
	expectedOutput := &dto.RoleOutput{
		Scopes:      []string{"role.edit"},
		Permissions: perms,
	}

	type args struct {
		ctx   context.Context
		input dto.RoleInput
	}

	tests := []struct {
		name    string
		args    args
		want    *dto.RoleOutput
		wantErr bool
	}{
		{
			name: "sad: unable to get logged in user",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad: unable to check if user has permissions",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad: user do not have required permissions",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad: unable to get user's profile",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad: unable to create role in database",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "happy:created role",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    expectedOutput,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "sad: unable to get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("unable to get logged in user")
				}
			}

			if tt.name == "sad: unable to check if user has permissions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, fmt.Errorf(
						"error unable to check if user has required permissions",
					)
				}
			}

			if tt.name == "sad: user do not have required permissions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, nil
				}
			}

			if tt.name == "sad: unable to get user's profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("error unable to get user profile")
				}
			}

			if tt.name == "sad: unable to create role in database" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}
				fakeInfraRepo.CreateRoleFn = func(ctx context.Context, profileID string, role dto.RoleInput) (*profileutils.Role, error) {
					return nil, fmt.Errorf("error un able to create role in db")
				}
			}

			if tt.name == "happy:created role" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}
				fakeInfraRepo.CreateRoleFn = func(ctx context.Context, profileID string, role dto.RoleInput) (*profileutils.Role, error) {
					return &profileutils.Role{
						Scopes: []string{"role.edit"},
					}, nil
				}
			}

			got, err := i.CreateRole(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleUseCaseImpl.CreateRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RoleUseCaseImpl.CreateRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleUseCaseImpl_GetAllRoles(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()

	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}
	allPerms, err := profileutils.AllPermissions(ctx)
	if err != nil {
		t.Errorf("failed to get all permissions")
		return
	}
	rolePerms := []profileutils.Permission{}
	for _, perm := range allPerms {
		if perm.Scope == "role.create" {
			perm.Allowed = true
		}
		rolePerms = append(rolePerms, perm)
	}

	expectedOutput := []*dto.RoleOutput{
		{
			ID:          "c9d62c7e-93e5-44a6-b503-6fc159c1782f",
			Scopes:      []string{"role.create"},
			Permissions: rolePerms,
			Users:       []*profileutils.UserProfile{},
		},
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    []*dto.RoleOutput
		wantErr bool
	}{
		{
			name:    "sad: did not get logged in user",
			args:    args{ctx: ctx},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "sad: unable to check if user has permission",
			args:    args{ctx: ctx},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "sad: user do not have required permission",
			args:    args{ctx: ctx},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "sad: did not get roles from database",
			args:    args{ctx: ctx},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "happy: got roles",
			args:    args{ctx: ctx},
			want:    expectedOutput,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "sad: did not get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("error, did not get logged in user")
				}
			}

			if tt.name == "sad: unable to check if user has permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, fmt.Errorf("error unable to check is user has permission")
				}
			}

			if tt.name == "sad: user do not have required permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, nil
				}
			}
			if tt.name == "sad: did not get roles from database" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetAllRolesFn = func(ctx context.Context) (*[]profileutils.Role, error) {
					return nil, fmt.Errorf("error, did not get roles from database")
				}
			}

			if tt.name == "happy: got roles" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetAllRolesFn = func(ctx context.Context) (*[]profileutils.Role, error) {
					return &[]profileutils.Role{
						{
							ID:     "c9d62c7e-93e5-44a6-b503-6fc159c1782f",
							Scopes: []string{"role.create"},
						},
					}, nil
				}
				fakeInfraRepo.GetUserProfilesByRoleIDFn = func(ctx context.Context, role string) ([]*profileutils.UserProfile, error) {

					return []*profileutils.UserProfile{}, nil
				}
			}
			got, err := i.GetAllRoles(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleUseCaseImpl.GetAllRoles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RoleUseCaseImpl.GetAllRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleUseCaseImpl_FindRoleByName(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()

	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	roleName := "Employee Role"

	allPerms, err := profileutils.AllPermissions(ctx)
	if err != nil {
		t.Errorf("failed to get all permissions")
		return
	}
	rolePerms := []profileutils.Permission{}
	for _, perm := range allPerms {
		if perm.Scope == "role.create" {
			perm.Allowed = true
		}
		rolePerms = append(rolePerms, perm)
	}

	expectedOutput := []*dto.RoleOutput{
		{
			Name:        roleName,
			ID:          "c9d62c7e-93e5-44a6-b503-6fc159c1782f",
			Scopes:      []string{"role.create"},
			Permissions: rolePerms,
		},
	}

	type args struct {
		ctx      context.Context
		roleName *string
	}
	tests := []struct {
		name    string
		args    args
		want    []*dto.RoleOutput
		wantErr bool
	}{
		{
			name:    "sad: did not get logged in user",
			args:    args{ctx: ctx, roleName: &roleName},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "sad: unable to check if user has permission",
			args:    args{ctx: ctx, roleName: &roleName},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "sad: user do not have required permission",
			args:    args{ctx: ctx, roleName: &roleName},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "sad: did not get roles from database",
			args:    args{ctx: ctx, roleName: &roleName},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "happy: got role",
			args:    args{ctx: ctx, roleName: &roleName},
			want:    expectedOutput,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "sad: did not get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("error, did not get logged in user")
				}
			}

			if tt.name == "sad: unable to check if user has permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, fmt.Errorf("error unable to check is user has permission")
				}
			}

			if tt.name == "sad: user do not have required permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, nil
				}
			}
			if tt.name == "sad: did not get roles from database" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetAllRolesFn = func(ctx context.Context) (*[]profileutils.Role, error) {
					return nil, fmt.Errorf("error, did not get roles from database")
				}
			}

			if tt.name == "happy: got role" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetAllRolesFn = func(ctx context.Context) (*[]profileutils.Role, error) {
					return &[]profileutils.Role{{
						Name:   roleName,
						ID:     "c9d62c7e-93e5-44a6-b503-6fc159c1782f",
						Scopes: []string{"role.create"},
					}}, nil
				}
				fakeInfraRepo.GetUserProfilesByRoleIDFn = func(ctx context.Context, role string) ([]*profileutils.UserProfile, error) {
					return []*profileutils.UserProfile{}, nil
				}
			}
			got, err := i.FindRoleByName(tt.args.ctx, tt.args.roleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleUseCaseImpl.FindRoleByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RoleUseCaseImpl.FindRoleByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleUseCaseImpl_DeleteRole(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()

	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	type args struct {
		ctx    context.Context
		roleID string
	}

	input := args{
		ctx:    ctx,
		roleID: "123",
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "sad: did not get logged in user",
			args:    input,
			want:    false,
			wantErr: true,
		},
		{
			name:    "sad: unable to check if user has permission",
			args:    input,
			want:    false,
			wantErr: true,
		},
		{
			name:    "sad: user do not have required permission",
			args:    input,
			want:    false,
			wantErr: true,
		},
		{
			name:    "sad: unable to delete role",
			args:    input,
			want:    false,
			wantErr: true,
		},
		{
			name:    "happy: deleted role",
			args:    input,
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "sad: did not get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("error, did not get logged in user")
				}
			}

			if tt.name == "sad: unable to check if user has permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, fmt.Errorf("error unable to check is user has permission")
				}
			}

			if tt.name == "sad: user do not have required permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, nil
				}
			}
			if tt.name == "sad: unable to delete role" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.DeleteRoleFn = func(ctx context.Context, roleID string) (bool, error) {
					return false, fmt.Errorf("error, unable to delete roles")
				}
			}

			if tt.name == "happy: deleted role" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.DeleteRoleFn = func(ctx context.Context, roleID string) (bool, error) {
					return true, nil
				}
			}
			got, err := i.DeleteRole(tt.args.ctx, tt.args.roleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleUseCaseImpl.DeleteRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RoleUseCaseImpl.DeleteRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleUseCaseImpl_UnauthorizedDeleteRole(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	type args struct {
		ctx    context.Context
		roleID string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "happy: remove test role",
			args: args{
				ctx:    ctx,
				roleID: uuid.NewString(),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "sad: remove non-test role",
			args: args{
				ctx:    ctx,
				roleID: uuid.NewString(),
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "sad: remove role error",
			args: args{
				ctx:    ctx,
				roleID: uuid.NewString(),
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		if tt.name == "happy: remove test role" {
			fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
				return &profileutils.Role{ID: uuid.NewString(), Name: "Happy Test Role"}, nil
			}
			fakeInfraRepo.DeleteRoleFn = func(ctx context.Context, roleID string) (bool, error) {
				return true, nil
			}
		}

		if tt.name == "sad: remove non-test role" {
			fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
				return &profileutils.Role{ID: uuid.NewString(), Name: "Crucial Role"}, nil
			}
			fakeInfraRepo.DeleteRoleFn = func(ctx context.Context, roleID string) (bool, error) {
				return true, nil
			}
		}

		if tt.name == "sad: remove role error" {
			fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
				return &profileutils.Role{ID: uuid.NewString(), Name: "Happy Test Role"}, nil
			}
			fakeInfraRepo.DeleteRoleFn = func(ctx context.Context, roleID string) (bool, error) {
				return true, fmt.Errorf("cannot remove role")
			}
		}

		t.Run(tt.name, func(t *testing.T) {
			got, err := i.UnauthorizedDeleteRole(tt.args.ctx, tt.args.roleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleUseCaseImpl.UnauthorizedDeleteRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RoleUseCaseImpl.UnauthorizedDeleteRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleUseCaseImpl_GetAllPermissions(t *testing.T) {

	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()

	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	allPerms, err := profileutils.AllPermissions(ctx)
	if err != nil {
		t.Error("error did not get all permissions")
		return
	}

	output := []*profileutils.Permission{}
	for _, perm := range allPerms {
		p := &profileutils.Permission{
			Scope:       perm.Scope,
			Group:       perm.Group,
			Description: perm.Description,
		}
		output = append(output, p)
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    []*profileutils.Permission
		wantErr bool
	}{
		{
			name:    "happy got all permissions",
			args:    args{ctx: ctx},
			want:    output,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := i.GetAllPermissions(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"RoleUseCaseImpl.GetAllPermissions() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RoleUseCaseImpl.GetAllPermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleUseCaseImpl_AddPermissionsToRole(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()

	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	input := dto.RolePermissionInput{
		RoleID: "123",
		Scopes: []string{"role.create"},
	}

	allPerms, err := profileutils.AllPermissions(ctx)
	if err != nil {
		t.Error("error did not get all permissions")
		return
	}

	perms := []profileutils.Permission{}
	for _, perm := range allPerms {
		if perm.Scope == "role.create" {
			perm.Allowed = true
		}
		perms = append(perms, perm)
	}

	expectedOutput := dto.RoleOutput{
		ID:          "123",
		Scopes:      []string{"role.create"},
		Permissions: perms,
	}

	type args struct {
		ctx   context.Context
		input dto.RolePermissionInput
	}

	tests := []struct {
		name    string
		args    args
		want    *dto.RoleOutput
		wantErr bool
	}{
		{
			name: "sad unable to get logged in user",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad unable to check if user has permissions",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad user do not have required permission",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad unable to get role by id",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad unable to get user profile",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad unable to update role details",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "happy added role permissions",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    &expectedOutput,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "sad unable to get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("unable to get logged in user")
				}
			}

			if tt.name == "sad unable to check if user has permissions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, fmt.Errorf("unable to check permissions")
				}
			}

			if tt.name == "sad user do not have required permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, nil
				}
			}

			if tt.name == "sad unable to get role by id" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return nil, fmt.Errorf("error unable to get role to edit")
				}
			}

			if tt.name == "sad unable to get user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get user profile")
				}
			}

			if tt.name == "sad unable to update role details" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}
				fakeInfraRepo.UpdateRoleDetailsFn = func(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error) {
					return nil, fmt.Errorf("error unable to update role")
				}
			}

			if tt.name == "happy added role permissions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}
				fakeInfraRepo.UpdateRoleDetailsFn = func(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "123",
						Scopes: []string{"role.create"},
					}, nil
				}
			}

			got, err := i.AddPermissionsToRole(tt.args.ctx, tt.args.input)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"RoleUseCaseImpl.AddPermissionsToRole() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RoleUseCaseImpl.AddPermissionsToRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleUseCaseImpl_RevokeRolePermissions(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()

	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	allPerms, err := profileutils.AllPermissions(ctx)
	if err != nil {
		t.Error("error did not get all permissions")
		return
	}

	perms := []profileutils.Permission{}
	for _, perm := range allPerms {
		if perm.Scope == "role.create" {
			perm.Allowed = true
		}
		perms = append(perms, perm)
	}

	expectedOutput := dto.RoleOutput{
		ID:          "123",
		Scopes:      []string{"role.create"},
		Permissions: perms,
	}

	type args struct {
		ctx       context.Context
		inputData dto.RolePermissionInput
	}

	input := args{
		ctx: ctx,
		inputData: dto.RolePermissionInput{
			RoleID: "123",
			Scopes: []string{"role.create"},
		},
	}

	tests := []struct {
		name    string
		args    args
		want    *dto.RoleOutput
		wantErr bool
	}{
		{
			name:    "sad unable to get logged in user",
			args:    input,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "sad unable to check if user has permissions",
			args:    input,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "sad user do not have required permission",
			args:    input,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "sad unable to get role by id",
			args:    input,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "sad unable to get user profile",
			args:    input,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "sad unable to update role details",
			args:    input,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "happy revoked role permissions",
			args:    input,
			want:    &expectedOutput,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "sad unable to get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("unable to get logged in user")
				}
			}

			if tt.name == "sad unable to check if user has permissions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, fmt.Errorf("unable to check permissions")
				}
			}

			if tt.name == "sad user do not have required permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, nil
				}
			}

			if tt.name == "sad unable to get role by id" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return nil, fmt.Errorf("error unable to get role to edit")
				}
			}

			if tt.name == "sad unable to get user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get user profile")
				}
			}

			if tt.name == "sad unable to update role details" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}
				fakeInfraRepo.UpdateRoleDetailsFn = func(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error) {
					return nil, fmt.Errorf("error unable to update role")
				}
			}

			if tt.name == "happy revoked role permissions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}
				fakeInfraRepo.UpdateRoleDetailsFn = func(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "123",
						Scopes: []string{"role.create"},
					}, nil
				}
			}

			got, err := i.RevokeRolePermission(tt.args.ctx, tt.args.inputData)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"RoleUseCaseImpl.AddPermissionsToRole() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RoleUseCaseImpl.AddPermissionsToRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleUseCaseImpl_AssignRole(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()

	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	type args struct {
		ctx    context.Context
		userID string
		roleID string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "fail: cannot get logged in user",
			args: args{
				ctx:    ctx,
				userID: uuid.NewString(),
				roleID: uuid.NewString(),
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "fail: user doesn't have the permission",
			args: args{
				ctx:    ctx,
				userID: uuid.NewString(),
				roleID: uuid.NewString(),
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "fail: role ID doesn't exist",
			args: args{
				ctx:    ctx,
				userID: uuid.NewString(),
				roleID: "invalid id",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "fail: cannot retrieve user profile",
			args: args{
				ctx:    ctx,
				userID: uuid.NewString(),
				roleID: uuid.NewString(),
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "fail: role already exists",
			args: args{
				ctx:    ctx,
				userID: uuid.NewString(),
				roleID: "0637333d-74b0-473d-95bd-0a03b1ae5e06",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "fail: error updating user profile role",
			args: args{
				ctx:    ctx,
				userID: uuid.NewString(),
				roleID: "17e6ea18-7147-4bdb-ad0b-d9ce03a8c0ac",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "success: add a new role to user",
			args: args{
				ctx:    ctx,
				userID: uuid.NewString(),
				roleID: "17e6ea18-7147-4bdb-ad0b-d9ce03a8c0ac",
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "fail: cannot get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("cannot get logged in user")
				}

				//remove
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return nil, fmt.Errorf("cannot get role ny id")
				}
			}

			if tt.name == "fail: user doesn't have the permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "",
						Scopes: []string{profileutils.CanRegisterAgent.Scope},
					}, nil
				}

				//remove
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return nil, fmt.Errorf("cannot get role ny id")
				}
			}

			if tt.name == "fail: role ID doesn't exist" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return nil, fmt.Errorf("cannot get role ny id")
				}
			}

			if tt.name == "fail: cannot retrieve user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "",
						Scopes: []string{profileutils.CanAssignRole.Scope},
					}, nil
				}

				fakeInfraRepo.GetUserProfileByIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("no user profile")
				}
			}

			if tt.name == "fail: role already exists" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "0637333d-74b0-473d-95bd-0a03b1ae5e06",
						Scopes: []string{profileutils.CanAssignRole.Scope},
					}, nil
				}

				fakeInfraRepo.GetUserProfileByIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:    "",
						Roles: []string{"0637333d-74b0-473d-95bd-0a03b1ae5e06"},
					}, nil
				}
			}

			if tt.name == "fail: error updating user profile role" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "",
						Scopes: []string{profileutils.CanAssignRole.Scope},
					}, nil
				}

				fakeInfraRepo.GetUserProfileByIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: ""}, nil
				}

				fakeInfraRepo.UpdateUserRoleIDsFn = func(ctx context.Context, id string, roleIDs []string) error {
					return fmt.Errorf("cannot update role ids")
				}
			}

			if tt.name == "success: add a new role to user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "",
						Scopes: []string{profileutils.CanAssignRole.Scope},
					}, nil
				}

				fakeInfraRepo.GetUserProfileByIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: ""}, nil
				}

				fakeInfraRepo.UpdateUserRoleIDsFn = func(ctx context.Context, id string, roleIDs []string) error {
					return nil
				}
			}

			got, err := i.AssignRole(tt.args.ctx, tt.args.userID, tt.args.roleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleUseCaseImpl.AssignRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RoleUseCaseImpl.AssignRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleUseCaseImpl_RevokeRole(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()

	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	type args struct {
		ctx    context.Context
		userID string
		roleID string
		reason string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "fail: cannot get logged in user",
			args: args{
				ctx:    ctx,
				userID: uuid.NewString(),
				roleID: uuid.NewString(),
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "fail: user doesn't have the permission",
			args: args{
				ctx:    ctx,
				userID: uuid.NewString(),
				roleID: uuid.NewString(),
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "fail: role ID doesn't exist",
			args: args{
				ctx:    ctx,
				userID: uuid.NewString(),
				roleID: "invalid id",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "fail: cannot retrieve user profile",
			args: args{
				ctx:    ctx,
				userID: uuid.NewString(),
				roleID: uuid.NewString(),
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "fail: user does not have the role",
			args: args{
				ctx:    ctx,
				userID: uuid.NewString(),
				roleID: "missing",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "fail: error updating user profile roles",
			args: args{
				ctx:    ctx,
				userID: uuid.NewString(),
				roleID: "17e6ea18-7147-4bdb-ad0b-d9ce03a8c0ac",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "success: remove a role from a user",
			args: args{
				ctx:    ctx,
				userID: uuid.NewString(),
				roleID: "17e6ea18-7147-4bdb-ad0b-d9ce03a8c0ac",
				reason: "no longer working for us",
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "fail: cannot get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("cannot get logged in user")
				}
			}

			if tt.name == "fail: user doesn't have the permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "",
						Scopes: []string{profileutils.CanRegisterAgent.Scope},
					}, nil
				}
			}

			if tt.name == "fail: role ID doesn't exist" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return nil, fmt.Errorf("cannot get role ny id")
				}
			}

			if tt.name == "fail: cannot retrieve user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "",
						Scopes: []string{profileutils.CanAssignRole.Scope},
					}, nil
				}

				fakeInfraRepo.GetUserProfileByIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("no user profile")
				}
			}

			if tt.name == "fail: user does not have the role" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "",
						Scopes: []string{"duplicate"},
					}, nil
				}

				fakeInfraRepo.GetUserProfileByIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: "", Roles: []string{"duplicate"}}, nil
				}
			}

			if tt.name == "fail: error updating user profile roles" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "17e6ea18-7147-4bdb-ad0b-d9ce03a8c0ac",
						Scopes: []string{profileutils.CanAssignRole.Scope},
					}, nil
				}

				fakeInfraRepo.GetUserProfileByIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "",
						Roles: []string{
							"17e6ea18-7147-4bdb-ad0b-d9ce03a8c0ac",
							"56e5e987-2f02-4455-9dde-ae15162d8bce",
						},
					}, nil
				}

				fakeInfraRepo.UpdateUserRoleIDsFn = func(ctx context.Context, id string, roleIDs []string) error {
					return fmt.Errorf("cannot update user profile roles")
				}
			}

			if tt.name == "success: remove a role from a user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "17e6ea18-7147-4bdb-ad0b-d9ce03a8c0ac",
						Scopes: []string{profileutils.CanAssignRole.Scope},
					}, nil
				}

				fakeInfraRepo.GetUserProfileByIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "",
						Roles: []string{
							"17e6ea18-7147-4bdb-ad0b-d9ce03a8c0ac",
							"56e5e987-2f02-4455-9dde-ae15162d8bce",
						},
					}, nil
				}

				fakeInfraRepo.UpdateUserRoleIDsFn = func(ctx context.Context, id string, roleIDs []string) error {
					return nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID: "",
						Roles: []string{
							"17e6ea18-7147-4bdb-ad0b-d9ce03a8c0ac",
							"56e5e987-2f02-4455-9dde-ae15162d8bce",
						},
					}, nil
				}

				fakeInfraRepo.SaveRoleRevocationFn = func(ctx context.Context, userID string, revocation dto.RoleRevocationInput) error {
					return nil
				}
			}

			got, err := i.RevokeRole(tt.args.ctx, tt.args.userID, tt.args.roleID, tt.args.reason)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleUseCaseImpl.RevokeRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RoleUseCaseImpl.RevokeRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleUseCaseImpl_DeactivateRole(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()

	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	type args struct {
		ctx    context.Context
		roleID string
	}
	tests := []struct {
		name    string
		args    args
		want    *dto.RoleOutput
		wantErr bool
	}{
		{
			name: "fail: cannot get logged in user",
			args: args{
				ctx:    ctx,
				roleID: uuid.NewString(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "fail: user has no permission permission",
			args: args{
				ctx:    ctx,
				roleID: uuid.NewString(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "fail: error retrieving role",
			args: args{
				ctx:    ctx,
				roleID: uuid.NewString(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "fail: error retrieving user profile",
			args: args{
				ctx:    ctx,
				roleID: uuid.NewString(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "fail: error updating role details",
			args: args{
				ctx:    ctx,
				roleID: uuid.NewString(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success: deactivating role",
			args: args{
				ctx:    ctx,
				roleID: uuid.NewString(),
			},
			want:    &dto.RoleOutput{ID: uuid.NewString()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "fail: cannot get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("cannot get logged in user")
				}
			}

			if tt.name == "fail: user has no permission permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: uuid.NewString()}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, nil
				}
			}

			if tt.name == "fail: error retrieving role" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: uuid.NewString()}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return nil, fmt.Errorf("cannot retrieve role")
				}
			}

			if tt.name == "fail: error retrieving user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: uuid.NewString()}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{ID: uuid.NewString()}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("cannot retrieve user profile")
				}
			}

			if tt.name == "fail: error updating role details" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: uuid.NewString()}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{ID: uuid.NewString()}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: uuid.NewString()}, nil
				}

				fakeInfraRepo.UpdateRoleDetailsFn = func(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error) {
					return nil, fmt.Errorf("cannot update role details")
				}
			}

			if tt.name == "success: deactivating role" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: uuid.NewString()}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{ID: uuid.NewString()}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: uuid.NewString()}, nil
				}

				fakeInfraRepo.UpdateRoleDetailsFn = func(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error) {
					return &profileutils.Role{ID: uuid.NewString()}, nil
				}
			}

			got, err := i.DeactivateRole(tt.args.ctx, tt.args.roleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleUseCaseImpl.DeactivateRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("RoleUseCaseImpl.DeactivateRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleUseCaseImpl_ActivateRole(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()

	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	type args struct {
		ctx    context.Context
		roleID string
	}
	tests := []struct {
		name    string
		args    args
		want    *dto.RoleOutput
		wantErr bool
	}{
		{
			name: "fail: cannot get logged in user",
			args: args{
				ctx:    ctx,
				roleID: uuid.NewString(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "fail: user has no permission permission",
			args: args{
				ctx:    ctx,
				roleID: uuid.NewString(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "fail: error retrieving role",
			args: args{
				ctx:    ctx,
				roleID: uuid.NewString(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "fail: error retrieving user profile",
			args: args{
				ctx:    ctx,
				roleID: uuid.NewString(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "fail: error updating role details",
			args: args{
				ctx:    ctx,
				roleID: uuid.NewString(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "pass: success activating role",
			args: args{
				ctx:    ctx,
				roleID: uuid.NewString(),
			},
			want:    &dto.RoleOutput{ID: uuid.NewString()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "fail: cannot get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("cannot get logged in user")
				}
			}

			if tt.name == "fail: user has no permission permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: uuid.NewString()}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, nil
				}
			}

			if tt.name == "fail: error retrieving role" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: uuid.NewString()}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return nil, fmt.Errorf("cannot retrieve role")
				}
			}

			if tt.name == "fail: error retrieving user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: uuid.NewString()}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{ID: uuid.NewString()}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("cannot retrieve user profile")
				}
			}

			if tt.name == "fail: error updating role details" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: uuid.NewString()}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{ID: uuid.NewString()}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: uuid.NewString()}, nil
				}

				fakeInfraRepo.UpdateRoleDetailsFn = func(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error) {
					return nil, fmt.Errorf("cannot update role details")
				}
			}

			if tt.name == "pass: success activating role" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: uuid.NewString()}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{ID: uuid.NewString()}, nil
				}

				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: uuid.NewString()}, nil
				}

				fakeInfraRepo.UpdateRoleDetailsFn = func(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error) {
					return &profileutils.Role{ID: uuid.NewString()}, nil
				}
			}

			got, err := i.ActivateRole(tt.args.ctx, tt.args.roleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleUseCaseImpl.ActivateRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("RoleUseCaseImpl.ActivateRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleUseCaseImpl_UpdateRolePermissions(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()

	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	input := dto.RolePermissionInput{
		RoleID: "123",
		Scopes: []string{"role.create"},
	}

	allPerms, err := profileutils.AllPermissions(ctx)
	if err != nil {
		t.Error("error did not get all permissions")
		return
	}

	perms := []profileutils.Permission{}
	for _, perm := range allPerms {
		if perm.Scope == "role.create" {
			perm.Allowed = true
		}
		perms = append(perms, perm)
	}

	expectedOutput := dto.RoleOutput{
		ID:          "123",
		Scopes:      []string{"role.create"},
		Permissions: perms,
	}

	type args struct {
		ctx   context.Context
		input dto.RolePermissionInput
	}

	tests := []struct {
		name    string
		args    args
		want    *dto.RoleOutput
		wantErr bool
	}{
		{
			name: "sad unable to get logged in user",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad unable to check if user has permissions",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad user do not have required permission",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad unable to get role by id",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad unable to get user profile",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad unable to update role details",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "happy added role permissions",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    &expectedOutput,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "sad unable to get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("unable to get logged in user")
				}
			}

			if tt.name == "sad unable to check if user has permissions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, fmt.Errorf("unable to check permissions")
				}
			}

			if tt.name == "sad user do not have required permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, nil
				}
			}

			if tt.name == "sad unable to get role by id" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return nil, fmt.Errorf("error unable to get role to edit")
				}
			}

			if tt.name == "sad unable to get user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("unable to get user profile")
				}
			}

			if tt.name == "sad unable to update role details" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}
				fakeInfraRepo.UpdateRoleDetailsFn = func(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error) {
					return nil, fmt.Errorf("error unable to update role")
				}
			}

			if tt.name == "happy added role permissions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: "123"}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}
				fakeInfraRepo.UpdateRoleDetailsFn = func(ctx context.Context, profileID string, role profileutils.Role) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "123",
						Scopes: []string{"role.create"},
					}, nil
				}
			}

			got, err := i.UpdateRolePermissions(tt.args.ctx, tt.args.input)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"RoleUseCaseImpl.UpdateRolePermissions() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RoleUseCaseImpl.UpdateRolePermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleUseCaseImpl_CreateUnauthorizedRole(t *testing.T) {
	ctx := context.Background()
	i, err := InitializeFakeOnboardingInteractor()

	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	input := dto.RoleInput{
		Name: "Agents",
	}

	allPerms, err := profileutils.AllPermissions(ctx)
	if err != nil {
		t.Error("error did not get all permissions")
		return
	}

	perms := []profileutils.Permission{}
	for _, perm := range allPerms {
		if perm.Scope == "role.edit" {
			perm.Allowed = true
		}
		perms = append(perms, perm)
	}
	expectedOutput := &dto.RoleOutput{
		Scopes:      []string{"role.edit"},
		Permissions: perms,
	}

	type args struct {
		ctx   context.Context
		input dto.RoleInput
	}

	tests := []struct {
		name    string
		args    args
		want    *dto.RoleOutput
		wantErr bool
	}{
		{
			name: "sad: unable to get logged in user",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad: unable to get user's profile",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "sad: unable to create role in database",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "happy:created role",
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    expectedOutput,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "sad: unable to get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("unable to get logged in user")
				}
			}

			if tt.name == "sad: unable to get user's profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("error unable to get user profile")
				}
			}

			if tt.name == "sad: unable to create role in database" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}
				fakeInfraRepo.CreateRoleFn = func(ctx context.Context, profileID string, role dto.RoleInput) (*profileutils.Role, error) {
					return nil, fmt.Errorf("error un able to create role in db")
				}
			}

			if tt.name == "happy:created role" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}
				fakeInfraRepo.CreateRoleFn = func(ctx context.Context, profileID string, role dto.RoleInput) (*profileutils.Role, error) {
					return &profileutils.Role{
						Scopes: []string{"role.edit"},
					}, nil
				}
			}

			got, err := i.CreateUnauthorizedRole(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleUseCaseImpl.CreateRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RoleUseCaseImpl.CreateRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleUseCaseImpl_AssignMultipleRoles(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	type args struct {
		ctx     context.Context
		userID  string
		roleIDs []string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "fail: cannot get logged in user",
			args: args{
				ctx:     ctx,
				userID:  uuid.NewString(),
				roleIDs: []string{uuid.NewString()},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "fail: user doesn't have the permission",
			args: args{
				ctx:     ctx,
				userID:  uuid.NewString(),
				roleIDs: []string{uuid.NewString()},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "fail: role ID doesn't exist",
			args: args{
				ctx:     ctx,
				userID:  uuid.NewString(),
				roleIDs: []string{"invalid id"},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "fail: cannot retrieve user profile",
			args: args{
				ctx:     ctx,
				userID:  uuid.NewString(),
				roleIDs: []string{uuid.NewString()},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "fail: role already exists",
			args: args{
				ctx:     ctx,
				userID:  uuid.NewString(),
				roleIDs: []string{"0637333d-74b0-473d-95bd-0a03b1ae5e06"},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "fail: error updating user profile role",
			args: args{
				ctx:     ctx,
				userID:  uuid.NewString(),
				roleIDs: []string{"17e6ea18-7147-4bdb-ad0b-d9ce03a8c0ac"},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "success: add a new role to user",
			args: args{
				ctx:     ctx,
				userID:  uuid.NewString(),
				roleIDs: []string{"17e6ea18-7147-4bdb-ad0b-d9ce03a8c0ac"},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "fail: cannot get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("cannot get logged in user")
				}

				//remove
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return nil, fmt.Errorf("cannot get role ny id")
				}
			}

			if tt.name == "fail: user doesn't have the permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "",
						Scopes: []string{profileutils.CanRegisterAgent.Scope},
					}, nil
				}

				//remove
				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return nil, fmt.Errorf("cannot get role ny id")
				}
			}

			if tt.name == "fail: role ID doesn't exist" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return nil, fmt.Errorf("cannot get role ny id")
				}
			}

			if tt.name == "fail: cannot retrieve user profile" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "",
						Scopes: []string{profileutils.CanAssignRole.Scope},
					}, nil
				}

				fakeInfraRepo.GetUserProfileByIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return nil, fmt.Errorf("no user profile")
				}
			}

			if tt.name == "fail: role already exists" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "0637333d-74b0-473d-95bd-0a03b1ae5e06",
						Scopes: []string{profileutils.CanAssignRole.Scope},
					}, nil
				}

				fakeInfraRepo.GetUserProfileByIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{
						ID:    "",
						Roles: []string{"0637333d-74b0-473d-95bd-0a03b1ae5e06"},
					}, nil
				}
			}

			if tt.name == "fail: error updating user profile role" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "",
						Scopes: []string{profileutils.CanAssignRole.Scope},
					}, nil
				}

				fakeInfraRepo.GetUserProfileByIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: ""}, nil
				}

				fakeInfraRepo.UpdateUserRoleIDsFn = func(ctx context.Context, id string, roleIDs []string) error {
					return fmt.Errorf("cannot update role ids")
				}
			}

			if tt.name == "success: add a new role to user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{UID: ""}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}

				fakeInfraRepo.GetRoleByIDFn = func(ctx context.Context, roleID string) (*profileutils.Role, error) {
					return &profileutils.Role{
						ID:     "",
						Scopes: []string{profileutils.CanAssignRole.Scope},
					}, nil
				}

				fakeInfraRepo.GetUserProfileByIDFn = func(ctx context.Context, id string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{ID: ""}, nil
				}

				fakeInfraRepo.UpdateUserRoleIDsFn = func(ctx context.Context, id string, roleIDs []string) error {
					return nil
				}
			}

			got, err := i.AssignMultipleRoles(tt.args.ctx, tt.args.userID, tt.args.roleIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleUseCaseImpl.AssignMultipleRoles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RoleUseCaseImpl.AssignMultipleRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}
