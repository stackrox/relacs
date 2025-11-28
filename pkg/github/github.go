package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v79/github"
	"github.com/stackrox/rel-acs/pkg/client"
)

// GetAuthenticatedUser returns the login name of the authenticated user
func GetAuthenticatedUser(ctx context.Context, c *client.Client) (string, error) {
	user, _, err := c.GetGitHub().Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("failed to get authenticated user: %w", err)
	}
	if user.Login == nil {
		return "", fmt.Errorf("user login is nil")
	}
	return *user.Login, nil
}

// IsTeamMember checks if a user is a member of a team in an organization
func IsTeamMember(ctx context.Context, c *client.Client, org, teamSlug, username string) (bool, error) {
	membership, _, err := c.GetGitHub().Teams.GetTeamMembershipBySlug(ctx, org, teamSlug, username)
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

// GetCommitForTag retrieves the commit SHA for a given tag
func GetCommitForTag(ctx context.Context, c *client.Client, org, repo, tag string) (string, error) {
	tagRef, _, err := c.GetGitHub().Git.GetRef(ctx, org, repo, "tags/"+tag)
	if err != nil {
		return "", err
	}
	tagSHA := *tagRef.Object.SHA

	tagDetails, _, err := c.GetGitHub().Git.GetTag(ctx, org, repo, tagSHA)
	if err != nil {
		return "", err
	}
	return *tagDetails.Object.SHA, nil
}
