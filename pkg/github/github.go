package github

import (
	"context"

	"github.com/google/go-github/v79/github"
	"github.com/stackrox/relacs/pkg/client"
)

// GetCommitForTag retrieves the commit SHA for a given tag
func GetCommitForTag(ctx context.Context, c *client.Client, org, repo, tag string) (github.GitObject, error) {
	tagRef, _, err := c.GetGitHub().Git.GetRef(ctx, org, repo, "tags/"+tag)
	if err != nil {
		return github.GitObject{}, err
	}
	tagSHA := *tagRef.Object.SHA

	tagDetails, _, err := c.GetGitHub().Git.GetTag(ctx, org, repo, tagSHA)
	if err != nil {
		return github.GitObject{}, err
	}
	return *tagDetails.Object, nil
}
