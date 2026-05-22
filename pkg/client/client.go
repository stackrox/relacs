package client

import (
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/v79/github"
	"github.com/stackrox/relacs/pkg/client/jira"
	"github.com/stackrox/relacs/pkg/client/konflux"
	"github.com/stackrox/relacs/pkg/config"
)

// Client is a base client that wraps the GitHub client with authentication
type Client struct {
	gh   *github.Client
	kflx *konflux.Client
	jira *jira.Client
	cfg  *config.Config
}

// NewClient creates a new base client with GitHub authentication.
// It requires the GH_TOKEN environment variable to be set.
func NewClient(cfg *config.Config) (*Client, error) {
	token := os.Getenv("GH_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GH_TOKEN environment variable is required")
	}

	gh := github.NewClient(nil).WithAuthToken(token)

	kflx, err := konflux.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Konflux client: %w", err)
	}

	// Jira client is optional - only create if credentials are provided
	jiraClient, err := jira.NewClient(cfg)
	if err != nil {
		// Jira is optional, so we don't fail if it's not configured
		jiraClient = nil
	}

	return &Client{gh: gh, kflx: kflx, jira: jiraClient, cfg: cfg}, nil
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
func NewTestClient(cfg *config.Config, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	// Create GitHub client with custom HTTP client
	gh := github.NewClient(httpClient).WithAuthToken("test-token")

	// For Konflux client, we need to handle the case where KUBECTL_PLUGIN_KA_HOST might not be set
	// In tests, we can create a mock Konflux client or skip it
	// For now, we'll try to create it, but it might fail - tests that need Konflux should set the env var
	kflx, err := konflux.NewClient(cfg)
	if err != nil {
		// In test mode, we can continue without Konflux client if it's not needed
		// The test will fail if Konflux is actually used
		kflx = nil
	}

	// Jira client is optional in tests
	jiraClient, err := jira.NewClient(cfg)
	if err != nil {
		jiraClient = nil
	}

	return &Client{gh: gh, kflx: kflx, jira: jiraClient, cfg: cfg}, nil
}

// GetConfig returns the configuration
func (c *Client) GetConfig() *config.Config {
	return c.cfg
}
