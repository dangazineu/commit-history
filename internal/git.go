package internal

import (
	"fmt"
	"os"
	"os/exec"
)

type GitService struct {
	repoPath string
	repoDir  string
}

func NewGitService(repoPath string) (*GitService, error) {
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository path not found: %s", repoPath)
	}
	return &GitService{
		repoPath: repoPath,
	}, nil
}

func (s *GitService) LocalClone(dir string) error {
	cmd := exec.Command("git", "clone", "--local", s.repoPath, dir)
	return cmd.Run()
}

func (s *GitService) SetRepoDir(dir string) {
	s.repoDir = dir
}

func (s *GitService) GetUnidiff(commitSHA string) (string, error) {
	cmd := exec.Command("git", "-C", s.repoDir, "show", "--format=%b", commitSHA)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting unidiff for commit %s: %w", commitSHA, err)
	}
	return string(output), nil
}