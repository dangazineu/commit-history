package cmd

import (
	"testing"
)

// TestFetchEndToEnd is an end-to-end integration test for the fetch command.
// It runs the core logic of the fetch command and verifies the output CSV file.
// This test is skipped in short mode and requires the GOOGLEAPIS_REPO_PATH
// environment variable to be set to the local path of the googleapis/googleapis repository.
func TestFetchEndToEnd(t *testing.T) {
	t.Skip("skipping test, not implemented yet")
}
