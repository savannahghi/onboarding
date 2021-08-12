package usecases_test

import (
	"context"
	"testing"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/profileutils"
)

func TestRoleUseCase_CreateRole(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v", err)
		return
	}

	validRoleInput := dto.RoleInput{
		Name:        "Test Role",
		Description: "Role Descriptions",
		Scopes:      []string{"scope1", "scope2"},
	}

	type args struct {
		ctx   context.Context
		input dto.RoleInput
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy:created role",
			args: args{
				ctx:   ctx,
				input: validRoleInput,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "happy:created role" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}
				fakeRepo.CreateRoleFn = func(ctx context.Context, profileID string, role dto.RoleInput) (*profileutils.Role, error) {
					return &profileutils.Role{
						Scopes: []string{"role.edit"},
					}, nil
				}
			}
			_, err := i.Role.CreateRole(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleUseCaseImpl.CreateRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("RoleUseCaseImpl.CreateRole() = %v, want %v", got, tt.want)
			// }
		})
	}

}
