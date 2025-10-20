#!/bin/bash

# webauto Release Binary Build Script
# This script builds cross-platform binaries for webauto plugin
# Platforms: Windows (amd64), macOS (amd64, arm64), Linux (amd64, arm64)

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script configuration
PLUGIN_NAME="webauto"
BUILD_DIR="build"
DIST_DIR="dist"

# Get version from git tag or use default
VERSION=${1:-$(git describe --tags --always 2>/dev/null || echo "dev")}
VERSION=${VERSION#v}  # Remove 'v' prefix if present

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}webauto Release Build Script${NC}"
echo -e "${BLUE}Version: ${VERSION}${NC}"
echo -e "${BLUE}========================================${NC}"
echo

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf "${BUILD_DIR}" "${DIST_DIR}"
mkdir -p "${BUILD_DIR}" "${DIST_DIR}"
echo -e "${GREEN}✓ Build directories created${NC}"
echo

# Platform build configurations
# Format: "GOOS GOARCH OUTPUT_NAME ARCHIVE_EXT"
PLATFORMS=(
    "windows amd64 webauto.exe zip"
    "darwin amd64 webauto tar.gz"
    "darwin arm64 webauto tar.gz"
    "linux amd64 webauto tar.gz"
    "linux arm64 webauto tar.gz"
)

# Build for each platform
for platform_config in "${PLATFORMS[@]}"; do
    read -r GOOS GOARCH OUTPUT ARCHIVE_EXT <<< "$platform_config"

    PLATFORM_DIR="${BUILD_DIR}/${GOOS}-${GOARCH}"
    OUTPUT_PATH="${PLATFORM_DIR}/${OUTPUT}"
    ARCHIVE_NAME="${PLUGIN_NAME}-${VERSION}-${GOOS}-${GOARCH}"

    echo -e "${BLUE}Building for ${GOOS}/${GOARCH}...${NC}"

    # Create platform-specific directory
    mkdir -p "${PLATFORM_DIR}"

    # Build binary
    echo "  - Compiling binary..."
    GOOS="${GOOS}" GOARCH="${GOARCH}" go build \
        -o "${OUTPUT_PATH}" \
        -ldflags "-s -w -X main.Version=${VERSION}" \
        ./cmd/webauto/main.go

    if [ $? -ne 0 ]; then
        echo -e "${RED}✗ Build failed for ${GOOS}/${GOARCH}${NC}"
        exit 1
    fi

    # Get binary size
    BINARY_SIZE=$(du -h "${OUTPUT_PATH}" | cut -f1)
    echo -e "  - Binary size: ${BINARY_SIZE}"

    # Create archive
    echo "  - Creating archive..."
    cd "${PLATFORM_DIR}"

    if [ "${ARCHIVE_EXT}" = "zip" ]; then
        # Windows: use zip
        zip -q "../../${DIST_DIR}/${ARCHIVE_NAME}.zip" "${OUTPUT}"
        ARCHIVE_PATH="../../${DIST_DIR}/${ARCHIVE_NAME}.zip"
    else
        # Unix: use tar.gz
        tar -czf "../../${DIST_DIR}/${ARCHIVE_NAME}.tar.gz" "${OUTPUT}"
        ARCHIVE_PATH="../../${DIST_DIR}/${ARCHIVE_NAME}.tar.gz"
    fi

    cd - > /dev/null

    # Calculate SHA256 checksum
    if command -v shasum > /dev/null; then
        SHA256=$(shasum -a 256 "${DIST_DIR}/${ARCHIVE_NAME}.${ARCHIVE_EXT}" | awk '{print $1}')
    elif command -v sha256sum > /dev/null; then
        SHA256=$(sha256sum "${DIST_DIR}/${ARCHIVE_NAME}.${ARCHIVE_EXT}" | awk '{print $1}')
    else
        echo -e "${YELLOW}  ⚠ Warning: shasum/sha256sum not found, skipping checksum${NC}"
        SHA256="N/A"
    fi

    # Append to checksums file
    echo "${SHA256}  ${ARCHIVE_NAME}.${ARCHIVE_EXT}" >> "${DIST_DIR}/SHA256SUMS.txt"

    echo -e "${GREEN}✓ Built ${GOOS}/${GOARCH} successfully${NC}"
    echo -e "  Archive: ${DIST_DIR}/${ARCHIVE_NAME}.${ARCHIVE_EXT}"
    echo -e "  SHA256: ${SHA256}"
    echo
done

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Build Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo
echo "Built binaries:"
ls -lh "${DIST_DIR}"
echo
echo "SHA256 Checksums:"
cat "${DIST_DIR}/SHA256SUMS.txt"
echo
echo -e "${GREEN}All artifacts are in: ${DIST_DIR}/${NC}"
echo

# Optional: Create release notes template
RELEASE_NOTES="${DIST_DIR}/RELEASE_NOTES.md"
cat > "${RELEASE_NOTES}" << EOF
# webauto v${VERSION}

## Installation

### Using OA CLI (Recommended)

\`\`\`bash
oa plugin install webauto
\`\`\`

### Manual Installation

Download the appropriate archive for your platform:

- **Windows (amd64)**: \`${PLUGIN_NAME}-${VERSION}-windows-amd64.zip\`
- **macOS (Intel)**: \`${PLUGIN_NAME}-${VERSION}-darwin-amd64.tar.gz\`
- **macOS (Apple Silicon)**: \`${PLUGIN_NAME}-${VERSION}-darwin-arm64.tar.gz\`
- **Linux (amd64)**: \`${PLUGIN_NAME}-${VERSION}-linux-amd64.tar.gz\`
- **Linux (arm64)**: \`${PLUGIN_NAME}-${VERSION}-linux-arm64.tar.gz\`

Extract and verify:

\`\`\`bash
# Extract (example for Linux)
tar -xzf ${PLUGIN_NAME}-${VERSION}-linux-amd64.tar.gz

# Verify checksum
sha256sum -c SHA256SUMS.txt

# Install
oa plugin install ./webauto
\`\`\`

## Platform Support

✅ **Windows 10/11** (amd64)
✅ **macOS 11+** (Intel & Apple Silicon)
✅ **Linux Ubuntu 20.04+** (amd64, arm64)

For platform-specific installation instructions, see [Platform Guide](https://github.com/oa-plugins/webauto/blob/main/docs/platform-guide.md).

## What's New

<!-- Add release notes here -->

## Checksums

See \`SHA256SUMS.txt\` for all checksums.

## Documentation

- [README](https://github.com/oa-plugins/webauto#readme)
- [Platform Guide](https://github.com/oa-plugins/webauto/blob/main/docs/platform-guide.md)
- [Architecture](https://github.com/oa-plugins/webauto/blob/main/ARCHITECTURE.md)
- [Issue Tracker](https://github.com/oa-plugins/webauto/issues)

## Support

- **Bug Reports**: [GitHub Issues](https://github.com/oa-plugins/webauto/issues)
- **Feature Requests**: [GitHub Discussions](https://github.com/oa-plugins/webauto/discussions)
EOF

echo -e "${BLUE}Release notes template created: ${RELEASE_NOTES}${NC}"
echo

exit 0
