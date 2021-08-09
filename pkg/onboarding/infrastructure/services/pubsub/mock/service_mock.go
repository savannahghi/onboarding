package mock

import (
	"context"
	"net/http"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"gitlab.slade360emr.com/go/commontools/crm/pkg/domain"
)

// FakeServicePubSub ...
type FakeServicePubSub struct {
	TopicIDsFn           func() []string
	AddPubSubNamespaceFn func(topicName string) string
	PublishToPubsubFn    func(
		ctx context.Context,
		topicID string,
		payload []byte,
	) error
	EnsureTopicsExistFn func(
		ctx context.Context,
		topicIDs []string,
	) error
	EnsureSubscriptionsExistFn func(
		ctx context.Context,
	) error
	SubscriptionIDsFn           func() map[string]string
	ReceivePubSubPushMessagesFn func(
		w http.ResponseWriter,
		r *http.Request,
	)
	AddEngagementPubsubNameSpaceFn func(topic string) string
	NotifyCreateContactFn          func(ctx context.Context, contact domain.CRMContact) error
	NotifyUpdateContactFn          func(ctx context.Context, contact domain.CRMContact) error
	NotifyCreateCustomerFn         func(ctx context.Context, data dto.CustomerPubSubMessage) error
	NotifyCreateSupplierFn         func(ctx context.Context, data dto.SupplierPubSubMessage) error
	NotifyCoverLinkingFn           func(ctx context.Context, data dto.LinkCoverPubSubMessage) error
	EDIMemberCoverLinkingFn        func(ctx context.Context, data dto.LinkCoverPubSubMessage) error
}

// AddPubSubNamespace ...
func (m *FakeServicePubSub) AddPubSubNamespace(topicName string) string {
	return m.AddPubSubNamespaceFn(topicName)
}

// TopicIDs returns the known (registered) topic IDs
func (m *FakeServicePubSub) TopicIDs() []string {
	return m.TopicIDsFn()
}

// PublishToPubsub ..
func (m *FakeServicePubSub) PublishToPubsub(
	ctx context.Context,
	topicID string,
	payload []byte,
) error {
	return m.PublishToPubsubFn(
		ctx,
		topicID,
		payload,
	)
}

// EnsureTopicsExist ...
func (m *FakeServicePubSub) EnsureTopicsExist(
	ctx context.Context,
	topicIDs []string,
) error {
	return m.EnsureTopicsExistFn(ctx, topicIDs)
}

// EnsureSubscriptionsExist ..
func (m *FakeServicePubSub) EnsureSubscriptionsExist(
	ctx context.Context,
) error {
	return m.EnsureSubscriptionsExistFn(ctx)
}

// SubscriptionIDs ..
func (m *FakeServicePubSub) SubscriptionIDs() map[string]string {
	return m.SubscriptionIDsFn()
}

// ReceivePubSubPushMessages ...
func (m *FakeServicePubSub) ReceivePubSubPushMessages(
	w http.ResponseWriter,
	r *http.Request,
) {
	m.ReceivePubSubPushMessagesFn(w, r)
}

// AddEngagementPubsubNameSpace ...
func (m *FakeServicePubSub) AddEngagementPubsubNameSpace(topic string) string {
	return m.AddEngagementPubsubNameSpaceFn(topic)
}

// NotifyCreateContact ..
func (m *FakeServicePubSub) NotifyCreateContact(ctx context.Context, contact domain.CRMContact) error {
	return m.NotifyCreateContactFn(ctx, contact)
}

// NotifyUpdateContact ..
func (m *FakeServicePubSub) NotifyUpdateContact(ctx context.Context, contact domain.CRMContact) error {
	return m.NotifyUpdateContactFn(ctx, contact)
}

// NotifyCreateCustomer ..
func (m *FakeServicePubSub) NotifyCreateCustomer(ctx context.Context, data dto.CustomerPubSubMessage) error {
	return m.NotifyCreateCustomerFn(ctx, data)
}

// NotifyCreateSupplier ..
func (m *FakeServicePubSub) NotifyCreateSupplier(ctx context.Context, data dto.SupplierPubSubMessage) error {
	return m.NotifyCreateSupplierFn(ctx, data)
}

// NotifyCoverLinking ..
func (m *FakeServicePubSub) NotifyCoverLinking(ctx context.Context, data dto.LinkCoverPubSubMessage) error {
	return m.NotifyCoverLinkingFn(ctx, data)
}

// EDIMemberCoverLinking represents a mock of the EDIMemberCoverLinking
func (m *FakeServicePubSub) EDIMemberCoverLinking(ctx context.Context, data dto.LinkCoverPubSubMessage) error {
	return m.EDIMemberCoverLinkingFn(ctx, data)
}
