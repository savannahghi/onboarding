package interservicelogin

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	interservice_login_mock "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/interservice_login/mock"
)

var fakeInterServiceLogin interservice_login_mock.FakeInterServiceLogin

func TestGetInterserviceBearerTokenHeader(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "get token",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
		{
			name: "get no token",
			args: args{
				ctx: nil,
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "get token" {
				fakeInterServiceLogin.GetInterserviceBearerTokenHeaderFn = func(ctx context.Context) (string, error) {
					return uuid.NewString(), nil
				}
			}

			if tt.name == "get no token" {
				fakeInterServiceLogin.GetInterserviceBearerTokenHeaderFn = func(ctx context.Context) (string, error) {
					return "", fmt.Errorf("unable to get token")
				}
			}

			got, err := GetInterserviceBearerTokenHeader(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInterserviceBearerTokenHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "some token" {
				t.Errorf("GetInterserviceBearerTokenHeader() = %v, want %v", got, tt.wantErr)
			}
		})
	}
}
