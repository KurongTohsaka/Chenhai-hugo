package main

import (
	"os"
	"os/exec"
	"testing"
)

// TestMainExitCodeOnError verifies that command errors propagate to a non-zero
// process exit code (regression for the discarded cli.Execute() return value,
// which made every error path exit 0 silently).
func TestMainExitCodeOnError(t *testing.T) {
	if os.Getenv("CHENHAI_MAIN_CHILD") == "1" {
		os.Args = []string{"chenhai", "nonexistent-command"}
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMainExitCodeOnError")
	cmd.Env = append(os.Environ(), "CHENHAI_MAIN_CHILD=1")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit for unknown command")
	}
	if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got %v", err)
	}
}
