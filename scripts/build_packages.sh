#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# Read version
if [ ! -f "packaging/version.txt" ]; then
    print_error "packaging/version.txt not found"
    exit 1
fi

VERSION=$(cat packaging/version.txt | tr -d '[:space:]')
# Native packages reject '-'/'+' as prerelease markers and sort them as newer
# than the release. Map +dev/-dev to '~dev', which dpkg/rpm sort *before* the
# release — correct prerelease ordering.
PKG_VERSION="${VERSION/+dev/~dev}"
PKG_VERSION="${PKG_VERSION/-dev/~dev}"
print_info "Building packages for version: $VERSION (package version: $PKG_VERSION)"

# Check if xkvm_app exists
if [ ! -f "bin/xkvm_app" ]; then
    print_error "bin/xkvm_app not found. Run 'make build_release' first."
    exit 1
fi

# Check for required .so files
REQUIRED_LIBS=(
    "internal/native/cgo/sdk/vendor/rockit/lib/lib64/librockit.so"
    "internal/native/cgo/sdk/vendor/rockit/lib/lib64/libgraphic_lsf.so"
)

for lib in "${REQUIRED_LIBS[@]}"; do
    if [ ! -f "$lib" ]; then
        print_error "Required library not found: $lib"
        exit 1
    fi
done

print_success "All required files found"

# Build DEB package
print_info "Building DEB package..."
    
    DEB_DIR="$PROJECT_ROOT/build/deb"
    rm -rf "$DEB_DIR"
    mkdir -p "$DEB_DIR"/{DEBIAN,usr/bin,usr/lib/xkvm,etc/systemd/system,usr/share/doc/xkvm}
    
    # Copy binary
    cp bin/xkvm_app "$DEB_DIR/usr/lib/xkvm/"
    chmod 755 "$DEB_DIR/usr/lib/xkvm/xkvm_app"
    
    # Copy libraries
    cp internal/native/cgo/sdk/vendor/rockit/lib/lib64/librockit.so "$DEB_DIR/usr/lib/xkvm/"
    cp internal/native/cgo/sdk/vendor/rockit/lib/lib64/libgraphic_lsf.so "$DEB_DIR/usr/lib/xkvm/"
    cp internal/native/cgo/sdk/mpp/lib/librockchip_mpp.so* "$DEB_DIR/usr/lib/xkvm/" 2>/dev/null || true
    
    # Copy wrapper script
    cp packaging/bin/xkvm-wrapper.sh "$DEB_DIR/usr/bin/xkvm"
    chmod 755 "$DEB_DIR/usr/bin/xkvm"
    
    # Copy systemd service
    cp packaging/systemd/xkvm.service "$DEB_DIR/etc/systemd/system/"
    
    # Copy documentation
    cp LICENSE "$DEB_DIR/usr/share/doc/xkvm/"
    cp packaging/deb/copyright "$DEB_DIR/usr/share/doc/xkvm/"
    
    # Create control file
    cat > "$DEB_DIR/DEBIAN/control" << EOF
Package: xkvm
Version: $PKG_VERSION
Section: utils
Priority: optional
Architecture: arm64
Maintainer: 134ARG <xen134@outlook.com>
Homepage: https://github.com/134ARG/xkvm
Depends: libmali-bifrost-g52-g13p0-gbm
Description: High-performance KVM-over-IP solution
 XKVM provides remote keyboard/mouse input and 1080p 60fps
 hardware-encoded video streaming for remote machine management.
 .
 Based on Rockchip RK3566 platform with tc358743 HDMI-CSI capture card.
EOF
    
    # Create postinst script
    cat > "$DEB_DIR/DEBIAN/postinst" << 'EOF'
#!/bin/bash
systemctl daemon-reload
mkdir -p /etc/xkvm /var/lib/xkvm /var/log/xkvm
echo ""
echo "XKVM has been installed successfully!"
echo ""
echo "To start the service:"
echo "  sudo systemctl start xkvm"
echo ""
echo "To enable on boot:"
echo "  sudo systemctl enable xkvm"
echo ""
EOF
    chmod 755 "$DEB_DIR/DEBIAN/postinst"
    
    # Create prerm script
    cat > "$DEB_DIR/DEBIAN/prerm" << 'EOF'
#!/bin/bash
if [ "$1" = "remove" ]; then
    systemctl stop xkvm.service 2>/dev/null || true
fi
EOF
    chmod 755 "$DEB_DIR/DEBIAN/prerm"
    
    # Create postrm script
    cat > "$DEB_DIR/DEBIAN/postrm" << 'EOF'
#!/bin/bash
systemctl daemon-reload
EOF
    chmod 755 "$DEB_DIR/DEBIAN/postrm"
    
    # Build package
    DEB_FILE="bin/xkvm_${PKG_VERSION}_arm64.deb"
    dpkg-deb --root-owner-group --build "$DEB_DIR" "$DEB_FILE"
    
    print_success "DEB package created: $DEB_FILE"
    echo "  Size: $(du -h "$DEB_FILE" | cut -f1)"

echo ""
print_success "Package build completed!"
echo ""
print_info "Testing commands:"
echo "  dpkg -c bin/xkvm_${PKG_VERSION}_arm64.deb"
echo "  dpkg -I bin/xkvm_${PKG_VERSION}_arm64.deb"
echo ""
