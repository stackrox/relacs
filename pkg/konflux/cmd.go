package konflux

import (
	"fmt"
	"os/exec"
)

func getK8sCommand(args ...string) (*exec.Cmd, error) {
	if _, err := exec.LookPath("kubectl"); err == nil {
		return exec.Command("kubectl", args...), nil
	}
	if _, err := exec.LookPath("oc"); err == nil {
		return exec.Command("oc", args...), nil
	}
	return nil, fmt.Errorf("neither kubectl nor oc is installed")
}
