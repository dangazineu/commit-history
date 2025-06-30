package cmd

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAnalyzeCmd(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a sample input CSV file
	inputPath := filepath.Join(tmpDir, "input.csv")
	file, err := os.Create(inputPath)
	if err != nil {
		t.Fatal(err)
	}

	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"commit_hash", "author_name", "author_email", "commit_date", "subject", "body", "unidiff", "gemini_subject", "gemini_body", "pr_number", "pr_url", "pr_title", "pr_body"}); err != nil {
		t.Fatal(err)
	}
	if err := writer.Write([]string{"abc", "John Doe", "john.doe@example.com", "2025-06-29T20:00:00Z", "feat: new feature", "This is a new feature.", `diff --git a/file.txt b/file.txt
--- a/file.txt
+++ b/file.txt
@@ -1 +1 @@
-hello
+hello world`, "feat: new feature", "This is a new feature.", "1", "http://example.com/1", "feat: new feature", "This is a new feature."}); err != nil {
		t.Fatal(err)
	}
	writer.Flush()
	file.Close()

	// Log input file content
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("could not read input file for logging: %v", err)
	}
	t.Logf("Input CSV content:\n%s", string(inputData))

	// Run the analyze command
	outPath := filepath.Join(tmpDir, "output.csv")
	geminiService := &mockGeminiService{}
	if err := runAnalyze(geminiService, inputPath, outPath); err != nil {
		t.Fatal(err)
	}

	// Log output file content
	outputData, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("could not read output file for logging: %v", err)
	}
	t.Logf("Output CSV content:=%s", string(outputData))

	// Verify the output CSV
	outputFile, err := os.Open(outPath)
	if err != nil {
		t.Fatal(err)
	}
	defer outputFile.Close()

	reader := csv.NewReader(outputFile)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Records read from output file: %#v", records)

	if len(records) != 2 {
		t.Fatalf("Expected 2 records, got %d. Records: %#v", len(records), records)
	}

	expected := []string{"abc", "John Doe", "john.doe@example.com", "2025-06-29T20:00:00Z", "feat: new feature", "This is a new feature.", `diff --git a/file.txt b/file.txt
--- a/file.txt
+++ b/file.txt
@@ -1 +1 @@
-hello
+hello world`, "feat: new feature", "This is a new feature.", "1", "http://example.com/1", "feat: new feature", "This is a new feature.", "gemini"}
	if !reflect.DeepEqual(records[1], expected) {
		t.Errorf("Expected record %#v, got %#v", expected, records[1])
	}
}
