package main

import (
	"strings"
	"testing"
)

// TestAdditionalEnvironmentVariables tests the new EnvVars functionality
func TestAdditionalEnvironmentVariables(t *testing.T) {
	// Set up test environment with additional env vars
	env := Environment{
		Name:   "test-with-envvars",
		URL:    "https://api.openai.com/v1",
		APIKey: "sk-test",
		Model:  "gpt-5",
		EnvVars: map[string]string{
			"OPENAI_TIMEOUT": "30",
			"CUSTOM_ENV_VAR": "test-value",
		},
	}

	// Test prepareEnvironment with additional env vars
	envVars, err := prepareEnvironment(env)
	if err != nil {
		t.Fatalf("prepareEnvironment() failed: %v", err)
	}

	// Check that standard variables are still set
	foundBaseURL := false
	foundAPIKey := false
	foundModel := false
	foundTimeout := false
	foundCustomVar := false

	for _, envVar := range envVars {
		if strings.HasPrefix(envVar, "OPENAI_BASE_URL=") {
			foundBaseURL = true
			if envVar != "OPENAI_BASE_URL=https://api.openai.com/v1" {
				t.Errorf("Unexpected OPENAI_BASE_URL value: %s", envVar)
			}
		}
		if strings.HasPrefix(envVar, "OPENAI_API_KEY=") {
			foundAPIKey = true
		}
		if strings.HasPrefix(envVar, "OPENAI_MODEL=") {
			foundModel = true
		}
		if strings.HasPrefix(envVar, "OPENAI_TIMEOUT=") {
			foundTimeout = true
			if envVar != "OPENAI_TIMEOUT=30" {
				t.Errorf("Unexpected OPENAI_TIMEOUT value: %s", envVar)
			}
		}
		if strings.HasPrefix(envVar, "CUSTOM_ENV_VAR=") {
			foundCustomVar = true
			if envVar != "CUSTOM_ENV_VAR=test-value" {
				t.Errorf("Unexpected CUSTOM_ENV_VAR value: %s", envVar)
			}
		}
	}

	if !foundBaseURL {
		t.Error("OPENAI_BASE_URL not found in environment variables")
	}
	if !foundAPIKey {
		t.Error("OPENAI_API_KEY not found in environment variables")
	}
	if !foundModel {
		t.Error("OPENAI_MODEL not found in environment variables")
	}
	if !foundTimeout {
		t.Error("OPENAI_TIMEOUT not found in environment variables")
	}
	if !foundCustomVar {
		t.Error("CUSTOM_ENV_VAR not found in environment variables")
	}
}

// TestEnvironmentEqualityWithEnvVars tests the equalEnvironments function
func TestEnvironmentEqualityWithEnvVars(t *testing.T) {
	env1 := Environment{
		Name:   "test",
		URL:    "https://api.openai.com/v1",
		APIKey: "sk-test",
		Model:  "gpt-5",
		EnvVars: map[string]string{
			"OPENAI_TIMEOUT": "30",
		},
	}

	env2 := Environment{
		Name:   "test",
		URL:    "https://api.openai.com/v1",
		APIKey: "sk-test",
		Model:  "gpt-5",
		EnvVars: map[string]string{
			"OPENAI_TIMEOUT": "30",
		},
	}

	env3 := Environment{
		Name:   "test",
		URL:    "https://api.openai.com/v1",
		APIKey: "sk-test",
		Model:  "gpt-5",
		EnvVars: map[string]string{
			"OPENAI_TIMEOUT": "60", // Different value
		},
	}

	// Test equal environments
	if !equalEnvironments(env1, env2) {
		t.Error("env1 and env2 should be equal")
	}

	// Test different environments
	if equalEnvironments(env1, env3) {
		t.Error("env1 and env3 should not be equal (different timeout)")
	}

	// Test with nil EnvVars
	env4 := Environment{
		Name:    "test",
		URL:     "https://api.openai.com/v1",
		APIKey:  "sk-test",
		Model:   "gpt-5",
		EnvVars: nil,
	}

	env5 := Environment{
		Name:    "test",
		URL:     "https://api.openai.com/v1",
		APIKey:  "sk-test",
		Model:   "gpt-5",
		EnvVars: make(map[string]string),
	}

	if !equalEnvironments(env4, env5) {
		t.Error("Environment with nil EnvVars should equal environment with empty EnvVars")
	}
}

// TestEmptyEnvVars tests behavior with empty or nil EnvVars
func TestEmptyEnvVars(t *testing.T) {
	env := Environment{
		Name:    "test-empty-envvars",
		URL:     "https://api.openai.com/v1",
		APIKey:  "sk-test",
		Model:   "gpt-5",
		EnvVars: nil, // nil EnvVars
	}

	envVars, err := prepareEnvironment(env)
	if err != nil {
		t.Fatalf("prepareEnvironment() with nil EnvVars failed: %v", err)
	}

	// Should still have the basic OPENAI variables
	foundBaseURL := false
	foundAPIKey := false
	foundModel := false

	for _, envVar := range envVars {
		if strings.HasPrefix(envVar, "OPENAI_BASE_URL=") {
			foundBaseURL = true
		}
		if strings.HasPrefix(envVar, "OPENAI_API_KEY=") {
			foundAPIKey = true
		}
		if strings.HasPrefix(envVar, "OPENAI_MODEL=") {
			foundModel = true
		}
	}

	if !foundBaseURL || !foundAPIKey || !foundModel {
		t.Error("Basic OPENAI environment variables should still be set with nil EnvVars")
	}
}

func TestValidateEnvVarNames(t *testing.T) {
	// Test valid environment variable names
	validNames := []string{
		"VALID_VAR",
		"_VALID_VAR",
		"VAR_123",
		"a",
		"_",
		"ABC_123_DEF",
		"lower_case_var",
		"MixedCase_123",
	}

	for _, name := range validNames {
		if !isValidEnvVarName(name) {
			t.Errorf("Expected '%s' to be valid, but validation failed", name)
		}
	}

	// Test invalid environment variable names
	invalidNames := []string{
		"",          // empty
		"123VAR",    // starts with number
		"VAR-NAME",  // contains dash
		"VAR NAME",  // contains space
		"VAR=VALUE", // contains equals
		"VAR@HOME",  // contains special character
		"-VAR",      // starts with dash
		"VAR.NAME",  // contains dot
	}

	for _, name := range invalidNames {
		if isValidEnvVarName(name) {
			t.Errorf("Expected '%s' to be invalid, but validation passed", name)
		}
	}
}

func TestCommonSystemVarDetection(t *testing.T) {
	// Test detection of common system variables
	commonVars := []string{
		"PATH",
		"HOME",
		"USER",
		"SHELL",
		"GOPATH",
		"JAVA_HOME",
		"path", // lowercase should also be detected
		"home",
		"java_home",
	}

	for _, varName := range commonVars {
		if !isCommonSystemVar(varName) {
			t.Errorf("Expected '%s' to be detected as common system variable", varName)
		}
	}

	// Test non-system variables
	nonSystemVars := []string{
		"OPENAI_API_KEY",
		"MY_CUSTOM_VAR",
		"APP_SECRET",
		"DATABASE_URL",
	}

	for _, varName := range nonSystemVars {
		if isCommonSystemVar(varName) {
			t.Errorf("Expected '%s' to NOT be detected as common system variable", varName)
		}
	}
}
