package github

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/rel-acs/pkg/client"
	ghpkg "github.com/stackrox/rel-acs/pkg/github"
)

type TeamMemberOutput struct {
	User         string `json:"user" yaml:"user"`
	Organization string `json:"organization" yaml:"organization"`
	Team         string `json:"team" yaml:"team"`
	IsMember     bool   `json:"is_member" yaml:"is_member"`
}

var isTeamMemberCmd = &cobra.Command{
	Use:   "is-team-member",
	Short: "Check if a user is a member of a GitHub team",
	Long:  `Check whether a specified user (or the authenticated user) is a member of a team in a GitHub organization`,
	RunE:  runIsTeamMember,
}

func init() {
	isTeamMemberCmd.Flags().String("org", "", "GitHub organization (required)")
	isTeamMemberCmd.Flags().String("team", "", "Team slug/name (required)")
	isTeamMemberCmd.Flags().String("user", "", "Username to check (defaults to authenticated user)")

	//nolint:errcheck
	isTeamMemberCmd.MarkFlagRequired("org")
	//nolint:errcheck
	isTeamMemberCmd.MarkFlagRequired("team")
}

func runIsTeamMember(cmd *cobra.Command, args []string) error {
	// Get flag values
	org, _ := cmd.Flags().GetString("org")
	team, _ := cmd.Flags().GetString("team")
	username, _ := cmd.Flags().GetString("user")

	outputFormat, err := getOutputFormat(cmd)
	if err != nil {
		return err
	}

	c, err := client.NewClient()
	if err != nil {
		return err
	}

	// Get username - if not provided, use authenticated user
	if username == "" {
		username, err = ghpkg.GetAuthenticatedUser(context.Background(), c)
		if err != nil {
			return fmt.Errorf("failed to get authenticated user: %w", err)
		}
	}

	// Check team membership
	isMember, err := ghpkg.IsTeamMember(context.Background(), c, org, team, username)
	if err != nil {
		return fmt.Errorf("failed to check team membership: %w", err)
	}

	// Format output
	output := TeamMemberOutput{
		User:         username,
		Organization: org,
		Team:         team,
		IsMember:     isMember,
	}

	return formatTeamMemberOutput(outputFormat, output)
}

func formatTeamMemberOutput(format string, output TeamMemberOutput) error {
	switch format {
	case "json", "yaml":
		return formatOutput(format, output)
	case "text":
		if output.IsMember {
			fmt.Printf("☑️  The user '%s' is a member of the '%s' team on the '%s' organization\n",
				output.User, output.Team, output.Organization)
		} else {
			fmt.Printf("⚠️  The user '%s' is not a member of the '%s' team on the '%s' organization\n",
				output.User, output.Team, output.Organization)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
