# Simplified Build for Claude Code Environment Switcher

.PHONY: build test clean help

# Use a repo-local Go build cache to avoid permission issues in sandboxes
GOCACHE_DIR ?= $(CURDIR)/.gocache
GOENV = GOCACHE=$(GOCACHE_DIR)

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
all: build

# Build the binary
build:
	$(GOENV) go build $(LDFLAGS) -o cde .

# Build with version info (for releases)
build-release:
	$(GOENV) go build $(LDFLAGS) -o cde .
	@echo "Built cde version $(VERSION)"

# Run tests
test:
	@mkdir -p $(GOCACHE_DIR)
	$(GOENV) go test -v ./...

# Test with coverage
test-coverage:
	@mkdir -p $(GOCACHE_DIR)
	$(GOENV) go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	@mkdir -p $(GOCACHE_DIR)
	$(GOENV) go test -bench=. -benchmem ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run security tests
test-security:
	@mkdir -p $(GOCACHE_DIR)
	$(GOENV) go test -v -run TestSecurity ./...

# Quality checks (format, vet, test)
quality: fmt vet test

# Clean build artifacts
clean:
	rm -f cde coverage.out coverage.html

# Install to system PATH
install: build
	sudo mv cde /usr/local/bin/

# Show help
help:
	@echo "Available targets:"
	@echo "  build         Build the CDE binary"
	@echo "  test          Run all tests"
	@echo "  test-coverage Generate test coverage report"
	@echo "  bench         Run performance benchmarks"
	@echo "  fmt           Format Go code"
	@echo "  vet           Run Go vet analysis"
	@echo "  test-security Run security-specific tests"
	@echo "  quality       Run format, vet, and test"
	@echo "  clean         Clean build artifacts"
	@echo "  install       Install to /usr/local/bin"
	@echo "  help          Show this help message"
