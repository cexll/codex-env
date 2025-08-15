# CDE Maintainers Notes

This file provides guidance to Codex Env (cde) maintainers when working with code in this repository.

## Project Overview

Codex Env (CDE) is a lightweight Go CLI tool that manages multiple Codex/OpenAI API endpoint configurations, allowing seamless switching between different environments (production, staging, custom API providers, etc.). The tool acts as a wrapper around the `codex` CLI, injecting appropriate environment variables before launching.

## Architecture

This simplified implementation follows KISS principles with just 4 core Go files:

- **`main.go`** (580+ lines): CLI interface, command routing, model validation (permissive), and **flag passthrough system**
- **`config.go`** (367 lines): Configuration management with atomic file operations, backup/recovery, and validation  
- **`ui.go`** (1000+ lines): User interface with **ANSI-free display management** and 4-tier progressive fallback
- **`launcher.go`** (174 lines): Process execution with comprehensive error handling and argument forwarding

### Key Design Patterns

**Configuration Management**: Uses atomic file operations (temp file + rename) with proper permissions (0600/0700). Includes automatic backup and recovery for corrupted configs.

**Terminal Compatibility**: 4-tier progressive fallback system with **ANSI-free display management**:
1. Full interactive (stateful rendering with arrow navigation)
2. Basic interactive (ANSI-free display with arrow support)  
3. Numbered selection (fallback for limited terminals)
4. Headless mode (CI/automation environments)

**Flag Passthrough System**: Two-phase argument parsing that separates CDE flags from codex arguments, allowing transparent command forwarding like `cde auto -e dev -- mcp` or `cde -- --help`.

**Display State Management**: Stateful rendering system with DisplayState tracking, TextPositioner for universal cursor control, and LineRenderer with differential updates.

**Model Validation**: Permissive model validation with basic safety only. `CCE_MODEL_PATTERNS` and `CCE_MODEL_STRICT` are supported but strict mode is disabled by default.

**Error Handling**: Structured error context with recovery suggestions and enhanced exit codes (4=terminal, 5=permission, 6=argument parsing, 7=argument validation).

## Recent Enhancements (2024)

### Flag Passthrough System
- **Two-phase argument parsing** separating CCE flags from Claude arguments
- **Support for `--` separator** for explicit argument separation
- **Security validation** preventing command injection while preserving functionality
- **Enhanced help system** with comprehensive flag passthrough examples

### ANSI-Free Display Management
- **DisplayState management** tracking screen content and changes
- **TextPositioner** providing universal cursor control using carriage return and padding
- **LineRenderer** with stateful menu rendering and differential updates
- **Terminal width detection** with responsive content formatting
- **Smart truncation algorithms** preserving essential information visibility

### UI Layout Improvements
- **Responsive design** adapting to terminal width (20-300+ columns tested)
- **Content overflow prevention** with guaranteed width compliance
- **Progressive fallback enhancement** maintaining compatibility across terminal types
- **Display stacking fix** preventing menu content accumulation during navigation

## Common Development Commands

### Build and Test
```bash
# Build binary
make build              # or: go build -o cde .

# Run tests
make test               # or: go test -v ./...
make test-coverage      # HTML coverage report
make test-security      # Security-specific tests
make bench              # Performance benchmarks

# Code quality
make quality            # fmt + vet + test
make fmt                # Format code
make vet                # Static analysis
```

### Installation and Usage
```bash
# Install to system PATH
make install            # or: sudo mv cde /usr/local/bin/

# Basic usage
./cde                   # Interactive environment selection
./cde --env production  # Use specific environment
./cde list              # List environments
./cde add               # Add new environment
./cde remove <name>     # Remove environment

# Flag passthrough examples
cde auto -e dev -- mcp         # Auto-approve + sandbox
cde --env staging -- proto     # Use staging env, pass to codex
cde -- --help                  # Show codex help (-- separates flags)
```

## Testing Strategy

The project has **95%+ test coverage** across multiple categories:

- **Unit Tests**: `main_test.go`, `config_test.go`, `ui_test.go`, `launcher_test.go`
- **Integration Tests**: `integration_test.go` - End-to-end workflows
- **Security Tests**: `security_test.go` - File permissions and input validation
- **Error Recovery**: `error_recovery_test.go` - Corrupted config handling
- **Platform Compatibility**: `platform_compatibility_test.go` - Cross-platform functionality
- **Enhancement Tests**: `enhancement_*_test.go` - Feature-specific test suites
- **Performance**: `performance_test.go` - Benchmarks for critical operations
- **Terminal Display**: `terminal_display_fix_test.go`, `ui_layout_test.go` - Display management
- **Display Stacking**: `display_stacking_fix_test.go` - Navigation behavior

## Security Implementation

- **API Key Protection**: Terminal raw mode input, masked display (first 6 + last 4 chars)
- **File Security**: Proper permissions (600 for files, 700 for directories)
- **Input Validation**: URL validation, API key format checking, name sanitization
- **Command Injection Prevention**: Argument validation with shell metacharacter detection
- **Process Isolation**: Clean environment variable handling with secure argument forwarding

## Configuration Structure

Environments stored in `~/.codex-env/config.json`:
```json
{
  "environments": [
    {
      "name": "production",
      "url": "https://api.openai.com/v1",
      "api_key": "sk-xxxxx",
      "model": "gpt-5"
    }
  ]
}
```

## Dependencies

**Minimal external dependencies:**
- `golang.org/x/term`: Secure terminal input (hidden API key entry)
- Go standard library for all other functionality

**No external CLI frameworks** - uses standard `flag` package only.

## Environment Configuration

**Model Validation Configuration:**
- `CCE_MODEL_PATTERNS`: Comma-separated custom regex patterns
- `CCE_MODEL_STRICT`: Set to "false" for permissive mode with warnings

## Troubleshooting

**Common Issues:**
1. "codex not found in PATH" - Install via `npm install -g @openai/codex` and ensure `codex --version` works
2. Permission denied - Check `~/.codex-env/` has 700 permissions
3. API key validation - No strict format enforcement; only unsafe characters rejected
4. Display issues - Try different terminal or check TERM environment variable
5. Flag not recognized - Use `--` to separate CCE flags from Claude flags

**Debug Commands:**
```bash
./cce list              # Test configuration loading
which codex             # Verify Codex CLI installation
ls -la ~/.codex-env/    # Check file permissions
./cde --help            # Show comprehensive usage including flag passthrough
```

## Development Principles

1. **KISS Principle**: Simple, direct implementations without unnecessary abstractions
2. **Security First**: Protect API keys and user data at all times  
3. **Comprehensive Error Handling**: All operations include descriptive error messages
4. **Backward Compatibility**: Existing configuration files work without modification
5. **Cross-Platform Support**: Works on macOS, Linux, and Windows
6. **Terminal Agnostic**: ANSI-free core functionality with enhanced features for capable terminals
7. **Performance Focused**: Sub-microsecond operations with minimal memory overhead
