package main_test

import (
	"errors"
	"os/exec"
	"testing"
)

func rungo(t *testing.T, args ...string) {
	t.Helper()

	cmd := exec.Command("go", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		if ee := (*exec.ExitError)(nil); errors.As(err, &ee) && len(ee.Stderr) > 0 {
			t.Fatalf("%v: %v\n%s", cmd, err, ee.Stderr)
		}
		t.Fatalf("%v: %v\n%s", cmd, err, output)
	}
}

func TestVulnerabilityCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping vulnerability check in short mode.")
	}
	rungo(t, "run", "golang.org/x/vuln/cmd/govulncheck@latest", "./...")
}

func TestStaticAnalysis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping static analysis in short mode.")
	}
	rungo(t, "run", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest", "run")
}
