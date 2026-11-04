package client

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/google/go-github/v79/github"
	"github.com/stackrox/relacs/pkg/client/jira"
	"github.com/stackrox/relacs/pkg/client/konflux"
)

type Client struct {
	gh     *github.Client
	ghOnce sync.Once
	ghErr  error

	kflx *konflux.Client

	jira     *jira.Client
	jiraOnce sync.Once
	jiraErr  error
}

func NewClient() *Client {
	return &Client{
		kflx: konflux.NewClient(),
	}
}

func (c *Client) GetGitHub() (*github.Client, error) {
	c.ghOnce.Do(func() {
		if c.gh != nil {
			return
		}
		token := os.Getenv("GH_TOKEN")
		if token == "" {
			c.ghErr = fmt.Errorf("GH_TOKEN environment variable is required")
			return
		}
		c.gh = github.NewClient(nil).WithAuthToken(token)
	})
	return c.gh, c.ghErr
}

func (c *Client) GetKonflux() *konflux.Client {
	return c.kflx
}

func (c *Client) GetJira() (*jira.Client, error) {
	c.jiraOnce.Do(func() {
		if c.jira != nil {
			return
		}
		var err error
		c.jira, err = jira.NewClient()
		if err != nil {
			c.jiraErr = err
		}
	})
	return c.jira, c.jiraErr
}

// NewTestClient creates a client with a custom HTTP client for testing.
// The HTTP client can be intercepted by gock for mocking API responses.
func NewTestClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	return &Client{
		gh:   github.NewClient(httpClient).WithAuthToken("test-token"),
		kflx: konflux.NewTestClient(httpClient),
		jira: jira.NewTestClient(httpClient),
	}
}
