#!/bin/bash
set -e

SCRIPT_PATH=$(realpath "$(dirname $(realpath "${BASH_SOURCE[0]}"))")
source ${SCRIPT_PATH}/build_utils.sh

CMAKE_BUILD_TYPE=${CMAKE_BUILD_TYPE:-Release}
CROSS_COMPILE=${CROSS_COMPILE:-auto}
BUILD_BINARY=${BUILD_BINARY:-ON}

CGO_PATH=$(realpath "${SCRIPT_PATH}/../internal/native/cgo")
BUILD_DIR=${CGO_PATH}/build
CLEAN_ALL=${CLEAN_ALL:-0}

# Detect architecture and set cross-compilation
HOST_ARCH=$(uname -m)
TARGET_ARCH=${TARGET_ARCH:-aarch64}

msg_info "Host architecture: $HOST_ARCH"
msg_info "Target architecture: $TARGET_ARCH"

# Determine if cross-compilation is needed
if [ "$CROSS_COMPILE" = "auto" ]; then
    if [ "$HOST_ARCH" = "aarch64" ] || [ "$HOST_ARCH" = "arm64" ]; then
        if [ "$TARGET_ARCH" = "aarch64" ] || [ "$TARGET_ARCH" = "arm64" ]; then
            CROSS_COMPILE=no
            msg_info "Native ARM64 build detected"
        else
            CROSS_COMPILE=yes
            msg_info "Cross-compilation required"
        fi
    else
        if [ "$TARGET_ARCH" = "aarch64" ] || [ "$TARGET_ARCH" = "arm64" ]; then
            CROSS_COMPILE=yes
            msg_info "Cross-compilation required (x86_64 -> ARM64)"
        else
            CROSS_COMPILE=no
            msg_info "Native build"
        fi
    fi
fi

if [ "$CLEAN_ALL" -eq 1 ]; then
    rm -rf "${BUILD_DIR}"
fi

TMP_DIR=$(mktemp -d)
pushd "${CGO_PATH}" > /dev/null

# Set up CMake arguments
CMAKE_ARGS=(
    -DCMAKE_BUILD_TYPE=${CMAKE_BUILD_TYPE}
    -DCMAKE_INSTALL_PREFIX="${TMP_DIR}"
    -DBUILD_BINARY=${BUILD_BINARY}
)

if [ "$CROSS_COMPILE" = "yes" ]; then
    msg_info "▶ Setting up cross-compilation for ARM64"
    
    # Check if cross-compiler is available
    if ! command -v aarch64-linux-gnu-gcc &> /dev/null; then
        msg_err "Cross-compiler aarch64-linux-gnu-gcc not found!"
        msg_err "Install it with: sudo dnf install gcc-aarch64-linux-gnu gcc-c++-aarch64-linux-gnu"
        exit 1
    fi
    
    # Check if ARM64_SYSROOT is set
    if [ -z "$ARM64_SYSROOT" ]; then
        msg_err "ARM64_SYSROOT environment variable is required for cross-compilation"
        msg_err "Set ARM64_SYSROOT to your ARM64 sysroot path:"
        msg_err "  export ARM64_SYSROOT=\"/path/to/your/arm64-sysroot\""
        msg_err "Or create a sysroot with: ./scripts/setup_arm64_sysroot.sh"
        exit 1
    fi
    
    # Check if sysroot exists
    if [ ! -d "$ARM64_SYSROOT" ]; then
        msg_err "ARM64 sysroot not found at: $ARM64_SYSROOT"
        msg_err "Create it with: ARM64_SYSROOT=\"$ARM64_SYSROOT\" ./scripts/setup_arm64_sysroot.sh"
        exit 1
    fi
    
    # Export ARM64_SYSROOT for CMake
    export ARM64_SYSROOT
    
    CMAKE_ARGS+=(-DCMAKE_TOOLCHAIN_FILE=cmake/aarch64-linux-gnu.cmake)
    msg_info "Using cross-compilation toolchain with sysroot"
    msg_info "Sysroot: ${ARM64_SYSROOT}"
else
    msg_info "▶ Using native compilation"
fi

msg_info "▶ Building native library (headless mode, RK3566)"
msg_info "Cross-compile: $CROSS_COMPILE"
msg_info "Build binary: $BUILD_BINARY"

VERBOSE=1 cmake -B "${BUILD_DIR}" "${CMAKE_ARGS[@]}" .

msg_info "▶ Compiling..."
cmake --build "${BUILD_DIR}" --parallel $(nproc)

if [ "$BUILD_BINARY" = "ON" ]; then
    msg_info "▶ Installing library and binary"
else
    msg_info "▶ Installing library only"
fi

cmake --build "${BUILD_DIR}" --target install

# Copy built artifacts
if [ -d "${TMP_DIR}/include" ]; then
    cp -r "${TMP_DIR}/include" "${CGO_PATH}" 2>/dev/null || true
fi
if [ -d "${TMP_DIR}/lib" ]; then
    cp -r "${TMP_DIR}/lib" "${CGO_PATH}" 2>/dev/null || true
fi
if [ -d "${TMP_DIR}/bin" ]; then
    cp -r "${TMP_DIR}/bin" "${CGO_PATH}" 2>/dev/null || true
fi

rm -rf "${TMP_DIR}"

popd > /dev/null

msg_info "✓ Build completed successfully"
if [ "$CROSS_COMPILE" = "yes" ]; then
    msg_info "✓ Cross-compiled for ARM64"
    if [ "$BUILD_BINARY" = "ON" ]; then
        msg_info "✓ Binary available at: ${CGO_PATH}/bin/jknative-bin"
    fi
else
    msg_info "✓ Native build completed"
fi
