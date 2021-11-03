package infrastructure

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/savannahghi/firebasetools"
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

// Infrastructure is an implementation of the infrastructure interface
// It combines each individual service implementation
type Infrastructure struct {
	Database   database.Repository
	Engagement engagement.ServiceEngagement
	Pubsub     pubsubmessaging.ServicePubSub
}

// NewInfrastructureInteractor initializes a new infrastructure interactor
func NewInfrastructureInteractor() Infrastructure {
	ctx := context.Background()

	db := database.NewDbService()

	baseExtension := extension.NewBaseExtensionImpl(&firebasetools.FirebaseClient{})

	projectID, err := serverutils.GetEnvVar(serverutils.GoogleCloudProjectIDEnvVarName)
	if err != nil {
		log.Fatal(err)
	}

	pubSubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}
	pubsub, err := pubsubmessaging.NewServicePubSubMessaging(pubSubClient, baseExtension, db)
	if err != nil {
		log.Fatal(err)
	}

	engagementClient := utils.NewInterServiceClient("engagement", baseExtension)
	engagement := engagement.NewServiceEngagementImpl(engagementClient, baseExtension)

	return Infrastructure{
		db,
		engagement,
		pubsub,
	}
}
