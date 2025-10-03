#!/bin/bash

# Build script for creating release binaries
# Usage: ./scripts/build-release.sh [version]

set -e

VERSION=${1:-"dev"}
OUTPUT_DIR="dist"

echo "🏗️  Building Initiat CLI v${VERSION}"

# Clean and create output directory
rm -rf ${OUTPUT_DIR}
mkdir -p ${OUTPUT_DIR}

# Build for multiple platforms
echo "📦 Building binaries..."

platforms=(
    "darwin/amd64"
    "darwin/arm64" 
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    
    output_name="initiat-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo "  → ${GOOS}/${GOARCH}"
    
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X github.com/InitiatDev/initiat-cli/cmd.version=${VERSION}" \
        -o "${OUTPUT_DIR}/${output_name}" \
        .
done

echo "✅ Binaries built in ${OUTPUT_DIR}/"
echo ""
echo "📋 Release files:"
ls -la ${OUTPUT_DIR}/

echo ""
echo "🚀 To create archives:"
echo "  cd ${OUTPUT_DIR}"
echo "  tar -czf initiat-darwin-amd64.tar.gz initiat-darwin-amd64"
echo "  tar -czf initiat-darwin-arm64.tar.gz initiat-darwin-arm64"
echo "  tar -czf initiat-linux-amd64.tar.gz initiat-linux-amd64"
echo "  tar -czf initiat-linux-arm64.tar.gz initiat-linux-arm64"
echo "  zip initiat-windows-amd64.zip initiat-windows-amd64.exe"
