# MySQL Manager - Icon Guide

## Overview

This guide explains how to create and manage application icons for different platforms.

## Icon Requirements

### Windows (.ico)
- **Format**: ICO file
- **Sizes**: Multiple sizes in one file (16x16, 32x32, 48x48, 64x64, 128x128, 256x256)
- **Location**: `build/windows/icon.ico`
- **Color Depth**: 32-bit with alpha channel

### macOS (.icns)
- **Format**: ICNS file
- **Sizes**: Multiple sizes (16x16@1x, 16x16@2x, 32x32@1x, 32x32@2x, 128x128@1x, 128x128@2x, 256x256@1x, 256x256@2x, 512x512@1x, 512x512@2x)
- **Location**: `build/darwin/icon.icns`
- **Color Depth**: 32-bit with alpha channel

### Linux (.png)
- **Format**: PNG file
- **Size**: 512x512 pixels (recommended)
- **Location**: `build/appicon.png`
- **Color Depth**: 32-bit with alpha channel

## Source Icon

The source icon should be:
- **Size**: 1024x1024 pixels minimum
- **Format**: PNG with transparency
- **Design**: Simple, recognizable at small sizes
- **Colors**: Works well on light and dark backgrounds

## Creating Icons

### Method 1: Using Online Tools

1. **Favicon Generator** (https://realfavicongenerator.net/)
   - Upload your 1024x1024 PNG
   - Download all platform icons
   - Place in appropriate directories

2. **App Icon Generator** (https://appicon.co/)
   - Upload source image
   - Generate all sizes
   - Download and organize

### Method 2: Using ImageMagick (Command Line)

#### Install ImageMagick

```bash
# macOS
brew install imagemagick

# Ubuntu/Debian
sudo apt-get install imagemagick

# Windows
choco install imagemagick
```

#### Generate Windows Icon

```bash
convert build/appicon.png \
  -define icon:auto-resize=256,128,64,48,32,16 \
  build/windows/icon.ico
```

#### Generate macOS Icon

```bash
# Create iconset directory
mkdir -p build/darwin/icon.iconset

# Generate all required sizes
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

### Method 3: Using Makefile

```bash
# Generate all icons from build/appicon.png
make icons
```

## Icon Design Guidelines

### General Principles

1. **Simplicity**: Keep the design simple and recognizable
2. **Scalability**: Design should work at 16x16 and 512x512
3. **Contrast**: Ensure good contrast for visibility
4. **Uniqueness**: Make it distinctive and memorable
5. **Relevance**: Reflect the application's purpose

### MySQL Manager Icon Suggestions

Consider these design elements:
- Database cylinder icon
- MySQL dolphin (if licensing allows)
- Gear/settings icon combined with database
- Table/grid icon representing data
- Connection/network icon

### Color Palette

Suggested colors for MySQL Manager:
- **Primary**: #00758F (MySQL blue)
- **Secondary**: #F29111 (MySQL orange)
- **Accent**: #E97826 (Warm orange)
- **Background**: White or transparent

### Testing Icons

Test your icons:
1. **Small sizes**: Verify clarity at 16x16 and 32x32
2. **Large sizes**: Check quality at 256x256 and 512x512
3. **Backgrounds**: Test on light, dark, and colored backgrounds
4. **Platforms**: View on actual Windows, macOS, and Linux systems

## Current Icon Status

### Existing Files

- `build/appicon.png` - ✓ Exists (base icon)
- `build/windows/icon.ico` - ⚠️ Needs generation
- `build/darwin/icon.icns` - ⚠️ Needs generation

### To-Do

1. [ ] Review and optimize `build/appicon.png`
2. [ ] Generate Windows icon: `make icons` or manual conversion
3. [ ] Generate macOS icon: `make icons` or manual conversion
4. [ ] Test icons on all platforms
5. [ ] Create installer/DMG background images (optional)

## Additional Assets

### Installer Graphics

#### Windows NSIS Installer
- **Header**: 150x57 pixels (BMP)
- **Wizard**: 164x314 pixels (BMP)
- **Location**: `build/windows/installer/`

#### macOS DMG Background
- **Size**: 800x400 pixels (PNG)
- **Location**: `build/darwin/dmg-background.png`
- **Design**: Include app icon and arrow to Applications folder

### Creating DMG Background

```bash
# Create a simple DMG background
convert -size 800x400 xc:white \
  -font Arial -pointsize 24 \
  -draw "text 300,200 'Drag to Applications'" \
  build/darwin/dmg-background.png
```

## Icon Verification

After generating icons, verify them:

### Windows
```bash
# View icon info
magick identify build/windows/icon.ico
```

### macOS
```bash
# View icns contents
iconutil -c iconset build/darwin/icon.icns -o temp.iconset
ls -la temp.iconset/
rm -rf temp.iconset
```

### Linux
```bash
# View PNG info
file build/appicon.png
identify build/appicon.png
```

## Troubleshooting

### Icon Not Showing in Windows

1. Clear icon cache:
   ```cmd
   ie4uinit.exe -show
   ```
2. Rebuild with `-clean` flag
3. Verify icon.ico exists and is valid

### Icon Not Showing in macOS

1. Clear icon cache:
   ```bash
   sudo rm -rf /Library/Caches/com.apple.iconservices.store
   sudo find /private/var/folders/ -name com.apple.iconservices -exec rm -rf {} \;
   killall Dock
   ```
2. Verify icon.icns is valid
3. Re-sign the application

### Icon Not Showing in Linux

1. Update icon cache:
   ```bash
   sudo gtk-update-icon-cache /usr/share/icons/hicolor
   ```
2. Verify PNG is in correct location
3. Check desktop entry file

## Resources

- [Apple Human Interface Guidelines - App Icons](https://developer.apple.com/design/human-interface-guidelines/app-icons)
- [Windows App Icon Guidelines](https://docs.microsoft.com/en-us/windows/apps/design/style/iconography/app-icon-design)
- [freedesktop.org Icon Theme Specification](https://specifications.freedesktop.org/icon-theme-spec/icon-theme-spec-latest.html)
- [ImageMagick Documentation](https://imagemagick.org/index.php)

## Quick Reference

```bash
# Generate all icons
make icons

# Generate Windows icon only
convert build/appicon.png -define icon:auto-resize=256,128,64,48,32,16 build/windows/icon.ico

# Generate macOS icon only (macOS only)
# See "Generate macOS Icon" section above

# Verify icon files
file build/windows/icon.ico
file build/darwin/icon.icns
file build/appicon.png
```
