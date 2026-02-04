.PHONY: build test install lint clean help

# Variables
BINARY_NAME=dotkeeper
BINARY_PATH=./cmd/$(BINARY_NAME)
OUTPUT_PATH=./bin/$(BINARY_NAME)
GO=go

# Default target
help:
	@echo "dotkeeper - Makefile targets"
	@echo ""
	@echo "Available targets:"
	@echo "  build       Build the dotkeeper binary"
	@echo "  test        Run tests"
	@echo "  install     Install dotkeeper to GOPATH/bin"
	@echo "  lint        Run linter (golangci-lint)"
	@echo "  clean       Remove build artifacts"
	@echo "  help        Show this help message"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@$(GO) build -o $(OUTPUT_PATH) $(BINARY_PATH)
	@echo "Build complete: $(OUTPUT_PATH)"

# Run tests
test:
	@echo "Running tests..."
	@$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

# Install to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	@$(GO) install $(BINARY_PATH)
	@echo "Installation complete"

# Run linter
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	@golangci-lint run ./...
	@echo "Linting complete"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@$(GO) clean
	@rm -f coverage.out
	@echo "Clean complete"
