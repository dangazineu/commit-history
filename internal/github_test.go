package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v62/github"
)

// setup sets up a test HTTP server along with a github.Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() (client *github.Client, mux *http.ServeMux, serverURL string, teardown func()) {
	// mux is the HTTP request multiplexer used with the test server.
	mux = http.NewServeMux()

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(mux)

	// client is the GitHub client being tested and is
	// configured to use test server.
	client = github.NewClient(nil)
	client.BaseURL, _ = url.Parse(server.URL + "/")

	return client, mux, server.URL, server.Close
}

func TestGitHubService_GetPullRequests_byNumber(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/repos/owner/repo/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"number": 1, "merged_at": "2025-06-29T19:53:00Z"}`)
	})

	service := &GitHubService{
		client: client,
		owner:  "owner",
		repo:   "repo",
	}

	prs, err := service.GetPullRequests("1")
	if err != nil {
		t.Errorf("GetPullRequests returned error: %v", err)
	}

	if len(prs) != 1 {
		t.Errorf("Expected 1 pull request, got %d", len(prs))
	}
}

func TestGitHubService_GetCommit(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/repos/owner/repo/git/commits/abc", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"sha": "abc"}`)
	})

	service := &GitHubService{
		client: client,
		owner:  "owner",
		repo:   "repo",
	}

	commit, err := service.GetCommit("abc")
	if err != nil {
		t.Errorf("GetCommit returned error: %v", err)
	}

	if *commit.SHA != "abc" {
		t.Errorf("Expected commit SHA to be 'abc', got '%s'", *commit.SHA)
	}
}

func TestGitHubService_IsSquashMerge(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	service := &GitHubService{
		client: client,
		owner:  "owner",
		repo:   "repo",
	}

	// Test case 1: True merge commit (more than one parent)
	mux.HandleFunc("/repos/owner/repo/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"number": 1, "merged_at": "2025-06-29T19:53:00Z", "merge_commit_sha": "merge_commit"}`)
	})
	mux.HandleFunc("/repos/owner/repo/git/commits/merge_commit", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"sha": "merge_commit", "parents": [{}, {}]}`)
	})

	pr := &github.PullRequest{
		Number:         github.Int(1),
		MergeCommitSHA: github.String("merge_commit"),
	}
	isSquash, err := service.IsSquashMerge(pr)
	if err != nil {
		t.Errorf("IsSquashMerge returned error: %v", err)
	}
	if isSquash {
		t.Errorf("Expected IsSquashMerge to be false for a true merge commit")
	}
}

func TestGitHubService_IsSquashMerge_Squash(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	service := &GitHubService{
		client: client,
		owner:  "owner",
		repo:   "repo",
	}

	mux.HandleFunc("/repos/owner/repo/pulls/2", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"number": 2, "merged_at": "2025-06-29T19:53:00Z", "merge_commit_sha": "squash_commit", "base": {"sha": "base_commit"}}`)
	})
	mux.HandleFunc("/repos/owner/repo/git/commits/squash_commit", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"sha": "squash_commit", "parents": [{"sha": "base_commit"}]}`)
	})

	pr := &github.PullRequest{
		Number:         github.Int(2),
		MergeCommitSHA: github.String("squash_commit"),
		Base:           &github.PullRequestBranch{SHA: github.String("base_commit")},
	}
	isSquash, err := service.IsSquashMerge(pr)
	if err != nil {
		t.Errorf("IsSquashMerge returned error: %v", err)
	}
	if !isSquash {
		t.Errorf("Expected IsSquashMerge to be true for a squash merge")
	}
}

func TestGitHubService_IsSquashMerge_Rebase(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	service := &GitHubService{
		client: client,
		owner:  "owner",
		repo:   "repo",
	}

	mux.HandleFunc("/repos/owner/repo/pulls/3", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"number": 3, "merged_at": "2025-06-29T19:53:00Z", "merge_commit_sha": "rebase_commit_2", "base": {"sha": "base_commit"}}`)
	})
	mux.HandleFunc("/repos/owner/repo/git/commits/rebase_commit_2", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"sha": "rebase_commit_2", "parents": [{"sha": "rebase_commit_1"}]}`)
	})
	mux.HandleFunc("/repos/owner/repo/git/commits/rebase_commit_1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"sha": "rebase_commit_1", "parents": [{"sha": "base_commit"}]}`)
	})

	pr := &github.PullRequest{
		Number:         github.Int(3),
		MergeCommitSHA: github.String("rebase_commit_2"),
		Base:           &github.PullRequestBranch{SHA: github.String("base_commit")},
	}
	isSquash, err := service.IsSquashMerge(pr)
	if err != nil {
		t.Errorf("IsSquashMerge returned error: %v", err)
	}
	if isSquash {
		t.Errorf("Expected IsSquashMerge to be false for a rebase and merge")
	}
}

