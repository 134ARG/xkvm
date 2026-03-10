# XKVM Packaging

This directory contains files for building DEB packages.

## Structure

- `version.txt` - Package version (updated by `scripts/bump_version.sh`)
- `bin/xkvm-wrapper.sh` - Wrapper script installed to `/usr/bin/xkvm`
- `systemd/xkvm.service` - Systemd service unit file
- `deb/copyright` - Debian copyright file

## Building Packages

```bash
# Build DEB package
make build_deb
```

## Dependencies

The package automatically installs required dependencies:

- `libmali-bifrost-g52-g13p0-gbm` - Mali GPU library for RK3566
  - Required by `libgraphic_lsf.so` (Rockchip graphics library)
  - Provides hardware-accelerated graphics support

If the Mali library is missing, you'll see errors like:
```
RTLibraryLoader failed to load library(libgraphic_lsf.so) error: No such file or directory
```

## Package Contents

The package installs:
- `/usr/bin/xkvm` - Wrapper script (sets LD_LIBRARY_PATH)
- `/usr/lib/xkvm/xkvm_app` - Main binary
- `/usr/lib/xkvm/*.so*` - Bundled libraries (librockit, libgraphic_lsf, librockchip_mpp)
- `/etc/systemd/system/xkvm.service` - Systemd service (not enabled by default)
- `/usr/share/doc/xkvm/LICENSE` - License file
- `/usr/share/doc/xkvm/copyright` - Copyright file

## Post-Install

The service is installed but NOT enabled or started automatically.

To start:
```bash
sudo systemctl start xkvm
```

To enable on boot:
```bash
sudo systemctl enable xkvm
```

## Version Management

Update version across all files:
```bash
./scripts/bump_version.sh 0.1.3
```

This updates:
- Makefile
- ota.go
- ui/src-tauri/Cargo.toml
- packaging/version.txt
- ui/package.json
- ui/src-tauri/tauri.conf.json

## Troubleshooting

### Missing Mali Library

If you see library loading errors, verify the Mali library is installed:

```bash
# Check if libmali is installed
ldconfig -p | grep libmali

# Install manually if needed
sudo apt install libmali-bifrost-g52-g13p0-gbm
```

### Library Path Issues

The wrapper script at `/usr/bin/xkvm` sets `LD_LIBRARY_PATH=/usr/lib/xkvm` to ensure bundled libraries are found. If you run `xkvm_app` directly, you may need to set this manually:

```bash
export LD_LIBRARY_PATH=/usr/lib/xkvm:$LD_LIBRARY_PATH
/usr/lib/xkvm/xkvm_app
```
