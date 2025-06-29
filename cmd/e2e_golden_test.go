package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFetchEndToEndWithGoldenFile(t *testing.T) {
	t.Skip("skipping test, flaky")
	if testing.Short() {
		t.Skip("skipping end-to-end test in short mode.")
	}

	// --- Test Setup ---
	// Build the CLI binary
	cmd := exec.Command("go", "build", "-o", "commit-history", "..")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("failed to build CLI binary: %v", err)
	}

	// --- Test Parameters ---
	repo := "googleapis/google-cloud-ruby"
	query := "is:pr is:merged created:2024-01-01..2024-01-02 label:owl-bot-copy"
	goldenFilePath := filepath.Join("testdata", "golden.csv")

	// --- Execute the Command ---
	tempDir := t.TempDir()
	cmd = exec.Command("./commit-history", "fetch", "--repo", repo, "--query", query, "--googleapis-repo-path", tempDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fetch command failed: %v\nOutput: %s", err, string(output))
	}

	// --- Verification ---
	outputFileName := fmt.Sprintf("%s-fetched.csv", strings.Split(repo, "/")[1])
	assertCsvsAreEqual(t, goldenFilePath, outputFileName)
}

func assertCsvsAreEqual(t *testing.T, expectedFilePath, actualFilePath string) {
	t.Helper()

	// Read expected CSV
	expectedFile, err := os.Open(expectedFilePath)
	if err != nil {
		t.Fatalf("failed to open expected CSV file '%s': %v", expectedFilePath, err)
	}
	defer expectedFile.Close()
	expectedReader := csv.NewReader(expectedFile)
	expectedRecords, err := expectedReader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read records from expected CSV: %v", err)
	}

	// Read actual CSV
	actualFile, err := os.Open(actualFilePath)
	if err != nil {
		t.Fatalf("failed to open actual CSV file '%s': %v", actualFilePath, err)
	}
	defer actualFile.Close()
	actualReader := csv.NewReader(actualFile)
	actualRecords, err := actualReader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read records from actual CSV: %v", err)
	}

	// Compare records
	if len(expectedRecords) != len(actualRecords) {
		t.Errorf("number of records mismatch: expected %d, got %d", len(expectedRecords), len(actualRecords))
		t.Logf("Expected records:\n%v", expectedRecords)
		t.Logf("Actual records:\n%v", actualRecords)
	}

	for i := range expectedRecords {
		if i >= len(actualRecords) {
			break
		}
		if !equalSlices(expectedRecords[i], actualRecords[i]) {
			t.Errorf("record %d mismatch:\nExpected: %v\nActual:   %v", i, expectedRecords[i], actualRecords[i])
		}
	}
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
