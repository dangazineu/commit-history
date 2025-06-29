package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/danielgazineu/commit-history/internal"
	"github.com/google/go-github/v62/github"
	"github.com/spf13/cobra"
)

var (
	repo               string
	query              string
	googleapisRepoPath string
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch pull requests from a GitHub repository.",
	Long:  `Fetch pull requests from a GitHub repository based on a query and saves them to a CSV file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runFetch(repo, query, googleapisRepoPath); err != nil {
			log.Fatal(err)
		}
	},
}

func runFetch(repo, query, googleapisRepoPath string) error {
	fmt.Printf("Fetching pull requests for repository: %s\n", repo)
	fmt.Printf("Query: %s\n", query)

	githubService, err := internal.NewGitHubService(repo)
	if err != nil {
		return err
	}

	prs, err := githubService.GetPullRequests(query)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d merged pull requests.\n", len(prs))

	outputFileName := fmt.Sprintf("%s-fetched.csv", strings.Split(repo, "/")[1])
	processedPRs := make(map[int]bool)
	if _, err := os.Stat(outputFileName); err == nil {
		fmt.Println("Output file already exists, resuming from last run.")
		file, err := os.Open(outputFileName)
		if err != nil {
			return err
		}
		defer file.Close()

		reader := csv.NewReader(file)
		// Skip header
		if _, err := reader.Read(); err != nil {
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
			prNumber, err := strconv.Atoi(record[0])
			if err != nil {
				return err
			}
			processedPRs[prNumber] = true
		}
	}

	csvWriter, err := internal.NewCSVWriter(outputFileName)
	if err != nil {
		return err
	}
	defer csvWriter.Close()

	re := regexp.MustCompile(`Source-Link: (?:https://github.com/)?googleapis/googleapis(?:/commit/|@)([a-f0-9]{7,40})`)

	gitService, err := internal.NewGitService(googleapisRepoPath)
	if err != nil {
		return err
	}

	tempGoogleapisDir, err := os.MkdirTemp("", "googleapis-copy")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempGoogleapisDir)

	fmt.Println("Creating a temporary local copy of googleapis/googleapis...")
	if err := gitService.LocalClone(tempGoogleapisDir); err != nil {
		return err
	}
	fmt.Println("Done creating copy.")
	gitService.SetRepoDir(tempGoogleapisDir)

	for _, pr := range prs {
		if processedPRs[*pr.Number] {
			fmt.Printf("Skipping already processed PR #%d\n", *pr.Number)
			continue
		}

		if err := processPullRequest(pr, githubService, csvWriter, re, gitService); err != nil {
			log.Printf("Error processing PR #%d: %v", *pr.Number, err)
		}
	}

	fmt.Println("Done.")
	return nil
}

func processPullRequest(pr *github.PullRequest, githubService *internal.GitHubService, csvWriter *internal.CSVWriter, re *regexp.Regexp, gitService *internal.GitService) error {
	isSquash, err := githubService.IsSquashMerge(pr)
	if err != nil {
		return fmt.Errorf("error checking if PR was squash merged: %w", err)
	}

	mergeCommit, err := githubService.GetCommit(*pr.MergeCommitSHA)
	if err != nil {
		return fmt.Errorf("error getting merge commit: %w", err)
	}

	var sourceLink, resolvedSourceLink, unidiff string
	if pr.Body != nil {
		matches := re.FindStringSubmatch(*pr.Body)
		if len(matches) > 1 {
			sourceLink = matches[0]
			commitSHA := matches[1]
			resolvedSourceLink = fmt.Sprintf("https://github.com/googleapis/googleapis/commit/%s", commitSHA)
			unidiff, err = gitService.GetUnidiff(commitSHA)
			if err != nil {
				return fmt.Errorf("error getting unidiff for commit %s: %w", commitSHA, err)
			}
		}
	}

	message := *mergeCommit.Message
	parts := strings.SplitN(message, "\n\n", 2)
	title := parts[0]
	body := ""
	if len(parts) > 1 {
		body = parts[1]
	}

	return csvWriter.Write(pr, isSquash, title, body, sourceLink, resolvedSourceLink, unidiff)
}

func init() {
	rootCmd.AddCommand(fetchCmd)
	fetchCmd.Flags().StringVar(&repo, "repo", "", "GitHub repository URL (e.g., 'owner/repo')")
	fetchCmd.Flags().StringVar(&query, "query", "", "GitHub search query for pull requests")
	fetchCmd.Flags().StringVar(&googleapisRepoPath, "googleapis-repo-path", "", "Local path to the googleapis/googleapis repository")
	fetchCmd.MarkFlagRequired("repo")
	fetchCmd.MarkFlagRequired("query")
	fetchCmd.MarkFlagRequired("googleapis-repo-path")
}