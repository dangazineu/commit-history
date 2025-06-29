package cmd

import (
	"encoding/csv"
	"os"
)

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
	}, nil
}

func (w *AugmentedCSVWriter) WriteAugmented(record []string, title, body string) error {
	newRecord := append(record, title, body)
	return w.writer.Write(newRecord)
}

func (w *AugmentedCSVWriter) Close() {
	w.writer.Flush()
	w.file.Close()
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
	}, nil
}

func (w *AnalyzedCSVWriter) WriteAnalyzed(record []string, analysis string) error {
	newRecord := append(record, analysis)
	return w.writer.Write(newRecord)
}

func (w *AnalyzedCSVWriter) Close() {
	w.writer.Flush()
	w.file.Close()
}
