# XKVM

[English](README.en.md) | [简体中文](README.md)

XKVM is a high-performance, open-source, 100% local KVM-over-IP solution running on a full Linux. It provides remote keyboard, video, and mouse control for efficient management of computers, servers, and workstations.

The project is currently for personal use.

## Overview

XKVM focuses on core KVM functionality with a clean, maintainable codebase. It is designed to run on a full Linux distro rather than a buildroot env compared with the base JetKVM. The annoying cloud features are removed, and native webui wrapper is provided for working around using VPNs like tailscale with WebRTC.

**The current implementation is only designed for Radxa Zero 3 with tc358743 HDMI->CSI capture chip.**

## Added/Improved Key Features

- **Low-latency Video** - H.264/H.265 hardware encoding for smooth remote control with selectable codec ultizing rk3566 hardware encdoing (H.265 only on MacOS client)
- **Adjustable Streaming bitrate with VBR** - Now the video streaming bitrate is adjustable from 1 Mbps to 20 Mbps
- **USB Gadget Emulation** - Keyboard, mouse, and mass storage device emulation with improved error handling and resetting
- **Web Interface and Tauri native encap** - Modern React-based UI for device management, with Linux/MacOS/Windows Tauri native wrapper
- **Network Monitoring** - Real-time network status and DHCP lease information. Network manipuation is removed, now it is read-only


## Architecture

XKVM consists of:
- **Backend** (Go) - Device management, video capture, USB gadget control, JSON-RPC API
- **Frontend** (React/TypeScript) - Web-based management interface
- **Native Layer** (C) - Hardware video capture and control via Rockchip SDK

## Development

XKVM is written in Go, TypeScript, and C. See **[DEVELOPMENT.md](DEVELOPMENT.md)** for comprehensive development information including setup, testing, and debugging.

### Building

To build XKVM for deployment:

```bash
# Set up ARM64 sysroot (one-time setup)
export ARM64_SYSROOT="/path/to/sysroot"
./scripts/setup_arm64_sysroot.sh

# Build everything (frontend + backend + installer)
make frontend
ARM64_SYSROOT=/path/to/sysroot make build_release

# Output: bin/xkvm_installer.sh (self-extracting installer)
```

## Changes from JetKVM

XKVM has been significantly refactored from the original JetKVM embedded system. Major changes include:

- **Full Linux Migration** - Standard OS instead of embedded system
- **Network Read-Only** - Network managed by OS, not application (1,260 lines removed)
- **Removed Features** - Cloud integration UI, display controls, developer mode toggle, OTA updates
- **USB Improvements** - Enhanced stability, health monitoring, automatic recovery
- **Simplified UI** - Focus on core KVM functionality

See **[CHANGELOG.md](CHANGELOG.md)** for detailed change history.

## Contributing

Contributions are welcome.

## License

See [LICENSE](LICENSE) for details.
