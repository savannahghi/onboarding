package mock

import (
	"context"
)

// FakeInterServiceLogin is an interservice login mock
type FakeInterServiceLogin struct {
	GetInterserviceBearerTokenHeaderFn func(ctx context.Context) (string, error)
}

// GetInterserviceBearerTokenHeader ...
func (f *FakeInterServiceLogin) GetInterserviceBearerTokenHeader(ctx context.Context) (string, error) {
	return f.GetInterserviceBearerTokenHeaderFn(ctx)
}
