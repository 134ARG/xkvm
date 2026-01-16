# XKVM Native App (Tauri)

## Overview

The XKVM Native App is a cross-platform desktop application built with Tauri that provides a native experience for managing XKVM devices. It eliminates browser restrictions and provides better integration with the operating system.

## Features

- **Native Desktop Experience**: Runs as a standalone application on macOS, Windows, and Linux
- **Multi-Backend Support**: Configure and manage multiple XKVM backend connections
- **No Browser Restrictions**: Direct network access without CORS limitations
- **Persistent Configuration**: Connection settings saved locally
- **Auto-versioning**: Version automatically synced with main application

## Architecture

### Frontend (TypeScript/React)
- **Mode Detection**: `isNative` flag distinguishes native app from web/device modes
- **Configuration Store**: Zustand store (`nativeConfigStore.ts`) manages backend connections
- **Dynamic Backend URL**: Backend URL loaded from configuration at runtime
- **Settings UI**: Native-only settings page for connection management

### Backend (Rust)
- **Configuration Management**: Stores connections in platform-specific app data directory
- **Tauri Commands**: Exposed commands for CRUD operations on connections
- **File System Access**: Uses `tauri-plugin-fs` for config persistence
- **HTTP Plugin**: Uses `tauri-plugin-http` for backend communication

## Configuration

### Config File Location

The configuration is stored in the platform-specific app data directory:

- **macOS**: `~/Library/Application Support/com.xkvm.native/config.json`
- **Linux**: `~/.local/share/xkvm-native/config.json` (or `$XDG_DATA_HOME/xkvm-native/config.json`)
- **Windows**: `C:\Users\<User>\AppData\Roaming\com.xkvm.native\config.json`

These locations follow platform conventions and are automatically managed by Tauri.

### Config Structure
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

## Development

### Prerequisites
- Node.js 22.x
- Rust (latest stable)
- Platform-specific dependencies:
  - **macOS**: Xcode Command Line Tools
  - **Linux**: See [TAURI_TEST.md](TAURI_TEST.md) for distribution-specific packages
  - **Windows**: Visual Studio Build Tools

### Setup
```bash
cd ui
npm install
```

### Development Mode
```bash
npm run tauri:dev
```

### Building

#### All Platforms (current platform)
```bash
npm run tauri:build
```

#### Platform-Specific Builds

**macOS:**
```bash
npm run tauri:build:mac          # Universal binary (Intel + Apple Silicon)
npm run tauri:build:mac:intel    # Intel only
npm run tauri:build:mac:arm      # Apple Silicon only
```

**Windows:**
```bash
npm run tauri:build:windows
```

**Linux:**
```bash
npm run tauri:build:linux              # Default bundle
npm run tauri:build:linux:appimage     # AppImage
npm run tauri:build:linux:deb          # Debian package
```

### Build Output Locations

- **macOS**: `ui/src-tauri/target/release/bundle/macos/XKVM.app`
- **Windows**: `ui/src-tauri/target/release/bundle/msi/XKVM_*.msi`
- **Linux**: `ui/src-tauri/target/release/bundle/[appimage|deb]/`

## Version Management

Version is automatically synced from `ota.go` during build:

```bash
# Version sync happens automatically during build
npm run build:tauri
```

The sync script (`ui/scripts/sync-version.cjs`) updates:
- `ui/src-tauri/tauri.conf.json`
- `ui/package.json`

## Key Files

### Configuration
- `ui/src-tauri/tauri.conf.json` - Tauri app configuration
- `ui/src-tauri/capabilities/default.json` - Security permissions
- `ui/src-tauri/Cargo.toml` - Rust dependencies

### Source Code
- `ui/src-tauri/src/config.rs` - Configuration management (Rust)
- `ui/src/stores/nativeConfigStore.ts` - Configuration store (TypeScript)
- `ui/src/routes/settings.connections.tsx` - Connection settings UI
- `ui/src/main.tsx` - Mode detection and routing
- `ui/src/ui.config.ts` - Dynamic backend URL configuration

### Build Scripts
- `ui/scripts/sync-version.cjs` - Version synchronization
- `ui/package.json` - Build scripts

## Usage

### First Launch

1. Launch the XKVM app
2. Navigate to Settings → Connections
3. Add your first backend connection:
   - **Name**: Friendly name for the connection
   - **URL**: Backend URL (e.g., `https://xkvm.example.com`)
4. The connection is automatically set as default

### Managing Connections

- **Add Connection**: Click "Add Connection" button
- **Remove Connection**: Click "Remove" button (requires at least one connection)
- **Set Default**: Click "Set as Default" to change the active connection
- **View Last Connected**: See when each connection was last used

### Switching Backends

The app uses the default connection on startup. To switch:
1. Go to Settings → Connections
2. Click "Set as Default" on the desired connection
3. Restart the app

## Security

### Permissions

The app requests the following permissions:
- **HTTP Access**: Connect to any HTTPS backend (user-configured)
- **File System**: Read/write config file in app data directory
- **WebSocket**: Real-time communication with backend

### Content Security Policy

CSP allows:
- HTTPS connections to any domain (for user-configured backends)
- WebSocket connections (WSS)
- Local resources (fonts, images, scripts)

## Troubleshooting

### App Won't Start
- Check console logs (Cmd+Option+I on macOS)
- Verify config file is valid JSON (see Config File Location above)
- Delete config file to reset (location varies by platform - see above)

### Connection Fails
- Verify backend URL is correct and accessible
- Check backend is running and reachable
- Ensure backend has CORS enabled (see backend setup)

### Build Fails
- Ensure all prerequisites are installed
- Run `cargo check` in `ui/src-tauri/` to verify Rust setup
- Run `npm install` to ensure dependencies are up to date

## Backend Requirements

The backend must support CORS for native app connections. The Go backend includes a global CORS middleware that handles this automatically.

### CORS Configuration (Already Implemented)

The backend (`web.go`) includes:
```go
// Global CORS middleware
r.Use(func(c *gin.Context) {
    origin := c.Request.Header.Get("Origin")
    if origin != "" {
        c.Header("Access-Control-Allow-Origin", origin)
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Header("Access-Control-Allow-Credentials", "true")
    }
    
    if c.Request.Method == "OPTIONS" {
        c.AbortWithStatus(http.StatusNoContent)
        return
    }
    
    c.Next()
})
```

## Future Enhancements

Potential features for future releases:
- System tray integration
- Auto-update mechanism
- Native notifications
- Global keyboard shortcuts
- Connection profiles
- Auto-discovery of local devices
- Connection health monitoring

## Related Documentation

- [TAURI_MACOS_INSTRUCTIONS.md](TAURI_MACOS_INSTRUCTIONS.md) - macOS setup guide
- [TAURI_TEST.md](TAURI_TEST.md) - Linux setup guide
- [DEVELOPMENT.md](DEVELOPMENT.md) - General development guide
- [README.md](README.md) - Project overview
