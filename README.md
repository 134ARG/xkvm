# XKVM

XKVM is a high-performance, open-source KVM-over-IP solution running on full Linux. It provides remote keyboard, video, and mouse control for efficient management of computers, servers, and workstations.

## Overview

XKVM has evolved from an embedded system to a full Linux-based solution, enabling standard system administration practices and better integration with existing infrastructure. The system focuses on core KVM functionality with a clean, maintainable codebase.

## Key Features

- **Low-latency Video** - H.264 hardware encoding for smooth remote control
- **USB Gadget Emulation** - Keyboard, mouse, and mass storage device emulation
- **Web Interface** - Modern React-based UI for device management
- **Network Monitoring** - Real-time network status and DHCP lease information
- **Standard Linux** - Full Linux OS enables standard administration tools

## Architecture

XKVM consists of:
- **Backend** (Go) - Device management, video capture, USB gadget control, JSON-RPC API
- **Frontend** (React/TypeScript) - Web-based management interface
- **Native Layer** (C) - Hardware video capture and control via Rockchip SDK

## System Management

XKVM integrates with standard Linux tools:

- **Network Configuration** - Use `nmcli`, `nmtui`, or `/etc/network/interfaces`
- **SSH Access** - Configure via `ssh-copy-id` or `/etc/ssh/authorized_keys`
- **System Control** - Use `systemctl` for service management
- **Updates** - Manage via standard Linux package managers

The web UI displays system status in read-only mode. Configuration changes should be made through standard Linux tools.

## USB Gadget Features

- Automatic recovery from crashes and stale state
- Health monitoring with automatic recovery
- Stable HID device operation during reconfiguration
- Mass storage mounting (disk mode)
- File-based locking prevents concurrent access issues

## Development

XKVM is written in Go, TypeScript, and C. See **[DEVELOPMENT.md](DEVELOPMENT.md)** for comprehensive development information including setup, testing, and debugging.

Quick device deployment:
```bash
./dev_deploy.sh --help
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

Contributions are welcome. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## License

See [LICENSE](LICENSE) for details.
