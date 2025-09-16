package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// Version information (set by ldflags during build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// modelValidator manages configurable model validation patterns
type modelValidator struct {
	patterns     []string
	customConfig map[string][]string
	strictMode   bool
}

// newModelValidator creates validator with built-in and custom patterns
func newModelValidator() *modelValidator {
	mv := &modelValidator{
		patterns:     []string{},
		customConfig: make(map[string][]string),
		strictMode:   false,
	}

	// Load custom patterns from environment variable
	if customPatterns := os.Getenv("CCE_MODEL_PATTERNS"); customPatterns != "" {
		patterns := strings.Split(customPatterns, ",")
		for _, pattern := range patterns {
			pattern = strings.TrimSpace(pattern)
			if pattern != "" {
				mv.patterns = append(mv.patterns, pattern)
			}
		}
	}

	// Check if strict mode is disabled
	if os.Getenv("CCE_MODEL_STRICT") == "false" {
		mv.strictMode = false
	}

	return mv
}

// newModelValidatorWithConfig creates validator with configuration file settings
func newModelValidatorWithConfig(config Config) *modelValidator {
	mv := newModelValidator()

	// Override with configuration file settings if available
	if config.Settings != nil && config.Settings.Validation != nil {
		validation := config.Settings.Validation

		// Add custom patterns from config
		if len(validation.ModelPatterns) > 0 {
			mv.patterns = append(mv.patterns, validation.ModelPatterns...)
		}

		// Override strict mode setting
		mv.strictMode = validation.StrictValidation
	}

	return mv
}

// validatePattern checks if a pattern compiles correctly
func (mv *modelValidator) validatePattern(pattern string) error {
	_, err := regexp.Compile(pattern)
	return err
}

// Environment represents a single Codex API configuration
type Environment struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	APIKey  string            `json:"api_key"`
	Model   string            `json:"model,omitempty"`
	EnvVars map[string]string `json:"env_vars,omitempty"`
}

// Config represents the complete configuration with all environments
type Config struct {
	Environments []Environment   `json:"environments"`
	Settings     *ConfigSettings `json:"settings,omitempty"`
}

// ConfigSettings holds optional configuration settings
type ConfigSettings struct {
	Terminal   *TerminalSettings   `json:"terminal,omitempty"`
	Validation *ValidationSettings `json:"validation,omitempty"`
}

// TerminalSettings configures terminal behavior
type TerminalSettings struct {
	ForceFallback     bool   `json:"force_fallback,omitempty"`
	DisableANSI       bool   `json:"disable_ansi,omitempty"`
	CompatibilityMode string `json:"compatibility_mode,omitempty"`
}

// ValidationSettings configures model validation behavior
type ValidationSettings struct {
	ModelPatterns    []string `json:"model_patterns,omitempty"`
	StrictValidation bool     `json:"strict_validation,omitempty"`
	// UnknownModelAction string   `json:"unknown_model_action,omitempty"`
}

// ArgumentParser manages two-phase argument parsing for CDE and codex flags
type ArgumentParser struct {
	cceFlags     map[string]string
	claudeArgs   []string
	separatorPos int // Position of -- separator if found
}

// ParseResult contains the results of argument parsing
type ParseResult struct {
	CCEFlags   map[string]string
	ClaudeArgs []string
	Subcommand string
	Error      error
}

// CCECommand represents a parsed command with environment and claude arguments
type CCECommand struct {
	Type        CommandType
	Environment string
	ClaudeArgs  []string
}

// CommandType represents the type of command being executed
type CommandType int

const (
	DefaultCommand CommandType = iota
	ListCommand
	AddCommand
	RemoveCommand
	HelpCommand
)

// errorContext provides structured error information with recovery guidance
type errorContext struct {
	Operation   string
	Component   string
	Context     map[string]string
	Suggestions []string
	Recovery    func() error
}

// newErrorContext creates a new error context
func newErrorContext(operation, component string) *errorContext {
	return &errorContext{
		Operation:   operation,
		Component:   component,
		Context:     make(map[string]string),
		Suggestions: []string{},
	}
}

// addContext adds contextual information to the error
func (ec *errorContext) addContext(key, value string) *errorContext {
	ec.Context[key] = value
	return ec
}

// addSuggestion adds a recovery suggestion
func (ec *errorContext) addSuggestion(suggestion string) *errorContext {
	ec.Suggestions = append(ec.Suggestions, suggestion)
	return ec
}

// withRecovery adds a recovery function
func (ec *errorContext) withRecovery(recovery func() error) *errorContext {
	ec.Recovery = recovery
	return ec
}

// formatError creates a comprehensive error message
func (ec *errorContext) formatError(baseErr error) error {
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("%s failed in %s: %v", ec.Operation, ec.Component, baseErr))

	if len(ec.Context) > 0 {
		msg.WriteString("\nContext:")
		for key, value := range ec.Context {
			msg.WriteString(fmt.Sprintf("\n  %s: %s", key, value))
		}
	}

	if len(ec.Suggestions) > 0 {
		msg.WriteString("\nSuggestions:")
		for _, suggestion := range ec.Suggestions {
			msg.WriteString(fmt.Sprintf("\n  • %s", suggestion))
		}
	}

	return fmt.Errorf("%s", msg.String())
}

// validateEnvironment performs comprehensive validation of environment data
func validateEnvironment(env Environment) error {
	if err := validateName(env.Name); err != nil {
		return fmt.Errorf("invalid name: %w", err)
	}
	if err := validateURL(env.URL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if err := validateAPIKey(env.APIKey); err != nil {
		return fmt.Errorf("invalid API key: %w", err)
	}
	if err := validateModel(env.Model); err != nil {
		return fmt.Errorf("invalid model: %w", err)
	}
	return nil
}

// validateName validates environment name format and length
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) > 50 {
		return fmt.Errorf("name too long (max 50 characters)")
	}
	// Allow alphanumeric, hyphens, underscores
	matched, err := regexp.MatchString("^[a-zA-Z0-9_-]+$", name)
	if err != nil {
		return fmt.Errorf("name validation failed: %w", err)
	}
	if !matched {
		return fmt.Errorf("name contains invalid characters (use only letters, numbers, hyphens, underscores)")
	}
	return nil
}

// validateURL validates URL using net/url.Parse with comprehensive error checking
func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	if parsed.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	// Disallow embedded credentials/userinfo or deceptive host components
	if parsed.User != nil || strings.Contains(parsed.Host, "@") {
		return fmt.Errorf("URL must not include credentials")
	}

	return nil
}

// validateAPIKey performs minimal API key checks (no format validation per requirements)
func validateAPIKey(apiKey string) error {
	// No format/length enforcement; just basic safety
	for _, r := range apiKey {
		if r < 32 || r == 127 {
			return fmt.Errorf("API key contains invalid characters")
		}
	}
	return nil
}

// validateModel allows any model name (no validation)
func validateModel(model string) error {
	if model == "" {
		return nil
	}
	// Basic injection/path traversal protections
	if strings.Contains(model, "$(") || strings.Contains(model, "`") || strings.Contains(model, ";") || strings.Contains(model, "../") || strings.Contains(model, "\\x") {
		return fmt.Errorf("model contains disallowed characters")
	}
	// Reject control characters
	for _, r := range model {
		if r < 32 || r == 127 {
			return fmt.Errorf("model contains invalid characters")
		}
	}
	// Reasonable length limit
	if len(model) > 200 {
		return fmt.Errorf("model name too long")
	}
	return nil
}

// validateModelAdaptive performs adaptive model validation with graceful degradation (relaxed for Codex)
func (mv *modelValidator) validateModelAdaptive(model string) error {
	if model == "" {
		return nil // Optional field
	}

	// For Codex, do not enforce naming patterns; only basic safety
	if strings.Contains(model, "$(") || strings.Contains(model, "`") || strings.Contains(model, ";") || strings.Contains(model, "../") || strings.Contains(model, "\\x") {
		return fmt.Errorf("model contains disallowed characters")
	}
	for _, r := range model {
		if r < 32 || r == 127 {
			return fmt.Errorf("model contains invalid characters")
		}
	}
	if len(model) > 200 {
		return fmt.Errorf("model name too long")
	}
	return nil
}

// parseArguments performs two-phase argument parsing to separate CDE flags from codex arguments
func parseArguments(args []string) ParseResult {
	result := ParseResult{
		CCEFlags:   make(map[string]string),
		ClaudeArgs: []string{},
	}

	if len(args) == 0 {
		return result
	}

	// Phase 1: Check for subcommands first
	switch args[0] {
	case "list":
		result.Subcommand = "list"
		return result
	case "add":
		result.Subcommand = "add"
		return result
	case "remove":
		if len(args) < 2 {
			result.Error = fmt.Errorf("remove command requires environment name")
			return result
		}
		result.Subcommand = "remove"
		result.CCEFlags["remove_target"] = args[1]
		return result
	case "help", "--help", "-h":
		result.Subcommand = "help"
		return result
	case "auto":
		result.Subcommand = "auto"
		return result
	}

	// Phase 1: Scan for CDE flags and -- separator
	i := 0
	separatorFound := false

	for i < len(args) {
		arg := args[i]

		// Check for -- separator
		if arg == "--" {
			separatorFound = true
			i++ // Skip the separator itself
			break
		}

		// Check for known CCE flags
		if arg == "--env" || arg == "-e" {
			if i+1 >= len(args) {
				result.Error = fmt.Errorf("flag %s requires a value", arg)
				return result
			}
			result.CCEFlags["env"] = args[i+1]
			i += 2 // Skip flag and its value
			continue
		}

		if arg == "--help" || arg == "-h" {
			result.Subcommand = "help"
			return result
		}

		// If we encounter an unknown flag or argument, stop CCE processing
		break
	}

	// Phase 2: Collect remaining arguments for codex
	if separatorFound || i < len(args) {
		result.ClaudeArgs = args[i:]
	}

	return result
}

// validatePassthroughArgs performs security validation on codex arguments
func validatePassthroughArgs(args []string) error {
	for _, arg := range args {
		// Check for potential command injection patterns
		if strings.Contains(arg, ";") || strings.Contains(arg, "&") ||
			strings.Contains(arg, "|") || strings.Contains(arg, "`") ||
			strings.Contains(arg, "$(") {
			// Allow these in quoted strings, but warn about potential risks
			fmt.Fprintf(os.Stderr, "Warning: Argument contains shell metacharacters: %s\n", arg)
		}

		// Block obvious command injection attempts
		if strings.Contains(arg, "rm -rf") || strings.Contains(arg, "sudo") ||
			strings.Contains(arg, "/etc/passwd") || strings.Contains(arg, "../") {
			return fmt.Errorf("potentially dangerous argument rejected: %s", arg)
		}
	}
	return nil
}

func main() {
	// Check for version flag first
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("cde version %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	if err := handleCommand(os.Args[1:]); err != nil {
		// Enhanced error categorization with clear messaging
		errorType := categorizeError(err)

		switch errorType {
		case "cde_argument":
			fmt.Fprintf(os.Stderr, "CDE Argument Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "Use 'cde help' for usage information.\n")
		case "cde_config":
			fmt.Fprintf(os.Stderr, "CDE Configuration Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "Check your environment configuration with 'cde list'.\n")
		case "codex_execution":
			fmt.Fprintf(os.Stderr, "Codex Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "This error originated from the codex command.\n")
		case "terminal":
			fmt.Fprintf(os.Stderr, "Terminal Compatibility Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "Try using a different terminal or check terminal capabilities.\n")
		case "permission":
			fmt.Fprintf(os.Stderr, "Permission Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "Check file permissions and access rights.\n")
		default:
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}

		// Enhanced error categorization with exit codes
		switch {
		case strings.Contains(err.Error(), "terminal"):
			os.Exit(4) // Terminal compatibility error
		case strings.Contains(err.Error(), "permission"):
			os.Exit(5) // Permission/access error
		case strings.Contains(err.Error(), "configuration"):
			os.Exit(2) // Configuration error
		case strings.Contains(strings.ToLower(err.Error()), "codex"):
			os.Exit(3) // Codex launcher error
		case strings.Contains(err.Error(), "argument parsing"):
			os.Exit(6) // CDE argument parsing error
		case strings.Contains(err.Error(), "argument validation"):
			os.Exit(7) // CDE argument validation error
		default:
			os.Exit(1) // General application error
		}
	}
}

// categorizeError determines the error category for appropriate handling
func categorizeError(err error) string {
	errStr := strings.ToLower(err.Error())

	// CDE argument-related errors
	if strings.Contains(errStr, "argument parsing") ||
		strings.Contains(errStr, "argument validation") ||
		(strings.Contains(errStr, "flag") && !strings.Contains(errStr, "codex")) {
		return "cde_argument"
	}

	// CDE configuration errors
	if strings.Contains(errStr, "configuration") ||
		(strings.Contains(errStr, "environment") && !strings.Contains(errStr, "codex")) {
		return "cde_config"
	}

	// Codex execution errors
	if strings.Contains(errStr, "codex") && (strings.Contains(errStr, "execution") || strings.Contains(errStr, "process")) {
		return "codex_execution"
	}

	// Terminal errors
	if strings.Contains(errStr, "terminal") ||
		strings.Contains(errStr, "tty") ||
		strings.Contains(errStr, "raw mode") {
		return "terminal"
	}

	// Permission errors
	if strings.Contains(errStr, "permission") ||
		strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "not executable") {
		return "permission"
	}

	return "general"
}

// handleCommand processes command line arguments using two-phase parsing and routes to appropriate handlers
func handleCommand(args []string) error {
	// Use new two-phase argument parsing
	parseResult := parseArguments(args)
	if parseResult.Error != nil {
		return fmt.Errorf("argument parsing failed: %w", parseResult.Error)
	}

	// Handle subcommands
	switch parseResult.Subcommand {
	case "list":
		return runList()
	case "add":
		return runAdd()
	case "remove":
		if target, exists := parseResult.CCEFlags["remove_target"]; exists {
			return runRemove(target)
		}
		return fmt.Errorf("remove command requires environment name")
	case "help":
		showHelp()
		return nil
	case "auto":
		// Validate passthrough arguments for security
		if err := validatePassthroughArgs(parseResult.ClaudeArgs); err != nil {
			return fmt.Errorf("argument validation failed: %w", err)
		}
		envName := parseResult.CCEFlags["env"]
		return runAuto(envName, parseResult.ClaudeArgs)
	}

	// Validate passthrough arguments for security
	if err := validatePassthroughArgs(parseResult.ClaudeArgs); err != nil {
		return fmt.Errorf("argument validation failed: %w", err)
	}

	// Handle default behavior with environment selection and codex arguments
	envName := parseResult.CCEFlags["env"]
	return runDefault(envName, parseResult.ClaudeArgs)
}

// showHelp displays usage information including flag passthrough capability
func showHelp() {
	fmt.Println("Codex Env (cde) Launcher")
	fmt.Println("\nUsage:")
	fmt.Println("  cde [command] [options] [-- codex-args...]")
	fmt.Println("\nCommands:")
	fmt.Println("  list                列出所有已配置环境")
	fmt.Println("  add                 新增环境配置（可选模型）")
	fmt.Println("  remove <name>       删除环境配置")
	fmt.Println("  auto                自动批准并使用沙箱（-a never --sandbox workspace-write）")
	fmt.Println("  help                显示帮助")
	fmt.Println("\nOptions:")
	fmt.Println("  -e, --env <name>    选择环境")
	fmt.Println("  -h, --help          显示帮助")
	fmt.Println("\n说明:")
	fmt.Println("  - 所有 CDE 选项之后的参数都会直接透传给 codex 命令。")
	fmt.Println("  - 使用 '--' 明确分隔 CDE 与 codex 参数。")
	fmt.Println("  - 如果环境配置了 model 且未在参数中指定 '-m/--model'，将自动追加 '-m <env.model>'（默认模型示例: gpt-5）。")
	fmt.Println("\n示例:")
	fmt.Println("  cde                              交互式选择并启动 Codex")
	fmt.Println("  cde --env prod                   使用 'prod' 环境启动 Codex")
	fmt.Println("  cde auto -e dev -- mcp           自动批准 + 沙箱，执行 mcp")
	fmt.Println("  cde -e staging -- --help         透传 '--help' 到 codex")
}

// runDefault handles the default behavior: environment selection and Codex launch with arguments
// prepareCodexArgs applies model injection rules to codex args
func prepareCodexArgs(selectedEnv Environment, codexArgs []string) []string {
	// If environment specifies model and user didn't pass -m/--model, prepend it
	hasModelFlag := false
	for i := 0; i < len(codexArgs); i++ {
		if codexArgs[i] == "-m" || codexArgs[i] == "--model" {
			hasModelFlag = true
			break
		}
	}
	if !hasModelFlag && strings.TrimSpace(selectedEnv.Model) != "" {
		codexArgs = append([]string{"-m", selectedEnv.Model}, codexArgs...)
	}
	return codexArgs
}

func runDefault(envName string, codexArgs []string) error {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("configuration loading failed: %w", err)
	}

	var selectedEnv Environment

	if envName != "" {
		// Use specified environment
		index, exists := findEnvironmentByName(config, envName)
		if !exists {
			return fmt.Errorf("environment '%s' not found", envName)
		}
		selectedEnv = config.Environments[index]
	} else {
		// Interactive selection
		selectedEnv, err = selectEnvironment(config)
		if err != nil {
			return fmt.Errorf("environment selection failed: %w", err)
		}
	}

	// Display selected environment
	if _, err := fmt.Printf("Using environment: %s (%s)\n", selectedEnv.Name, selectedEnv.URL); err != nil {
		return fmt.Errorf("failed to display selected environment: %w", err)
	}

	// Prepare final codex args with model injection if needed
	codexArgs = prepareCodexArgs(selectedEnv, codexArgs)

	// Launch Codex with arguments
	return launchCodex(selectedEnv, codexArgs)
}

// runAuto appends auto-approval and sandbox flags then launches Codex
// applyAutoFlags prepends automatic approval and sandbox flags
func applyAutoFlags(args []string) []string {
	return append([]string{"-a", "never", "--sandbox", "workspace-write"}, args...)
}

func runAuto(envName string, codexArgs []string) error {
	autoArgs := applyAutoFlags(codexArgs)
	return runDefault(envName, autoArgs)
}

// runList displays all configured environments
func runList() error {
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("configuration loading failed: %w", err)
	}

	return displayEnvironments(config)
}

// runAdd adds a new environment configuration
func runAdd() error {
	// Load existing configuration
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("configuration loading failed: %w", err)
	}

	// Prompt for new environment details
	env, err := promptForEnvironment(config)
	if err != nil {
		return fmt.Errorf("environment input failed: %w", err)
	}

	// Add environment to configuration
	if err := addEnvironmentToConfig(&config, env); err != nil {
		return fmt.Errorf("failed to add environment: %w", err)
	}

	// Save updated configuration
	if err := saveConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	if _, err := fmt.Printf("Environment '%s' added successfully.\n", env.Name); err != nil {
		return fmt.Errorf("failed to display success message: %w", err)
	}

	return nil
}

// runRemove removes an environment configuration
func runRemove(name string) error {
	// Validate name parameter
	if err := validateName(name); err != nil {
		return fmt.Errorf("invalid environment name: %w", err)
	}

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("configuration loading failed: %w", err)
	}

	// Remove environment from configuration
	if err := removeEnvironmentFromConfig(&config, name); err != nil {
		return fmt.Errorf("failed to remove environment: %w", err)
	}

	// Save updated configuration
	if err := saveConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	if _, err := fmt.Printf("Environment '%s' removed successfully.\n", name); err != nil {
		return fmt.Errorf("failed to display success message: %w", err)
	}

	return nil
}
