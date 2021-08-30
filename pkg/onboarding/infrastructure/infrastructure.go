package infrastructure

import (
	"context"

	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/database"
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
