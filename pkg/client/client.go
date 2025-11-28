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

// NewClient creates a new GitHub client.
// If GH_TOKEN is provided, it will be used for authentication.
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

// GetAuthenticatedUser returns the login name of the authenticated user
func (c *Client) GetAuthenticatedUser(ctx context.Context) (string, error) {
	user, _, err := c.gh.Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("failed to get authenticated user: %w", err)
	}
	if user.Login == nil {
		return "", fmt.Errorf("user login is nil")
	}
	return *user.Login, nil
}

// IsTeamMember checks if a user is a member of a team in an organization
func (c *Client) IsTeamMember(ctx context.Context, org, teamSlug, username string) (bool, error) {
	membership, _, err := c.gh.Teams.GetTeamMembershipBySlug(ctx, org, teamSlug, username)
	if err != nil {
		// 404 means the user is not a member
		if resp, ok := err.(*github.ErrorResponse); ok && resp.Response.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("failed to get team membership: %w", err)
	}

	// Check if the membership is active
	if membership.State != nil && *membership.State == "active" {
		return true, nil
	}

	return false, nil
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
