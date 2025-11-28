package selftest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stackrox/rel-acs/pkg/client"
	ghpkg "github.com/stackrox/rel-acs/pkg/github"
	"github.com/stackrox/rel-acs/pkg/konflux"
	"gopkg.in/yaml.v3"
)

// CheckResult represents the result of a single check
type CheckResult struct {
	Name        string `json:"name" yaml:"name"`
	Status      string `json:"status" yaml:"status"` // "pass", "fail", "warning"
	Message     string `json:"message" yaml:"message"`
	Remediation string `json:"remediation,omitempty" yaml:"remediation,omitempty"`
}

// SetupCheckOutput represents the output of the setup check
type SetupCheckOutput struct {
	Checks      []CheckResult `json:"checks" yaml:"checks"`
	OverallPass bool          `json:"overall_pass" yaml:"overall_pass"`
}

var checkSetupCmd = &cobra.Command{
	Use:   "check-setup",
	Short: "Check if the current user has the necessary setup to prepare a release",
	Long: `Verifies the following prerequisites:
- GitHub token (GH_TOKEN) is configured
- Connected to Konflux cluster (oc/kubectl)
- KubeArchive plugin and configuration
- GitHub teams membership
- Basic tool availability`,
	RunE: runCheckSetup,
}

func runCheckSetup(cmd *cobra.Command, args []string) error {
	format, err := getOutputFormat(cmd)
	if err != nil {
		return err
	}

	c, err := client.NewClient()
	if err != nil {
		return err
	}

	checks := []CheckResult{}

	// Check 1: GitHub Token
	checks = append(checks, checkGitHubToken(c))

	// Check 1.1: GitHub team membership (optional)
	githubOrg := "stackrox"
	githubTeams := []string{"release-mgmt", "release-mgmt-no-notifications"}
	for _, team := range githubTeams {
		checks = append(checks, checkGitHubTeamMembership(c, githubOrg, team))
	}

	// Check 2: Konflux cluster
	// Check 2.1: kubectl/oc availability
	checks = append(checks, checkKubectlInstalled())

	// Check 2.2: Konflux cluster connection
	checks = append(checks, checkKonfluxConnection())

	// Check 2.3: Konflux maintainer role
	checks = append(checks, checkKonfluxMaintainerRole())

	// Check 2.4: KubeArchive configuration
	checks = append(checks, checkKubeArchive())

	// Determine overall status
	overallPass := true
	for _, check := range checks {
		if check.Status == "fail" {
			overallPass = false
			break
		}
	}

	output := SetupCheckOutput{
		Checks:      checks,
		OverallPass: overallPass,
	}

	return formatOutput(format, output, overallPass)
}

func checkGitHubToken(c *client.Client) CheckResult {
	_, err := ghpkg.GetAuthenticatedUser(context.Background(), c)
	if err != nil {
		return CheckResult{
			Name:        "GitHub Token",
			Status:      "fail",
			Message:     fmt.Sprintf("Failed to get authenticated user: %v", err),
			Remediation: "Verify your GH_TOKEN is valid",
		}
	}

	return CheckResult{
		Name:    "GitHub Token",
		Status:  "pass",
		Message: "GH_TOKEN is configured",
	}
}

func checkKubectlInstalled() CheckResult {
	// Check for kubectl first
	if _, err := exec.LookPath("kubectl"); err == nil {
		return CheckResult{
			Name:    "kubectl Tool",
			Status:  "pass",
			Message: "kubectl is installed and available",
		}
	}

	// Check for oc as fallback
	if _, err := exec.LookPath("oc"); err == nil {
		return CheckResult{
			Name:    "kubectl Tool",
			Status:  "pass",
			Message: "oc is installed and available (kubectl alternative)",
		}
	}

	return CheckResult{
		Name:        "kubectl Tool",
		Status:      "fail",
		Message:     "Neither kubectl nor oc is installed",
		Remediation: "Install kubectl from https://kubernetes.io/docs/tasks/tools/ or oc from https://docs.openshift.com/container-platform/latest/cli_reference/openshift_cli/getting-started-cli.html",
	}
}

func checkKonfluxConnection() CheckResult {
	// Check if we're connected to the right cluster by looking for stone-prd in context
	currentContext, err := konflux.GetCurrentContext()
	if err != nil {
		return CheckResult{
			Name:        "Konflux Cluster Connection",
			Status:      "fail",
			Message:     fmt.Sprintf("Failed to get current context: %v", err),
			Remediation: "Verify you're connected to the correct Konflux tenant: oc login --web https://api.stone-prd-rh01.pg1f.p1.openshiftapps.com:6443/",
		}
	}
	if !strings.Contains(currentContext, "stone-prd-rh01") {
		return CheckResult{
			Name:        "Konflux Cluster Connection",
			Status:      "warning",
			Message:     fmt.Sprintf("Connected to a different cluster than the expected one: %s", currentContext),
			Remediation: "Verify you're connected to the correct Konflux tenant: oc login --web https://api.stone-prd-rh01.pg1f.p1.openshiftapps.com:6443/",
		}
	}
	return CheckResult{
		Name:    "Konflux Cluster Connection",
		Status:  "pass",
		Message: fmt.Sprintf("Connected to Konflux cluster: %s", currentContext),
	}
}

func checkKonfluxMaintainerRole() CheckResult {
	// Check if the user has the konflux-maintainer role
	// based on the rolebindings in the rh-acs-tenant namespace
	// use the last part of the context name to find the user name,
	// then check if there is a rolebinding by name for konflux-maintainer-<username>-user-actions or konflux-admin-<username>-user-actions
	// if either exists, return pass, otherwise return fail
	// Remediation: Request the Konflux Maintainer role from #forum-acs-eng-release

	username, err := konflux.GetCurrentUsername()
	if err != nil {
		return CheckResult{
			Name:        "Konflux Maintainer Role",
			Status:      "fail",
			Message:     fmt.Sprintf("Failed to get current username: %v", err),
			Remediation: "Verify you're connected to the correct Konflux tenant: oc login --web https://api.stone-prd-rh01.pg1f.p1.openshiftapps.com:6443/",
		}
	}

	hasMaintainerRole, _ := konflux.HasKonfluxMaintainerRole(username, "rh-acs-tenant")
	hasAdminRole, _ := konflux.HasKonfluxAdminRole(username, "rh-acs-tenant")
	if !hasMaintainerRole && !hasAdminRole {
		return CheckResult{
			Name:        "Konflux Maintainer Role",
			Status:      "fail",
			Message:     fmt.Sprintf("User %s does not have the Konflux Maintainer role", username),
			Remediation: "Request the Konflux Maintainer role from #forum-acs-eng-release",
		}
	}

	return CheckResult{
		Name:    "Konflux Maintainer Role",
		Status:  "pass",
		Message: fmt.Sprintf("User %s has the Konflux Maintainer or Admin role", username),
	}
}

func checkKubeArchive() CheckResult {
	// Check if kubectl-ka plugin is installed
	cmd := exec.Command("kubectl", "ka", "--help")
	if err := cmd.Run(); err != nil {
		return CheckResult{
			Name:        "KubeArchive Plugin",
			Status:      "fail",
			Message:     "kubectl-ka plugin is not installed",
			Remediation: "Install the KubeArchive kubectl plugin from https://kubearchive.github.io/kubearchive/main/cli/installation.html",
		}
	}

	// Check if KUBECTL_PLUGIN_KA_HOST is set
	kaHost := os.Getenv("KUBECTL_PLUGIN_KA_HOST")
	if kaHost == "" {
		return CheckResult{
			Name:        "KubeArchive Configuration",
			Status:      "fail",
			Message:     "KUBECTL_PLUGIN_KA_HOST environment variable is not set",
			Remediation: "Set KUBECTL_PLUGIN_KA_HOST to the KubeArchive API server, e.g.: export KUBECTL_PLUGIN_KA_HOST=https://kubearchive-api-server-product-kubearchive.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com",
		}
	}

	return CheckResult{
		Name:    "KubeArchive Configuration",
		Status:  "pass",
		Message: fmt.Sprintf("KubeArchive is configured: %s", kaHost),
	}
}

func checkGitHubTeamMembership(c *client.Client, org, team string) CheckResult {
	// Get authenticated user
	username, err := ghpkg.GetAuthenticatedUser(context.Background(), c)
	if err != nil {
		return CheckResult{
			Name:        fmt.Sprintf("GitHub Team Membership (%s/%s)", org, team),
			Status:      "fail",
			Message:     fmt.Sprintf("Failed to get authenticated user: %v", err),
			Remediation: "Verify your GH_TOKEN is valid",
		}
	}

	// Check team membership
	isMember, err := ghpkg.IsTeamMember(context.Background(), c, org, team, username)
	if err != nil {
		return CheckResult{
			Name:        fmt.Sprintf("GitHub Team Membership (%s/%s)", org, team),
			Status:      "fail",
			Message:     fmt.Sprintf("Failed to check team membership: %v", err),
			Remediation: "Verify your GitHub token has the necessary permissions and the team exists",
		}
	}

	if !isMember {
		return CheckResult{
			Name:        fmt.Sprintf("GitHub Team Membership (%s/%s)", org, team),
			Status:      "fail",
			Message:     fmt.Sprintf("User %s is not a member of team %s/%s", username, org, team),
			Remediation: fmt.Sprintf("Request access to the %s/%s team from @acs-eng-staff", org, team),
		}
	}

	return CheckResult{
		Name:    fmt.Sprintf("GitHub Team Membership (%s/%s)", org, team),
		Status:  "pass",
		Message: fmt.Sprintf("User %s is a member of team %s/%s", username, org, team),
	}
}

func formatOutput(format string, output SetupCheckOutput, overallPass bool) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		encoder.SetIndent(2)
		return encoder.Encode(output)
	case "text":
		fmt.Println("Release Preparation Setup Check")
		fmt.Println("================================")
		fmt.Println()

		for _, check := range output.Checks {
			var statusSymbol string
			switch check.Status {
			case "pass":
				statusSymbol = "✓"
			case "fail":
				statusSymbol = "✗"
			case "warning":
				statusSymbol = "⚠"
			default:
				statusSymbol = "?"
			}

			fmt.Printf("%s %s: %s\n", statusSymbol, check.Name, check.Message)
			if check.Remediation != "" {
				fmt.Printf("  → %s\n", check.Remediation)
			}
		}

		fmt.Println()
		if overallPass {
			fmt.Println("✓ All critical checks passed!")
		} else {
			fmt.Println("✗ Some checks failed. Please address the issues above before preparing a release.")
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
