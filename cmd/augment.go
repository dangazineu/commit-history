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

type augmentProcessor struct{}

func (p *augmentProcessor) ProcessRecord(record []string, geminiService internal.GeminiService) ([]string, error) {
	prTitle := record[3]
	prBody := record[4]
	unidiff := record[10]

	tempDir, err := os.MkdirTemp("", "gemini")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	title, body, err := geminiService.GenerateCommitMessage(tempDir, prTitle, prBody, unidiff)
	if err != nil {
		log.Printf("Error generating commit message: %v", err)
		return append(record, "error", "error"), nil
	}
	return append(record, title, body), nil
}

func (p *augmentProcessor) GetOutputHeaders() []string {
	return []string{"pr_number", "before_merge_commit_hash", "after_merge_commit_hash", "pr_title", "pr_body", "is_squash_merge", "merge_commit_title", "merge_commit_body", "source_link", "resolved_source_link", "source_link_unidiff", "gemini_proposed_title", "gemini_proposed_body"}
}

func (p *augmentProcessor) ShouldSkip(record []string) bool {
	return len(record) > 11 && record[11] != "" && record[12] != ""
}

func runAugment(geminiService internal.GeminiService, inputPath, outputPath string) error {
	processor := &augmentProcessor{}
	return processCSV(processor, geminiService, inputPath, outputPath)
}


func init() {
	rootCmd.AddCommand(augmentCmd)
	augmentCmd.Flags().StringVar(&inputCsvAugment, "input-csv", "", "Path to the fetched CSV file")
	augmentCmd.Flags().StringVar(&repoAugment, "repo", "", "GitHub repository URL (e.g., 'owner/repo')")
	augmentCmd.MarkFlagRequired("input-csv")
	augmentCmd.MarkFlagRequired("repo")
}
