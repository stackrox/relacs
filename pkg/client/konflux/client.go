package konflux

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/stackrox/relacs/pkg/config"
)

type Client struct {
	k8sCommand func(args ...string) (*exec.Cmd, error)
}

func getK8sCommand(args ...string) (*exec.Cmd, error) {

	defaultArgs := []string{"ka", "-n", config.KonfluxTenant, "-o", "json"}
	allArgs := append(defaultArgs, args...)
	if _, err := exec.LookPath("kubectl"); err == nil {
		return exec.Command("kubectl", allArgs...), nil
	}
	if _, err := exec.LookPath("oc"); err == nil {
		return exec.Command("oc", allArgs...), nil
	}
	return nil, fmt.Errorf("neither kubectl nor oc is installed")
}

func NewClient() (*Client, error) {
	if os.Getenv("KUBECTL_PLUGIN_KA_HOST") == "" {
		return nil, fmt.Errorf("KUBECTL_PLUGIN_KA_HOST environment variable is required")
	}
	return &Client{k8sCommand: getK8sCommand}, nil
}
