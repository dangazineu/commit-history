package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/danielgazineu/commit-history/internal"
	"github.com/spf13/cobra"
)

var (
	inputCsvAugment string
	repoAugment     string
)

var augmentCmd = &cobra.Command{
	Use:   "augment",
	Short: "Augment a CSV file with AI-generated commit messages.",
	Long:  `Augment a CSV file with AI-generated commit messages using the Gemini CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Augmenting CSV file: %s\n", inputCsvAugment)
		fmt.Printf("Repository: %s\n", repoAugment)

		geminiService, err := internal.NewGeminiService()
		if err != nil {
			log.Fatal(err)
		}

		outputFileName := strings.Replace(inputCsvAugment, "-fetched.csv", "-augmented.csv", 1)
		if err := runAugment(geminiService, inputCsvAugment, outputFileName); err != nil {
			log.Fatal(err)
		}
	},
}

func runAugment(geminiService internal.GeminiService, inputPath, outputPath string) error {
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

	csvWriter, err := internal.NewAugmentedCSVWriter(outputPath)
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


func init() {
	rootCmd.AddCommand(augmentCmd)
	augmentCmd.Flags().StringVar(&inputCsvAugment, "input-csv", "", "Path to the fetched CSV file")
	augmentCmd.Flags().StringVar(&repoAugment, "repo", "", "GitHub repository URL (e.g., 'owner/repo')")
	augmentCmd.MarkFlagRequired("input-csv")
	augmentCmd.MarkFlagRequired("repo")
}
