.PHONY: build clean test coverage test-verbose test-run test-race test-pkg install fmt lint check help

# Build variables
BINARY_NAME=tkube
BUILD_DIR=build
VERSION?=1.2.0

# Build the application
build:
	@echo "🔨 Building tkube..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/tkube

# Build for multiple platforms
build-all: build-linux build-darwin build-darwin-arm64

build-linux:
	@echo "🔨 Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/tkube

build-darwin:
	@echo "🔨 Building for macOS (Intel)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/tkube

build-darwin-arm64:
	@echo "🔨 Building for macOS (Apple Silicon)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/tkube

# Install locally
install: build
	@echo "📦 Installing tkube..."
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "✅ tkube installed to /usr/local/bin/"

# Run tests
test:
	@echo "🧪 Running tests..."
	go test ./internal/...

# Run tests with coverage
coverage:
	@echo "🧪 Running tests with coverage..."
	go test -coverprofile=coverage.out ./internal/...
	@go tool cover -func=coverage.out | grep total | awk '{print "Coverage: " $$3}'

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running tests with verbose output..."
	go test -v ./internal/...

# Run specific test
test-run:
	@echo "🧪 Running specific test pattern..."
	go test -run $(TEST) ./internal/...

# Run tests with race detection
test-race:
	@echo "🧪 Running tests with race detection..."
	go test -race ./internal/...

# Run tests for a specific package
test-pkg:
	@echo "🧪 Running tests for package $(PKG)..."
	go test ./internal/$(PKG)/...

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "🔍 Linting code..."
	golangci-lint run

# Run all checks
check: fmt lint test

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  build-all      - Build for multiple platforms"
	@echo "  install        - Install locally"
	@echo "  test           - Run tests"
	@echo "  coverage       - Run tests with coverage percentage"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  test-run       - Run specific test pattern (TEST=pattern)"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-pkg       - Run tests for specific package (PKG=package)"
	@echo "  clean          - Clean build artifacts"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  check          - Run all checks (fmt, lint, test)"
	@echo "  help           - Show this help message"