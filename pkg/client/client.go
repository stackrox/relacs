package client

import (
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/v79/github"
	"github.com/stackrox/relacs/pkg/client/jira"
	"github.com/stackrox/relacs/pkg/client/konflux"
)

// Client is a base client that wraps the GitHub client with authentication
type Client struct {
	gh   *github.Client
	kflx *konflux.Client
	jira *jira.Client
}

// NewClient creates a new base client with GitHub authentication.
// It requires the GH_TOKEN environment variable to be set.
func NewClient() (*Client, error) {
	token := os.Getenv("GH_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GH_TOKEN environment variable is required")
	}

	gh := github.NewClient(nil).WithAuthToken(token)

	kflx, err := konflux.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Konflux client: %w", err)
	}

	// Jira client is optional - only create if credentials are provided
	jiraClient, err := jira.NewClient()
	if err != nil {
		// Jira is optional, so we don't fail if it's not configured
		jiraClient = nil
	}

	return &Client{gh: gh, kflx: kflx, jira: jiraClient}, nil
}

// GetGitHub returns the underlying GitHub client for advanced usage
func (c *Client) GetGitHub() *github.Client {
	return c.gh
}

func (c *Client) GetKonflux() *konflux.Client {
	return c.kflx
}

func (c *Client) GetJira() *jira.Client {
	return c.jira
}

// NewTestClient creates a client with a custom HTTP client for testing.
// The HTTP client can be intercepted by gock for mocking API responses.
func NewTestClient(httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	// Create GitHub client with custom HTTP client
	gh := github.NewClient(httpClient).WithAuthToken("test-token")

	// Create a test Konflux client with default mock data
	// Tests can customize this by providing specific mock application lists
	mockApplicationsJSON := `{"items":[
		{"metadata":{"name":"acs-4-4"}},
		{"metadata":{"name":"acs-4-5"}},
		{"metadata":{"name":"acs-4-6"}},
		{"metadata":{"name":"acs-4-7"}},
		{"metadata":{"name":"other-app"}},
		{"metadata":{"name":"acs-5-0"}}
	]}`
	kflx := konflux.NewTestClient(mockApplicationsJSON)

	// Always create a Jira test client with the shared HTTP client so gock can intercept requests
	jiraClient := jira.NewTestClient(httpClient)

	return &Client{gh: gh, kflx: kflx, jira: jiraClient}, nil
}

// NewTestClientWithKonfluxApps creates a test client with custom Konflux application data.
// The apps parameter is a slice of application names (e.g., ["acs-4-5", "acs-4-6"]).
func NewTestClientWithKonfluxApps(httpClient *http.Client, apps []string) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	// Create GitHub client with custom HTTP client
	gh := github.NewClient(httpClient).WithAuthToken("test-token")

	// Build JSON for the apps
	mockApplicationsJSON := buildMockApplicationsJSON(apps)
	kflx := konflux.NewTestClient(mockApplicationsJSON)

	// Jira client is optional in tests
	jiraClient, err := jira.NewClient()
	if err != nil {
		jiraClient = nil
	}

	return &Client{gh: gh, kflx: kflx, jira: jiraClient}, nil
}

// buildMockApplicationsJSON creates a JSON string representing kubectl application output
func buildMockApplicationsJSON(apps []string) string {
	items := ""
	for i, app := range apps {
		if i > 0 {
			items += ","
		}
		items += `{"metadata":{"name":"` + app + `"}}`
	}
	return `{"items":[` + items + `]}`
}
