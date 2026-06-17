package versioncheck

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v79/github"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
)

// fetchMinimumRequiredVersion fetches the MINIMUM_REQUIRED_VERSION file from GitHub
func fetchMinimumRequiredVersion(ctx context.Context, c *client.Client) (string, error) {
	fileContent, _, _, err := c.GetGitHub().Repositories.GetContents(
		ctx,
		config.GitHubOrganization,
		config.RelacsRepository,
		config.RelacsMinimumVersionFile,
		&github.RepositoryContentGetOptions{Ref: config.RelacsBranch},
	)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s from GitHub: %w", config.RelacsMinimumVersionFile, err)
	}

	if fileContent == nil {
		return "", fmt.Errorf("%s file not found", config.RelacsMinimumVersionFile)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", fmt.Errorf("failed to decode file content: %w", err)
	}

	version := strings.TrimSpace(content)
	if version == "" {
		return "", fmt.Errorf("%s file is empty", config.RelacsMinimumVersionFile)
	}

	return version, nil
}
