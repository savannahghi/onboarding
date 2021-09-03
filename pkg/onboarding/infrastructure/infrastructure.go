package infrastructure

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/utils"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/database"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/engagement"
	pubsubmessaging "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/pubsub"
	"github.com/savannahghi/serverutils"
)

const (
	// ServiceName ..
	ServiceName = "onboarding"

	// TopicVersion ...
	TopicVersion = "v1"
)

// Infrastructure defines the contract provided by the infrastructure layer
// It's a combination of interactions with external services/dependencies
type Infrastructure interface {
	database.Repository
	engagement.ServiceEngagement
	pubsubmessaging.ServicePubSub
}

// Interactor is an implementation of the infrastructure interface
// It combines each individual service implementation
type Interactor struct {
	database.Repository
	engagement.ServiceEngagement
	pubsubmessaging.ServicePubSub
}

// NewInfrastructureInteractor initializes a new infrastructure interactor
func NewInfrastructureInteractor() (Infrastructure, error) {
	ctx := context.Background()

	db := database.NewDbService()

	baseExtension := extension.NewBaseExtensionImpl()

	projectID, err := serverutils.GetEnvVar(serverutils.GoogleCloudProjectIDEnvVarName)
	if err != nil {
		return nil, err
	}

	pubSubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	pubsub, err := pubsubmessaging.NewServicePubSubMessaging(pubSubClient, baseExtension, db)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize new pubsub messaging service: %w", err)
	}

	engagementClient := utils.NewInterServiceClient("engagement", baseExtension)
	engagement := engagement.NewServiceEngagementImpl(engagementClient, baseExtension)

	return &Interactor{
		db,
		engagement,
		pubsub,
	}, nil
}
