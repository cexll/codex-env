package main

import (
	"os"
	"testing"
)

// TestModelValidatorCreation tests enhanced model validator creation and configuration
func TestModelValidatorCreation(t *testing.T) {
	t.Run("newModelValidator default creation", func(t *testing.T) {
		mv := newModelValidator()

		if mv == nil {
			t.Fatal("Expected non-nil model validator")
		}
		// Codex mode: no default patterns, non-strict
		if len(mv.patterns) != 0 {
			t.Error("Expected no default patterns for Codex")
		}
		if mv.strictMode {
			t.Error("Expected strict mode disabled by default for Codex")
		}
	})

	t.Run("newModelValidator with custom patterns", func(t *testing.T) {
		// Save original environment
		originalPatterns := os.Getenv("CCE_MODEL_PATTERNS")
		originalStrict := os.Getenv("CCE_MODEL_STRICT")
		defer func() {
			if originalPatterns == "" {
				os.Unsetenv("CCE_MODEL_PATTERNS")
			} else {
				os.Setenv("CCE_MODEL_PATTERNS", originalPatterns)
			}
			if originalStrict == "" {
				os.Unsetenv("CCE_MODEL_STRICT")
			} else {
				os.Setenv("CCE_MODEL_STRICT", originalStrict)
			}
		}()

		// Set custom patterns and non-strict mode
		os.Setenv("CCE_MODEL_PATTERNS", "custom-pattern-.*,another-pattern-[0-9]+")
		os.Setenv("CCE_MODEL_STRICT", "false")

		mv := newModelValidator()

		if mv.strictMode {
			t.Error("Expected strict mode to be disabled")
		}

		// Check that custom patterns were added
		found := false
		for _, pattern := range mv.patterns {
			if pattern == "custom-pattern-.*" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Custom pattern not found in validator")
		}
	})

	t.Run("newModelValidatorWithConfig", func(t *testing.T) {
		config := Config{
			Settings: &ConfigSettings{
				Validation: &ValidationSettings{
					ModelPatterns:    []string{"config-pattern-.*"},
					StrictValidation: false,
				},
			},
		}

		mv := newModelValidatorWithConfig(config)

		if mv.strictMode {
			t.Error("Expected strict mode to be disabled from config")
		}

		// Check that config patterns were added
		found := false
		for _, pattern := range mv.patterns {
			if pattern == "config-pattern-.*" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Config pattern not found in validator")
		}
	})
}

// TestModelValidation tests enhanced adaptive model validation
func TestModelValidation(t *testing.T) {
	testCases := []struct {
		name        string
		model       string
		strictMode  bool
		expectError bool
		description string
	}{
		// In Codex mode, all model names are allowed unless unsafe
		{"openai-gpt", "gpt-5", true, false, "openai model"},
		{"openai-o4", "o4-mini", true, false, "openai o-series"},
		{"custom", "my-custom-model-202501", true, false, "custom model allowed"},
		{"empty-strict", "", true, false, "empty model allowed"},
		{"unsafe backtick", "bad`name", true, true, "unsafe character blocked"},
		{"unsafe subcmd", "$(whoami)", true, true, "substitution blocked"},
		{"path traversal", "../bad", true, true, "path traversal blocked"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mv := newModelValidator()
			mv.strictMode = tc.strictMode

			err := mv.validateModelAdaptive(tc.model)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s but got none", tc.description)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.description, err)
			}
		})
	}
}

// TestPatternValidation tests pattern compilation and validation
func TestPatternValidation(t *testing.T) {
	mv := newModelValidator()

	testCases := []struct {
		name        string
		pattern     string
		expectError bool
	}{
		{"valid pattern", `^claude-.*$`, false},
		{"complex pattern", `^claude-[0-9]+-[a-z]+-[0-9]{8}$`, false},
		{"invalid regex", `[unclosed`, true},
		{"empty pattern", "", false}, // Empty is valid regex
		{"malformed bracket", `[abc`, true},
		{"invalid escape", `\x`, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := mv.validatePattern(tc.pattern)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for pattern '%s'", tc.pattern)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for pattern '%s': %v", tc.pattern, err)
			}
		})
	}
}

// TestModelValidationWithConfig tests model validation integration with configuration
func TestModelValidationWithConfig(t *testing.T) {
	t.Run("validateModel function integration", func(t *testing.T) {
		testModels := []string{
			"claude-3-5-sonnet-20241022",
			"claude-3-haiku-20240307",
			"claude-opus-20250101", // Future format
			"",                     // Empty (should be valid)
		}

		for _, model := range testModels {
			err := validateModel(model)
			if err != nil {
				t.Errorf("validateModel('%s') failed: %v", model, err)
			}
		}
	})

	t.Run("any models now accepted", func(t *testing.T) {
		anyModels := []string{
			"gpt-5",    // OpenAI model
			"o4-mini",  // OpenAI o-series
			"kimi",     // Kimi
			"deepseek", // DeepSeek
			"glm-4",    // GLM
			"custom",   // custom
		}

		for _, model := range anyModels {
			err := validateModel(model)
			if err != nil {
				t.Errorf("validateModel('%s') should now be accepted: %v", model, err)
			}
		}
	})
}

// TestValidationSettingsIntegration tests validation settings with configuration
func TestValidationSettingsIntegration(t *testing.T) {
	t.Run("config with validation settings", func(t *testing.T) {
		config := Config{
			Environments: []Environment{
				{
					Name:   "test",
					URL:    "https://api.openai.com/v1",
					APIKey: "sk-test",
					Model:  "custom-model-pattern",
				},
			},
			Settings: &ConfigSettings{
				Validation: &ValidationSettings{
					ModelPatterns:    []string{"custom-model-.*"},
					StrictValidation: false,
				},
			},
		}

		mv := newModelValidatorWithConfig(config)

		// Test that custom pattern allows the model
		err := mv.validateModelAdaptive("custom-model-pattern")
		if err != nil {
			t.Errorf("Custom pattern should validate: %v", err)
		}
	})

	t.Run("config without validation settings", func(t *testing.T) {
		config := Config{
			Environments: []Environment{
				{
					Name:   "test",
					URL:    "https://api.openai.com/v1",
					APIKey: "sk-test",
					Model:  "gpt-5",
				},
			},
			// No Settings specified
		}

		mv := newModelValidatorWithConfig(config)

		// Should use defaults
		// Codex mode default: non-strict
		if mv.strictMode {
			t.Error("Expected non-strict mode by default in Codex")
		}

		// Any model should be accepted unless unsafe
		err := mv.validateModelAdaptive("gpt-5")
		if err != nil {
			t.Errorf("Model should be accepted in Codex mode: %v", err)
		}
	})
}

// TestEnvironmentValidationWithModel tests environment validation including model
func TestEnvironmentValidationWithModel(t *testing.T) {
	testCases := []struct {
		name        string
		env         Environment
		expectError bool
	}{
		{
			"valid environment with model",
			Environment{
				Name:   "test",
				URL:    "https://api.anthropic.com",
				APIKey: "sk-ant-test123456789",
				Model:  "claude-3-5-sonnet-20241022",
			},
			false,
		},
		{
			"valid environment without model",
			Environment{
				Name:   "test",
				URL:    "https://api.anthropic.com",
				APIKey: "sk-ant-test123456789",
				Model:  "",
			},
			false,
		},
		{
			"any model in environment now accepted",
			Environment{
				Name:   "test",
				URL:    "https://api.anthropic.com",
				APIKey: "sk-ant-test123456789",
				Model:  "any-model-name",
			},
			false, // Now accepted
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateEnvironment(tc.env)

			if tc.expectError && err == nil {
				t.Error("Expected validation error")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}

			// Note: Model validation is now disabled, so no model-specific errors expected
		})
	}
}

// BenchmarkModelValidation benchmarks model validation performance
func BenchmarkModelValidation(b *testing.B) {
	mv := newModelValidator()
	testModel := "gpt-5"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mv.validateModelAdaptive(testModel)
	}
}

// BenchmarkPatternCompilation benchmarks pattern compilation performance
func BenchmarkPatternCompilation(b *testing.B) {
	mv := newModelValidator()
	testPattern := `^gpt-[0-9]+$`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mv.validatePattern(testPattern)
	}
}
