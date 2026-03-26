package research

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"context"

	"github.com/stackrox/relacs/pkg/research"
)

type SecurityTrackersOutput struct {
	Version string `json:"version" yaml:"version"`
	Trackers []research.SecurityTracker `json:"trackers" yaml:"trackers"`
}

var findSecurityTrackersCmd = &cobra.Command{
	Use:   "find-security-trackers",
	Short: "Find security trackers for a given ACS release version",
	Long:  `Find security trackers for a given ACS release version`,
	RunE:  common.RunCommand(runFindSecurityTrackers),
}

func init() {
	findSecurityTrackersCmd.Flags().String("version", "", "ACS release version (required)")

	findSecurityTrackersCmd.MarkFlagRequired("version")
}

func runFindSecurityTrackers(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	version, _ := cmd.Flags().GetString("version")

	trackers, err := research.FindSecurityTrackers(ctx, c, version)
	if err != nil {
		return nil, fmt.Errorf("failed to find security trackers: %w", err)
	}

	return SecurityTrackersOutput{
		Version: version,
		Trackers: trackers,
	}, nil
}
