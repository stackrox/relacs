package konflux

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCommand_Success(t *testing.T) {
	cmd := exec.Command("echo", "hello")
	output, err := runCommand(cmd)
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(output))
}

func TestRunCommand_StderrIncludedInError(t *testing.T) {
	cmd := exec.Command("sh", "-c", "echo 'some stderr message' >&2; exit 1")
	output, err := runCommand(cmd)
	assert.Empty(t, output)
	assert.ErrorContains(t, err, "some stderr message")
	assert.ErrorContains(t, err, "exit status 1")
}

func TestRunCommand_NoStderr(t *testing.T) {
	cmd := exec.Command("sh", "-c", "exit 1")
	output, err := runCommand(cmd)
	assert.Empty(t, output)
	assert.ErrorContains(t, err, "exit status 1")
}

func TestRunCommand_CommandNotFound(t *testing.T) {
	cmd := exec.Command("nonexistent-command-that-does-not-exist")
	output, err := runCommand(cmd)
	assert.Empty(t, output)
	assert.Error(t, err)
}

func TestRunCommand_MultilineStderr(t *testing.T) {
	cmd := exec.Command("sh", "-c", "echo 'line1' >&2; echo 'line2' >&2; exit 42")
	output, err := runCommand(cmd)
	assert.Empty(t, output)
	require.Error(t, err)
	assert.ErrorContains(t, err, "line1")
	assert.ErrorContains(t, err, "line2")
	assert.ErrorContains(t, err, "exit status 42")
}
