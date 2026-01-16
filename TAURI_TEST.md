# Tauri WebRTC Test - Quick Start (Linux)

## Setup Complete ✓

Tauri has been configured and is ready to test. Here's what was done:

1. ✓ Tauri CLI installed (`@tauri-apps/cli`)
2. ✓ Tauri project initialized in `ui/src-tauri/`
3. ✓ Configuration updated for xKVM
4. ✓ Scripts added to `package.json`

## Prerequisites - Install System Dependencies

Tauri on Linux requires GTK and WebKit dependencies. Install them first:

### Ubuntu/Debian:
```bash
sudo apt update
sudo apt install libwebkit2gtk-4.1-dev \
  build-essential \
  curl \
  wget \
  file \
  libxdo-dev \
  libssl-dev \
  libayatana-appindicator3-dev \
  librsvg2-dev
```

### Arch Linux:
```bash
sudo pacman -Syu
sudo pacman -S webkit2gtk-4.1 base-devel curl wget file openssl appmenu-gtk-module gtk3 libappindicator-gtk3 librsvg libvips
```

### Fedora:
```bash
sudo dnf check-update
sudo dnf install webkit2gtk4.1-devel openssl-devel curl wget file libappindicator-gtk3-devel librsvg2-devel
sudo dnf group install "C Development Tools and Libraries"
```

## How to Test

### Step 1: Install Dependencies (see above)

### Step 2: Build the UI (if not already built)

```bash
cd ui
npm run build
```

This creates the static files in `../static/` directory.

### Step 3: Run Tauri Dev Mode

```bash
cd ui
npm run tauri:dev
```

This will:
- Open your UI in a native Linux window
- Enable dev tools (F12 to open console)

### Step 4: Test WebRTC Connection

1. In the Tauri window, navigate to your device
2. Try connecting to the backend via Tailscale IP
3. Open the dev console (F12)
4. Look for WebRTC logs

### What to Look For

**✅ SUCCESS - If you see:**
- WebRTC connects to `100.78.203.15` (or your Tailscale IP)
- NO "address type mis-match" errors
- Video stream appears
- HID control works

**❌ FAILURE - If you see:**
- Same "Skipping TURN server because of address type mis-match" errors
- Connection fails
- No video

## Alternative: Build and Run

If dev mode has issues, you can build a full app:

```bash
cd ui
npm run tauri:build
```

The app will be in `ui/src-tauri/target/release/bundle/`

## Troubleshooting

### If dependencies are missing:
Follow the installation commands above for your Linux distribution.

### If build fails:
Check that `../static/` directory exists and has your built UI files.

### If window doesn't open:
Check console output for errors.

## Next Steps

- **If it works:** Report success and we'll add features
- **If it fails:** Report the error and we'll try WebSocket fallback

## Files Modified

- `ui/src-tauri/tauri.conf.json` - Updated config
- `ui/package.json` - Added tauri scripts

## Files Created

- `ui/src-tauri/` - Tauri project directory (auto-generated)
- This file (`TAURI_TEST.md`)

## Note

This test is for Linux. The original plan mentioned macOS, but your system is Linux.
The WebRTC behavior should be the same - if Tauri bypasses browser restrictions,
it will work on both platforms.
