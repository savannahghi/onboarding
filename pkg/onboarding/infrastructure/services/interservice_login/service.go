package interservicelogin

import (
	"context"
	"fmt"

	"github.com/savannahghi/interserviceclient"
)

// GetInterserviceBearerTokenHeader ...
func GetInterserviceBearerTokenHeader(ctx context.Context) (string, error) {
	service := interserviceclient.ISCService{} // name and domain not necessary for our use case
	isc, err := interserviceclient.NewInterserviceClient(service)
	if err != nil {
		return "", fmt.Errorf("can't initialize interservice client: %w", err)
	}

	authToken, err := isc.CreateAuthToken(ctx)
	if err != nil {
		return "", fmt.Errorf("can't get auth token: %w", err)
	}
	bearerHeader := fmt.Sprintf("Bearer %s", authToken)
	return bearerHeader, nil
}
