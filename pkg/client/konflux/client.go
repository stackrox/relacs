package konflux

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/stackrox/relacs/pkg/config"
)

type Client struct {
	k8sCommand            func(args ...string) (*exec.Cmd, error)
	k8sCommandPlainOutput func(args ...string) (*exec.Cmd, error)
}

func NewClient() (*Client, error) {
	if os.Getenv("KUBECTL_PLUGIN_KA_HOST") == "" {
		return nil, fmt.Errorf("KUBECTL_PLUGIN_KA_HOST environment variable is required")
	}
	return &Client{k8sCommand: getK8sCommandWithJSONOutput, k8sCommandPlainOutput: getK8sCommandWithPlainOutput}, nil
}

func getK8sCommand(withJSONOutput bool, args ...string) (*exec.Cmd, error) {
	defaultArgs := []string{"ka", "-n", config.KonfluxTenant}
	if withJSONOutput {
		defaultArgs = append(defaultArgs, "-o", "json")
	}
	// TODO: Properly handle pagination for get commands
	if args[0] == "get" {
		args = append(args, "--limit", "1000")
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

func getK8sCommandWithJSONOutput(args ...string) (*exec.Cmd, error) {
	return getK8sCommand(true, args...)
}

func getK8sCommandWithPlainOutput(args ...string) (*exec.Cmd, error) {
	return getK8sCommand(false, args...)
}
