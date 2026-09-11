package usecases_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/profileutils"
)

func TestRoleUseCaseImpl_CreatePermission(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	input := dto.PermissionInput{
		Name:  "Agents",
		Scope: "agents.create",
	}

	expectedOutput := &domain.RolePermission{
		Scope: "agent.create",
		Name:  "RESGISTER_AGENT",
	}

	type args struct {
		ctx   context.Context
		input dto.PermissionInput
	}

	tests := map[string]struct {
		name    string
		args    args
		want    *domain.RolePermission
		wantErr bool
	}{
		"sad: unable to get logged in user": {
			args: args{
				ctx:   ctx,
				input: input,
			},

			wantErr: true,
		},
		"sad: unable to check if user has permissions": {
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		"sad: user do not have required permissions": {
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		"sad: unable to get user's profile": {
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		"sad: unable to create permission in database": {
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    nil,
			wantErr: true,
		},
		"happy:created permission": {
			args: args{
				ctx:   ctx,
				input: input,
			},
			want:    expectedOutput,
			wantErr: false,
		},
	}
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {

			if testName == "sad: unable to get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("unable to get logged in user")
				}
			}

			if testName == "sad: unable to check if user has permissions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, fmt.Errorf(
						"error unable to check if user has required permissions",
					)
				}
			}

			if testName == "sad: user do not have required permissions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}

				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, nil
				}
			}

			if testName == "sad: unable to get user's profile" {
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

			if testName == "sad: unable to create role in database" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}
				fakeInfraRepo.CreatePermissionFn = func(ctx context.Context, profileID string, input dto.PermissionInput) (*domain.RolePermission, error) {
					return nil, fmt.Errorf("error un able to create permission")
				}
			}

			if testName == "happy:created permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetUserProfileByUIDFn = func(ctx context.Context, uid string, suspended bool) (*profileutils.UserProfile, error) {
					return &profileutils.UserProfile{}, nil
				}
				fakeInfraRepo.CreatePermissionFn = func(ctx context.Context, profileID string, input dto.PermissionInput) (*domain.RolePermission, error) {
					return &domain.RolePermission{
						Scope: "agent.create",
						Name:  "RESGISTER_AGENT",
					}, nil
				}
			}

			got, err := i.CreatePermission(test.args.ctx, test.args.input)
			if (err != nil) != test.wantErr {
				t.Errorf("RoleUseCaseImpl.CreatePermission() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("RoleUseCaseImpl.CreatePermission() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestRoleUseCaseImpl_GetPermission(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	type args struct {
		ctx context.Context
	}

	tests := map[string]struct {
		name    string
		args    args
		want    []*dto.RoleOutput
		wantErr bool
	}{
		"sad: did not get logged in user": {
			args:    args{ctx: ctx},
			wantErr: true,
		},
		"sad: unable to check if user has permission": {
			args:    args{ctx: ctx},
			wantErr: true,
		},
		"sad: user do not have required permission": {
			args:    args{ctx: ctx},
			wantErr: true,
		},
		"sad: did not get permissions from database": {
			args:    args{ctx: ctx},
			wantErr: true,
		},
		"happy: got permissions": {
			args:    args{ctx: ctx},
			wantErr: false,
		},
	}
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			if testName == "sad: did not get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("error, did not get logged in user")
				}
			}

			if testName == "sad: unable to check if user has permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, fmt.Errorf("error unable to check is user has permission")
				}
			}

			if testName == "sad: user do not have required permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, nil
				}
			}
			if testName == "sad: did not get permissions from database" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetAllPermissionsFn = func(ctx context.Context) (*[]domain.RolePermission, error) {
					return nil, fmt.Errorf("error, did not get permissions from database")
				}
			}

			if testName == "happy: got permissions" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetAllPermissionsFn = func(ctx context.Context) (*[]domain.RolePermission, error) {
					return &[]domain.RolePermission{}, nil
				}
				fakeInfraRepo.GetUserProfilesByRoleIDFn = func(ctx context.Context, role string) ([]*profileutils.UserProfile, error) {

					return []*profileutils.UserProfile{}, nil
				}
			}
			got, err := i.GetPermissions(test.args.ctx)
			if (err != nil) != test.wantErr {
				t.Errorf(
					"RoleUseCaseImpl.DeletePermission() error = %v, wantErr %v",
					err,
					test.wantErr,
				)
				return
			}
			if got != nil && test.wantErr {
				t.Errorf("expcted the operation to fail ")
			}
		})
	}
}

func TestRoleUseCaseImpl_DeletePermission(t *testing.T) {
	ctx := context.Background()

	i, err := InitializeFakeOnboardingInteractor()
	if err != nil {
		t.Errorf("failed to fake initialize onboarding interactor: %v",
			err,
		)
		return
	}

	type args struct {
		ctx   context.Context
		scope string
	}
	tests := map[string]struct {
		name    string
		args    args
		want    []*dto.RoleOutput
		wantErr bool
	}{
		"sad: did not get logged in user": {
			args:    args{ctx: ctx},
			wantErr: true,
		},
		"sad: unable to check if user has permission": {
			args:    args{ctx: ctx},
			wantErr: true,
		},
		"sad: user do not have required permission": {
			args:    args{ctx: ctx},
			wantErr: true,
		},
		"sad: did not get permissions from database": {
			args:    args{ctx: ctx},
			wantErr: true,
		},
		"happy: permission deleted successfully": {
			args:    args{ctx: ctx},
			wantErr: false,
		},
	}
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			if testName == "sad: did not get logged in user" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return nil, fmt.Errorf("error, did not get logged in user")
				}
			}

			if testName == "sad: unable to check if user has permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, fmt.Errorf("error unable to check is user has permission")
				}
			}

			if testName == "sad: user do not have required permission" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return false, nil
				}
			}
			if testName == "sad: did not get permissions from database" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.DeletePermissionFn = func(ctx context.Context, permissionScope, profileID string) (bool, error) {
					return false, fmt.Errorf("error, did not get roles from database")
				}
			}

			if testName == "happy: permission deleted successfully" {
				fakeBaseExt.GetLoggedInUserFn = func(ctx context.Context) (*dto.UserInfo, error) {
					return &dto.UserInfo{}, nil
				}
				fakeInfraRepo.CheckIfUserHasPermissionFn = func(ctx context.Context, UID string, requiredPermission profileutils.Permission) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.DeletePermissionFn = func(ctx context.Context, permissionScope, profileID string) (bool, error) {
					return true, nil
				}
				fakeInfraRepo.GetUserProfilesByRoleIDFn = func(ctx context.Context, role string) ([]*profileutils.UserProfile, error) {

					return []*profileutils.UserProfile{}, nil
				}
			}
			got, err := i.DeletePermission(test.args.ctx, test.args.scope)
			if (err != nil) != test.wantErr {
				t.Errorf(
					"RoleUseCaseImpl.DeletePermission() error = %v, wantErr %v",
					err,
					test.wantErr,
				)
				return
			}
			if got == true && test.wantErr {
				t.Errorf("expcted the operation to fail ")
			}
		})
	}
}
