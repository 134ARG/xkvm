# XKVM Development Guide

This guide covers building, testing, and developing XKVM.

## Prerequisites

- **Go 1.24.4+** - [Download](https://go.dev/doc/install)
- **Node.js 22.15.0+** - [Download](https://nodejs.org/en/download/)
- **Git** - [Download](https://git-scm.com/downloads)
- **ARM64 cross-compilation toolchain** (for building)
- **XKVM device** (for testing)

### Platform Support

Development works best on Linux or macOS. Windows users should use WSL (Windows Subsystem for Linux).

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/134ARG/xkvm.git
cd kvm
```

### 2. Set Up ARM64 Sysroot

XKVM requires cross-compilation to ARM64. Set up the sysroot once:

```bash
export ARM64_SYSROOT="/opt/arm64-sysroot"
./scripts/setup_arm64_sysroot.sh
```

This downloads and configures the necessary ARM64 libraries for cross-compilation.

### 3. Build Frontend

```bash
cd ui
npm install
npm run build:device
cd ..
```

Or use the Makefile:

```bash
make frontend
```

### 4. Build Backend

```bash
ARM64_SYSROOT=/opt/arm64-sysroot make build_release
```

This creates:
- `bin/xkvm_app` - The main executable
- `bin/xkvm_installer.sh` - Self-extracting installer

### 5. Deploy to Device

```bash
# Copy installer to device
scp bin/xkvm_installer.sh root@<device-ip>:/tmp/

# SSH to device and install
ssh root@<device-ip>
cd /opt
sudo /tmp/xkvm_installer.sh --dir /opt/xkvm
sudo /opt/xkvm/xkvm_app
```

### 6. Access Web UI

Open `http://<device-ip>` in your browser.

## Project Structure

```
/kvm/
├── main.go                   # Application entry point
├── cmd/                      # Command-line tools
├── internal/                 # Internal packages
│   ├── native/               # Native C code interface
│   │   ├── cgo/              # C implementation (video, HDMI, etc.)
│   │   └── proto/            # gRPC protocol definitions
│   ├── usbgadget/            # USB gadget management
│   ├── network/              # Network monitoring
│   ├── ota/                  # Update system (legacy)
│   └── ...                   # Other internal packages
├── pkg/                      # Public packages
│   ├── nmlite/               # Network monitoring lite
│   └── myip/                 # IP detection
├── ui/                       # React frontend
│   ├── src/                  # Source code
│   │   ├── routes/           # Pages
│   │   ├── components/       # UI components
│   │   └── hooks/            # React hooks
│   └── localization/         # i18n translations
├── scripts/                  # Build scripts
├── static/                   # Built frontend (generated)
└── resource/                 # Embedded resources
```

## Development Workflows

### Frontend Development

For rapid UI development without rebuilding the backend:

```bash
cd ui
npm install
npm run dev
```

This starts a development server at `http://localhost:5173` with hot reload.

To connect to a real device, set the proxy:

```bash
cd ui
./dev_device.sh <device-ip>
```

### Backend Development

For backend changes:

```bash
# Make your changes to Go files
# Rebuild
ARM64_SYSROOT=/opt/arm64-sysroot make build_dev

# Deploy to device
scp bin/xkvm_app root@<device-ip>:/opt/xkvm/
ssh root@<device-ip> "systemctl restart xkvm"
```

### Native Code Development

The native layer (C code) handles video capture and hardware control:

```bash
# Build native library only
CMAKE_BUILD_TYPE=Debug ./scripts/build_cgo.sh

# Full rebuild
ARM64_SYSROOT=/opt/arm64-sysroot make build_dev
```

For debugging native code with GDB:

1. Update `TARGET_IP` in `.vscode/settings.json`
2. Set breakpoints in C code
3. Start "Debug Native" configuration in VSCode

## Build Targets

### Development Build

```bash
ARM64_SYSROOT=/opt/arm64-sysroot make build_dev
```

Creates `bin/xkvm_app` with version `0.5.1-dev<timestamp>`.

### Release Build

```bash
ARM64_SYSROOT=/opt/arm64-sysroot make build_release
```

Creates:
- `bin/xkvm_app` - Release binary
- `bin/xkvm_installer.sh` - Self-extracting installer with embedded libraries

### Frontend Only

```bash
make frontend
```

Builds the React UI and places output in `static/`.

### Skip Builds

```bash
# Skip frontend if already built
SKIP_UI_BUILD=1 make build_dev

# Skip native if already built
SKIP_NATIVE_IF_EXISTS=1 make build_dev
```

## Testing

### Unit Tests

```bash
make test
```

Runs all Go unit tests.

### E2E Tests

```bash
make test_e2e
```

Prompts for device IP and runs Playwright end-to-end tests against the device.

### Linting

```bash
make lint
```

Runs `go vet` on all packages.

## Configuration

### Application Config

Configuration is stored in `/etc/xkvm/kvm_config.json` on the device. The config includes:

- Network settings (read-only display)
- USB gadget configuration
- Video codec preferences
- Access credentials

### Environment Variables

Development environment variables:

```bash
# Enable trace logging
export LOG_TRACE_SCOPES="xkvm,cloud,websocket,native,jsonrpc"

# Frontend proxy URL
export XKVM_PROXY_URL="ws://<device-ip>"

# Enable SSL in development
export USE_SSL=true

# Enable sync tracing (debugging)
export ENABLE_SYNC_TRACE=1
```

## Cross-Compilation

XKVM uses cross-compilation from x86_64 to ARM64 (RK3566).

### Toolchain Setup

Install the cross-compilation toolchain:

```bash
# Fedora/RHEL
sudo dnf install gcc-aarch64-linux-gnu gcc-c++-aarch64-linux-gnu

# Ubuntu/Debian
sudo apt install gcc-aarch64-linux-gnu g++-aarch64-linux-gnu
```

### Sysroot Setup

The sysroot provides ARM64 system libraries:

```bash
export ARM64_SYSROOT="/opt/arm64-sysroot"
./scripts/setup_arm64_sysroot.sh
```

This creates a minimal ARM64 root filesystem with required libraries.

### Build Process

The Makefile automatically configures cross-compilation when `ARM64_SYSROOT` is set:

- Sets `CC=aarch64-linux-gnu-gcc`
- Sets `CXX=aarch64-linux-gnu-g++`
- Configures CGO with sysroot paths
- Links against Rockchip vendor libraries

See `internal/native/cgo/README_CROSS_COMPILATION.md` for details.

## Embedded Libraries

XKVM uses vendor-provided libraries for video encoding:

- `librockit.so` - Rockchip video encoding
- `libgraphic_lsf.so` - Graphics layer composition
- `librockchip_mpp.so*` - Media Process Platform

These are embedded in the self-extracting installer. See `EMBEDDED_LIBRARIES.md` for details.

## Localization

The UI uses [paraglide-js](https://inlang.com/m/gerre34r/library-inlang-paraglideJs) for internationalization.

### Adding Translations

1. Add key/value to `ui/localization/messages/en.json`
2. Run `npm run i18n` to validate and sort
3. Use in code: `m.your_key_name()`
4. Run `npm run i18n:machine-translate` to auto-translate other languages

### Translation Commands

```bash
cd ui

# Validate translations
npm run i18n:validate

# Find unused keys
npm run i18n:find-unused

# Find duplicate values
npm run i18n:find-dupes

# Machine translate missing keys
npm run i18n:machine-translate

# Full audit
npm run i18n:audit
```

## Debugging

### View Logs

```bash
# On device
ssh root@<device-ip>
tail -f /var/log/xkvm.log

# Or if running as systemd service
journalctl -u xkvm -f
```

### Common Issues

**Build fails with "ARM64_SYSROOT not set"**
```bash
export ARM64_SYSROOT="/opt/arm64-sysroot"
./scripts/setup_arm64_sysroot.sh
```

**"cannot open shared object file" on device**
- Ensure installer extracted properly
- Check `lib/` directory exists next to `xkvm_app`
- Verify RPATH: `readelf -d xkvm_app | grep RPATH`

**Frontend not updating**
```bash
cd ui
rm -rf node_modules dist
npm install
npm run build:device
```

**USB gadget not working**
```bash
# On device
ssh root@<device-ip>
rm -rf /sys/kernel/config/usb_gadget/xkvm
systemctl restart xkvm
```

### Performance Profiling

Enable developer mode on the device, then access profiling at:

```
http://api:<password>@<device-ip>/developer/pprof/
```

## Code Style

- **Go**: Follow standard Go conventions, use `gofmt`
- **TypeScript**: Use TypeScript for type safety
- **React**: Keep components small and focused
- **C**: Follow Linux kernel style for native code

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly on device
5. Submit a pull request

### Pull Request Checklist

- [ ] Code builds successfully
- [ ] Tests pass (`make test`)
- [ ] Tested on actual device
- [ ] UI strings are localized
- [ ] Documentation updated if needed

## Additional Resources

- **CHANGELOG.md** - Version history and changes
- **EMBEDDED_LIBRARIES.md** - Third-party library information
- **internal/native/cgo/README_CROSS_COMPILATION.md** - Cross-compilation details
