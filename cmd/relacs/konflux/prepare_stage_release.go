package konflux

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/konflux"
)

var prepareStageReleaseCmd = &cobra.Command{
	Use:   "prepare-stage-release",
	Short: "Prepare a given Konflux stage release",
	Long:  `Prepare a custom resource YAML for a given Konflux stage release from a snapshot`,
	RunE:  common.RunCommand(runPrepareStageRelease),
}

func init() {
	prepareStageReleaseCmd.Flags().String("application-prefix", config.KonfluxApplicationPrefix, "Application prefix (optional)")
	prepareStageReleaseCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	prepareStageReleaseCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	prepareStageReleaseCmd.Flags().String("tag", "", "Git tag (required)")
	prepareStageReleaseCmd.Flags().Int("attempt", 1, "Attempt number (optional)")

	// Default to YAML output for this command since it generates a Kubernetes CR
	prepareStageReleaseCmd.Flags().StringP("output", "o", "yaml", "Output format (text|json|yaml)")

	//nolint:errcheck
	prepareStageReleaseCmd.MarkFlagRequired("tag")
}

func runPrepareStageRelease(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	applicationPrefix, _ := cmd.Flags().GetString("application-prefix")
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")
	attempt, _ := cmd.Flags().GetInt("attempt")

	// Fetch issues and CVEs from Jira
	jiraClient := c.GetJira()
	if jiraClient == nil {
		return nil, fmt.Errorf("jira client not configured. Please set JIRA_EMAIL and JIRA_TOKEN environment variables")
	}

	// Build the default JQL query using the tag as the version
	jql := fmt.Sprintf(config.JiraReleaseDataQuery, tag)

	releaseData, err := jiraClient.FetchReleaseData(jql)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release data from Jira: %w", err)
	}

	snapshot, err := konflux.GetCustomSnapshot(ctx, c, applicationPrefix, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom snapshot: %w", err)
	}
	release, err := konflux.RenderStageRelease(snapshot, releaseData.Issues, releaseData.CVEs, attempt)
	if err != nil {
		return nil, fmt.Errorf("failed to render release: %w", err)
	}

	return release, nil
}
