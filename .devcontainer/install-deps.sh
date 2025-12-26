#!/bin/bash

SUDO_PATH=$(which sudo)
function sudo() {
  if [ "$UID" -eq 0 ]; then
    "$@"
  else
    ${SUDO_PATH} "$@"
  fi
}

set -ex

export DEBIAN_FRONTEND=noninteractive
sudo apt-get update && \
sudo apt-get install -y --no-install-recommends \
  iputils-ping \
  build-essential \
  device-tree-compiler \
  gperf g++-multilib gcc-multilib \
  gdb-multiarch \
  libnl-3-dev libdbus-1-dev libelf-dev libmpc-dev dwarves \
  bc openssl flex bison libssl-dev python3 python-is-python3 texinfo kmod cmake \
  wget zstd \
  python3-venv python3-kconfiglib \
  gcc-aarch64-linux-gnu g++-aarch64-linux-gnu \
  && sudo rm -rf /var/lib/apt/lists/*

# Note: Buildkit installation removed for RK3566 - using local SDK instead
# The project now uses extracted MPP libraries from .deb packages

echo "✅ Development dependencies installed"
echo "🔧 Cross-compilation toolchain for ARM64 ready"
