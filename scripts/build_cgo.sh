#!/bin/bash
set -e

SCRIPT_PATH=$(realpath "$(dirname $(realpath "${BASH_SOURCE[0]}"))")
source ${SCRIPT_PATH}/build_utils.sh

CMAKE_BUILD_TYPE=${CMAKE_BUILD_TYPE:-Release}

CGO_PATH=$(realpath "${SCRIPT_PATH}/../internal/native/cgo")
BUILD_DIR=${CGO_PATH}/build

# Updated for RK3566 - no external toolchain needed for local development
# CMAKE_TOOLCHAIN_FILE=/opt/jetkvm-native-buildkit/rv1106-jetkvm-v2.cmake
CLEAN_ALL=${CLEAN_ALL:-0}

if [ "$CLEAN_ALL" -eq 1 ]; then
    rm -rf "${BUILD_DIR}"
fi

TMP_DIR=$(mktemp -d)
pushd "${CGO_PATH}" > /dev/null

msg_info "▶ Building native library (headless mode, RK3566)"
VERBOSE=1 cmake -B "${BUILD_DIR}" \
    -DCMAKE_BUILD_TYPE=${CMAKE_BUILD_TYPE} \
    -DCMAKE_INSTALL_PREFIX="${TMP_DIR}" \
    .

msg_info "▶ Copying built library and header files"
cmake --build "${BUILD_DIR}" --target install
cp -r "${TMP_DIR}/include" "${CGO_PATH}" 2>/dev/null || true
cp -r "${TMP_DIR}/lib" "${CGO_PATH}" 2>/dev/null || true
rm -rf "${TMP_DIR}"

popd > /dev/null
