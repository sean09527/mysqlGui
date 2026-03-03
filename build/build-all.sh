#!/bin/bash

# MySQL Manager - Multi-Platform Build Script
# This script builds the application for Windows, macOS, and Linux

set -e

echo "======================================"
echo "MySQL Manager - Multi-Platform Build"
echo "======================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if wails is installed
if ! command -v wails &> /dev/null; then
    echo -e "${RED}Error: Wails CLI is not installed${NC}"
    echo "Install with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    exit 1
fi

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf build/bin/*
echo -e "${GREEN}✓ Clean complete${NC}"
echo ""

# Detect current platform
PLATFORM=$(uname -s)
echo "Current platform: $PLATFORM"
echo ""

# Build for current platform first
echo -e "${YELLOW}Building for current platform...${NC}"
wails build -clean
echo -e "${GREEN}✓ Current platform build complete${NC}"
echo ""

# Build for other platforms based on current OS
case "$PLATFORM" in
    "Darwin")
        echo -e "${YELLOW}Building macOS Universal Binary...${NC}"
        wails build -platform darwin/universal
        echo -e "${GREEN}✓ macOS Universal build complete${NC}"
        echo ""
        
        echo -e "${YELLOW}Building for Windows (amd64)...${NC}"
        wails build -platform windows/amd64
        echo -e "${GREEN}✓ Windows build complete${NC}"
        echo ""
        
        echo -e "${YELLOW}Building for Linux (amd64)...${NC}"
        wails build -platform linux/amd64
        echo -e "${GREEN}✓ Linux build complete${NC}"
        echo ""
        ;;
        
    "Linux")
        echo -e "${YELLOW}Building for Linux (amd64)...${NC}"
        wails build -platform linux/amd64
        echo -e "${GREEN}✓ Linux build complete${NC}"
        echo ""
        
        echo -e "${YELLOW}Building for Windows (amd64)...${NC}"
        if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
            wails build -platform windows/amd64
            echo -e "${GREEN}✓ Windows build complete${NC}"
        else
            echo -e "${RED}✗ Windows cross-compilation requires mingw-w64${NC}"
            echo "  Install with: sudo apt-get install mingw-w64"
        fi
        echo ""
        
        echo -e "${YELLOW}Note: macOS builds require building on macOS${NC}"
        echo ""
        ;;
        
    "MINGW"*|"MSYS"*|"CYGWIN"*)
        echo -e "${YELLOW}Building for Windows (amd64)...${NC}"
        wails build -platform windows/amd64
        echo -e "${GREEN}✓ Windows build complete${NC}"
        echo ""
        
        echo -e "${YELLOW}Building for Linux (amd64)...${NC}"
        wails build -platform linux/amd64
        echo -e "${GREEN}✓ Linux build complete${NC}"
        echo ""
        
        echo -e "${YELLOW}Note: macOS builds require building on macOS${NC}"
        echo ""
        ;;
        
    *)
        echo -e "${RED}Unknown platform: $PLATFORM${NC}"
        exit 1
        ;;
esac

# List built files
echo "======================================"
echo "Build Summary"
echo "======================================"
echo ""
echo "Built files in build/bin/:"
ls -lh build/bin/ | grep -v "^total" | grep -v "^d"
echo ""

# Calculate total size
TOTAL_SIZE=$(du -sh build/bin/ | cut -f1)
echo "Total size: $TOTAL_SIZE"
echo ""

echo -e "${GREEN}======================================"
echo "Build Complete!"
echo "======================================${NC}"
echo ""
echo "Next steps:"
echo "1. Test the applications on target platforms"
echo "2. Create installers (see BUILD.md for instructions)"
echo "3. Sign and notarize (macOS) or sign (Windows) if distributing"
echo ""
