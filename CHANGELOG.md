# XKVM Changelog

This document summarizes the major changes made to XKVM during its evolution from the original JetKVM embedded system to a full Linux-based KVM-over-IP solution.

## Recent Updates

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
- SSH keys in `/userdata/dropbear/.ssh/authorized_keys` still work
- USB gadget automatically recovers from crashes
