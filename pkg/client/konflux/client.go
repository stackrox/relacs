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
