package helpers

import (
	"net/http"
	"strings"

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
	mockApplicationsJSON := BuildMockApplicationsJSON(apps)
	return client.NewTestClientWithKonfluxJSON(httpClient, mockApplicationsJSON)
}

// SetupTestClientWithKonfluxApps creates a test client with gock-intercepted HTTP client
// and custom Konflux application data.
func SetupTestClientWithKonfluxApps(apps []string) (*client.Client, error) {
	httpClient := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(httpClient)
	return NewTestClientWithKonfluxApps(httpClient, apps)
}

// BuildMockApplicationsJSON creates a JSON string representing kubectl application output
func BuildMockApplicationsJSON(apps []string) string {
	var items strings.Builder
	for i, app := range apps {
		if i > 0 {
			items.WriteString(",")
		}
		items.WriteString(`{"metadata":{"name":"`)
		items.WriteString(app)
		items.WriteString(`"}}`)
	}
	return `{"items":[` + items.String() + `]}`
}
