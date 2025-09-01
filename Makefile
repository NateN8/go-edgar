# EDGAR CLI Tool Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=edgar
BINARY_PATH=bin/$(BINARY_NAME)
CMD_PATH=./cmd/edgar

# Test parameters
TEST_TIMEOUT=30s
INTEGRATION_TEST_TIMEOUT=5m

.PHONY: all build clean test test-unit test-integration test-all coverage benchmark help deps tidy

# Default target
all: deps build test

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_PATH) $(CMD_PATH)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_PATH)
	rm -f coverage.out
	rm -f coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Tidy go modules
tidy:
	@echo "Tidying go modules..."
	$(GOMOD) tidy

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -timeout $(TEST_TIMEOUT) -short ./...

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	INTEGRATION_TESTS=true $(GOTEST) -v -timeout $(INTEGRATION_TEST_TIMEOUT) -tags=integration ./pkg/edgar/

# Run all tests
test-all: test-unit test-integration

# Run tests with default settings (unit tests only)
test: test-unit

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -timeout $(TEST_TIMEOUT) -short -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -v -timeout $(TEST_TIMEOUT) -bench=. -benchmem ./...

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	$(GOTEST) -v -timeout $(TEST_TIMEOUT) -short -race ./...

# Run tests verbosely
test-verbose:
	@echo "Running tests verbosely..."
	$(GOTEST) -v -timeout $(TEST_TIMEOUT) -short ./...

# Install the binary to $GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BINARY_PATH) $(GOPATH)/bin/

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Run linter (requires golangci-lint to be installed)
lint:
	@echo "Running linter..."
	golangci-lint run

# Vet the code
vet:
	@echo "Vetting code..."
	$(GOCMD) vet ./...

# Run full validation (format, vet, lint, test)
validate: fmt vet test

# Quick test for development
quick-test:
	@echo "Running quick tests..."
	$(GOTEST) -timeout 10s -short ./pkg/edgar/

# Test specific package
test-client:
	@echo "Testing client package..."
	$(GOTEST) -v -timeout $(TEST_TIMEOUT) -short ./pkg/edgar/

test-main:
	@echo "Testing main package..."
	$(GOTEST) -v -timeout $(TEST_TIMEOUT) -short ./cmd/edgar/

# Run example commands (requires a valid CIK)
examples: build
	@echo "Running example commands..."
	@echo "Note: These examples use Apple Inc. (CIK: 0000320193)"
	@echo "Running basic cash flow analysis..."
	./$(BINARY_PATH) -cik 0000320193 || true
	@echo ""
	@echo "Running quarterly cash flow analysis..."
	./$(BINARY_PATH) -cik 0000320193 -quarterly || true
	@echo ""
	@echo "Running EBITDA analysis..."
	./$(BINARY_PATH) -cik 0000320193 -ebitda || true

# Development setup
dev-setup: deps
	@echo "Setting up development environment..."
	@echo "Installing development tools..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Docker build (if needed in the future)
docker-build:
	@echo "Building Docker image..."
	docker build -t edgar-cli .

# Generate test coverage badge (requires additional tools)
coverage-badge: coverage
	@echo "Coverage report available at coverage.html"

# Watch mode for tests (requires entr or similar tool)
test-watch:
	@echo "Watching for changes and running tests..."
	find . -name "*.go" | entr -c make test-unit

# Performance profiling
profile-cpu:
	@echo "Running CPU profiling..."
	$(GOTEST) -cpuprofile cpu.prof -bench=. ./pkg/edgar/
	$(GOCMD) tool pprof cpu.prof

profile-mem:
	@echo "Running memory profiling..."
	$(GOTEST) -memprofile mem.prof -bench=. ./pkg/edgar/
	$(GOCMD) tool pprof mem.prof

# Check for security vulnerabilities (requires gosec)
security:
	@echo "Running security check..."
	gosec ./...

# Release build with optimizations
release: clean
	@echo "Building release version..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o $(BINARY_PATH)-linux-amd64 $(CMD_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o $(BINARY_PATH)-darwin-amd64 $(CMD_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o $(BINARY_PATH)-windows-amd64.exe $(CMD_PATH)

# Help target
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run unit tests"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-all       - Run all tests"
	@echo "  coverage       - Run tests with coverage report"
	@echo "  benchmark      - Run benchmarks"
	@echo "  test-race      - Run tests with race detection"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  vet            - Vet the code"
	@echo "  validate       - Run full validation (fmt, vet, test)"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  deps           - Download dependencies"
	@echo "  tidy           - Tidy go modules"
	@echo "  examples       - Run example commands"
	@echo "  dev-setup      - Setup development environment"
	@echo "  release        - Build release binaries for multiple platforms"
	@echo "  help           - Show this help message"
