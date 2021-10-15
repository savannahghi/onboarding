package pubsubmessaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/common"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
)

func (ps *ServicePubSubMessaging) newPublish(
	ctx context.Context,
	data interface{},
	topic string,
) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("unable to marshal data received: %v", err)
	}
	return ps.PublishToPubsub(
		ctx,
		ps.AddPubSubNamespace(topic),
		payload,
	)
}

// NotifyCreateCustomer publishes to customers.create topic
func (ps *ServicePubSubMessaging) NotifyCreateCustomer(
	ctx context.Context,
	data dto.CustomerPubSubMessage,
) error {
	return ps.newPublish(ctx, data, common.CreateCustomerTopic)
}

// NotifyCreateSupplier publishes to suppliers.create topic
func (ps *ServicePubSubMessaging) NotifyCreateSupplier(
	ctx context.Context,
	data dto.SupplierPubSubMessage,
) error {
	return ps.newPublish(ctx, data, common.CreateCustomerTopic)
}

// NotifyCoverLinking pushes to covers.link topic
func (ps *ServicePubSubMessaging) NotifyCoverLinking(
	ctx context.Context,
	data dto.LinkCoverPubSubMessage,
) error {
	return ps.newPublish(ctx, data, common.LinkCoverTopic)
}

// EDIMemberCoverLinking publishes to the edi.covers.link topic. The reason for this is
// to Auto-link the Sladers who get text messages from EDI. If a slader is converted
// and creates an account on Be.Well app, we should automatically append a cover to their profile.
func (ps *ServicePubSubMessaging) EDIMemberCoverLinking(
	ctx context.Context,
	data dto.LinkCoverPubSubMessage,
) error {
	return ps.newPublish(ctx, data, common.LinkEDIMemberCoverTopic)
}
