package internal

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

// mockExecCommand is a helper function to mock exec.Command for testing.
func mockExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestGenerateCommitMessage(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()

	s := &geminiServiceImpl{geminiPath: "gemini"}
	title, body, err := s.GenerateCommitMessage("/tmp", "test title", "test body", "test diff")
	if err != nil {
		t.Fatalf("GenerateCommitMessage failed: %v", err)
	}

	if want := "feat: new feature"; title != want {
		t.Errorf("got title %q, want %q", title, want)
	}
	if want := "This is the body."; body != want {
		t.Errorf("got body %q, want %q", body, want)
	}
}

func TestAnalyzeCommitMessages(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()

	s := &geminiServiceImpl{geminiPath: "gemini"}
	analysis, err := s.AnalyzeCommitMessages("/tmp", "original title", "original body", "gemini title", "gemini body")
	if err != nil {
		t.Fatalf("AnalyzeCommitMessages failed: %v", err)
	}

	if want := "feat: new feature\n\nThis is the body."; analysis != want {
		t.Errorf("got analysis %q, want %q", analysis, want)
	}
}

// TestHelperProcess isn't a real test. It's used as a helper process
// for TestGenerateCommitMessage.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd := args[0]
	if cmd == "gemini" {
		// Simulate the gemini executable's behavior
		fmt.Fprintln(os.Stdout, "feat: new feature\n\nThis is the body.")
	} else {
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", cmd)
		os.Exit(2)
	}
}
