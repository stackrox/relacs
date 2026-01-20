package client

import (
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/v79/github"
	"github.com/stackrox/relacs/pkg/client/konflux"
)

// Client is a base client that wraps the GitHub client with authentication
type Client struct {
	gh   *github.Client
	kflx *konflux.Client
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

	return &Client{gh: gh, kflx: kflx}, nil
}

// GetGitHub returns the underlying GitHub client for advanced usage
func (c *Client) GetGitHub() *github.Client {
	return c.gh
}

func (c *Client) GetKonflux() *konflux.Client {
	return c.kflx
}

// NewTestClient creates a client with a custom HTTP client for testing.
// The HTTP client can be intercepted by gock for mocking API responses.
func NewTestClient(httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	// Create GitHub client with custom HTTP client
	gh := github.NewClient(httpClient).WithAuthToken("test-token")

	// For Konflux client, we need to handle the case where KUBECTL_PLUGIN_KA_HOST might not be set
	// In tests, we can create a mock Konflux client or skip it
	// For now, we'll try to create it, but it might fail - tests that need Konflux should set the env var
	kflx, err := konflux.NewClient()
	if err != nil {
		// In test mode, we can continue without Konflux client if it's not needed
		// The test will fail if Konflux is actually used
		kflx = nil
	}

	return &Client{gh: gh, kflx: kflx}, nil
}
