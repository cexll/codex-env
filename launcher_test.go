package main

import (
	"os"
	"strings"
	"testing"
)

func TestPrepareEnvironment(t *testing.T) {
	env := Environment{
		Name:   "test",
		URL:    "https://api.openai.com/v1",
		APIKey: "sk-test",
	}

	envVars, err := prepareEnvironment(env)
	if err != nil {
		t.Fatalf("prepareEnvironment() failed: %v", err)
	}

	// Check that OpenAI variables are set
	foundBaseURL := false
	foundAPIKey := false
	foundOtherOpenAIVar := false

	for _, envVar := range envVars {
		if strings.HasPrefix(envVar, "OPENAI_BASE_URL=") {
			foundBaseURL = true
			expected := "OPENAI_BASE_URL=" + env.URL
			if envVar != expected {
				t.Errorf("Expected %s, got %s", expected, envVar)
			}
		}
		if strings.HasPrefix(envVar, "OPENAI_API_KEY=") {
			foundAPIKey = true
			expected := "OPENAI_API_KEY=" + env.APIKey
			if envVar != expected {
				t.Errorf("Expected %s, got %s", expected, envVar)
			}
		}
		// Check that existing OpenAI variables are filtered out
		if strings.HasPrefix(envVar, "OPENAI_") &&
			!strings.HasPrefix(envVar, "OPENAI_BASE_URL=") &&
			!strings.HasPrefix(envVar, "OPENAI_API_KEY=") {
			foundOtherOpenAIVar = true
		}
	}

	if !foundBaseURL {
		t.Error("OPENAI_BASE_URL not found in environment")
	}
	if !foundAPIKey {
		t.Error("OPENAI_API_KEY not found in environment")
	}
	if foundOtherOpenAIVar {
		t.Error("Other OPENAI variables should be filtered out")
	}
}

func TestPrepareEnvironmentInvalid(t *testing.T) {
	invalidEnv := Environment{
		Name:   "",
		URL:    "invalid-url",
		APIKey: "invalid",
	}

	_, err := prepareEnvironment(invalidEnv)
	if err == nil {
		t.Error("Expected error with invalid environment")
	}
}

func TestCheckCodexExists(t *testing.T) {
	// This test depends on whether claude-code is actually installed
	// We'll test both scenarios

	err := checkCodexExists()

	// If claude-code is not installed, we should get a specific error
	if err != nil {
		if !strings.Contains(err.Error(), "not found in PATH") {
			t.Errorf("Expected PATH error, got: %v", err)
		}
	}

	// Test with definitely non-existent command by temporarily changing PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Set PATH to empty to ensure codex is not found
	os.Setenv("PATH", "")

	err = checkCodexExists()
	if err == nil {
		t.Error("Expected error when codex is not in PATH")
	}
	if !strings.Contains(err.Error(), "not found in PATH") {
		t.Errorf("Expected PATH error, got: %v", err)
	}
}

// Mock launcher tests would require more complex setup,
// but we can test the error paths and validation logic

func TestLaunchCodexValidation(t *testing.T) {
	// Test with invalid environment
	invalidEnv := Environment{
		Name:   "",
		URL:    "invalid",
		APIKey: "invalid",
	}

	// This should fail during environment preparation
	err := launchCodex(invalidEnv, []string{})
	if err == nil {
		t.Error("Expected error with invalid environment")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "launcher failed") {
		t.Errorf("Expected launcher error, got: %v", err)
	}
}

func TestLaunchCodexWithOutputValidation(t *testing.T) {
	// Test with invalid environment
	invalidEnv := Environment{
		Name:   "",
		URL:    "invalid",
		APIKey: "invalid",
	}

	// This should fail during environment preparation
	err := launchCodexWithOutput(invalidEnv, []string{})
	if err == nil {
		t.Error("Expected error with invalid environment")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "launcher failed") {
		t.Errorf("Expected launcher error, got: %v", err)
	}
}
