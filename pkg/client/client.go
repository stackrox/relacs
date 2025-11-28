package client

import (
	"context"

	"github.com/google/go-github/v79/github"
)

type Client struct {
	gh *github.Client
}

func NewClient() *Client {
	gh := github.NewClient(nil)
	return &Client{
		gh: gh,
	}
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
