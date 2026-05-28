package konflux

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
)

type ReleasesForCommitOutput struct {
	Commit   string          `json:"commit" yaml:"commit"`
	Releases []ReleaseResult `json:"releases" yaml:"releases"`
	wide     bool
}

func (o ReleasesForCommitOutput) String() string {
	output := fmt.Sprintf("Commit: %s\n", o.Commit)
	output += fmt.Sprintln("Releases:")
	output += renderReleasesTable(o.Releases, o.wide)
	return output
}

type ReleaseResult struct {
	Name              string    `json:"name" yaml:"name"`
	ReleasePlan       string    `json:"releasePlan" yaml:"releasePlan"`
	Target            string    `json:"target" yaml:"target"`
	Snapshot          string    `json:"snapshot" yaml:"snapshot"`
	State             string    `json:"state" yaml:"state"`
	CreationTimestamp time.Time `json:"creationTimestamp" yaml:"creationTimestamp"`
	URL               string    `json:"url" yaml:"url"`
}

var listReleasesForCommitCmd = &cobra.Command{
	Use:   "list-releases-for-commit",
	Short: "List Konflux releases for a given commit SHA",
	Long:  `Retrieve the Konflux releases associated with a specific commit SHA`,
	RunE:  common.RunCommand(runListReleasesForCommit),
}

func init() {
	listReleasesForCommitCmd.Flags().String("commit", "", "Commit SHA (required)")
	listReleasesForCommitCmd.Flags().StringSlice("state", nil, "Filter by release state, e.g. succeeded, failed, progressing (comma-separated)")
	listReleasesForCommitCmd.Flags().Bool("wide", false, "Show additional columns (snapshot, URL)")
	listReleasesForCommitCmd.Flags().Bool("prod", false, "Show only production releases")
	listReleasesForCommitCmd.Flags().Bool("stage", false, "Show only stage releases")

	//nolint:errcheck
	listReleasesForCommitCmd.MarkFlagRequired("commit")

	//nolint:errcheck
	listReleasesForCommitCmd.RegisterFlagCompletionFunc("state", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"succeeded", "failed", "progressing"}, cobra.ShellCompDirectiveNoFileComp
	})
}

func runListReleasesForCommit(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	commit, _ := cmd.Flags().GetString("commit")
	stateFilters, _ := cmd.Flags().GetStringSlice("state")
	wide, _ := cmd.Flags().GetBool("wide")
	prodOnly, _ := cmd.Flags().GetBool("prod")
	stageOnly, _ := cmd.Flags().GetBool("stage")

	if prodOnly && stageOnly {
		return nil, fmt.Errorf("--prod and --stage are mutually exclusive")
	}

	labels := []string{
		fmt.Sprintf("pac.test.appstudio.openshift.io/sha=%s", commit),
	}

	releases, err := c.GetKonflux().GetReleases(labels)
	if err != nil {
		return nil, fmt.Errorf("failed to list releases for commit: %w", err)
	}

	stateSet := make(map[string]bool, len(stateFilters))
	for _, s := range stateFilters {
		stateSet[strings.ToLower(s)] = true
	}

	var results []ReleaseResult
	for _, release := range releases {
		state := release.GetState()
		target := release.GetTarget()

		if len(stateSet) > 0 && !stateSet[state] {
			continue
		}
		if prodOnly && target != "prod" {
			continue
		}
		if stageOnly && target != "stage" {
			continue
		}

		url, _ := release.GetURL()
		var creationTimestamp time.Time
		if release.Metadata.CreationTimestamp != nil {
			creationTimestamp = *release.Metadata.CreationTimestamp
		}
		results = append(results, ReleaseResult{
			Name:              release.Metadata.Name,
			ReleasePlan:       release.Spec.ReleasePlan,
			Target:            target,
			Snapshot:          release.Spec.Snapshot,
			State:             state,
			CreationTimestamp: creationTimestamp,
			URL:               url,
		})
	}

	return ReleasesForCommitOutput{
		Commit:   commit,
		Releases: results,
		wide:     wide,
	}, nil
}

func renderReleasesTable(releases []ReleaseResult, wide bool) string {
	if len(releases) == 0 {
		return "  No releases found."
	}

	var buf bytes.Buffer
	headers := []string{"Name", "Target", "State", "Created"}
	if wide {
		headers = append(headers, "Snapshot", "URL")
	}
	table := tablewriter.NewTable(&buf, tablewriter.WithHeader(headers))

	for _, r := range releases {
		row := []string{
			r.Name,
			strings.ToUpper(r.Target),
			r.State,
			r.CreationTimestamp.Format("Jan 02 15:04 MST"),
		}
		if wide {
			row = append(row, r.Snapshot, r.URL)
		}
		//nolint:errcheck
		table.Append(row)
	}

	//nolint:errcheck
	table.Render()
	return buf.String()
}
