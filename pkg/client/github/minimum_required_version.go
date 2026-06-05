package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v79/github"
	"github.com/stackrox/relacs/pkg/client"
)

const (
	relacsOrg  = "stackrox"
	relacsRepo = "relacs"
	// TODO: Change to "main" before merging
	relacsBranch = "tm/minimum-required-relacs-version"
	versionFile  = "MINIMUM_REQUIRED_VERSION"
)

// FetchMinimumRequiredVersion fetches the MINIMUM_REQUIRED_VERSION file from GitHub
func FetchMinimumRequiredVersion(ctx context.Context, c *client.Client) (string, error) {
	fileContent, _, _, err := c.GetGitHub().Repositories.GetContents(
		ctx,
		relacsOrg,
		relacsRepo,
		versionFile,
		&github.RepositoryContentGetOptions{Ref: relacsBranch},
	)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s from GitHub: %w", versionFile, err)
	}

	if fileContent == nil {
		return "", fmt.Errorf("%s file not found", versionFile)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", fmt.Errorf("failed to decode file content: %w", err)
	}

	version := strings.TrimSpace(content)
	if version == "" {
		return "", fmt.Errorf("%s file is empty", versionFile)
	}

	return version, nil
}
