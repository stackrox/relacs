package konflux

import (
	"os/exec"
	"strings"
	"testing"
)

func TestRunCommand_Success(t *testing.T) {
	cmd := exec.Command("echo", "hello")
	output, err := runCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(output) != "hello\n" {
		t.Errorf("expected %q, got %q", "hello\n", string(output))
	}
}

func TestRunCommand_StderrIncludedInError(t *testing.T) {
	cmd := exec.Command("sh", "-c", "echo 'some stderr message' >&2; exit 1")
	_, err := runCommand(cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "some stderr message") {
		t.Errorf("expected error to contain stderr output, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "exit status 1") {
		t.Errorf("expected error to contain exit status, got: %s", errMsg)
	}
}

func TestRunCommand_NoStderr(t *testing.T) {
	cmd := exec.Command("sh", "-c", "exit 1")
	_, err := runCommand(cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errMsg := err.Error()
	if errMsg != "exit status 1" {
		t.Errorf("expected %q, got %q", "exit status 1", errMsg)
	}
}

func TestRunCommand_CommandNotFound(t *testing.T) {
	cmd := exec.Command("nonexistent-command-that-does-not-exist")
	_, err := runCommand(cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunCommand_MultilineStderr(t *testing.T) {
	cmd := exec.Command("sh", "-c", "echo 'line1' >&2; echo 'line2' >&2; exit 42")
	_, err := runCommand(cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "line1") || !strings.Contains(errMsg, "line2") {
		t.Errorf("expected error to contain all stderr lines, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "exit status 42") {
		t.Errorf("expected error to contain exit status 42, got: %s", errMsg)
	}
}

