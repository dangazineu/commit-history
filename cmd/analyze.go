package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/danielgazineu/commit-history/internal"
	"github.com/spf13/cobra"
)

var (
	inputCsvAnalyze string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze the augmented CSV file.",
	Long:  `Analyze the augmented CSV file and compares the original commit message with the AI-generated one.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Analyzing CSV file: %s\n", inputCsvAnalyze)

		geminiService, err := internal.NewGeminiService()
		if err != nil {
			log.Fatal(err)
		}

		outputFileName := strings.Replace(inputCsvAnalyze, "-augmented.csv", "-analyzed.csv", 1)
		if err := runAnalyze(geminiService, inputCsvAnalyze, outputFileName); err != nil {
			log.Fatal(err)
		}
	},
}

type analyzeProcessor struct{}

func (p *analyzeProcessor) ProcessRecord(record []string, geminiService internal.GeminiService) ([]string, error) {
	originalTitle := record[6]
	originalBody := record[7]
	geminiTitle := record[11]
	geminiBody := record[12]

	tempDir, err := os.MkdirTemp("", "gemini")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	analysis, err := geminiService.AnalyzeCommitMessages(tempDir, originalTitle, originalBody, geminiTitle, geminiBody)
	if err != nil {
		log.Printf("Error analyzing commit messages: %v", err)
		return append(record, "error"), nil
	}
	return append(record, analysis), nil
}

func (p *analyzeProcessor) GetOutputHeaders() []string {
	return []string{"commit_hash", "author_name", "author_email", "commit_date", "subject", "body", "unidiff", "gemini_subject", "gemini_body", "pr_number", "pr_url", "pr_title", "pr_body", "gemini_analysis"}
}

func (p *analyzeProcessor) ShouldSkip(record []string) bool {
	return len(record) > 13 && record[13] != ""
}

func runAnalyze(geminiService internal.GeminiService, inputPath, outputPath string) error {
	processor := &analyzeProcessor{}
	return processCSV(processor, geminiService, inputPath, outputPath)
}


func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().StringVar(&inputCsvAnalyze, "input-csv", "", "Path to the augmented CSV file")
	if err := analyzeCmd.MarkFlagRequired("input-csv"); err != nil {
		log.Fatalf("failed to mark flag as required: %v", err)
	}
}
