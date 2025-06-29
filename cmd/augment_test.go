package cmd

import (
	"encoding/csv"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAugmentCmd(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := ioutil.TempDir("", "test")
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

	writer.Write([]string{"pr_number", "before_merge_commit_hash", "after_merge_commit_hash", "pr_title", "pr_body", "is_squash_merge", "merge_commit_title", "merge_commit_body", "source_link", "resolved_source_link", "source_link_unidiff"})
	writer.Write([]string{"1", "abc", "def", "feat: new feature", "This is a new feature.", "true", "feat: new feature", "This is a new feature.", "http://example.com/1", "http://example.com/1", `diff --git a/file.txt b/file.txt
--- a/file.txt
+++ b/file.txt
@@ -1 +1 @@
-hello
+hello world`})
	writer.Flush()
	file.Close()

	// Log input file content
	inputData, err := ioutil.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("could not read input file for logging: %v", err)
	}
	t.Logf("Input CSV content:\n%s", string(inputData))

	// Run the augment command
	outPath := filepath.Join(tmpDir, "output.csv")
	geminiService := &mockGeminiService{}
	if err := runAugment(geminiService, inputPath, outPath); err != nil {
		t.Fatal(err)
	}

	// Log output file content
	outputData, err := ioutil.ReadFile(outPath)
	if err != nil {
		t.Fatalf("could not read output file for logging: %v", err)
	}
	t.Logf("Output CSV content:%s", string(outputData))

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

	expected := []string{"1", "abc", "def", "feat: new feature", "This is a new feature.", "true", "feat: new feature", "This is a new feature.", "http://example.com/1", "http://example.com/1", `diff --git a/file.txt b/file.txt
--- a/file.txt
+++ b/file.txt
@@ -1 +1 @@
-hello
+hello world`, "feat: new feature", "This is a new feature."}
	if !reflect.DeepEqual(records[1], expected) {
		t.Errorf("Expected record %#v, got %#v", expected, records[1])
	}
}
