#!/bin/bash

# Build script for mdatlas
# This script builds the project with proper version information

set -e

# Get version from git or use default
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
BINARY_NAME="mdatlas"

# Build flags
LDFLAGS="-X main.version=${VERSION} -X main.buildDate=${BUILD_DATE}"

echo "Building ${BINARY_NAME} version ${VERSION}..."

# Create bin directory
mkdir -p bin

# Build the binary
go build -ldflags="${LDFLAGS}" -o bin/${BINARY_NAME} cmd/mdatlas/main.go

echo "Build completed successfully!"
echo "Binary: bin/${BINARY_NAME}"
echo "Version: ${VERSION}"
echo "Build date: ${BUILD_DATE}"