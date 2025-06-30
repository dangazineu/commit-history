package main_test

import (
	"flag"
	"os/exec"
	"testing"
)

var runIntegration = flag.Bool("integration", false, "run integration tests")

func TestDependencies(t *testing.T) {
	if !*runIntegration {
		t.Skip("skipping dependency check in non-integration mode.")
	}
	cmd := exec.Command("go", "mod", "tidy", "-diff")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go mod tidy -diff failed: %v\n%s", err, output)
	}
	if len(output) > 0 {
		t.Errorf("go mod tidy -diff produced output, meaning the go.mod and go.sum files are not up-to-date:\n%s", output)
	}
}
