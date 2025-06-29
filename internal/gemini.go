package internal

import (
	"fmt"
	"os/exec"
	"strings"
)

type GeminiService struct {
	geminiPath string
}

func NewGeminiService() (*GeminiService, error) {
	path, err := exec.LookPath("gemini")
	if err != nil {
		return nil, fmt.Errorf("gemini executable not found in PATH")
	}
	return &GeminiService{geminiPath: path}, nil
}

var execCommand = exec.Command

func (s *GeminiService) GenerateCommitMessage(tempDir, prTitle, prBody, unidiff string) (string, string, error) {
	prompt := fmt.Sprintf(`
Given the following pull request information and a unidiff of the changes from the source repository, please generate a conventional commit message.
The commit message should describe how the changes in the source repository affect the client library.
Sometimes a feature change in an API may only result in a documentation change in the library, so ensure the resulting commit is relevant to the library.

PR Title: %s
PR Body: %s
Unidiff:
%s
`, prTitle, prBody, unidiff)

	cmd := execCommand(s.geminiPath, "propose-commit", "--prompt", prompt)
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("error executing gemini command: %w\nOutput: %s", err, string(output))
	}

	parts := strings.SplitN(string(output), "\n\n", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected output from gemini: %s", string(output))
	}

	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func (s *GeminiService) AnalyzeCommitMessages(tempDir, originalTitle, originalBody, geminiTitle, geminiBody string) (string, error) {
	prompt := fmt.Sprintf(`
Given the following two commit messages, which one is more comprehensive and accurately describes the changes to the library?

Original Commit:
Title: %s
Body: %s

Gemini Commit:
Title: %s
Body: %s
`, originalTitle, originalBody, geminiTitle, geminiBody)

	cmd := execCommand(s.geminiPath, "analyze-commit", "--prompt", prompt)
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error executing gemini command: %w\nOutput: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}
