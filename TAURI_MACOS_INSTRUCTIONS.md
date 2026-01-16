# XKVM Native App - macOS Build & Release Guide

## Quick Start

### Prerequisites
- **Xcode Command Line Tools**: `xcode-select --install`
- **Node.js 22.x**: Install via Homebrew or from nodejs.org
- **Rust**: `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh`

### Development
```bash
cd ui
npm install
npm run tauri:dev
```

### Building for Release

#### Universal Binary (Intel + Apple Silicon)
```bash
cd ui
npm run tauri:build:mac
```

**Output:** `ui/src-tauri/target/release/bundle/macos/XKVM.app`

#### Platform-Specific Builds
```bash
# Intel only
npm run tauri:build:mac:intel

# Apple Silicon only
npm run tauri:build:mac:arm
```

## First Run Setup

1. Launch XKVM.app
2. Add your backend connection:
   - **Name**: Friendly name (e.g., "My XKVM Device")
   - **URL**: Backend URL (e.g., `https://xkvm.example.com`)
3. Click "Connect"
4. App connects to your backend

## Managing Connections

- **Settings → Connections**: Manage multiple backends
- **Add Connection**: Add new backend
- **Set as Default**: Change active backend
- **Remove**: Delete connection (requires at least one)

## Configuration

Config stored at: `~/.xkvm-native/config.json`

```json
{
  "version": "1.0",
  "connections": [
    {
      "id": "conn_1234567890",
      "name": "My XKVM Device",
      "url": "https://xkvm.example.com",
      "is_default": true,
      "last_connected": "2026-01-16T15:30:00Z"
    }
  ],
  "settings": {
    "auto_connect": true,
    "remember_last_connection": true
  }
}
```

## Distribution

### For Testing
Share the `.app` bundle directly. Users may need to right-click → Open on first launch (macOS Gatekeeper).

### For Production
1. **Code Signing** (requires Apple Developer account):
   ```bash
   # Sign the app
   codesign --deep --force --verify --verbose --sign "Developer ID Application: Your Name" XKVM.app
   
   # Verify signature
   codesign --verify --verbose XKVM.app
   ```

2. **Notarization** (required for distribution):
   ```bash
   # Create DMG
   hdiutil create -volname XKVM -srcfolder XKVM.app -ov -format UDZO XKVM.dmg
   
   # Submit for notarization
   xcrun notarytool submit XKVM.dmg --apple-id your@email.com --team-id TEAMID --wait
   
   # Staple notarization ticket
   xcrun stapler staple XKVM.app
   ```

3. **Create DMG installer**:
   ```bash
   # Tauri can create DMG automatically
   npm run tauri:build:mac -- --bundles dmg
   ```

## Troubleshooting

### App Won't Open
- Right-click → Open (first time only)
- Check Console.app for errors
- Verify config: `cat ~/.xkvm-native/config.json`

### Connection Fails
- Verify backend URL is correct
- Ensure backend is running and accessible
- Check backend has CORS enabled

### Build Fails
- Update Xcode Command Line Tools: `xcode-select --install`
- Update Rust: `rustup update`
- Clean build: `cd ui/src-tauri && cargo clean`

## Version Management

Version is automatically synced from `ota.go` during build. No manual updates needed.

## Backend Requirements

Backend must support CORS. The Go backend includes this automatically via global CORS middleware in `web.go`.

## Related Documentation

- [TAURI_NATIVE_APP.md](TAURI_NATIVE_APP.md) - Complete native app documentation
- [DEVELOPMENT.md](DEVELOPMENT.md) - General development guide
- [README.md](README.md) - Project overview
