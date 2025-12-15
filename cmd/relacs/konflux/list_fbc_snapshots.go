package konflux

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	kflx "github.com/stackrox/relacs/pkg/client/konflux"
	"github.com/stackrox/relacs/pkg/konflux"
)

type SnapshotsOutput struct {
	Commit    string          `json:"commit" yaml:"commit"`
	Branch    string          `json:"branch" yaml:"branch"`
	Snapshots []kflx.Snapshot `json:"snapshots" yaml:"snapshots"`
}

func (s SnapshotsOutput) String() string {
	output := fmt.Sprintf("Commit: %s\n", s.Commit)
	output += fmt.Sprintf("Branch: %s\n", s.Branch)
	output += fmt.Sprintln("Snapshots:")
	for _, snapshot := range s.Snapshots {
		output += fmt.Sprintf("  %s: %s (created at: %s)\n", snapshot.Spec.Application, snapshot.Metadata.Name, snapshot.Metadata.CreationTimestamp.Format(time.RFC3339))
	}
	return output
}

var listFbcSnapshotsCmd = &cobra.Command{
	Use:   "list-fbc-snapshots",
	Short: "List the FBC snapshots for a given Git commit and branch",
	Long:  `Retrieve the FBC snapshots associated with a specific Git commit and branch`,
	RunE:  common.RunCommand(runListFbcSnapshots),
}

func init() {
	listFbcSnapshotsCmd.Flags().String("commit", "", "Git commit (required)")
	listFbcSnapshotsCmd.Flags().String("branch", "", "Git branch (required)")

	//nolint:errcheck
	listFbcSnapshotsCmd.MarkFlagRequired("commit")
	//nolint:errcheck
	listFbcSnapshotsCmd.MarkFlagRequired("branch")
}

func runListFbcSnapshots(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	// Get flag values
	commit, _ := cmd.Flags().GetString("commit")
	branch, _ := cmd.Flags().GetString("branch")

	snapshots, err := konflux.ListFbcSnapshots(ctx, c, commit, branch)
	if err != nil {
		return nil, fmt.Errorf("failed to get FBC snapshots: %w", err)
	}

	return SnapshotsOutput{
		Commit:    commit,
		Branch:    branch,
		Snapshots: snapshots,
	}, nil
}
