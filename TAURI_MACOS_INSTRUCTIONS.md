# Tauri WebRTC Test - macOS Instructions

## Quick Start (5 minutes)

### Prerequisites Check

1. **Xcode Command Line Tools** (required)
   ```bash
   xcode-select --install
   ```
   If already installed, you'll see: "command line tools are already installed"

2. **Node.js** (required - v22.x)
   
   **Option A: Using Homebrew (recommended)**
   ```bash
   # Install Homebrew if not installed
   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
   
   # Install Node.js
   brew install node@22
   ```
   
   **Option B: Direct Download**
   - Download from: https://nodejs.org/en/download/
   - Install the macOS installer (.pkg)
   - Choose LTS version (v22.x)
   
   Verify:
   ```bash
   node --version  # Should show v22.x
   npm --version   # Should show v10.x
   ```

3. **Rust** (required)
   ```bash
   curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
   source $HOME/.cargo/env
   ```
   
   Verify:
   ```bash
   rustc --version
   ```

### Setup Steps

**Step 1: Clone/Pull Latest Code**
```bash
cd /path/to/xkvm
git pull  # or however you sync code
```

The Tauri setup is already in the repo:
- `ui/src-tauri/` - Tauri project
- `ui/package.json` - Has tauri scripts
- `ui/src-tauri/tauri.conf.json` - Configured

**Step 2: Install Dependencies**
```bash
cd ui
npm install
```

**Step 3: Build UI**
```bash
npm run build
```

This creates files in `../static/` directory.

**Step 4: Run Tauri Dev Mode**
```bash
npm run tauri:dev
```

First run will take 5-10 minutes to compile Rust dependencies.
Subsequent runs are much faster (~30 seconds).

A native macOS window will open with your xKVM UI.

### Testing WebRTC

**Step 5: Test Connection**

1. In the Tauri window, navigate to your device
2. Try connecting via Tailscale IP (e.g., `100.78.203.15`)
3. Open dev console: **Cmd+Option+I**
4. Watch for WebRTC logs

**What to Look For:**

✅ **SUCCESS:**
```
[ICE] New candidate gathered: type: "relay", address: "100.78.203.15"
[WebRTC] Connection state: connected
```
- NO "address type mis-match" errors
- Video appears
- Mouse/keyboard works

❌ **FAILURE:**
```
Skipping TURN server because of address type mis-match
ICE failed
```
- Same errors as browser
- No connection

### Alternative: Build Release App

If dev mode has issues:

```bash
cd ui
npm run tauri:build
```

The app will be at:
```
ui/src-tauri/target/release/bundle/macos/xKVM.app
```

Double-click to run.

## Troubleshooting

### "xcrun: error: invalid active developer path"
Install Xcode Command Line Tools:
```bash
xcode-select --install
```

### "cargo: command not found"
Install Rust:
```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env
```

### "npm run tauri:dev" fails
1. Make sure UI is built: `npm run build`
2. Check `../static/` exists and has files
3. Check Rust is installed: `rustc --version`

### Window opens but shows blank/error
1. Check `../static/index.html` exists
2. Try building UI again: `npm run build`
3. Check console for errors

### "Developer cannot be verified" warning
Right-click the app → Open (instead of double-clicking)

## What Happens Next

### If WebRTC Works ✅
Report success! We'll then:
1. Add native features (system tray, auto-update)
2. Set up proper distribution
3. Code signing for release

### If WebRTC Fails ❌
We'll try:
1. WebSocket fallback approach
2. External TURN server
3. Other workarounds

## Commands Summary

```bash
# One-time setup
cd ui
npm install

# Every time you want to test
npm run build          # Build UI
npm run tauri:dev      # Run in dev mode

# Or build release app
npm run tauri:build    # Creates .app bundle
```

## File Locations

- **Source code:** `ui/src-tauri/src/`
- **Config:** `ui/src-tauri/tauri.conf.json`
- **Built app (dev):** Opens automatically
- **Built app (release):** `ui/src-tauri/target/release/bundle/macos/xKVM.app`

## Notes

- First build takes 5-10 minutes (compiling Rust)
- Subsequent builds are fast (~30 seconds)
- Dev mode has hot reload for UI changes
- Rust changes require restart
- The app has full network access (no browser restrictions)

## Expected Behavior

The Tauri app should behave exactly like a browser, except:
- It can access Tailscale IPs without restrictions
- WebRTC should work with `100.78.203.15` addresses
- No "address type mis-match" errors

This is the whole point of the test!

## Quick Test Checklist

- [ ] Xcode Command Line Tools installed
- [ ] Rust installed
- [ ] `npm install` completed
- [ ] `npm run build` completed
- [ ] `npm run tauri:dev` opens window
- [ ] Can navigate to device in UI
- [ ] Try connecting via Tailscale IP
- [ ] Check console for WebRTC logs
- [ ] Report: ✅ works or ❌ fails

---

**Ready to test? Just run:**
```bash
cd ui
npm install
npm run build
npm run tauri:dev
```
