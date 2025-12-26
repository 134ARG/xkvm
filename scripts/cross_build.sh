#!/bin/bash
# Convenience script for cross-compilation to ARM64

set -e

SCRIPT_PATH=$(realpath "$(dirname $(realpath "${BASH_SOURCE[0]}"))")

echo "🔧 Cross-compiling JetKVM native components for ARM64..."

# Force cross-compilation
export CROSS_COMPILE=yes
export TARGET_ARCH=aarch64
export BUILD_BINARY=ON

# Run the main build script
"${SCRIPT_PATH}/build_cgo.sh"

echo "✅ Cross-compilation completed!"
echo "📦 ARM64 binaries are ready for deployment to RK3566"