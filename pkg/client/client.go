package client

import (
	"fmt"
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
