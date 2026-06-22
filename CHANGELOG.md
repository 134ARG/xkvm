# XKVM Changelog

This document summarizes the major changes made to XKVM during its evolution from the original JetKVM embedded system to a full Linux-based KVM-over-IP solution.

## Version 0.2.0 (2026-06-22)

### Version Display and Auto-Update
- Added a Versions section in General settings showing backend and Connector versions with update checking against GitHub releases
- Added one-click backend self-update for Debian installs, with a download link fallback for other setups

### Web Terminal
- Dropped the interactive shell to an unprivileged user while the daemon keeps root for hardware access
- Fixed the terminal not fitting its container on open

### Video
- Added the "XKVM Ultimate" EDID preset (1080p60 default, 4K60, 1440p/1600p 120 Hz)

### Build and Release
- Updated UI dependencies

## Version 0.1.9 (2026-06-21)

### ATX Power Control
- Moved ATX power control out of the extensions panel into a dedicated action bar badge
- Added ATX support-status feedback so the control reflects whether the device reports ATX capability
- Removed the redundant separate power-on indicator now covered by the badge

### Native Desktop Connection Profiles
- Reworked connection profile management in the native app with editable profiles and a preferences dialog
- Added startup behavior options to connect to a default or last-used profile on launch
- Added a backend reachability test with success and error reporting, backed by new Tauri config and store handling

### USB Reliability
- Fixed a USB gadget race condition and hardened concurrent HID handling
- Replaced the changeset-based gadget configuration with a simpler direct gadget config and HID device model
- Fixed swallowed errors in gadget cleanup and configuration paths

### Video
- Added a 2K 120Hz EDID preset

### VFD
- Fixed a silent VFD failure path so renderer errors surface instead of leaving the display dead

### UI
- Fixed click pass-through behavior on the environment metrics badge

### Build and Release
- Reworked the version bump script for consistent version handling across all packages, including correct deb/rpm prerelease versions
- Updated CI workflow action versions

## Version 0.1.8 (2026-06-15)

### Host Monitoring and VFD
- Made the host metrics receiver an independent device-side feature, so the WebUI host monitor can be used without enabling a VFD screen
- Kept VFD display enablement tied to host metrics internally, with clearer startup paths and logging for the listener and renderer
- Renamed the device config for the metrics listener from VFD-specific wording to host metrics wording

### Network Status
- Restored public IP lookup as explicit IPv4 and IPv6 endpoint queries instead of showing local interface addresses as public IPs
- Added per-family public IP status, timeout/error reporting, and address-family validation in the WebUI card

### UI
- Reworked the host metrics and VFD settings layout so each enable control is paired with its related input
- Fixed disabled-state behavior and alignment for the metrics port and VFD device path controls
- Clarified the host metrics and VFD labels in English and Chinese

### Packaging and Release
- Renamed the host-side metric collector package from `xkvm-vfd-agent` to `xkvm-host-agent`
- Updated the host agent RPM, systemd service, config file, build script, release workflow, and documentation around the new package name
- Renamed host agent CLI and environment settings to use `xkvm` address, port, and update interval terminology

## Version 0.1.7 (2026-06-07)

### Native Desktop and Video
- Fixed Tauri relative mouse locking and video-frame alignment with or without CAS
- Added a 4K HDMI EDID preset and fixed EDID preset matching
- Fixed Tauri language persistence

### USB and VFD
- Simplified USB gadget lifecycle handling and improved keyboard HID recovery
- Improved VFD host metrics networking, refresh behavior, and renderer logging

### Build and Release
- Added GitHub release automation for the macOS connector DMG, arm64 Debian package, and VFD agent RPM
- Moved device package builds into a Fedora container and checked in the required native SDK layout

### UI and Reliability
- Localized remaining hard-coded UI text and refined Chinese translations
- Fixed concurrent logger creation crashes

## Version 0.1.6 (2026-05-26)

### New Features
- Added optional VFD display support for CH347-driven VFD modules
- Added a host metrics receiver for CPU, RAM, GPU, temperature, uptime, network counters, and failed systemd units
- Added a host-side `xkvm-vfd-agent` metric collector for controlled machines
- Added lm-sensors based environment metrics for case temperature, humidity, and SoC temperature
- Added WebGL CAS-based video sharpening with existing brightness, contrast, and saturation adjustments

### UI
- Added a metrics badge in the action bar showing device environment metrics and connected host metrics
- Added hardware settings for VFD enablement, CH347 device path, and host metrics receiver port
- Added hardware settings for mapping lm-sensors chips and features used by the environment metrics badge
- Added video sharpness configuration backed by the new CAS renderer

### Packaging
- Added RPM packaging, systemd service, and config file for `xkvm-vfd-agent`
- Included the VFD agent RPM in the normal package build flow
- Made the VFD agent RPM version follow `packaging/version.txt`
- Updated packaging documentation for the split xKVM device package and host-side VFD agent package

### VFD Runtime
- Runs VFD rendering in a dedicated xKVM subprocess so VFD failures are isolated from the main service
- Added dense-map and numeric VFD themes for compact status display
- Added smooth numeric rendering, hex uptime encoding, fractional network readouts, and failure-state uptime blinking
- Added CH347 timeout, retry, SPI re-init, VRAM shadowing, and lazy grid update handling

## Version 0.1.5 (2026-04-28)

### New Features
- Added automatic H.264 fallback for clients without H.265 WebRTC support
- Added one-click build script support with environment-based ARM64 sysroot configuration

### Bug Fixes
- Fixed HDD LED state reporting so HDD activity is only shown while power LED state is active

### Video System
- Improved streaming color accuracy by switching capture back to UYVY 4:2:2
- Added BT.709 limited-range color metadata for H.264 and H.265 encoder output
- Updated encoder buffer sizing and virtual stride handling to match the real V4L2 capture layout

### USB Gadget
- Hardened USB gadget reconfiguration with lifecycle locking around config updates, soft reset, hard reset, and recovery
- Suspended and resumed HID operations safely during gadget reconfiguration
- Improved HID error propagation and reduced repeated timeout log noise
- Respected enabled HID device configuration when opening, verifying, and writing keyboard and mouse reports

### UI
- Disabled H.265 selection on unsupported browser/platform combinations and labeled it as Safari-only when unavailable
- Sent preferred video codec during WebRTC session setup so compatible clients can use H.265 automatically

### Documentation
- Adjusted README demo image scaling

## Version 0.1.4 (2026-03-23)

### New Features
- Added ATX power control with GPIO and serial port selection
- Added PWR and HDD LED state reading and ATX state feedback
- Added polarity configuration for GPIOs
- Added experimental soft and hard reset for USB

### UI
- Synced scroll bar theme with the UI theme
- Updated native Tauri wrapper to dark mode window

### Deprecations
- Retired Wake-on-LAN completely (previously marked deprecated)

### System
- Updated file paths to respect the Linux FHS

### Documentation
- Updated README

## Version 0.1.3 (2026-03-11)

### Bug Fixes
- Fixed USB HID not working on Windows
- Fixed USB serial number default configuration
- Fixed keyboard HID report descriptor for proper LOGICAL_MAXIMUM/USAGE_MAXIMUM encoding

### USB Gadget
- Simplified USB gadget implementation by removing unnecessary retry/backoff mechanisms
- Disabled USB gadget health check monitoring
- Changed absolute mouse HID protocol from 2 to 0
- Changed mass storage stall attribute from 1 to 0

### UI
- Added USB config settings back to hardware settings page
- Added icon margin to Tauri app icons

### Native Desktop App (Tauri)
- Allow HTTP access for macOS native wrapper

### Packaging
- Added Debian package build support with systemd service

### Documentation
- Added Chinese (zh-CN) version of README

### Video
- Updated encoding parameters in native video capture

## Version 0.1.2 (2026-03-02)

### UI
- Updated application icon

## Version 0.1.1 (2026-03-02)

### Bug Fixes
- Fixed missing connection config in native mode
- Fixed UI lint problems
- Added page reload after codec switch to ensure proper video stream restart
- Fixed pending H.265 startup issues

### Native Desktop App (Tauri)
- Added experimental Tauri-based native desktop application
- Changed Windows build target from MSVC to GNU toolchain for better compatibility
- Moved configuration to application directory for better cross-platform support

### Video System
- Changed pixel format to UYUV for improved color accuracy
- Added experimental CBR (Constant Bitrate) mode support in backend (not yet exposed in UI)

### Development
- Added version bump automation script (`scripts/bump_version.sh`)
- Improved build tooling and cross-platform support

## JetKVM to XKVM Updates

### Video Codec Selection (H.265 Support)
Added support for H.265 (HEVC) video encoding alongside the existing H.264 (AVC) codec. Users can now select their preferred codec from the video settings page.

**Features:**
- Selectable H.264/H.265 codec via web UI (Settings → Video)
- Hardware-accelerated encoding for both codecs using Rockchip MPP
- Automatic stream restart when codec is changed
- Codec preference persists across reboots
- Full GRPC proxy mode support for remote codec switching

**Technical Details:**
- H.265 provides better compression efficiency (up to 50% bitrate reduction)
- H.264 remains the default for maximum browser compatibility
- Both codecs use 60 fps GOP and configurable bitrate (1000-20000 kbps)
- H.265 uses 16-byte alignment, H.264 uses 2-byte alignment
- WebRTC track dynamically created based on selected codec

**Browser Compatibility:**
- H.264: Universal support across all modern browsers
- H.265: Best support in Safari, limited support in Chrome/Firefox

### Video Bitrate Configuration
Changed video quality configuration from a 0-1 factor to direct bitrate specification in kbps.

**Changes:**
- Bitrate now specified directly in kbps (1000-20000 range)
- UI displays bitrate slider with kbps values
- Backend validates bitrate range and applies directly to encoder
- Default bitrate set to 5000 kbps
- Legacy "quality factor" terminology replaced with "bitrate" throughout codebase

## Architecture Changes

### Full Linux Migration
XKVM has been migrated from an embedded system to a full Linux OS, enabling standard Linux system administration and better integration with existing infrastructure.

### Network Management (Read-Only)
Network configuration is now managed by the operating system (NetworkManager, systemd-networkd, etc.) rather than the application. The web UI displays network status in read-only mode. Configure network settings using standard Linux tools (`nmcli`, `nmtui`, `/etc/network/interfaces`, etc.).

**Changes:**
- Removed network write operations
- Preserved all network monitoring and status display
- DHCP lease information read from system DHCP clients
- Network state changes monitored via netlink

## Feature Removals

### Cloud Integration
The application supports three deployment modes: on-device (local), cloud-hosted UI, and native desktop app (Tauri). Cloud infrastructure is fully functional with WebRTC signaling, session management, and TURN server coordination. However, there is no device-side cloud adoption/configuration UI - devices cannot be "adopted" into a cloud service through the web interface. Cloud features are available when the UI itself is deployed in cloud mode.

### Display/Screen Management
Removed all physical display and backlight controls for headless operation. XKVM focuses on remote KVM functionality without local display requirements.

### Developer Mode & SSH Management
Removed embedded-device-specific developer mode toggle and SSH key management UI. SSH access should be configured using standard Linux tools (`ssh-copy-id`, `/etc/ssh/authorized_keys`, etc.).

### Reboot Control
Removed web UI reboot functionality. System management operations should be performed through standard Linux tools (`systemctl reboot`, SSH access, etc.).

### OTA Updates
Over-the-air update functionality has been disabled in the main loop but the infrastructure remains in the codebase. System updates should be managed through standard Linux package managers and update mechanisms.

## USB Gadget Improvements

### Stability Enhancements
- Added automatic cleanup of stale USB gadget state on startup
- Implemented file-based locking to prevent concurrent access issues
- Added retry logic with exponential backoff for transient failures
- Implemented health monitoring with automatic recovery
- Fixed HID device stability during USB reconfiguration

### Mass Storage
- Changed default mode from CDROM to Disk for better compatibility
- CDROM toggle option hidden in UI to prevent USB stability issues (but backend still supports CDROM mode)
- Automatic cleanup of stale mass storage configuration on restart
- ISO files now mount as disk mode by default
- CDROM mode can still be set via RPC calls for advanced use cases

### HID Devices
- Proper HID file lifecycle management during USB rebind
- Suspension mechanism prevents operations during reconfiguration
- Retry logic for HID device reopening after USB changes
- Eliminated "transport endpoint shutdown" errors

## UI Changes

### Mount Interface
- Simplified to show only "XKVM Storage" option
- Removed URL mount option (experimental feature)
- All images mount in Disk mode by default

### Settings Pages
- Network settings display read-only information
- Removed hardware settings page (display controls)
- No cloud adoption/configuration UI on device (cloud features available when UI is cloud-hosted)
- Removed developer mode and SSH key management
- Removed reboot device control

## Code Quality

### Reduction
- Network management: -1,260 lines (64% reduction)
- Total codebase simplified through removal of embedded-specific features
- Cleaner separation between monitoring and management

### Improvements
- Better error handling throughout USB gadget subsystem
- Comprehensive health monitoring and diagnostics
- Idempotent operations for safer USB management
- Enhanced logging for troubleshooting

## Migration Notes

### For Users
- **Network Configuration:** Use OS tools (`nmcli`, `nmtui`, etc.) instead of web UI
- **SSH Access:** Configure via standard Linux methods (`ssh-copy-id`, etc.)
- **System Management:** Use `systemctl` and other standard Linux commands
- **Existing Configurations:** Preserved and automatically migrated where applicable

### For Developers
- Network write operations removed from `pkg/nmlite`
- USB gadget initialization now returns errors properly
- Health monitoring available via RPC endpoint
- See `DEVELOPMENT.md` for updated development workflow

## Documentation

Detailed documentation for specific changes:
- `EMBEDDED_LIBRARIES.md` - Third-party library information
- `DEVELOPMENT.md` - Development setup and guidelines
- `CODE_OF_CONDUCT.md` - Community guidelines

## Breaking Changes

- Network configuration can no longer be modified through web UI
- No cloud adoption/configuration UI on device side (cloud features work when UI is cloud-hosted)
- Developer mode toggle removed
- SSH key management removed from UI
- Reboot control removed from UI
- Hardware/display settings removed
- CDROM mode toggle hidden in UI (disk mode default, but CDROM still supported via RPC)
- Video quality factor replaced with direct bitrate specification (kbps)

## Compatibility

- Existing network configurations preserved but ignored
- Cloud tokens and URLs preserved in config files
- SSH keys in `/etc/xkvm/dropbear/.ssh/authorized_keys` still work
- USB gadget automatically recovers from crashes
