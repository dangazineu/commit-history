package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type GitService struct {
	repoDir string
}

func NewGitService(repoDir string) (*GitService, error) {
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository path not found: %s", repoDir)
	}
	return &GitService{
		repoDir: repoDir,
	}, nil
}

func (s *GitService) Clone(url, dir string) error {
	cmd := exec.Command("git", "clone", url, dir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error cloning repository: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func (s *GitService) GetUnidiff(commitSHA string) (string, error) {
	cmd := exec.Command("git", "-C", s.repoDir, "show", "--format=%b", commitSHA)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error getting unidiff for commit %s: %w\nOutput: %s", commitSHA, err, string(output))
	}
	return string(output), nil
}
