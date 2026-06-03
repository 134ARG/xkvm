# XKVM

[English](README.en.md) | [简体中文](README.md)

XKVM is a high-performance, open-source, 100% local KVM-over-IP solution running on a full Linux system. It provides remote keyboard/mouse input and 1080p 60fps hardware-encoded video streaming for remote machine management.

This thing is currently only adapted to the KVM platform I built myself: Radxa Zero 3 (rk3566) + tc358743 HDMI-CSI capture card. My capture card hardware is opensourced on OSHWHub [here](https://oshwhub.com/gth38/hdmi-csi). It should be fairly easy to port to other RK platforms.

## Overview

XKVM is based on JetKVM. I removed a bunch of stuff I do not need, such as cloud connection, network configuration, device management, and so on. After all, this runs on a full Linux system instead of buildroot. I added 2M-20M bitrate control, H.265 encoding, and customizable GPIO-based ATX power control/monitoring. I also added Linux, Windows, and macOS native frontend apps based on Tauri, mainly to work around the annoying browser WebRTC security restrictions when connecting through VPNs like Tailscale.

This project is mainly for personal use. I may drop it at any time.

## File Paths

XKVM follows standard Linux FHS conventions. Paths can be overridden via environment variables.

| Path | Env Override | Purpose |
|---|---|---|
| `/etc/xkvm/` | `XKVM_CONFIG_DIR` | Configuration files |
| `/var/lib/xkvm/` | `XKVM_DATA_DIR` | Data files |
| `/var/log/xkvm/` | `XKVM_LOG_DIR` | Logs |

Main files:

| File | Description |
|---|---|
| `/etc/xkvm/kvm_config.json` | Main config (USB, video, auth, macros, etc.) |
| `/etc/xkvm/tls/` | TLS certificates |
| `/etc/xkvm/.native-debug-mode` | Create this file to enable native debug mode |
| `/var/lib/xkvm/images/` | Virtual media images (ISO/disk) |
| `/var/lib/xkvm/crashdump/` | Crash logs |
| `/var/log/xkvm/last.log` | Application stdout/stderr log |

When running through systemd, these directories are automatically created by `ConfigurationDirectory`, `StateDirectory`, and `LogsDirectory`.

## Demo

Screenshots below are from the Tauri native desktop UI (macOS):

Video settings: 2M-20M bitrate control and H.264/H.265 codec switching (H.265 is macOS Safari only).

<img src="images/video-demo.png" />

Hardware settings: USB device types, GPIO ATX control, case sensors, and VFD display configuration.

<img src="images/gpio-sensor-vfd-settings.png" />

Sensors and host metrics: the main UI can show device environment sensors, host load, temperature, network traffic, and runtime state.

<img src="images/sensors-demo.png" />

## Contributing

Forks are welcome.

## License

See [LICENSE](LICENSE) for details.
