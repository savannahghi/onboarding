package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/presentation/graph/generated"
	"github.com/savannahghi/profileutils"
	"github.com/savannahghi/scalarutils"
)

func (r *permissionResolver) Group(ctx context.Context, obj *profileutils.Permission) (profileutils.PermissionGroup, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *userProfileResolver) RoleDetails(ctx context.Context, obj *profileutils.UserProfile) ([]*dto.RoleOutput, error) {
	return r.usecases.GetRolesByIDs(ctx, obj.Roles)
}

func (r *verifiedIdentifierResolver) Timestamp(ctx context.Context, obj *profileutils.VerifiedIdentifier) (*scalarutils.Date, error) {
	return &scalarutils.Date{
		Year:  obj.Timestamp.Year(),
		Day:   obj.Timestamp.Day(),
		Month: int(obj.Timestamp.Month()),
	}, nil
}

// Permission returns generated.PermissionResolver implementation.
func (r *Resolver) Permission() generated.PermissionResolver { return &permissionResolver{r} }

// UserProfile returns generated.UserProfileResolver implementation.
func (r *Resolver) UserProfile() generated.UserProfileResolver { return &userProfileResolver{r} }

// VerifiedIdentifier returns generated.VerifiedIdentifierResolver implementation.
func (r *Resolver) VerifiedIdentifier() generated.VerifiedIdentifierResolver {
	return &verifiedIdentifierResolver{r}
}

type permissionResolver struct{ *Resolver }
type userProfileResolver struct{ *Resolver }
type verifiedIdentifierResolver struct{ *Resolver }
