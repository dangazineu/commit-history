package cmd

import (
	"encoding/csv"
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

func runAnalyze(geminiService internal.GeminiService, inputPath, outputPath string) error {
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

	csvWriter, err := internal.NewAnalyzedCSVWriter(outputPath)
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
		log.Printf("Wrote record: %v", record)
	}
	return nil
}


func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().StringVar(&inputCsvAnalyze, "input-csv", "", "Path to the augmented CSV file")
	analyzeCmd.MarkFlagRequired("input-csv")
}
