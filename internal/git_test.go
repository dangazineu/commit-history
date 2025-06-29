package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewGitService(t *testing.T) {
	t.Run("repository exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0755); err != nil {
			t.Fatalf("failed to create git dir: %v", err)
		}

		_, err := NewGitService(tmpDir)
		if err != nil {
			t.Errorf("NewGitService() error = %v, wantErr %v", err, false)
		}
	})

	t.Run("repository does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := NewGitService(tmpDir)
		if err == nil {
			t.Errorf("NewGitService() error = %v, wantErr %v", err, true)
		}
	})
}

func TestGetUnidiff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}
	tmpDir := t.TempDir()
	service, err := NewGitService(tmpDir)
	if err == nil {
		t.Fatalf("NewGitService() error = %v, wantErr %v", err, true)
	}

	err = service.Clone("https://github.com/dangazineu/commit-history.git", tmpDir)
	if err != nil {
		t.Fatalf("Clone() failed: %v", err)
	}

	service, err = NewGitService(tmpDir)
	if err != nil {
		t.Fatalf("NewGitService() after clone failed: %v", err)
	}

	// This is the merge commit of the previous PR
	unidiff, err := service.GetUnidiff("e0dc0d4")
	if err != nil {
		t.Fatalf("GetUnidiff() failed: %v", err)
	}

	if unidiff == "" {
		t.Errorf("GetUnidiff() returned empty string")
	}
}
