.PHONY: build test clean install release help

BINARY_NAME=mdatlas
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags="-X main.version=${VERSION} -X main.buildDate=${BUILD_DATE}"

# Default target
all: build

# Build the binary
build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p bin
ifeq ($(OS),Windows_NT)
	go build ${LDFLAGS} -o bin/${BINARY_NAME}.exe cmd/mdatlas/main.go
else
	go build ${LDFLAGS} -o bin/${BINARY_NAME} cmd/mdatlas/main.go
endif

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install the binary to GOPATH/bin
install:
	@echo "Installing ${BINARY_NAME}..."
	go install ${LDFLAGS} cmd/mdatlas/main.go

# Cross-compile for multiple platforms
release:
	@echo "Building release binaries..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 cmd/mdatlas/main.go
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-amd64 cmd/mdatlas/main.go
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-arm64 cmd/mdatlas/main.go
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-windows-amd64.exe cmd/mdatlas/main.go

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping lint"; \
	fi

# Run the application
run:
	@echo "Running ${BINARY_NAME}..."
	go run cmd/mdatlas/main.go

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  release       - Build release binaries for multiple platforms"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  run           - Run the application"
	@echo "  help          - Show this help message"