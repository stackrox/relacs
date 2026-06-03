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

type ListReleasesOutput struct {
	Commit   string                  `json:"commit,omitempty" yaml:"commit,omitempty"`
	Tag      string                  `json:"tag,omitempty" yaml:"tag,omitempty"`
	Releases []konflux.ReleaseResult `json:"releases" yaml:"releases"`
	wide     bool
}

func (o ListReleasesOutput) String() string {
	var output string
	if o.Commit != "" {
		output = fmt.Sprintf("Commit: %s\n", o.Commit)
	}
	if o.Tag != "" {
		output = fmt.Sprintf("Tag: %s\n", o.Tag)
	}
	output += fmt.Sprintln("Releases:")
	output += renderReleasesTable(o.Releases, o.wide)
	return output
}

var listReleasesCmd = &cobra.Command{
	Use:   "list-releases",
	Short: "List Konflux releases by commit SHA or tag",
	Long:  `Retrieve the Konflux releases associated with a specific commit SHA or release tag`,
	RunE:  common.RunCommand(runListReleases),
}

func init() {
	listReleasesCmd.Flags().String("commit", "", "Commit SHA")
	listReleasesCmd.Flags().String("tag", "", "Release tag (e.g. 4.10.1)")
	listReleasesCmd.Flags().String("application", "", "Filter by Konflux application name")
	listReleasesCmd.Flags().StringSlice("state", nil, "Filter by release state, e.g. succeeded, failed, progressing (comma-separated)")
	listReleasesCmd.Flags().Bool("wide", false, "Show additional columns (snapshot, URL)")
	listReleasesCmd.Flags().Bool("prod", false, "Show only production releases")
	listReleasesCmd.Flags().Bool("stage", false, "Show only stage releases")

	listReleasesCmd.MarkFlagsOneRequired("commit", "tag")
	listReleasesCmd.MarkFlagsMutuallyExclusive("commit", "tag")

	//nolint:errcheck
	listReleasesCmd.RegisterFlagCompletionFunc("state", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"succeeded", "failed", "progressing"}, cobra.ShellCompDirectiveNoFileComp
	})
}

func runListReleases(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	commit, _ := cmd.Flags().GetString("commit")
	tag, _ := cmd.Flags().GetString("tag")
	application, _ := cmd.Flags().GetString("application")
	stateFilters, _ := cmd.Flags().GetStringSlice("state")
	wide, _ := cmd.Flags().GetBool("wide")
	prodOnly, _ := cmd.Flags().GetBool("prod")
	stageOnly, _ := cmd.Flags().GetBool("stage")

	if prodOnly && stageOnly {
		return nil, fmt.Errorf("--prod and --stage are mutually exclusive")
	}

	results, err := konflux.ListReleases(ctx, c, konflux.ListReleasesOptions{
		Commit:       commit,
		Tag:          tag,
		Application:  application,
		StateFilters: stateFilters,
		ProdOnly:     prodOnly,
		StageOnly:    stageOnly,
	})
	if err != nil {
		return nil, err
	}

	return ListReleasesOutput{
		Commit:   commit,
		Tag:      tag,
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
