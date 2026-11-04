package helpers

import (
	"net/http"

	"github.com/stackrox/relacs/pkg/client"
)

// NewTestClient creates a client with a custom HTTP client for testing.
// The HTTP client can be intercepted by gock for mocking API responses.
func NewTestClient(httpClient *http.Client) *client.Client {
	return client.NewTestClient(httpClient)
}
