package konflux

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/konflux"
)

type PruneSnapshotsOutput struct {
	Version   string   `json:"version" yaml:"version"`
	DryRun    bool     `json:"dry-run" yaml:"dry-run"`
	Snapshots []string `json:"snapshots" yaml:"snapshots"`
}

func (p PruneSnapshotsOutput) String() string {
	output := fmt.Sprintf("Version: %s\n", p.Version)
	output += fmt.Sprintf("Dry Run: %t\n", p.DryRun)
	if p.DryRun {
		output += fmt.Sprintf("Snapshots that would be pruned (%d):\n", len(p.Snapshots))
	} else {
		output += fmt.Sprintf("Snapshots pruned (%d):\n", len(p.Snapshots))
	}
	for _, snapshot := range p.Snapshots {
		output += fmt.Sprintf("  - %s\n", snapshot)
	}
	return output
}

var pruneSnapshotsCmd = &cobra.Command{
	Use:   "prune-snapshots",
	Short: "Remove keep-snapshot annotation from snapshots matching a version",
	Long: `Remove the "test.appstudio.openshift.io/keep-snapshot" annotation from snapshots
that match the specified ACS product version. This allows the snapshots to be
automatically cleaned up by the system.

By default, this command runs in dry-run mode to show which snapshots would be affected.
Use --no-dry-run to actually remove the annotations.`,
	RunE: common.RunCommand(runPruneSnapshots),
}

func init() {
	pruneSnapshotsCmd.Flags().String("version", "", "ACS product version prefix (e.g., 4.10.0) (required)")
	pruneSnapshotsCmd.Flags().Bool("dry-run", true, "Show snapshots that would be pruned without making changes (default: true)")

	//nolint:errcheck
	pruneSnapshotsCmd.MarkFlagRequired("version")
}

func runPruneSnapshots(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	version, _ := cmd.Flags().GetString("version")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	snapshots, err := konflux.PruneSnapshots(ctx, c, version, dryRun)
	if err != nil {
		return nil, fmt.Errorf("failed to prune snapshots: %w", err)
	}

	return PruneSnapshotsOutput{
		Version:   version,
		DryRun:    dryRun,
		Snapshots: snapshots,
	}, nil
}
