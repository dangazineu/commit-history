package internal

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

type GitHubService struct {
	client *github.Client
	owner  string
	repo   string
}

func NewGitHubService(repoURL string) (*GitHubService, error) {
	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_PERSONAL_ACCESS_TOKEN environment variable not set")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	parts := strings.Split(repoURL, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository URL format, expected 'owner/repo'")
	}

	return &GitHubService{
		client: github.NewClient(tc),
		owner:  parts[0],
		repo:   parts[1],
	}, nil
}

func (s *GitHubService) GetPullRequests(query string) ([]*github.PullRequest, error) {
	if prNumber, err := strconv.Atoi(query); err == nil {
		pr, _, err := s.client.PullRequests.Get(context.Background(), s.owner, s.repo, prNumber)
		if err != nil {
			return nil, err
		}
		if pr.MergedAt != nil {
			return []*github.PullRequest{pr}, nil
		}
		return []*github.PullRequest{}, nil
	}

	var allPRs []*github.PullRequest
	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		result, resp, err := s.client.Search.Issues(context.Background(), fmt.Sprintf("repo:%s/%s %s", s.owner, s.repo, query), opts)
		if err != nil {
			return nil, err
		}

		for _, issue := range result.Issues {
			if issue.IsPullRequest() {
				pr, _, err := s.client.PullRequests.Get(context.Background(), s.owner, s.repo, *issue.Number)
				if err != nil {
					return nil, err
				}
				if pr.MergedAt != nil {
					allPRs = append(allPRs, pr)
				}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPRs, nil
}

func (s *GitHubService) GetCommit(sha string) (*github.Commit, error) {
	commit, _, err := s.client.Git.GetCommit(context.Background(), s.owner, s.repo, sha)
	return commit, err
}

func (s *GitHubService) IsSquashMerge(pr *github.PullRequest) (bool, error) {
	commits, _, err := s.client.PullRequests.ListCommits(context.Background(), s.owner, s.repo, *pr.Number, nil)
	if err != nil {
		return false, err
	}

	return len(commits) == 1, nil
}