package cmd

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/danielgazineu/commit-history/internal"
	"github.com/google/go-github/v62/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGitHubService struct {
	mock.Mock
}

func (m *MockGitHubService) GetPullRequests(query string) ([]*github.PullRequest, error) {
	args := m.Called(query)
	return args.Get(0).([]*github.PullRequest), args.Error(1)
}

func (m *MockGitHubService) GetCommit(sha string) (*github.Commit, error) {
	args := m.Called(sha)
	return args.Get(0).(*github.Commit), args.Error(1)
}

func (m *MockGitHubService) IsSquashMerge(pr *github.PullRequest) (bool, error) {
	args := m.Called(pr)
	return args.Bool(0), args.Error(1)
}

type MockGitService struct {
	mock.Mock
}

func (m *MockGitService) Clone(url, dir string) error {
	args := m.Called(url, dir)
	return args.Error(0)
}

func (m *MockGitService) GetUnidiff(commitSHA string) (string, error) {
	args := m.Called(commitSHA)
	return args.String(0), args.Error(1)
}

func TestFetch(t *testing.T) {
	githubService := new(MockGitHubService)
	gitService := new(MockGitService)

	pr := &github.PullRequest{
		Number:         github.Int(1),
		Title:          github.String("test title"),
		Body:           github.String("Source-Link: https://github.com/googleapis/googleapis/commit/12345"),
		Head:           &github.PullRequestBranch{SHA: github.String("head_sha")},
		MergeCommitSHA: github.String("merge_sha"),
	}
	githubService.On("GetPullRequests", mock.Anything).Return([]*github.PullRequest{pr}, nil)
	githubService.On("IsSquashMerge", pr).Return(false, nil)
	githubService.On("GetCommit", mock.Anything).Return(&github.Commit{Message: github.String("commit message")}, nil)
	gitService.On("GetUnidiff", "12345").Return("unidiff", nil)
	gitService.On("Clone", mock.Anything, mock.Anything).Return(nil)

	err := runFetchForTest("owner/repo", "is:pr", githubService, gitService)
	assert.NoError(t, err)
}

func runFetchForTest(repo, query string, githubService internal.GitHubServiceInterface, gitService internal.GitServiceInterface) error {
	log.Println("Running fetch command")
	fmt.Printf("Fetching pull requests for repository: %s\n", repo)
	fmt.Printf("Query: %s\n", query)

	prs, err := githubService.GetPullRequests(query)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d merged pull requests.\n", len(prs))

	outputFileName := fmt.Sprintf("%s-fetched.csv", strings.Split(repo, "/")[1])
	csvWriter, err := internal.NewCSVWriter(outputFileName, []string{
		"pr_number",
		"before_merge_commit_hash",
		"after_merge_commit_hash",
		"pr_title",
		"pr_body",
		"is_squash_merge",
		"merge_commit_title",
		"merge_commit_body",
		"source_link",
		"resolved_source_link",
		"source_link_unidiff",
	})
	if err != nil {
		return err
	}
	defer csvWriter.Close()

	re := regexp.MustCompile(`Source-Link: (?:https://github.com/)?googleapis/googleapis(?:/commit/|@)([a-f0-9]{7,40})`)

	for _, pr := range prs {
		if err := processPullRequest(pr, githubService, csvWriter, re, gitService); err != nil {
			log.Printf("Error processing PR #%d: %v", *pr.Number, err)
		}
	}

	fmt.Println("Done.")
	return nil
}
