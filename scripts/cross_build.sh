#!/bin/bash
# Convenience script for cross-compilation to ARM64

set -e

SCRIPT_PATH=$(realpath "$(dirname $(realpath "${BASH_SOURCE[0]}"))")

echo "🔧 Cross-compiling JetKVM native components for ARM64..."

# Force cross-compilation
export CROSS_COMPILE=yes
export TARGET_ARCH=aarch64
export BUILD_BINARY=ON

# Check if ARM64_SYSROOT is set
if [ -z "$ARM64_SYSROOT" ]; then
    echo "❌ ARM64_SYSROOT environment variable is required"
    echo "💡 Set ARM64_SYSROOT to your ARM64 sysroot path:"
    echo "   export ARM64_SYSROOT=\"/path/to/your/arm64-sysroot\""
    echo "   $0"
    echo ""
    echo "💡 Or create a sysroot first:"
    echo "   ./scripts/setup_arm64_sysroot.sh"
    exit 1
else
    echo "ℹ️  Using ARM64_SYSROOT: $ARM64_SYSROOT"
fi

# Validate sysroot exists
if [ ! -d "$ARM64_SYSROOT" ]; then
    echo "❌ ARM64 sysroot not found at: $ARM64_SYSROOT"
    echo "💡 Create it with: ARM64_SYSROOT=\"$ARM64_SYSROOT\" ./scripts/setup_arm64_sysroot.sh"
    exit 1
fi

# Run the main build script
"${SCRIPT_PATH}/build_cgo.sh"

echo "✅ Cross-compilation completed!"
echo "📦 ARM64 binaries are ready for deployment to RK3566"