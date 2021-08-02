package utils

import (
	"testing"

	"github.com/savannahghi/profileutils"
)

func TestCheckUserHasPermission(t *testing.T) {
	type args struct {
		roles      []profileutils.Role
		permission profileutils.Permission
	}
	tests := []struct {
		name string
		args args
		want bool
	}{

		{
			name: "sad: user do not have permission, role deactivated",
			args: args{
				roles: []profileutils.Role{
					{Name: "Employee Role", Scopes: []string{"agent.view"}, Active: false},
				},
				permission: profileutils.CanViewAgent,
			},
			want: false,
		},

		{
			name: "sad: user do not have permission, no such scope",
			args: args{
				roles: []profileutils.Role{
					{Name: "Employee Role", Scopes: []string{"patient.create"}, Active: true},
				},
				permission: profileutils.CanViewAgent,
			},
			want: false,
		},
		{
			name: "happy: user has permission",
			args: args{
				roles: []profileutils.Role{
					{Name: "Employee Role", Scopes: []string{"agent.view"}, Active: true},
				},
				permission: profileutils.CanViewAgent,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckUserHasPermission(tt.args.roles, tt.args.permission); got != tt.want {
				t.Errorf("CheckUserHasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}
