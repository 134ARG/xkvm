#!/bin/bash
# Script to set up ARM64 sysroot for cross-compilation

set -e

SCRIPT_PATH=$(realpath "$(dirname $(realpath "${BASH_SOURCE[0]}"))")
source ${SCRIPT_PATH}/build_utils.sh

# Sysroot location - must be provided by user
if [ -z "$ARM64_SYSROOT" ]; then
    msg_err "ARM64_SYSROOT environment variable is required"
    msg_info "Set ARM64_SYSROOT to your desired sysroot location:"
    msg_info "  export ARM64_SYSROOT=\"/opt/arm64-sysroot\""
    msg_info "  $0"
    msg_info ""
    msg_info "Or run with inline variable:"
    msg_info "  ARM64_SYSROOT=\"/opt/arm64-sysroot\" $0"
    exit 1
fi

SYSROOT_PATH="$ARM64_SYSROOT"

msg_info "Setting up ARM64 sysroot for cross-compilation"
msg_info "Target location: $SYSROOT_PATH"

# Check if running as root for system-wide installation
if [ "$EUID" -ne 0 ] && [[ "$SYSROOT_PATH" == /opt/* ]]; then
    msg_err "Root privileges required for system-wide sysroot installation"
    msg_err "Run with sudo or set ARM64_SYSROOT to a user-writable location"
    msg_err "Example: ARM64_SYSROOT=~/arm64-sysroot $0"
    exit 1
fi

# Check if debootstrap is available
if ! command -v debootstrap &> /dev/null; then
    msg_err "debootstrap not found!"
    msg_info "Install it with:"
    msg_info "  Fedora/RHEL: sudo dnf install debootstrap"
    msg_info "  Ubuntu/Debian: sudo apt install debootstrap"
    exit 1
fi

# Check if sysroot already exists
if [ -d "$SYSROOT_PATH" ]; then
    msg_warn "Sysroot already exists at $SYSROOT_PATH"
    read -p "Remove and recreate? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        msg_info "Removing existing sysroot..."
        rm -rf "$SYSROOT_PATH"
    else
        msg_info "Keeping existing sysroot"
        exit 0
    fi
fi

# Create sysroot directory
msg_info "Creating sysroot directory..."
mkdir -p "$SYSROOT_PATH"

# Create ARM64 sysroot with debootstrap
msg_info "Creating ARM64 sysroot (this may take several minutes)..."
debootstrap --arch=arm64 --variant=minbase \
    --include=build-essential,libc6-dev,libasound2-dev,libdrm-dev \
    jammy "$SYSROOT_PATH" \
    http://ports.ubuntu.com/ubuntu-ports

# Set proper permissions
if [ "$EUID" -eq 0 ]; then
    chown -R root:root "$SYSROOT_PATH"
    chmod -R 755 "$SYSROOT_PATH"
fi

msg_info "✅ ARM64 sysroot created successfully!"
msg_info "📍 Location: $SYSROOT_PATH"
msg_info ""
msg_info "To use this sysroot for cross-compilation:"
msg_info "  export ARM64_SYSROOT=\"$SYSROOT_PATH\""
msg_info "  ./scripts/cross_build.sh"
msg_info ""
msg_info "Or build the full project:"
msg_info "  ARM64_SYSROOT=\"$SYSROOT_PATH\" make build_release"