package konflux

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/konflux"
)

type ReleasesForCommitOutput struct {
	Commit   string                  `json:"commit" yaml:"commit"`
	Releases []konflux.ReleaseResult `json:"releases" yaml:"releases"`
	wide     bool
}

func (o ReleasesForCommitOutput) String() string {
	output := fmt.Sprintf("Commit: %s\n", o.Commit)
	output += fmt.Sprintln("Releases:")
	output += renderReleasesTable(o.Releases, o.wide)
	return output
}

var listReleasesForCommitCmd = &cobra.Command{
	Use:   "list-releases-for-commit",
	Short: "List Konflux releases for a given commit SHA",
	Long:  `Retrieve the Konflux releases associated with a specific commit SHA`,
	RunE:  common.RunCommand(runListReleasesForCommit),
}

func init() {
	listReleasesForCommitCmd.Flags().String("commit", "", "Commit SHA (required)")
	listReleasesForCommitCmd.Flags().String("application", "", "Filter by Konflux application name")
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
	application, _ := cmd.Flags().GetString("application")
	stateFilters, _ := cmd.Flags().GetStringSlice("state")
	wide, _ := cmd.Flags().GetBool("wide")
	prodOnly, _ := cmd.Flags().GetBool("prod")
	stageOnly, _ := cmd.Flags().GetBool("stage")

	if prodOnly && stageOnly {
		return nil, fmt.Errorf("--prod and --stage are mutually exclusive")
	}

	results, err := konflux.ListReleasesForCommit(ctx, c, commit, konflux.ListReleasesOptions{
		Application:  application,
		StateFilters: stateFilters,
		ProdOnly:     prodOnly,
		StageOnly:    stageOnly,
	})
	if err != nil {
		return nil, err
	}

	return ReleasesForCommitOutput{
		Commit:   commit,
		Releases: results,
		wide:     wide,
	}, nil
}

func renderReleasesTable(releases []konflux.ReleaseResult, wide bool) string {
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
