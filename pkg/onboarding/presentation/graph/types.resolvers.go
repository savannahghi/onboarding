package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/presentation/graph/generated"
	"github.com/savannahghi/profileutils"
	"github.com/savannahghi/scalarutils"
)

func (r *navigationActionResolver) OnTapRoute(ctx context.Context, obj *domain.NavigationAction) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *userProfileResolver) RoleDetails(ctx context.Context, obj *profileutils.UserProfile) ([]*dto.RoleOutput, error) {
	return r.interactor.Role.GetRolesByIDs(ctx, obj.Roles)
}

func (r *verifiedIdentifierResolver) Timestamp(ctx context.Context, obj *profileutils.VerifiedIdentifier) (*scalarutils.Date, error) {
	return &scalarutils.Date{
		Year:  obj.Timestamp.Year(),
		Day:   obj.Timestamp.Day(),
		Month: int(obj.Timestamp.Month()),
	}, nil
}

// NavigationAction returns generated.NavigationActionResolver implementation.
func (r *Resolver) NavigationAction() generated.NavigationActionResolver {
	return &navigationActionResolver{r}
}

// UserProfile returns generated.UserProfileResolver implementation.
func (r *Resolver) UserProfile() generated.UserProfileResolver { return &userProfileResolver{r} }

// VerifiedIdentifier returns generated.VerifiedIdentifierResolver implementation.
func (r *Resolver) VerifiedIdentifier() generated.VerifiedIdentifierResolver {
	return &verifiedIdentifierResolver{r}
}

type navigationActionResolver struct{ *Resolver }
type userProfileResolver struct{ *Resolver }
type verifiedIdentifierResolver struct{ *Resolver }
