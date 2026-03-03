# MySQL Manager - Makefile

.PHONY: help dev build build-all clean test install-deps icons

# Default target
help:
	@echo "MySQL Manager - Build Commands"
	@echo ""
	@echo "Development:"
	@echo "  make dev          - Run in development mode with hot reload"
	@echo "  make install-deps - Install all dependencies"
	@echo ""
	@echo "Building:"
	@echo "  make build        - Build for current platform"
	@echo "  make build-all    - Build for all platforms"
	@echo "  make build-windows - Build for Windows"
	@echo "  make build-macos  - Build for macOS (Universal)"
	@echo "  make build-linux  - Build for Linux"
	@echo ""
	@echo "Packaging:"
	@echo "  make package-windows - Build Windows with NSIS installer"
	@echo "  make package-macos   - Build macOS and create DMG"
	@echo "  make package-linux   - Build Linux and create DEB package"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make test         - Run all tests"
	@echo "  make icons        - Generate platform-specific icons"
	@echo ""

# Development
dev:
	wails dev

install-deps:
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "Dependencies installed!"

# Building
build:
	wails build

build-all:
	@chmod +x build/build-all.sh
	@./build/build-all.sh

build-windows:
	wails build -platform windows/amd64

build-macos:
	wails build -platform darwin/universal

build-linux:
	wails build -platform linux/amd64

# Packaging
package-windows:
	wails build -platform windows/amd64 -nsis

package-macos: build-macos
	@echo "Creating DMG installer..."
	@if command -v create-dmg >/dev/null 2>&1; then \
		create-dmg \
			--volname "MySQL Manager" \
			--window-pos 200 120 \
			--window-size 800 400 \
			--icon-size 100 \
			--icon "MySQL Manager.app" 200 190 \
			--hide-extension "MySQL Manager.app" \
			--app-drop-link 600 185 \
			"build/bin/MySQL-Manager-Installer.dmg" \
			"build/bin/MySQL Manager.app"; \
	else \
		echo "create-dmg not found. Install with: brew install create-dmg"; \
	fi

package-linux:
	wails build -platform linux/amd64 -deb

# Utilities
clean:
	@echo "Cleaning build artifacts..."
	rm -rf build/bin/*
	rm -rf frontend/dist/*
	rm -rf frontend/wailsjs/go/*
	@echo "Clean complete!"

test:
	@echo "Running backend tests..."
	go test ./backend/... -v
	@echo ""
	@echo "Running frontend tests..."
	cd frontend && npm run test
	@echo ""
	@echo "All tests complete!"

test-backend:
	go test ./backend/... -v -cover

test-frontend:
	cd frontend && npm run test

# Icon generation (requires ImageMagick and iconutil on macOS)
icons:
	@echo "Generating platform-specific icons..."
	@if [ -f build/appicon.png ]; then \
		echo "Generating Windows icon..."; \
		convert build/appicon.png -define icon:auto-resize=256,128,64,48,32,16 build/windows/icon.ico 2>/dev/null || echo "ImageMagick not found, skipping Windows icon"; \
		echo "Generating macOS icon..."; \
		if [ "$$(uname)" = "Darwin" ]; then \
			mkdir -p build/darwin/icon.iconset; \
			sips -z 16 16     build/appicon.png --out build/darwin/icon.iconset/icon_16x16.png; \
			sips -z 32 32     build/appicon.png --out build/darwin/icon.iconset/icon_16x16@2x.png; \
			sips -z 32 32     build/appicon.png --out build/darwin/icon.iconset/icon_32x32.png; \
			sips -z 64 64     build/appicon.png --out build/darwin/icon.iconset/icon_32x32@2x.png; \
			sips -z 128 128   build/appicon.png --out build/darwin/icon.iconset/icon_128x128.png; \
			sips -z 256 256   build/appicon.png --out build/darwin/icon.iconset/icon_128x128@2x.png; \
			sips -z 256 256   build/appicon.png --out build/darwin/icon.iconset/icon_256x256.png; \
			sips -z 512 512   build/appicon.png --out build/darwin/icon.iconset/icon_256x256@2x.png; \
			sips -z 512 512   build/appicon.png --out build/darwin/icon.iconset/icon_512x512.png; \
			sips -z 1024 1024 build/appicon.png --out build/darwin/icon.iconset/icon_512x512@2x.png; \
			iconutil -c icns build/darwin/icon.iconset -o build/darwin/icon.icns; \
			rm -rf build/darwin/icon.iconset; \
			echo "macOS icon generated!"; \
		else \
			echo "macOS icon generation requires macOS"; \
		fi; \
		echo "Icons generated!"; \
	else \
		echo "Error: build/appicon.png not found"; \
	fi

# Version bump
bump-version:
	@read -p "Enter new version (current: $$(grep 'productVersion' wails.json | cut -d'"' -f4)): " version; \
	sed -i.bak "s/\"productVersion\": \".*\"/\"productVersion\": \"$$version\"/" wails.json; \
	sed -i.bak "s/\"packageVersion\": \".*\"/\"packageVersion\": \"$$version\"/" wails.json; \
	sed -i.bak "s/\"version\": \".*\"/\"version\": \"$$version\"/" frontend/package.json; \
	rm -f wails.json.bak frontend/package.json.bak; \
	echo "Version updated to $$version"

# Quick build and run
run: build
	./build/bin/MySQL-Manager
