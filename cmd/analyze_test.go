package cmd

import (
	"encoding/csv"
	"fmt"
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

type AnalyzedCSVWriter struct {
	file   *os.File
	writer *csv.Writer
}

func NewAnalyzedCSVWriter(filename string) (*AnalyzedCSVWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(file)
	return &AnalyzedCSVWriter{
		file:   file,
		writer: writer,
	},
	nil
}

func (w *AnalyzedCSVWriter) WriteAnalyzed(record []string, analysis string) error {
	newRecord := append(record, analysis)
	return w.writer.Write(newRecord)
}

func (w *AnalyzedCSVWriter) Close() {
	w.writer.Flush()
	w.file.Close()
}

func runAnalyze(geminiService GeminiService, inputPath, outputPath string) error {
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

	csvWriter, err := NewAnalyzedCSVWriter(outputPath)
	if err != nil {
		return err
	}
	defer csvWriter.Close()

	if err := csvWriter.WriteAnalyzed(header, "gemini_analysis"); err != nil {
		return err
	}

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}

		if len(record) > 13 && record[13] != "" {
			fmt.Printf("Skipping already analyzed record for PR #%s\n", record[0])
			if err := csvWriter.WriteAnalyzed(record, record[13]); err != nil {
				log.Printf("Error writing to CSV: %v", err)
			}
			continue
		}

		originalTitle := record[6]
		originalBody := record[7]
		geminiTitle := record[11]
		geminiBody := record[12]

		tempDir, err := os.MkdirTemp("", "gemini")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)

		analysis, err := geminiService.AnalyzeCommitMessages(tempDir, originalTitle, originalBody, geminiTitle, geminiBody)
		if err != nil {
			log.Printf("Error analyzing commit messages: %v", err)
			if err := csvWriter.WriteAnalyzed(record, "error"); err != nil {
				log.Printf("Error writing to CSV: %v", err)
			}
			continue
		}

		if err := csvWriter.WriteAnalyzed(record, analysis); err != nil {
			log.Printf("Error writing to CSV: %v", err)
		}
	}
	return nil
}

func TestAnalyzeCmd_OutputFile(t *testing.T) {
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

	headers := []string{"commit_hash", "author_name", "author_email", "commit_date", "subject", "body", "unidiff", "gemini_subject", "gemini_body", "pr_number", "pr_url", "pr_title", "pr_body"}
	writer.Write(headers)
	writer.Write([]string{"abc", "John Doe", "john.doe@example.com", "2025-06-29T20:00:00Z", "feat: new feature", "This is a new feature.", "diff --git a/file.txt b/file.txt\n--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-hello\n+hello world", "feat: new feature", "This is a new feature.", "1", "http://example.com/1", "feat: new feature", "This is a new feature."})
	writer.Flush()

	// Run the analyze command
	outputPath := filepath.Join(tmpDir, "output.csv")
	geminiService := &mockGeminiService{}
	if err := runAnalyze(geminiService, inputPath, outputPath); err != nil {
		t.Fatal(err)
	}

	// Verify that the output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Expected output file to be created, but it was not")
	}

	// Verify the output CSV
	outputFile, err := os.Open(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer outputFile.Close()

	reader := csv.NewReader(outputFile)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(records))
	}

	expected := []string{"abc", "John Doe", "john.doe@example.com", "2025-06-29T20:00:00Z", "feat: new feature", "This is a new feature.", "diff --git a/file.txt b/file.txt\n--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-hello\n+hello world", "feat: new feature", "This is a new feature.", "1", "http://example.com/1", "feat: new feature", "This is a new feature.", "gemini"}
	if !reflect.DeepEqual(records[1], expected) {
		t.Errorf("Expected %v, got %v", expected, records[1])
	}
}
