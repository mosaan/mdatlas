#!/bin/bash

# Release script for mdatlas
# This script builds cross-platform binaries for release

set -e

# Get version from git or use default
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
BINARY_NAME="mdatlas"

# Build flags
LDFLAGS="-X main.version=${VERSION} -X main.buildDate=${BUILD_DATE}"

echo "Building ${BINARY_NAME} version ${VERSION} for multiple platforms..."

# Create bin directory
mkdir -p bin

# Build for different platforms
platforms=(
    "linux/amd64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r -a platform_split <<< "$platform"
    GOOS="${platform_split[0]}"
    GOARCH="${platform_split[1]}"
    
    output_name="${BINARY_NAME}-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo "Building for ${GOOS}/${GOARCH}..."
    
    env GOOS="$GOOS" GOARCH="$GOARCH" go build \
        -ldflags="${LDFLAGS}" \
        -o "bin/${output_name}" \
        cmd/mdatlas/main.go
    
    if [ $? -ne 0 ]; then
        echo "Failed to build for ${GOOS}/${GOARCH}"
        exit 1
    fi
done

echo "Release builds completed successfully!"
echo "Built binaries:"
ls -la bin/

# Generate checksums
echo "Generating checksums..."
cd bin
sha256sum * > checksums.txt
cd ..

echo "Checksums generated in bin/checksums.txt"