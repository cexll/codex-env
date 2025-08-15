package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// retryConfig holds retry configuration
type retryConfig struct {
	maxRetries int
	baseDelay  time.Duration
}

// defaultRetryConfig returns sensible defaults
func defaultRetryConfig() retryConfig {
	return retryConfig{
		maxRetries: 3,
		baseDelay:  100 * time.Millisecond,
	}
}

// exponentialBackoff calculates delay for attempt
func (rc retryConfig) exponentialBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return rc.baseDelay
	}
	delay := rc.baseDelay
	for i := 0; i < attempt; i++ {
		delay *= 2
	}
	return delay
}

// checkCodexExists verifies that codex is available in PATH with enhanced error guidance
func checkCodexExists() error {
	path, err := exec.LookPath("codex")
	if err != nil {
		errorCtx := newErrorContext("codex verification", "launcher")
		errorCtx.addContext("command", "codex")
		errorCtx.addSuggestion("Install Codex CLI via: npm install -g @openai/codex")
		errorCtx.addSuggestion("Ensure 'codex' is in your PATH environment variable")
		errorCtx.addSuggestion("Try running 'codex --version' to verify installation")

		return errorCtx.formatError(fmt.Errorf("codex not found in PATH"))
	}

	// Additional check to ensure the file is executable with permission guidance
	if info, err := os.Stat(path); err != nil {
		errorCtx := newErrorContext("permission verification", "launcher")
		errorCtx.addContext("path", path)
		errorCtx.addSuggestion("Check file permissions with: ls -la " + path)
		errorCtx.addSuggestion("Reinstall Codex if file is corrupted")

		return errorCtx.formatError(fmt.Errorf("codex path verification failed: %w", err))
	} else if info.Mode()&0111 == 0 {
		errorCtx := newErrorContext("permission check", "launcher")
		errorCtx.addContext("path", path)
		errorCtx.addContext("permissions", info.Mode().String())
		errorCtx.addSuggestion("Fix permissions with: chmod +x " + path)
		errorCtx.addSuggestion("Reinstall Codex if permission issues persist")

		return errorCtx.formatError(fmt.Errorf("codex found but not executable"))
	}

	return nil
}

// prepareEnvironment sets up environment variables for Codex execution
func prepareEnvironment(env Environment) ([]string, error) {
	// Validate environment before setting variables
	if err := validateEnvironment(env); err != nil {
		return nil, fmt.Errorf("environment preparation failed: %w", err)
	}

	// Get current environment
	currentEnv := os.Environ()

	// Calculate capacity for new environment slice
	envVarsCount := len(env.EnvVars)
	newEnv := make([]string, 0, len(currentEnv)+3+envVarsCount)

	// Copy existing environment variables (filter out OpenAI and legacy Anthropic ones)
	for _, envVar := range currentEnv {
		// Filter out OPENAI_* and ANTHROPIC_* to avoid conflicts
		if strings.HasPrefix(envVar, "OPENAI_") || strings.HasPrefix(envVar, "ANTHROPIC_") {
			continue
		}
		newEnv = append(newEnv, envVar)
	}

	// Add OpenAI-specific environment variables
	newEnv = append(newEnv, fmt.Sprintf("OPENAI_BASE_URL=%s", env.URL))
	newEnv = append(newEnv, fmt.Sprintf("OPENAI_API_KEY=%s", env.APIKey))

	// Add model indicator if specified (CLI -m remains primary)
	if env.Model != "" {
		newEnv = append(newEnv, fmt.Sprintf("OPENAI_MODEL=%s", env.Model))
	}

	// Add additional environment variables
	if env.EnvVars != nil {
		for key, value := range env.EnvVars {
			if key != "" && value != "" {
				newEnv = append(newEnv, fmt.Sprintf("%s=%s", key, value))
			}
		}
	}

	return newEnv, nil
}

// launchCodex executes codex with the specified environment and arguments
func launchCodex(env Environment, args []string) error {
	// Check if codex exists and is executable
	if err := checkCodexExists(); err != nil {
		return fmt.Errorf("Codex launcher failed: %w", err)
	}

	// Prepare environment variables
	envVars, err := prepareEnvironment(env)
	if err != nil {
		return fmt.Errorf("Codex launcher failed: %w", err)
	}

	// Find codex executable path
	codexPath, err := exec.LookPath("codex")
	if err != nil {
		return fmt.Errorf("Codex launcher failed - executable not found: %w", err)
	}

	// Prepare command arguments
	cmdArgs := append([]string{"codex"}, args...)

	// Execute codex and replace current process (Unix exec behavior)
	if err := syscall.Exec(codexPath, cmdArgs, envVars); err != nil {
		return fmt.Errorf("Codex execution failed: %w", err)
	}

	// This point should never be reached if exec succeeds
	return fmt.Errorf("unexpected return from Codex execution")
}

// launchCodexWithOutput executes codex and waits for it to complete (for testing)
func launchCodexWithOutput(env Environment, args []string) error {
	// Check if codex exists and is executable
	if err := checkCodexExists(); err != nil {
		return fmt.Errorf("Codex launcher failed: %w", err)
	}

	// Prepare environment variables
	envVars, err := prepareEnvironment(env)
	if err != nil {
		return fmt.Errorf("Codex launcher failed: %w", err)
	}

	// Create command
	cmd := exec.Command("codex", args...)
	cmd.Env = envVars
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Codex process start failed: %w", err)
	}

	// Wait for completion and handle exit code
	if err := cmd.Wait(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Get exit code from the process
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				// Exit with the same code as codex
				os.Exit(status.ExitStatus())
			}
		}
		return fmt.Errorf("Codex execution failed: %w", err)
	}

	return nil
}

// --- Backward-compatibility wrappers for legacy tests ---
// These map old Claude-named functions to new Codex implementations.
func checkClaudeCodeExists() error                          { return checkCodexExists() }
func launchClaudeCode(env Environment, args []string) error { return launchCodex(env, args) }
func launchClaudeCodeWithOutput(env Environment, args []string) error {
	return launchCodexWithOutput(env, args)
}
