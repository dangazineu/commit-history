package cmd

import (
	"testing"
)

func TestRootCmd(t *testing.T) {
	// This test is just to ensure the command initializes without errors.
	// We can add more specific tests later.
	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("rootCmd.Execute() returned an error: %v", err)
	}
}
