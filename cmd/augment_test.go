package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type GeminiService interface {
	GenerateCommitMessage(tempDir, prTitle, prBody, unidiff string) (string, string, error)
	AnalyzeCommitMessages(tempDir, originalTitle, originalBody, geminiTitle, geminiBody string) (string, error)
}

type mockGeminiService struct{}

func (m *mockGeminiService) GenerateCommitMessage(tempDir, prTitle, prBody, unidiff string) (string, string, error) {
	return "feat: new feature", "This is a new feature.", nil
}

func (m *mockGeminiService) AnalyzeCommitMessages(tempDir, originalTitle, originalBody, geminiTitle, geminiBody string) (string, error) {
	return "gemini", nil
}

type AugmentedCSVWriter struct {
	file   *os.File
	writer *csv.Writer
}

func NewAugmentedCSVWriter(filename string) (*AugmentedCSVWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(file)
	return &AugmentedCSVWriter{
		file:   file,
		writer: writer,
	},	nil
}

func (w *AugmentedCSVWriter) WriteAugmented(record []string, title, body string) error {
	newRecord := append(record, title, body)
	return w.writer.Write(newRecord)
}

func (w *AugmentedCSVWriter) Close() {
	w.writer.Flush()
	w.file.Close()
}

func runAugment(geminiService GeminiService, inputPath, outputPath string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Read header
	header, err := reader.Read()
	if err != nil {
		return err
	}

	csvWriter, err := NewAugmentedCSVWriter(outputPath)
	if err != nil {
		return err
	}
	defer csvWriter.Close()

	if err := csvWriter.WriteAugmented(header, "gemini_proposed_title", "gemini_proposed_body"); err != nil {
		return err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if len(record) > 11 && record[11] != "" && record[12] != "" {
			fmt.Printf("Skipping already augmented record for PR #%s\n", record[0])
			if err := csvWriter.WriteAugmented(record, record[11], record[12]); err != nil {
				log.Printf("Error writing to CSV: %v", err)
			}
			continue
		}

		prTitle := record[3]
		prBody := record[4]
		unidiff := record[10]

		tempDir, err := os.MkdirTemp("", "gemini")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)

		title, body, err := geminiService.GenerateCommitMessage(tempDir, prTitle, prBody, unidiff)
		if err != nil {
			log.Printf("Error generating commit message: %v", err)
			// Write the original record even if Gemini fails
			if err := csvWriter.WriteAugmented(record, "error", "error"); err != nil {
				log.Printf("Error writing to CSV: %v", err)
			}
			continue
		}

		if err := csvWriter.WriteAugmented(record, title, body); err != nil {
			log.Printf("Error writing to CSV: %v", err)
		}
	}
	return nil
}

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
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"pr_number", "before_merge_commit_hash", "after_merge_commit_hash", "pr_title", "pr_body", "is_squash_merge", "merge_commit_title", "merge_commit_body", "source_link", "resolved_source_link", "source_link_unidiff"}
	writer.Write(headers)
	writer.Write([]string{"1", "abc", "def", "feat: new feature", "This is a new feature.", "true", "feat: new feature", "This is a new feature.", "http://example.com/1", "http://example.com/1", "diff --git a/file.txt b/file.txt\n--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-hello\n+hello world"})
	writer.Flush()

	// Run the augment command
	outPath := filepath.Join(tmpDir, "output.csv")
	geminiService := &mockGeminiService{}
	if err := runAugment(geminiService, inputPath, outPath); err != nil {
		t.Fatal(err)
	}

	// Verify the output CSV
	outFile, err := os.Open(outPath)
	if err != nil {
		t.Fatal(err)
	}
	defer outFile.Close()

	reader := csv.NewReader(outFile)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(records))
	}

	expected := []string{"1", "abc", "def", "feat: new feature", "This is a new feature.", "true", "feat: new feature", "This is a new feature.", "http://example.com/1", "http://example.com/1", "diff --git a/file.txt b/file.txt\n--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-hello\n+hello world", "feat: new feature", "This is a new feature."}
	if !reflect.DeepEqual(records[1], expected) {
		t.Errorf("Expected %v, got %v", expected, records[1])
	}
}

