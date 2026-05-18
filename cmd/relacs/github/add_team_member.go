package github

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/github"
)

type AddTeamMemberOutput struct {
	github.TeamMembershipInfo `json:",inline" yaml:",inline"`
}

func (o AddTeamMemberOutput) String() string {
	return fmt.Sprintf("User %q added to team %q in org %q (role: %s, state: %s)",
		o.Username, o.Team, o.Org, o.Role, o.State)
}

var addTeamMemberCmd = &cobra.Command{
	Use:   "add-team-member",
	Short: "Add a user to a GitHub team",
	Long: `Add or invite a user to a team in a GitHub organization.

The GH_TOKEN must have admin:org scope (or the fine-grained
organization:members:write permission).

This operation is idempotent: if the user is already a member,
the command returns the current membership state without error.`,
	RunE: common.RunCommand(RunAddTeamMember),
}

func init() {
	addTeamMemberCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization")
	addTeamMemberCmd.Flags().String("team", "", "Team slug (required)")
	addTeamMemberCmd.Flags().String("user", "", "GitHub username (required)")
	addTeamMemberCmd.Flags().String("role", "member", "Role: member or maintainer")

	//nolint:errcheck
	addTeamMemberCmd.MarkFlagRequired("team")
	//nolint:errcheck
	addTeamMemberCmd.MarkFlagRequired("user")
}

// RunAddTeamMember is the main function of this command and exported for testing purposes.
func RunAddTeamMember(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	org, _ := cmd.Flags().GetString("org")
	team, _ := cmd.Flags().GetString("team")
	user, _ := cmd.Flags().GetString("user")
	role, _ := cmd.Flags().GetString("role")

	info, err := github.AddTeamMember(ctx, c, org, team, user, role)
	if err != nil {
		return nil, err
	}

	return AddTeamMemberOutput{TeamMembershipInfo: info}, nil
}
