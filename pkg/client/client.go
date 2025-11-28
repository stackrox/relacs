package client

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v79/github"
)

type Client struct {
	gh *github.Client
}

// NewClient creates a new GitHub client. If token is provided, it will be used for authentication.
func NewClient() (*Client, error) {
	var gh *github.Client

	token := os.Getenv("GH_TOKEN")
	if token != "" {
		gh = github.NewClient(nil).WithAuthToken(token)
	} else {
		return nil, fmt.Errorf("no token provided")
	}

	return &Client{
		gh: gh,
	}, nil
}

func (c *Client) GetCommitForTag(context context.Context, org string, repo string, tag string) (string, error) {
	tagRef, _, err := c.gh.Git.GetRef(context, org, repo, "tags/"+tag)
	if err != nil {
		return "", err
	}
	tagSHA := *tagRef.Object.SHA

	tagDetails, _, err := c.gh.Git.GetTag(context, org, repo, tagSHA)
	if err != nil {
		return "", err
	}
	return *tagDetails.Object.SHA, nil
}
