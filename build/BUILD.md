# MySQL Manager - Build Guide

## Overview

This guide explains how to build and package the MySQL Manager application for different platforms.

## Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- Platform-specific tools:
  - **Windows**: NSIS installer (optional, for installer creation)
  - **macOS**: Xcode Command Line Tools
  - **Linux**: Standard build tools (gcc, pkg-config)

## Build Commands

### Development Build

```bash
# Run in development mode with hot reload
wails dev
```

### Production Builds

#### Build for Current Platform

```bash
# Build for your current operating system
wails build
```

The output will be in `build/bin/` directory.

#### Build for Windows

```bash
# On Windows or with cross-compilation support
wails build -platform windows/amd64

# With NSIS installer
wails build -platform windows/amd64 -nsis
```

Output:
- Executable: `build/bin/MySQL-Manager.exe`
- Installer (if -nsis): `build/bin/MySQL-Manager-amd64-installer.exe`

#### Build for macOS

```bash
# On macOS
wails build -platform darwin/amd64

# For Apple Silicon (M1/M2)
wails build -platform darwin/arm64

# Universal binary (both Intel and Apple Silicon)
wails build -platform darwin/universal
```

Output:
- App Bundle: `build/bin/MySQL Manager.app`

To create a DMG installer:
```bash
# Install create-dmg if not already installed
brew install create-dmg

# Create DMG
create-dmg \
  --volname "MySQL Manager" \
  --window-pos 200 120 \
  --window-size 800 400 \
  --icon-size 100 \
  --icon "MySQL Manager.app" 200 190 \
  --hide-extension "MySQL Manager.app" \
  --app-drop-link 600 185 \
  "MySQL-Manager-Installer.dmg" \
  "build/bin/MySQL Manager.app"
```

#### Build for Linux

```bash
# Build for Linux AMD64
wails build -platform linux/amd64

# Build for Linux ARM64
wails build -platform linux/arm64
```

Output:
- Executable: `build/bin/mysql-manager`

To create a Debian package:
```bash
# Build with debian package
wails build -platform linux/amd64 -deb
```

Output:
- DEB Package: `build/bin/mysql-manager_1.0.0_amd64.deb`

## Build Options

### Clean Build

```bash
# Clean previous builds
wails build -clean
```

### Skip Frontend Build

```bash
# Skip frontend rebuild (use existing dist)
wails build -skipbindings
```

### Debug Build

```bash
# Build with debug information
wails build -debug
```

### Optimized Build

```bash
# Build with optimizations and smaller binary size
wails build -ldflags "-s -w"
```

## Icon Files

The application uses platform-specific icon files:

- **Windows**: `build/windows/icon.ico` (256x256, .ico format)
- **macOS**: `build/darwin/icon.icns` (512x512@2x, .icns format)
- **Linux**: `build/appicon.png` (512x512, .png format)

### Creating Icon Files

#### From PNG to ICO (Windows)

```bash
# Using ImageMagick
convert build/appicon.png -define icon:auto-resize=256,128,64,48,32,16 build/windows/icon.ico
```

#### From PNG to ICNS (macOS)

```bash
# Create iconset directory
mkdir -p build/darwin/icon.iconset

# Generate different sizes
sips -z 16 16     build/appicon.png --out build/darwin/icon.iconset/icon_16x16.png
sips -z 32 32     build/appicon.png --out build/darwin/icon.iconset/icon_16x16@2x.png
sips -z 32 32     build/appicon.png --out build/darwin/icon.iconset/icon_32x32.png
sips -z 64 64     build/appicon.png --out build/darwin/icon.iconset/icon_32x32@2x.png
sips -z 128 128   build/appicon.png --out build/darwin/icon.iconset/icon_128x128.png
sips -z 256 256   build/appicon.png --out build/darwin/icon.iconset/icon_128x128@2x.png
sips -z 256 256   build/appicon.png --out build/darwin/icon.iconset/icon_256x256.png
sips -z 512 512   build/appicon.png --out build/darwin/icon.iconset/icon_256x256@2x.png
sips -z 512 512   build/appicon.png --out build/darwin/icon.iconset/icon_512x512.png
sips -z 1024 1024 build/appicon.png --out build/darwin/icon.iconset/icon_512x512@2x.png

# Create icns file
iconutil -c icns build/darwin/icon.iconset -o build/darwin/icon.icns

# Clean up
rm -rf build/darwin/icon.iconset
```

## Cross-Platform Building

### From macOS

```bash
# Build for all platforms
wails build -platform darwin/universal
wails build -platform windows/amd64
wails build -platform linux/amd64
```

### From Linux

```bash
# Build for all platforms (requires mingw-w64 for Windows)
wails build -platform linux/amd64
wails build -platform windows/amd64
# macOS cross-compilation is complex, build on macOS if possible
```

### From Windows

```bash
# Build for Windows and Linux
wails build -platform windows/amd64
wails build -platform linux/amd64
# macOS cross-compilation requires additional setup
```

## Build Automation Script

A convenience script is provided for building all platforms:

```bash
# Make script executable
chmod +x build/build-all.sh

# Run build script
./build/build-all.sh
```

## Distribution

### Windows

Distribute either:
- `MySQL-Manager.exe` (standalone executable)
- `MySQL-Manager-amd64-installer.exe` (NSIS installer)

### macOS

Distribute either:
- `MySQL Manager.app` (app bundle, can be zipped)
- `MySQL-Manager-Installer.dmg` (disk image)

For distribution outside the App Store, you may need to sign and notarize the app.

### Linux

Distribute either:
- `mysql-manager` (standalone executable)
- `mysql-manager_1.0.0_amd64.deb` (Debian package)
- Create RPM, AppImage, or Flatpak for other distributions

## Troubleshooting

### Build Fails on Windows

- Ensure you have a C compiler (MinGW-w64 or TDM-GCC)
- Install: `choco install mingw` or download from mingw-w64.org

### Build Fails on macOS

- Install Xcode Command Line Tools: `xcode-select --install`
- Accept Xcode license: `sudo xcodebuild -license accept`

### Build Fails on Linux

- Install build essentials: `sudo apt-get install build-essential pkg-config`
- Install GTK3 development files: `sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev`

### Icon Not Showing

- Verify icon files exist in correct locations
- Rebuild with `-clean` flag
- Check icon file formats and sizes

## File Sizes

Typical build sizes (approximate):

- Windows: ~25-35 MB (executable)
- macOS: ~30-40 MB (app bundle)
- Linux: ~25-35 MB (executable)

Sizes can be reduced with:
- UPX compression (not recommended for production)
- Build flags: `-ldflags "-s -w"` (strips debug info)

## Version Management

Update version in:
1. `wails.json` - `info.productVersion`
2. `wails.json` - `debianPackage.packageVersion`
3. `frontend/package.json` - `version`
4. `README.md` - version references

## Next Steps

After building:
1. Test the application on target platforms
2. Create release notes
3. Upload to distribution channels
4. Update documentation with download links
