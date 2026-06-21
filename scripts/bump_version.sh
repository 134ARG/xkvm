#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# File paths
MAKEFILE="$PROJECT_ROOT/Makefile"
OTA_GO="$PROJECT_ROOT/ota.go"
CARGO_TOML="$PROJECT_ROOT/ui/src-tauri/Cargo.toml"
SYNC_SCRIPT="$PROJECT_ROOT/ui/scripts/sync-version.cjs"
PACKAGING_VERSION="$PROJECT_ROOT/packaging/version.txt"
UI_PACKAGE_LOCK="$PROJECT_ROOT/ui/package-lock.json"
CARGO_LOCK="$PROJECT_ROOT/ui/src-tauri/Cargo.lock"

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Function to extract current version from a file
get_makefile_version() {
    grep -E "^VERSION\s*:=\s*" "$MAKEFILE" | sed -E 's/^VERSION\s*:=\s*(.+)/\1/' | tr -d ' '
}

get_ota_version() {
    grep -E 'var builtAppVersion = "' "$OTA_GO" | sed -E 's/.*var builtAppVersion = "(.+)"/\1/'
}

get_cargo_version() {
    grep -E '^version = "' "$CARGO_TOML" | head -1 | sed -E 's/^version = "(.+)"/\1/'
}

# Version of the local "app" package as pinned in Cargo.lock
get_cargo_lock_version() {
    sed -nE '/^name = "app"$/{n;s/^version = "(.+)"/\1/p;}' "$CARGO_LOCK"
}

# Validate version format (supports semver with optional +dev or -dev suffix)
validate_version() {
    local version="$1"
    if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(\+dev|-dev)?$ ]]; then
        print_error "Invalid version format: $version"
        print_info "Expected format: X.Y.Z or X.Y.Z+dev (e.g., 0.1.1 or 0.1.1+dev)"
        return 1
    fi
    return 0
}

# Main script
main() {
    # Check if version argument is provided
    if [ $# -eq 0 ]; then
        print_error "No version specified"
        echo "Usage: $0 <new-version>"
        echo "Example: $0 0.1.1+dev"
        exit 1
    fi

    NEW_VERSION="$1"

    # Validate new version format
    if ! validate_version "$NEW_VERSION"; then
        exit 1
    fi

    # Check if all required files exist
    print_info "Checking required files..."
    for file in "$MAKEFILE" "$OTA_GO" "$CARGO_TOML" "$SYNC_SCRIPT" "$PACKAGING_VERSION" "$UI_PACKAGE_LOCK" "$CARGO_LOCK"; do
        if [ ! -f "$file" ]; then
            print_error "Required file not found: $file"
            exit 1
        fi
    done
    print_success "All required files found"

    # Extract current versions
    print_info "Reading current versions..."
    CURRENT_MAKEFILE_VERSION=$(get_makefile_version)
    CURRENT_OTA_VERSION=$(get_ota_version)
    CURRENT_CARGO_VERSION=$(get_cargo_version)
    CURRENT_PACKAGING_VERSION=$(cat "$PACKAGING_VERSION" | tr -d '[:space:]')

    echo ""
    echo "Current versions:"
    echo "  Makefile:         $CURRENT_MAKEFILE_VERSION"
    echo "  ota.go:           $CURRENT_OTA_VERSION"
    echo "  Cargo.toml:       $CURRENT_CARGO_VERSION"
    echo "  packaging/version: $CURRENT_PACKAGING_VERSION"
    echo ""
    echo "New version:        $NEW_VERSION"
    echo ""

    # Derive per-tool version variants (single place for suffix rules):
    #   canonical (Makefile/ota.go/version.txt): keep as-is (may have +dev)
    #   Cargo:        strip +dev/-dev (Cargo rejects build metadata here)
    #   npm/Tauri:    +dev -> -dev (they reject '+' in versions)
    CARGO_VERSION="${NEW_VERSION%+dev}"
    CARGO_VERSION="${CARGO_VERSION%-dev}"
    NPM_VERSION="${NEW_VERSION/+dev/-dev}"

    # Ask for confirmation
    read -p "$(echo -e ${YELLOW}Proceed with version bump? [y/N]:${NC} )" -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Version bump cancelled"
        exit 0
    fi

    echo ""
    print_info "Starting version bump..."

    # Step 1: Update Makefile
    print_info "Updating Makefile..."
    sed -i.bak -E "s/^VERSION\s*:=\s*.+/VERSION := $NEW_VERSION/" "$MAKEFILE"
    rm -f "$MAKEFILE.bak"
    print_success "Makefile updated to $NEW_VERSION"

    # Step 2: Update ota.go
    print_info "Updating ota.go..."
    sed -i.bak -E "s/var builtAppVersion = \".+\"/var builtAppVersion = \"$NEW_VERSION\"/" "$OTA_GO"
    rm -f "$OTA_GO.bak"
    print_success "ota.go updated to $NEW_VERSION"

    # Step 3: Update Cargo.toml
    print_info "Updating Cargo.toml..."
    sed -i.bak -E "0,/^version = \".+\"/s//version = \"$CARGO_VERSION\"/" "$CARGO_TOML"
    rm -f "$CARGO_TOML.bak"
    print_success "Cargo.toml updated to $CARGO_VERSION"

    # Step 3b: Update the local "app" package version in Cargo.lock
    print_info "Updating Cargo.lock..."
    sed -i.bak -E "/^name = \"app\"$/{n;s/^version = \".+\"/version = \"$CARGO_VERSION\"/;}" "$CARGO_LOCK"
    rm -f "$CARGO_LOCK.bak"
    print_success "Cargo.lock updated to $CARGO_VERSION"

    # Step 4: Update packaging/version.txt
    print_info "Updating packaging/version.txt..."
    echo "$NEW_VERSION" > "$PACKAGING_VERSION"
    print_success "packaging/version.txt updated to $NEW_VERSION"

    # Step 5: Run sync script (single source of truth: version computed here)
    print_info "Running sync script to update package.json and tauri.conf.json..."
    if command -v node &> /dev/null; then
        node "$SYNC_SCRIPT" "$NPM_VERSION"
        print_success "Sync script completed"
    else
        print_error "Node.js not found. Please run manually: node ui/scripts/sync-version.cjs $NPM_VERSION"
        exit 1
    fi

    # Step 6: Sync package-lock.json version (NOT audit fix — that would pull
    # unrelated dependency upgrades into a version bump).
    print_info "Syncing package-lock.json version..."
    if command -v npm &> /dev/null; then
        cd "$PROJECT_ROOT/ui"
        npm install --package-lock-only
        cd "$PROJECT_ROOT"
        print_success "package-lock.json synced"
    else
        print_error "npm not found. Please run manually: cd ui && npm install --package-lock-only"
        exit 1
    fi

    # Step 7: Verify every file now reports the expected version.
    print_info "Verifying all files agree..."
    verify_failed=0
    check() {
        local name="$1" expected="$2" actual="$3"
        if [ "$actual" != "$expected" ]; then
            print_error "$name is '$actual', expected '$expected'"
            verify_failed=1
        fi
    }
    check "Makefile"          "$NEW_VERSION"   "$(get_makefile_version)"
    check "ota.go"            "$NEW_VERSION"   "$(get_ota_version)"
    check "Cargo.toml"        "$CARGO_VERSION" "$(get_cargo_version)"
    check "Cargo.lock"        "$CARGO_VERSION" "$(get_cargo_lock_version)"
    check "packaging/version" "$NEW_VERSION"   "$(tr -d '[:space:]' < "$PACKAGING_VERSION")"
    check "package.json"      "$NPM_VERSION"   "$(node -p "require('$PROJECT_ROOT/ui/package.json').version")"
    check "tauri.conf.json"   "$NPM_VERSION"   "$(node -p "require('$PROJECT_ROOT/ui/src-tauri/tauri.conf.json').version")"
    if [ "$verify_failed" -ne 0 ]; then
        print_error "Verification failed — files are inconsistent. Review the changes above."
        exit 1
    fi
    print_success "All files verified at $NEW_VERSION"

    echo ""
    print_success "Version bump completed successfully!"
    echo ""
    print_info "Updated files:"
    echo "  • Makefile → $NEW_VERSION"
    echo "  • ota.go → $NEW_VERSION"
    echo "  • ui/src-tauri/Cargo.toml → $CARGO_VERSION"
    echo "  • ui/src-tauri/Cargo.lock → $CARGO_VERSION"
    echo "  • packaging/version.txt → $NEW_VERSION"
    echo "  • ui/package.json (via sync script)"
    echo "  • ui/package-lock.json (via npm audit fix)"
    echo "  • ui/src-tauri/tauri.conf.json (via sync script)"
    echo ""
    print_warning "Don't forget to update CHANGELOG.md!"
}

main "$@"
