# Makefile for File Viewer

# Binary name
BINARY_NAME=bullseye

# Build directory
BUILD_DIR=build

# Source directories
CMD_DIR=cmd/bullseye
INTERNAL_DIR=internal
PKG_DIR=pkg

# Go build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build clean run test install uninstall help

# Default target
all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run without building (if already built)
run-only:
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Install to system
install: build
	@echo "Installing $(BINARY_NAME)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

# Uninstall from system
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstallation complete"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	gosec ./...

# Generate documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060

# Dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  build-all     - Build for Linux, macOS, and Windows"
	@echo "  run           - Build and run the application"
	@echo "  run-only      - Run without building (if already built)"
	@echo "  install       - Install to system"
	@echo "  uninstall     - Uninstall from system"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  security      - Check for security vulnerabilities"
	@echo "  docs          - Generate documentation"
	@echo "  deps          - Download dependencies"
	@echo "  deps-update   - Update dependencies"
	@echo "  help          - Show this help message"


