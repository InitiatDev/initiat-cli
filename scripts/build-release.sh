#!/bin/bash

# Build script for creating release binaries
# Usage: ./scripts/build-release.sh [version]

set -e

VERSION=${1:-"dev"}
OUTPUT_DIR="dist"

echo "🏗️  Building Initiat CLI v${VERSION}"

# Install Linux dependencies if on Linux
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "📦 Installing Linux dependencies for clipboard support..."
    if command -v apt-get &> /dev/null; then
        sudo apt-get update
        sudo apt-get install -y xvfb libx11-dev x11-utils libegl1-mesa-dev libgles2-mesa-dev libxrandr-dev libxinerama-dev libxcursor-dev libxi-dev
        echo "🖥️  Starting Xvfb for headless X11 support..."
        # Start Xvfb and wait for it to be ready
        Xvfb :0 -screen 0 1024x768x24 > /dev/null 2>&1 &
        MAX_ATTEMPTS=120 # About 60 seconds
        COUNT=0
        echo -n "Waiting for Xvfb to be ready..."
        while ! xdpyinfo -display "${DISPLAY}" >/dev/null 2>&1; do
            echo -n "."
            sleep 0.50s
            COUNT=$(( COUNT + 1 ))
            if [ "${COUNT}" -ge "${MAX_ATTEMPTS}" ]; then
                echo "  Gave up waiting for X server on ${DISPLAY}"
                exit 1
            fi
        done
        echo "Done - Xvfb is ready!"
    elif command -v yum &> /dev/null; then
        sudo yum install -y libX11-devel libXrandr-devel libXinerama-devel libXcursor-devel libXi-devel
    elif command -v dnf &> /dev/null; then
        sudo dnf install -y libX11-devel libXrandr-devel libXinerama-devel libXcursor-devel libXi-devel
    elif command -v pacman &> /dev/null; then
        sudo pacman -S --noconfirm libx11 libxrandr libxinerama libxcursor libxi
    else
        echo "⚠️  Warning: Could not detect package manager. You may need to install X11 development libraries manually."
        echo "   Required packages: libx11-dev, libxrandr-dev, libxinerama-dev, libxcursor-dev, libxi-dev"
    fi
fi

# Clean and create output directory
rm -rf ${OUTPUT_DIR}
mkdir -p ${OUTPUT_DIR}

# Build binaries
echo "📦 Building binaries..."

host_goos="$(go env GOOS)"
host_goarch="$(go env GOARCH)"

if [ -n "${PLATFORMS}" ]; then
    IFS=',' read -r -a platforms <<< "${PLATFORMS}"
else
    platforms=("${host_goos}/${host_goarch}")
fi

for platform in "${platforms[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    
    output_name="initiat-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo "  → ${GOOS}/${GOARCH}"
    
    if [ "${GOOS}/${GOARCH}" != "${host_goos}/${host_goarch}" ] && [ "${ALLOW_CROSS}" != "1" ]; then
        echo "❌ Refusing to cross-compile ${GOOS}/${GOARCH} from ${host_goos}/${host_goarch} (set ALLOW_CROSS=1 to override)"
        exit 1
    fi

    CGO_ENABLED=1 GOOS=$GOOS GOARCH=$GOARCH go build \
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
