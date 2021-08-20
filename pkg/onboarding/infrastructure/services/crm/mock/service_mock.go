package mock

import (
	"context"

	hubspotDomain "gitlab.slade360emr.com/go/commontools/crm/pkg/domain"
)

// FakeServiceCrm ..
type FakeServiceCrm struct {
	OptOutFn                  func(ctx context.Context, phoneNumber string) (*hubspotDomain.CRMContact, error)
	CreateHubSpotContactFn    func(ctx context.Context, contact *hubspotDomain.CRMContact) (*hubspotDomain.CRMContact, error)
	UpdateHubSpotContactFn    func(ctx context.Context, contact *hubspotDomain.CRMContact) (*hubspotDomain.CRMContact, error)
	GetContactByPhoneFn       func(ctx context.Context, phone string) (*hubspotDomain.CRMContact, error)
	SearchContactByPhoneFn    func(phone string) (*hubspotDomain.ContactSearchResponse, error)
	CreateHubspotEngagementFn func(ctx context.Context, eng *hubspotDomain.EngagementData) (*hubspotDomain.EngagementData, error)
}

// OptOut ..
func (f *FakeServiceCrm) OptOut(ctx context.Context, phoneNumber string) (*hubspotDomain.CRMContact, error) {
	return f.OptOutFn(ctx, phoneNumber)
}

// CreateHubSpotContact ..
func (f *FakeServiceCrm) CreateHubSpotContact(ctx context.Context, contact *hubspotDomain.CRMContact) (*hubspotDomain.CRMContact, error) {
	return f.CreateHubSpotContactFn(ctx, contact)
}

// UpdateHubSpotContact ..
func (f *FakeServiceCrm) UpdateHubSpotContact(ctx context.Context, contact *hubspotDomain.CRMContact) (*hubspotDomain.CRMContact, error) {
	return f.UpdateHubSpotContactFn(ctx, contact)
}

// GetContactByPhone ..
func (f *FakeServiceCrm) GetContactByPhone(ctx context.Context, phone string) (*hubspotDomain.CRMContact, error) {
	return f.GetContactByPhoneFn(ctx, phone)
}

// SearchContactByPhone ..
func (f *FakeServiceCrm) SearchContactByPhone(phone string) (*hubspotDomain.ContactSearchResponse, error) {
	return f.SearchContactByPhoneFn(phone)
}

// CreateHubspotEngagement ..
func (f *FakeServiceCrm) CreateHubspotEngagement(ctx context.Context, eng *hubspotDomain.EngagementData) (*hubspotDomain.EngagementData, error) {
	return f.CreateHubspotEngagementFn(ctx, eng)
}
