#!/bin/bash

# MySQL Manager - Version Management Script
# This script helps manage version numbers across the project

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get current version from wails.json
get_current_version() {
    grep -o '"productVersion": "[^"]*"' "$PROJECT_ROOT/wails.json" | cut -d'"' -f4
}

# Update version in a file
update_version_in_file() {
    local file=$1
    local old_version=$2
    local new_version=$3
    local pattern=$4
    
    if [ -f "$file" ]; then
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS
            sed -i '' "s/$pattern/$new_version/" "$file"
        else
            # Linux
            sed -i "s/$pattern/$new_version/" "$file"
        fi
        echo -e "${GREEN}✓${NC} Updated $file"
    else
        echo -e "${RED}✗${NC} File not found: $file"
    fi
}

# Show current version
show_version() {
    local version=$(get_current_version)
    echo -e "${BLUE}Current version:${NC} $version"
    echo ""
    echo "Version locations:"
    echo "  - wails.json (productVersion)"
    echo "  - wails.json (debianPackage.packageVersion)"
    echo "  - frontend/package.json (version)"
}

# Update version
update_version() {
    local new_version=$1
    local current_version=$(get_current_version)
    
    echo -e "${YELLOW}Updating version from $current_version to $new_version${NC}"
    echo ""
    
    # Update wails.json (productVersion)
    update_version_in_file \
        "$PROJECT_ROOT/wails.json" \
        "$current_version" \
        "$new_version" \
        "\"productVersion\": \"[^\"]*\""
    
    # Update wails.json (debianPackage.packageVersion)
    update_version_in_file \
        "$PROJECT_ROOT/wails.json" \
        "$current_version" \
        "$new_version" \
        "\"packageVersion\": \"[^\"]*\""
    
    # Update frontend/package.json
    update_version_in_file \
        "$PROJECT_ROOT/frontend/package.json" \
        "$current_version" \
        "$new_version" \
        "\"version\": \"[^\"]*\""
    
    echo ""
    echo -e "${GREEN}Version updated successfully!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Review changes: git diff"
    echo "2. Update CHANGELOG.md"
    echo "3. Commit changes: git commit -am 'Bump version to $new_version'"
    echo "4. Create tag: git tag -a v$new_version -m 'Release $new_version'"
    echo "5. Push changes: git push && git push --tags"
}

# Bump version automatically
bump_version() {
    local bump_type=$1
    local current_version=$(get_current_version)
    
    # Parse version
    IFS='.' read -r major minor patch <<< "$current_version"
    
    case "$bump_type" in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            echo -e "${RED}Invalid bump type: $bump_type${NC}"
            echo "Valid types: major, minor, patch"
            exit 1
            ;;
    esac
    
    local new_version="$major.$minor.$patch"
    update_version "$new_version"
}

# Validate version format
validate_version() {
    local version=$1
    if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo -e "${RED}Invalid version format: $version${NC}"
        echo "Version must be in format: MAJOR.MINOR.PATCH (e.g., 1.0.0)"
        exit 1
    fi
}

# Show help
show_help() {
    echo "MySQL Manager - Version Management"
    echo ""
    echo "Usage:"
    echo "  $0 show                    Show current version"
    echo "  $0 set <version>           Set specific version (e.g., 1.2.3)"
    echo "  $0 bump <type>             Bump version (major|minor|patch)"
    echo ""
    echo "Examples:"
    echo "  $0 show                    # Show current version"
    echo "  $0 set 1.2.3               # Set version to 1.2.3"
    echo "  $0 bump major              # 1.0.0 -> 2.0.0"
    echo "  $0 bump minor              # 1.0.0 -> 1.1.0"
    echo "  $0 bump patch              # 1.0.0 -> 1.0.1"
    echo ""
    echo "Semantic Versioning:"
    echo "  MAJOR: Breaking changes"
    echo "  MINOR: New features (backwards compatible)"
    echo "  PATCH: Bug fixes (backwards compatible)"
}

# Main
main() {
    case "${1:-}" in
        show)
            show_version
            ;;
        set)
            if [ -z "${2:-}" ]; then
                echo -e "${RED}Error: Version required${NC}"
                echo "Usage: $0 set <version>"
                exit 1
            fi
            validate_version "$2"
            update_version "$2"
            ;;
        bump)
            if [ -z "${2:-}" ]; then
                echo -e "${RED}Error: Bump type required${NC}"
                echo "Usage: $0 bump <major|minor|patch>"
                exit 1
            fi
            bump_version "$2"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            echo -e "${RED}Invalid command${NC}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"
