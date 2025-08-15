package main

import (
	"os"
	"strings"
	"testing"
)

// Ensure missing codex binary yields helpful npm install guidance.
func TestCodexMissingGivesInstallHint(t *testing.T) {
	orig := os.Getenv("PATH")
	defer os.Setenv("PATH", orig)
	os.Setenv("PATH", "") // force not found

	err := checkCodexExists()
	if err == nil {
		t.Skip("codex found in PATH in this environment; skipping")
	}

	msg := err.Error()
	if !strings.Contains(msg, "npm install -g @openai/codex") {
		t.Errorf("expected npm install guidance, got: %v", msg)
	}
}
