# MySQL Manager - Packaging Guide

## Overview

This guide provides detailed instructions for creating distributable packages for MySQL Manager on different platforms.

## Table of Contents

1. [Windows Packaging](#windows-packaging)
2. [macOS Packaging](#macos-packaging)
3. [Linux Packaging](#linux-packaging)
4. [Code Signing](#code-signing)
5. [Distribution Checklist](#distribution-checklist)

## Windows Packaging

### NSIS Installer

Wails can automatically create an NSIS installer:

```bash
wails build -platform windows/amd64 -nsis
```

This creates:
- `MySQL-Manager-amd64-installer.exe` - Full installer with uninstaller

### Customizing NSIS Installer

Create `build/windows/installer/project.nsi` for custom installer settings:

```nsis
!define PRODUCT_NAME "MySQL Manager"
!define PRODUCT_VERSION "1.0.0"
!define PRODUCT_PUBLISHER "Your Company"
!define PRODUCT_WEB_SITE "https://yourwebsite.com"

; Additional customizations
!define MUI_ICON "icon.ico"
!define MUI_UNICON "icon.ico"
!define MUI_WELCOMEFINISHPAGE_BITMAP "installer-banner.bmp"
```

### Portable Version

For a portable version, simply distribute the `MySQL-Manager.exe` file with a README:

```
MySQL-Manager-Portable/
├── MySQL-Manager.exe
├── README.txt
└── LICENSE.txt
```

### MSI Installer (Advanced)

For enterprise deployment, create an MSI using WiX Toolset:

1. Install WiX Toolset
2. Create `product.wxs` file
3. Build with: `candle product.wxs && light product.wixobj`

## macOS Packaging

### App Bundle

The build process creates `MySQL Manager.app` automatically. To distribute:

```bash
# Zip the app bundle
cd build/bin
zip -r MySQL-Manager-macOS.zip "MySQL Manager.app"
```

### DMG Installer

Create a professional DMG installer:

```bash
# Install create-dmg
brew install create-dmg

# Create DMG
create-dmg \
  --volname "MySQL Manager" \
  --volicon "build/darwin/icon.icns" \
  --window-pos 200 120 \
  --window-size 800 400 \
  --icon-size 100 \
  --icon "MySQL Manager.app" 200 190 \
  --hide-extension "MySQL Manager.app" \
  --app-drop-link 600 185 \
  --background "build/darwin/dmg-background.png" \
  "MySQL-Manager-Installer.dmg" \
  "build/bin/MySQL Manager.app"
```

### Code Signing (Required for Distribution)

```bash
# Sign the app
codesign --deep --force --verify --verbose \
  --sign "Developer ID Application: Your Name (TEAM_ID)" \
  --options runtime \
  "build/bin/MySQL Manager.app"

# Verify signature
codesign --verify --verbose "build/bin/MySQL Manager.app"
spctl --assess --verbose "build/bin/MySQL Manager.app"
```

### Notarization (Required for macOS 10.15+)

```bash
# Create a zip for notarization
ditto -c -k --keepParent "build/bin/MySQL Manager.app" "MySQL-Manager.zip"

# Submit for notarization
xcrun notarytool submit "MySQL-Manager.zip" \
  --apple-id "your@email.com" \
  --team-id "TEAM_ID" \
  --password "app-specific-password" \
  --wait

# Staple the notarization ticket
xcrun stapler staple "build/bin/MySQL Manager.app"

# Verify notarization
spctl --assess -vv --type install "build/bin/MySQL Manager.app"
```

### PKG Installer (Alternative)

```bash
# Create PKG installer
pkgbuild --root "build/bin/MySQL Manager.app" \
  --identifier "com.mysqlmanager.app" \
  --version "1.0.0" \
  --install-location "/Applications/MySQL Manager.app" \
  "MySQL-Manager.pkg"

# Sign the PKG
productsign --sign "Developer ID Installer: Your Name (TEAM_ID)" \
  "MySQL-Manager.pkg" \
  "MySQL-Manager-Signed.pkg"
```

## Linux Packaging

### Debian Package (.deb)

Wails can create a DEB package automatically:

```bash
wails build -platform linux/amd64 -deb
```

This uses the configuration in `wails.json`:

```json
"debianPackage": {
  "packageName": "mysql-manager",
  "packageDescription": "A modern MySQL/MariaDB database management tool",
  "packageVersion": "1.0.0",
  "maintainer": "Your Name <your@email.com>",
  "homepage": "https://yourwebsite.com"
}
```

### Manual DEB Creation

Create `build/linux/DEBIAN/control`:

```
Package: mysql-manager
Version: 1.0.0
Section: database
Priority: optional
Architecture: amd64
Maintainer: Your Name <your@email.com>
Description: MySQL Manager
 A modern MySQL/MariaDB database management tool
 with intuitive interface and powerful features.
```

Build the package:

```bash
# Create directory structure
mkdir -p build/linux/mysql-manager/usr/local/bin
mkdir -p build/linux/mysql-manager/usr/share/applications
mkdir -p build/linux/mysql-manager/usr/share/icons/hicolor/512x512/apps

# Copy files
cp build/bin/mysql-manager build/linux/mysql-manager/usr/local/bin/
cp build/appicon.png build/linux/mysql-manager/usr/share/icons/hicolor/512x512/apps/mysql-manager.png

# Create desktop entry
cat > build/linux/mysql-manager/usr/share/applications/mysql-manager.desktop << EOF
[Desktop Entry]
Name=MySQL Manager
Comment=MySQL/MariaDB Database Management Tool
Exec=/usr/local/bin/mysql-manager
Icon=mysql-manager
Terminal=false
Type=Application
Categories=Development;Database;
EOF

# Build DEB
dpkg-deb --build build/linux/mysql-manager mysql-manager_1.0.0_amd64.deb
```

### RPM Package

Create `mysql-manager.spec`:

```spec
Name:           mysql-manager
Version:        1.0.0
Release:        1%{?dist}
Summary:        MySQL/MariaDB Database Management Tool

License:        MIT
URL:            https://yourwebsite.com
Source0:        %{name}-%{version}.tar.gz

%description
A modern MySQL/MariaDB database management tool
with intuitive interface and powerful features.

%prep
%setup -q

%install
mkdir -p %{buildroot}%{_bindir}
install -m 755 mysql-manager %{buildroot}%{_bindir}/mysql-manager

%files
%{_bindir}/mysql-manager

%changelog
* Mon Jan 01 2024 Your Name <your@email.com> - 1.0.0-1
- Initial release
```

Build RPM:

```bash
rpmbuild -ba mysql-manager.spec
```

### AppImage

Create a portable AppImage:

```bash
# Download appimagetool
wget https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage
chmod +x appimagetool-x86_64.AppImage

# Create AppDir structure
mkdir -p MySQL-Manager.AppDir/usr/bin
mkdir -p MySQL-Manager.AppDir/usr/share/applications
mkdir -p MySQL-Manager.AppDir/usr/share/icons/hicolor/512x512/apps

# Copy files
cp build/bin/mysql-manager MySQL-Manager.AppDir/usr/bin/
cp build/appicon.png MySQL-Manager.AppDir/usr/share/icons/hicolor/512x512/apps/mysql-manager.png
cp build/appicon.png MySQL-Manager.AppDir/mysql-manager.png

# Create desktop entry
cat > MySQL-Manager.AppDir/mysql-manager.desktop << EOF
[Desktop Entry]
Name=MySQL Manager
Exec=mysql-manager
Icon=mysql-manager
Type=Application
Categories=Development;Database;
EOF

# Create AppRun
cat > MySQL-Manager.AppDir/AppRun << 'EOF'
#!/bin/bash
SELF=$(readlink -f "$0")
HERE=${SELF%/*}
export PATH="${HERE}/usr/bin:${PATH}"
exec "${HERE}/usr/bin/mysql-manager" "$@"
EOF
chmod +x MySQL-Manager.AppDir/AppRun

# Build AppImage
./appimagetool-x86_64.AppImage MySQL-Manager.AppDir MySQL-Manager-x86_64.AppImage
```

### Flatpak (Advanced)

Create `com.mysqlmanager.app.yml`:

```yaml
app-id: com.mysqlmanager.app
runtime: org.freedesktop.Platform
runtime-version: '22.08'
sdk: org.freedesktop.Sdk
command: mysql-manager
finish-args:
  - --share=network
  - --socket=x11
  - --socket=wayland
  - --filesystem=home
modules:
  - name: mysql-manager
    buildsystem: simple
    build-commands:
      - install -D mysql-manager /app/bin/mysql-manager
    sources:
      - type: file
        path: build/bin/mysql-manager
```

Build Flatpak:

```bash
flatpak-builder --repo=repo build-dir com.mysqlmanager.app.yml
flatpak build-bundle repo mysql-manager.flatpak com.mysqlmanager.app
```

## Code Signing

### Windows Code Signing

```bash
# Using signtool (Windows SDK)
signtool sign /f certificate.pfx /p password /t http://timestamp.digicert.com MySQL-Manager.exe

# Verify signature
signtool verify /pa MySQL-Manager.exe
```

### macOS Code Signing

See macOS Packaging section above.

### Linux Code Signing

Linux packages are typically signed with GPG:

```bash
# Sign DEB package
dpkg-sig --sign builder mysql-manager_1.0.0_amd64.deb

# Sign RPM package
rpm --addsign mysql-manager-1.0.0-1.x86_64.rpm
```

## Distribution Checklist

Before distributing your application:

### Pre-Release Testing

- [ ] Test on clean Windows 10/11 installation
- [ ] Test on macOS 10.15+ (Intel and Apple Silicon)
- [ ] Test on Ubuntu 20.04/22.04 LTS
- [ ] Test on other Linux distributions (Fedora, Arch, etc.)
- [ ] Verify all features work correctly
- [ ] Test database connections (MySQL 5.7, 8.0, MariaDB)
- [ ] Test SSH tunnel connections
- [ ] Verify data import/export functionality
- [ ] Test schema synchronization

### Security

- [ ] Code signed (Windows and macOS)
- [ ] Notarized (macOS)
- [ ] No hardcoded credentials
- [ ] Secure password storage verified
- [ ] SQL injection protection verified
- [ ] Dependencies scanned for vulnerabilities

### Documentation

- [ ] README.md updated
- [ ] User guide created
- [ ] Installation instructions for each platform
- [ ] Troubleshooting guide
- [ ] License file included
- [ ] Changelog/Release notes

### Legal

- [ ] License file included
- [ ] Third-party licenses documented
- [ ] Privacy policy (if collecting data)
- [ ] Terms of service (if applicable)

### Distribution

- [ ] GitHub Release created
- [ ] Download links tested
- [ ] File sizes documented
- [ ] SHA256 checksums provided
- [ ] Installation instructions verified
- [ ] Update mechanism tested (if applicable)

### Marketing

- [ ] Screenshots prepared
- [ ] Demo video created
- [ ] Website updated
- [ ] Social media announcement
- [ ] Product Hunt launch (optional)

## File Size Optimization

To reduce binary size:

```bash
# Build with size optimization
wails build -ldflags "-s -w"

# Use UPX compression (use with caution)
upx --best --lzma build/bin/MySQL-Manager.exe
```

## Checksums

Generate checksums for verification:

```bash
# SHA256
sha256sum build/bin/* > SHA256SUMS

# MD5 (legacy)
md5sum build/bin/* > MD5SUMS
```

## Automated Release Script

Create `build/release.sh`:

```bash
#!/bin/bash
VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Usage: ./release.sh <version>"
    exit 1
fi

# Update version
sed -i "s/\"productVersion\": \".*\"/\"productVersion\": \"$VERSION\"/" wails.json

# Build all platforms
./build/build-all.sh

# Create checksums
cd build/bin
sha256sum * > SHA256SUMS
cd ../..

# Create release notes
echo "Release $VERSION" > RELEASE_NOTES.md
echo "" >> RELEASE_NOTES.md
echo "## Changes" >> RELEASE_NOTES.md
git log --oneline $(git describe --tags --abbrev=0)..HEAD >> RELEASE_NOTES.md

echo "Release $VERSION prepared!"
echo "Next steps:"
echo "1. Review RELEASE_NOTES.md"
echo "2. Create git tag: git tag -a v$VERSION -m 'Release $VERSION'"
echo "3. Push tag: git push origin v$VERSION"
echo "4. Upload files from build/bin/ to GitHub Release"
```

## Support

For build issues, check:
- [Wails Documentation](https://wails.io/docs/introduction)
- [GitHub Issues](https://github.com/yourusername/mysql-manager/issues)
- Build logs in `build/` directory
