package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFetchEndToEnd is an end-to-end integration test for the fetch command.
// It runs the core logic of the fetch command and verifies the output CSV file.
// This test is skipped in short mode and requires the GOOGLEAPIS_REPO_PATH
// environment variable to be set to the local path of the googleapis/googleapis repository.
func TestFetchEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test in short mode.")
	}

	googleapisRepoPath := os.Getenv("GOOGLEAPIS_REPO_PATH")
	if googleapisRepoPath == "" {
		t.Skip("skipping end-to-end test, GOOGLEAPIS_REPO_PATH not set")
	}

	// --- Test Parameters ---
	repo := "googleapis/google-cloud-ruby"
	query := "is:pr is:merged created:2024-01-01..2024-01-02 label:owl-bot-copy"
	expectedRows := 4
	expectedPrNumber := "23726" // We'll check this specific PR

	// --- Test Setup ---
	// Create a temporary directory for the test output
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	// Change working directory to the temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory to temp dir: %v", err)
	}
	// Ensure we change back to the original directory after the test
	defer os.Chdir(originalWd)

	// --- Execute the Command's Logic ---
	if err := runFetch(repo, query, googleapisRepoPath); err != nil {
		t.Fatalf("runFetch() returned an error: %v", err)
	}

	// --- Verification ---
	outputFileName := fmt.Sprintf("%s-fetched.csv", strings.Split(repo, "/")[1])
	outputFilePath := filepath.Join(tempDir, outputFileName)

	file, err := os.Open(outputFilePath)
	if err != nil {
		t.Fatalf("failed to open output file '%s': %v", outputFilePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Verify header
	header, err := reader.Read()
	if err != nil {
		t.Fatalf("failed to read header from CSV: %v", err)
	}
	expectedHeader := []string{
		"pr_number", "before_merge_commit_hash", "after_merge_commit_hash", "pr_title", "pr_body",
		"is_squash_merge", "merge_commit_title", "merge_commit_body", "source_link",
		"resolved_source_link", "source_link_unidiff",
	}
	if !equalSlices(header, expectedHeader) {
		t.Errorf("CSV header is incorrect.\nGot:  %v\nWant: %v", header, expectedHeader)
	}

	// Verify rows
	var records [][]string
	var specificRecord []string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to read record from CSV: %v", err)
		}
		records = append(records, record)
		if record[0] == expectedPrNumber {
			specificRecord = record
		}
	}

	if len(records) != expectedRows {
		t.Errorf("expected %d rows in CSV, but got %d", expectedRows, len(records))
	}

	if specificRecord == nil {
		t.Fatalf("expected to find record for PR #%s, but it was not found", expectedPrNumber)
	}

	// Verify content of a specific row
	if specificRecord[5] != "false" {
		t.Errorf("expected 'is_squash_merge' to be 'false' for PR #%s, but got '%s'", expectedPrNumber, specificRecord[5])
	}
	if specificRecord[8] == "" {
		t.Errorf("expected 'source_link' to be populated for PR #%s, but it was empty", expectedPrNumber)
	}
	if specificRecord[10] == "" {
		t.Errorf("expected 'source_link_unidiff' to be populated for PR #%s, but it was empty", expectedPrNumber)
	}

	t.Logf("Successfully verified %d rows in %s", len(records), outputFileName)
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
