package helpers

import (
	"net/http"

	"github.com/h2non/gock"
	"github.com/stackrox/relacs/pkg/client"
)

// NewTestClient creates a client with a custom HTTP client for testing.
// The HTTP client can be intercepted by gock for mocking API responses.
func NewTestClient(httpClient *http.Client) (*client.Client, error) {
	return client.NewTestClient(httpClient)
}

// NewTestClientWithKonfluxApps creates a test client with custom Konflux application data.
// The apps parameter is a slice of application names (e.g., ["acs-4-5", "acs-4-6"]).
func NewTestClientWithKonfluxApps(httpClient *http.Client, apps []string) (*client.Client, error) {
	return client.NewTestClientWithKonfluxApps(httpClient, apps)
}

// SetupTestClientWithKonfluxApps creates a test client with gock-intercepted HTTP client
// and custom Konflux application data.
func SetupTestClientWithKonfluxApps(apps []string) (*client.Client, error) {
	httpClient := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(httpClient)
	return NewTestClientWithKonfluxApps(httpClient, apps)
}
