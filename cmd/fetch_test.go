package cmd

import (
	"os"
	"testing"

	"github.com/danielgazineu/commit-history/internal"
)

func TestFetchIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	if os.Getenv("GOOGLEAPIS_REPO_PATH") == "" {
		t.Skip("skipping integration test, GOOGLEAPIS_REPO_PATH not set")
	}

	repo := "googleapis/google-cloud-ruby"
	query := "is:pr is:merged created:2024-01-01..2024-01-02 label:owl-bot-copy"

	githubService, err := internal.NewGitHubService(repo)
	if err != nil {
		t.Fatalf("NewGitHubService() error = %v", err)
	}

	prs, err := githubService.GetPullRequests(query)
	if err != nil {
		t.Fatalf("GetPullRequests() error = %v", err)
	}

	if len(prs) != 4 {
		t.Errorf("expected 4 pull requests, got %d", len(prs))
	}
}

func TestFetchSquashMergeIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	if os.Getenv("GOOGLEAPIS_REPO_PATH") == "" {
		t.Skip("skipping integration test, GOOGLEAPIS_REPO_PATH not set")
	}

	repo := "google-gemini/gemini-cli"
	query := "85"

	githubService, err := internal.NewGitHubService(repo)
	if err != nil {
		t.Fatalf("NewGitHubService() error = %v", err)
	}

	prs, err := githubService.GetPullRequests(query)
	if err != nil {
		t.Fatalf("GetPullRequests() error = %v", err)
	}

	if len(prs) != 1 {
		t.Fatalf("expected 1 pull request, got %d", len(prs))
	}

	isSquash, err := githubService.IsSquashMerge(prs[0])
	if err != nil {
		t.Fatalf("IsSquashMerge() error = %v", err)
	}

	if !isSquash {
		t.Errorf("expected PR to be squash-merged, but it wasn't")
	}
}
