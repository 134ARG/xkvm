# Developer Mode and SSH Key Management Removal Summary

## Overview
Removed redundant advanced settings that are not needed on a full functional Linux system:
1. **Dev Channel Updates** - Obsolete since OTA functionality was already removed
2. **Developer Mode** - Essentially just SSH on/off control, not needed on full Linux
3. **SSH Key Management** - Should be managed via standard Linux tools

## Changes Made

### Backend (Go)

#### `jsonrpc.go`
- **Removed constants:**
  - `devModeFile` (`/userdata/xkvm/devmode.enable`)
  - `sshKeyDir` (`/userdata/dropbear/.ssh`)
  - `sshKeyFile` (`/userdata/dropbear/.ssh/authorized_keys`)

- **Removed types:**
  - `DevModeState` struct
  - `SSHKeyState` struct

- **Removed RPC handler functions:**
  - `rpcGetDevModeState()` - Checked if developer mode file exists
  - `rpcSetDevModeState(enabled bool)` - Created/removed dev mode file and controlled dropbear SSH
  - `rpcGetSSHKeyState()` - Read SSH authorized_keys file
  - `rpcSetSSHKeyState(sshKey string)` - Wrote SSH authorized_keys file with validation

- **Removed RPC handler registrations:**
  - `getDevModeState`
  - `setDevModeState`
  - `getSSHKeyState`
  - `setSSHKeyState`
  - Note: `getDevChannelState` and `setDevChannelState` were already commented out

#### `web.go`
- **Modified `basicAuthProtectedMiddleware` function:**
  - Removed `requireDeveloperMode` parameter
  - Removed developer mode file check logic
  - Now only performs basic auth check

- **Updated debug/profiling routes:**
  - Renamed `developerModeRouter` to `debugRouter`
  - Removed developer mode requirement
  - Routes now protected by password only

#### `internal/utils/ssh.go` and `internal/utils/ssh_test.go`
- **Deleted files:**
  - Removed SSH key validation utility
  - Removed associated tests

### Frontend (UI)

#### `ui/src/routes/devices.$id.settings.advanced.tsx`
- **Removed state variables:**
  - `sshKey` - SSH public key input
  - `devChannel` - Dev channel toggle state
  - `setDeveloperMode` from useSettingsStore

- **Removed useEffect RPC calls:**
  - `getDevModeState`
  - `getSSHKeyState`
  - `getDevChannelState`

- **Removed handler functions:**
  - `handleUpdateSSHKey()` - Updated SSH key via RPC
  - `handleDevModeChange()` - Toggled developer mode
  - `handleDevChannelChange()` - Toggled dev channel

- **Removed UI sections:**
  - Dev Channel checkbox and settings item
  - Developer Mode checkbox and settings item
  - Developer Mode warning card with security warnings
  - Entire SSH access section (textarea, button, instructions)
  - Nested settings group that was gated by developer mode

- **Removed imports:**
  - `JsonRpcError`, `CheckboxWithLabel`, `GridCard`
  - `TextAreaWithLabel`, `InputFieldWithLabel`, `SelectMenuBasic`
  - `isOnDevice`, `checkUpdateComponents`, `UpdateComponents`
  - `SystemVersionInfo`, `FeatureFlag`, `useDeviceUiNavigation`

#### `ui/src/hooks/stores.ts`
- **Removed from settings store:**
  - `developerMode: boolean` state property
  - `setDeveloperMode: (enabled: boolean) => void` function

#### `ui/src/components/ActionBar.tsx`
- **Removed developer mode gating:**
  - Removed `developerMode` from `useSettingsStore`
  - Removed conditional rendering around web terminal button
  - Web terminal button now always visible

#### Localization Files (`ui/localization/messages/*.json`)
- **Removed message keys from all language files:**
  - `advanced_dev_channel_title`
  - `advanced_dev_channel_description`
  - `advanced_developer_mode_title`
  - `advanced_developer_mode_description`
  - `advanced_developer_mode_enabled_title`
  - `advanced_developer_mode_warning_security`
  - `advanced_developer_mode_warning_risks`
  - `advanced_developer_mode_warning_advanced`
  - `advanced_error_set_dev_channel`
  - `advanced_error_set_dev_mode`
  - `advanced_error_update_ssh_key`
  - `advanced_ssh_access_title`
  - `advanced_ssh_access_description`
  - `advanced_ssh_public_key_label`
  - `advanced_ssh_public_key_placeholder`
  - `advanced_ssh_default_user`
  - `advanced_update_ssh_key_button`
  - `advanced_success_update_ssh_key`
  - Plus OTA-related version update keys (already obsolete)

## Behavioral Changes

### Before
1. **Developer Mode:** Required enabling via UI toggle, created `/userdata/xkvm/devmode.enable` file
2. **SSH Access:** Managed through UI with public key textarea, validated and written to dropbear config
3. **Web Terminal:** Only visible when developer mode was enabled
4. **Debug Endpoints:** Required both password AND developer mode file to exist
5. **Dev Channel:** Toggle for receiving development channel updates (OTA already removed)

### After
1. **Developer Mode:** Completely removed - not needed on full Linux
2. **SSH Access:** Managed via standard Linux tools (`ssh-copy-id`, manual authorized_keys editing, etc.)
3. **Web Terminal:** Always visible in action bar
4. **Debug Endpoints:** Protected by password only (no dev mode file check)
5. **Dev Channel:** Completely removed

## Migration Notes

### For Users
- **SSH Access:** If you previously configured SSH keys through the UI, they should still work (existing `/userdata/dropbear/.ssh/authorized_keys` file is preserved). Future SSH key management should be done via standard Linux tools.
- **Web Terminal:** Now always available - no need to enable developer mode first.
- **Debug/Profiling:** Access to `/developer/pprof/*` endpoints now only requires the web interface password.

### For Developers
- The `/userdata/xkvm/devmode.enable` file is no longer checked or created
- SSH key management should be done via standard Linux SSH configuration
- Debug endpoints are always available (when password is set)

## Files Modified
- `jsonrpc.go` - Removed RPC handlers and types
- `web.go` - Simplified middleware
- `ui/src/routes/devices.$id.settings.advanced.tsx` - Removed UI sections
- `ui/src/hooks/stores.ts` - Removed state
- `ui/src/components/ActionBar.tsx` - Removed conditional rendering
- `ui/localization/messages/*.json` - Removed obsolete keys (8 language files)
- `REBOOT_CONTROL_REMOVAL_SUMMARY.md` - Updated status

## Files Deleted
- `internal/utils/ssh.go` - SSH key validation utility
- `internal/utils/ssh_test.go` - SSH key validation tests

## Testing Recommendations
1. Verify web terminal is always visible in action bar
2. Verify advanced settings page loads without errors
3. Verify debug endpoints (`/developer/pprof/*`) are accessible with password
4. Verify no references to removed RPC methods in logs
5. Test that existing SSH access still works (if previously configured)

## Summary
This removal simplifies the codebase by eliminating embedded-device-specific features that are redundant on a full Linux system. SSH and system access should now be managed through standard Linux tools and practices, making the system more maintainable and aligned with standard Linux administration.
