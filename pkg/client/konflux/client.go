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
	cfg                   *config.Config
	k8sCommand            func(args ...string) (*exec.Cmd, error)
	k8sCommandPlainOutput func(args ...string) (*exec.Cmd, error)
}

func NewClient(cfg *config.Config) (*Client, error) {
	if os.Getenv("KUBECTL_PLUGIN_KA_HOST") == "" {
		return nil, fmt.Errorf("KUBECTL_PLUGIN_KA_HOST environment variable is required")
	}
	client := &Client{cfg: cfg}
	client.k8sCommand = func(args ...string) (*exec.Cmd, error) {
		return client.getK8sCommandWithJSONOutput(args...)
	}
	client.k8sCommandPlainOutput = func(args ...string) (*exec.Cmd, error) {
		return client.getK8sCommandWithPlainOutput(args...)
	}
	return client, nil
}

func (c *Client) getK8sCommand(useKubeArchive bool, withJSONOutput bool, args ...string) (*exec.Cmd, error) {
	var defaultArgs []string
	if useKubeArchive {
		defaultArgs = []string{"ka", "-n", c.cfg.KonfluxTenant}
		// TODO: Properly handle pagination for get commands
		if args[0] == "get" {
			args = append(args, "--limit", "1000")
		}
	} else {
		defaultArgs = []string{"-n", c.cfg.KonfluxTenant}
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

func (c *Client) getK8sCommandWithJSONOutput(args ...string) (*exec.Cmd, error) {
	return c.getK8sCommand(true, true, args...)
}

func (c *Client) getK8sCommandWithPlainOutput(args ...string) (*exec.Cmd, error) {
	return c.getK8sCommand(true, false, args...)
}

func (c *Client) getDirectK8sCommandWithJSONOutput(args ...string) (*exec.Cmd, error) {
	return c.getK8sCommand(false, true, args...)
}

func (c *Client) getDirectK8sCommandWithPlainOutput(args ...string) (*exec.Cmd, error) {
	return c.getK8sCommand(false, false, args...)
}
