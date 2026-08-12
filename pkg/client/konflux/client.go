package konflux

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/stackrox/relacs/pkg/config"
)

type Client struct {
	k8sCommand            func(args ...string) (*exec.Cmd, error)
	k8sCommandPlainOutput func(args ...string) (*exec.Cmd, error)
	// directK8sCommandJSON is used for direct kubectl commands (not kubearchive)
	directK8sCommandJSON func(args ...string) (*exec.Cmd, error)
}

func NewClient() (*Client, error) {
	if os.Getenv("KUBECTL_PLUGIN_KA_HOST") == "" {
		return nil, fmt.Errorf("KUBECTL_PLUGIN_KA_HOST environment variable is required")
	}
	return &Client{
		k8sCommand:            getK8sCommandWithJSONOutput,
		k8sCommandPlainOutput: getK8sCommandWithPlainOutput,
		directK8sCommandJSON:  getDirectK8sCommandWithJSONOutput,
	}, nil
}

func getK8sCommand(useKubeArchive bool, withJSONOutput bool, args ...string) (*exec.Cmd, error) {
	var defaultArgs []string
	if useKubeArchive {
		defaultArgs = []string{"ka", "-n", config.KonfluxTenant}
		// TODO: Properly handle pagination for get commands
		if args[0] == "get" {
			args = append(args, "--limit", "1000")
		}
	} else {
		defaultArgs = []string{"-n", config.KonfluxTenant}
	}

	if withJSONOutput {
		defaultArgs = append(defaultArgs, "-o", "json")
	}

	allArgs := append(defaultArgs, args...)
	if _, err := exec.LookPath("kubectl"); err == nil {
		return exec.Command("kubectl", allArgs...), nil
	}
	if _, err := exec.LookPath("oc"); err == nil {
		return exec.Command("oc", allArgs...), nil
	}
	return nil, fmt.Errorf("neither kubectl nor oc is installed")
}

func runCommand(cmd *exec.Cmd) ([]byte, error) {
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return output, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return output, err
	}
	return output, nil
}

func getK8sCommandWithJSONOutput(args ...string) (*exec.Cmd, error) {
	return getK8sCommand(true, true, args...)
}

func getK8sCommandWithPlainOutput(args ...string) (*exec.Cmd, error) {
	return getK8sCommand(true, false, args...)
}

func getDirectK8sCommandWithJSONOutput(args ...string) (*exec.Cmd, error) {
	return getK8sCommand(false, true, args...)
}

func getDirectK8sCommandWithPlainOutput(args ...string) (*exec.Cmd, error) {
	return getK8sCommand(false, false, args...)
}

// NewTestClient creates a test Konflux client with mock application data.
// The mockApplications parameter is a JSON string representing kubectl application output.
func NewTestClient(mockOutput string) *Client {
	mockCmd := func(args ...string) (*exec.Cmd, error) {
		// Return a mock command that outputs the provided mock data
		// We use echo to simulate kubectl output
		return exec.Command("echo", mockOutput), nil
	}

	return &Client{
		k8sCommand:            mockCmd,
		k8sCommandPlainOutput: mockCmd,
		directK8sCommandJSON:  mockCmd,
	}
}
