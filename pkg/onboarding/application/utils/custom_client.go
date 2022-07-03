package utils

import (
	"context"
	"net/http"

	"golang.org/x/time/rate"
)

// RatedHTTPClient Rate Limited HTTP Client
// This client will be used in-place of the default client which may be throttled for send
// too many requests
// The usecase of this client is simple: A concurrent process that produces streams to request that the
// servicing server may not handle as quickly as the requests are been sent
//
type RatedHTTPClient struct {
	client      *http.Client
	Ratelimiter *rate.Limiter
}

//Do dispatches the HTTP request to the network
func (c *RatedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	ctx := context.Background()
	err := c.Ratelimiter.Wait(ctx) // This is a blocking call. Honors the rate limit
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

//NewClient return http client with a ratelimiter
func NewClient(rl *rate.Limiter) *RatedHTTPClient {
	c := &RatedHTTPClient{
		client:      http.DefaultClient,
		Ratelimiter: rl,
	}
	return c
}
