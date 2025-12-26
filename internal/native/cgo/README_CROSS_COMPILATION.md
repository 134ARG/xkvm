# Cross-Compilation Setup for JetKVM Native Components

This document describes the cross-compilation setup for building JetKVM native components for ARM64 (RK3566) from x86_64 development machines.

## Overview

The cross-compilation system supports:
- **Target**: ARM64 (aarch64) with Cortex-A55 optimizations for RK3566
- **Host**: x86_64 development machines
- **Compatibility**: Native ARM64 builds still work on target hardware
- **SDK**: Local RK3566 MPP SDK with ARM64 libraries
- **Mode**: Headless operation (LVGL display components removed)

## Quick Start

### Cross-compile for ARM64:
```bash
# Set up sysroot first
export ARM64_SYSROOT="/opt/arm64-sysroot"
./scripts/setup_arm64_sysroot.sh

# Then cross-compile
./scripts/cross_build.sh

# Or with inline variable
ARM64_SYSROOT="/opt/arm64-sysroot" ./scripts/cross_build.sh
```

### Native build (on ARM64 machine):
```bash
./scripts/build_cgo.sh
```

## Architecture

### Cross-Compilation Toolchain
- **Compiler**: `aarch64-linux-gnu-gcc` (GCC 15.2.1)
- **Sysroot**: Full ARM64 Ubuntu Jammy sysroot at `internal/native/cgo/sysroot/`
- **Target CPU**: Cortex-A55 with ARMv8.2-A+fp16 optimizations
- **Dependencies**: ALSA (libasound2) and DRM (libdrm) libraries included

### Build Outputs
- **Library**: `internal/native/cgo/lib/libjknative.a` (ARM64 static library)
- **Binary**: `internal/native/cgo/bin/jknative-bin` (ARM64 executable)

### SDK Integration
- **Location**: `internal/native/cgo/sdk/`
- **MPP Libraries**: RK3566-specific ARM64 libraries
- **Headers**: Rockchip MPP and TGI headers included

## Build System

### Auto-Detection
The build system automatically detects:
- Host architecture (x86_64 vs ARM64)
- Cross-compilation requirements
- Available toolchain and sysroot

### Environment Variables
- `CROSS_COMPILE`: Force cross-compilation (yes/no/auto)
- `TARGET_ARCH`: Target architecture (aarch64/x86_64)
- `BUILD_BINARY`: Enable/disable binary building (ON/OFF)
- `ARM64_SYSROOT`: **Required** - Path to ARM64 sysroot for cross-compilation

## Prerequisites

### Host System Requirements
- x86_64 Linux development machine
- Cross-compilation toolchain for ARM64
- CMake 3.14 or later

### Install Cross-Compilation Toolchain

#### On Fedora/RHEL:
```bash
sudo dnf install gcc-aarch64-linux-gnu gcc-c++-aarch64-linux-gnu
```

#### On Ubuntu/Debian:
```bash
sudo apt install gcc-aarch64-linux-gnu g++-aarch64-linux-gnu
```

### Create ARM64 Sysroot

The cross-compilation setup requires a full ARM64 sysroot with the necessary libraries. You must set the `ARM64_SYSROOT` environment variable to specify the location:

```bash
# Set sysroot location (required)
export ARM64_SYSROOT="/opt/arm64-sysroot"

# Install debootstrap if not available
sudo dnf install debootstrap  # Fedora/RHEL
# or
sudo apt install debootstrap  # Ubuntu/Debian

# Create ARM64 sysroot at the specified location
sudo debootstrap --arch=arm64 --variant=minbase --include=build-essential,libc6-dev jammy $ARM64_SYSROOT http://ports.ubuntu.com/ubuntu-ports

# Install additional dependencies needed by Rockchip SDK
sudo chroot $ARM64_SYSROOT /bin/bash -c "apt update && apt install -y libasound2-dev libdrm-dev"
```

**Or use the provided setup script:**
```bash
export ARM64_SYSROOT="/opt/arm64-sysroot"
./scripts/setup_arm64_sysroot.sh
```

## Files Modified

### Core Build Files
- `internal/native/cgo/CMakeLists.txt`: Updated for RK3566 SDK and cross-compilation
- `internal/native/cgo/cmake/aarch64-linux-gnu.cmake`: Cross-compilation toolchain
- `scripts/build_cgo.sh`: Enhanced with cross-compilation detection
- `scripts/cross_build.sh`: Convenience script for cross-compilation

### SDK Structure
- `internal/native/cgo/sdk/mpp/`: RK3566 MPP libraries and headers
- `internal/native/cgo/sdk/vendor/rockit/`: Rockchip SDK components

### Sysroot
- `internal/native/cgo/sysroot/`: Full ARM64 Ubuntu Jammy sysroot
- Includes ALSA, DRM, and standard C libraries for ARM64

## Compatibility

### Cross-Compilation (x86_64 → ARM64)
✅ Library compilation  
✅ Binary linking  
✅ Rockchip SDK integration  
✅ ALSA/DRM dependencies resolved  

### Native Compilation (ARM64 → ARM64)
✅ Library compilation  
✅ Binary linking  
✅ SDK compatibility maintained  

## Deployment

Cross-compiled binaries are ready for deployment to RK3566 hardware:
- Copy `internal/native/cgo/lib/libjknative.a` for Go CGO linking
- Copy `internal/native/cgo/bin/jknative-bin` for standalone testing
- Ensure target system has ALSA and DRM runtime libraries

## Integration with Go

The cross-compiled library integrates with Go CGO:

```go
// Use the cross-compiled library in Go
/*
#cgo CFLAGS: -I${SRCDIR}/internal/native/cgo
#cgo LDFLAGS: -L${SRCDIR}/internal/native/cgo/lib -ljknative
*/
import "C"
```

## Troubleshooting

### Missing Dependencies
If linking fails with missing library errors, install dependencies in sysroot:
```bash
sudo chroot $ARM64_SYSROOT /bin/bash -c "apt update && apt install -y <package-name>"
```

### Toolchain Issues
Ensure cross-compilation toolchain is installed:
```bash
sudo dnf install gcc-aarch64-linux-gnu gcc-c++-aarch64-linux-gnu
```

### Sysroot Problems
The sysroot can be rebuilt using debootstrap. Set `ARM64_SYSROOT` to your desired location:
```bash
export ARM64_SYSROOT="/path/to/your/sysroot"
sudo debootstrap --arch=arm64 --variant=minbase --include=build-essential,libc6-dev,libasound2-dev,libdrm-dev jammy $ARM64_SYSROOT http://ports.ubuntu.com/ubuntu-ports
```

### Verifying Cross-Compilation
```bash
# Check architecture of built files
file internal/native/cgo/lib/libjknative.a
file internal/native/cgo/bin/jknative-bin

# Should show ARM64/aarch64 architecture
```