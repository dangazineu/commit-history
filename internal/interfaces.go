package internal

import "github.com/google/go-github/v62/github"

type GitHubServiceInterface interface {
	GetPullRequests(query string) ([]*github.PullRequest, error)
	GetCommit(sha string) (*github.Commit, error)
	IsSquashMerge(pr *github.PullRequest) (bool, error)
}

type GitServiceInterface interface {
	Clone(url, dir string) error
	GetUnidiff(commitSHA string) (string, error)
}
