package database_test

import (
	"context"
	"testing"

	"cloud.google.com/go/firestore"
	"gitlab.slade360emr.com/go/profile/pkg/onboarding/infrastructure/database"
	extMock "gitlab.slade360emr.com/go/profile/pkg/onboarding/infrastructure/database/mock"
)

var fakeFireBaseClientExt extMock.FirebaseClientExtension
var fireBaseClientExt database.FirebaseClientExtension = &fakeFireBaseClientExt

var fakeFireStoreClientExt extMock.FirestoreClientExtension

func TestRepository_UpdateUserName(t *testing.T) {
	ctx := context.Background()
	var fireStoreClientExt database.FirestoreClientExtension = &fakeFireStoreClientExt
	repo := database.NewFirebaseRepository(fireStoreClientExt, fireBaseClientExt)

	type args struct {
		ctx      context.Context
		id       string
		userName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid:update_user_name_failed_to_get_a user_profile",
			args: args{
				ctx:      ctx,
				id:       "12333",
				userName: "mwas",
			},
			wantErr: true,
		},
		{
			name: "invalid:user_name_already_exists",
			args: args{
				ctx:      ctx,
				id:       "12333",
				userName: "mwas",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid:update_user_name_failed_to_get_a user_profile" {
				fakeFireStoreClientExt.GetAllFn = func(ctx context.Context, query *database.GetAllQuery) ([]*firestore.DocumentSnapshot, error) {
					docs := []*firestore.DocumentSnapshot{}
					return docs, nil
				}

				fakeFireStoreClientExt.UpdateFn = func(ctx context.Context, command *database.UpdateCommand) error {
					return nil
				}
			}

			if tt.name == "invalid:user_name_already_exists" {
				fakeFireStoreClientExt.GetAllFn = func(ctx context.Context, query *database.GetAllQuery) ([]*firestore.DocumentSnapshot, error) {
					docs := []*firestore.DocumentSnapshot{
						{
							Ref: &firestore.DocumentRef{
								ID: "5555",
							},
						},
					}
					return docs, nil
				}

				fakeFireStoreClientExt.UpdateFn = func(ctx context.Context, command *database.UpdateCommand) error {
					return nil
				}
			}

			err := repo.UpdateUserName(tt.args.ctx, tt.args.id, tt.args.userName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected got %v", err)
					return
				}
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("error not expected got %v", err)
					return
				}
			}

		})
	}
}
