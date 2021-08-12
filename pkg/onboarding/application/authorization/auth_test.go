package authorization

import (
	"testing"

	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
)

func TestCheckPemissions(t *testing.T) {
	type args struct {
		subject string
		input   dto.PermissionInput
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "valid: permission is set and subject has permission",
			args: args{
				subject: "254711223344",
				input: dto.PermissionInput{
					Resource: "update_primary_phone",
					Action:   "edit",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "valid: unknown subject with unknown resource",
			args: args{
				subject: "mail@example.com",
				input: dto.PermissionInput{
					Resource: "unknown_resource",
					Action:   "edit",
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckPemissions(tt.args.subject, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPemissions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckPemissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckAuthorization(t *testing.T) {
	type args struct {
		subject    string
		permission dto.PermissionInput
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "valid: permission is set and subject has permission",
			args: args{
				subject: "254711223344",
				permission: dto.PermissionInput{
					Resource: "update_primary_phone",
					Action:   "edit",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "valid: unknown subject with unknown resource",
			args: args{
				subject: "mail@example.com",
				permission: dto.PermissionInput{
					Resource: "unknown_resource",
					Action:   "edit",
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckAuthorization(tt.args.subject, tt.args.permission)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAuthorization() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckAuthorization() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAuthorized(t *testing.T) {
	userInfo := dto.UserInfo{PhoneNumber: interserviceclient.TestUserPhoneNumber, Email: firebasetools.TestUserEmail}

	permission := dto.PermissionInput{
		Action:   "test",
		Resource: "http://example.com",
	}

	type args struct {
		user       *dto.UserInfo
		permission dto.PermissionInput
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid: args",
			args: args{
				user:       &userInfo,
				permission: permission,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		_, err := IsAuthorized(tt.args.user, tt.args.permission)

		if !tt.wantErr && err != nil {
			t.Errorf("error not expected, got %v", err)
		}

	}
}
