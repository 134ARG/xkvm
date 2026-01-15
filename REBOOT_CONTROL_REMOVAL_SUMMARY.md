# Reboot Control Removal Summary

## Overview
Removed the manual reboot device control functionality from the settings/general page as part of the transition to a full functional Linux system. System management functions like reboot should be handled through standard Linux tools rather than through the web UI.

## Changes Made

### Backend Changes

#### 1. `jsonrpc.go`
- **Removed** `rpcReboot()` function (previously stubbed out)
- **Removed** `"reboot"` entry from `rpcHandlers` map

#### 2. `native.go`
- **Removed** `"reboot"` case from the native RPC event handler switch statement
- **Removed** reboot call after `resetConfig` case (line that called `rpcReboot(true)`)

#### 3. `network.go`
- **Removed** `PostRebootAction` struct (no longer needed after OTA removal)

#### 4. `hw.go`
- No changes needed - `hwReboot()` function was already removed during OTA cleanup

### Frontend Changes

#### 1. Route Files
- **Deleted** `ui/src/routes/devices.$id.settings.general.reboot.tsx` (entire file)
- **Modified** `ui/src/main.tsx`:
  - Removed lazy import for `SettingsGeneralRebootRoute`
  - Removed route definition with path `"reboot"`

#### 2. Settings UI
- **Modified** `ui/src/routes/devices.$id.settings.general._index.tsx`:
  - Removed reboot button and associated `SettingsItem` component

#### 3. Main Device Route
- **Modified** `ui/src/routes/devices.$id.tsx`:
  - Removed `PostRebootAction` import
  - Removed `RebootingOverlay` import and usage
  - Removed `rebootState` and `setRebootState` from store destructuring
  - Removed `willReboot` event handler
  - Removed `setRebootState` call in WebSocket onOpen handler
  - Removed reboot-related dependencies from useEffect dependency array
  - Cleaned up commented code referencing `rebootState`

#### 4. Video Overlay Component
- **Modified** `ui/src/components/VideoOverlay.tsx`:
  - Removed `PostRebootAction` import
  - Removed `RebootingOverlay` component (entire component and interface)

#### 5. Failsafe Mode Overlay
- **Modified** `ui/src/components/FailSafeModeOverlay.tsx`:
  - Removed "Reboot Device" button from failsafe mode overlay

#### 6. State Management
- **Modified** `ui/src/hooks/stores.ts`:
  - Removed `PostRebootAction` type definition
  - Removed `rebootState` property from `UIState` interface
  - Removed `setRebootState` method from `UIState` interface
  - Removed `rebootState` initialization in `useUiStore`

#### 7. Localization Files
Removed the following keys from all 8 language files (`en.json`, `da.json`, `de.json`, `es.json`, `fr.json`, `it.json`, `nb.json`, `sv.json`, `zh.json`):
- `general_reboot_device`
- `general_reboot_device_description`
- `general_reboot_title`
- `general_reboot_description`
- `general_reboot_yes_button`
- `general_reboot_no_button`

## Files Modified

### Deleted
1. `ui/src/routes/devices.$id.settings.general.reboot.tsx`

### Modified
1. `jsonrpc.go`
2. `native.go`
3. `network.go`
4. `ui/src/main.tsx`
5. `ui/src/routes/devices.$id.settings.general._index.tsx`
6. `ui/src/routes/devices.$id.tsx`
7. `ui/src/components/VideoOverlay.tsx`
8. `ui/src/components/FailSafeModeOverlay.tsx`
9. `ui/src/hooks/stores.ts`
10. `ui/localization/messages/en.json`
11. `ui/localization/messages/da.json`
12. `ui/localization/messages/de.json`
13. `ui/localization/messages/es.json`
14. `ui/localization/messages/fr.json`
15. `ui/localization/messages/it.json`
16. `ui/localization/messages/nb.json`
17. `ui/localization/messages/sv.json`
18. `ui/localization/messages/zh.json`

## Rationale

As the system transitions to a full functional Linux environment, system management operations like rebooting should be handled through standard Linux tools and interfaces (e.g., `systemctl reboot`, SSH access) rather than through a proprietary web UI. This change:

1. **Aligns with Linux conventions** - System administration is typically done through standard tools
2. **Reduces attack surface** - Removes a privileged operation from the web interface
3. **Simplifies codebase** - Removes unnecessary abstraction layers for system operations
4. **Improves security** - System-level operations should require appropriate system-level access

## Related Changes

This removal was done after the OTA (Over-The-Air update) functionality was removed. The OTA system previously used the reboot functionality for automatic reboots after updates, which is why some reboot-related code (like `RebootingOverlay` and `PostRebootAction`) existed. With OTA removed, the manual reboot control was the only remaining use case.

## Testing Recommendations

1. Verify the settings/general page loads without errors
2. Confirm no broken navigation links to `/settings/general/reboot`
3. Test that failsafe mode overlay displays correctly without the reboot button
4. Verify no console errors related to missing reboot state or functions
5. Test that the application builds successfully (both frontend and backend)

## Future Considerations

Other system administration functions that may need similar treatment:
- Reset Configuration (`resetConfig`) - Should use standard Linux config management
- SSH Key Management (`setSSHKeyState`) - **REMOVED** - Use standard SSH configuration files
- Developer Mode (`setDevModeState`) - **REMOVED** - Not needed on full Linux systems
- Loopback-Only Mode (`setLocalLoopbackOnly`) - Should use standard firewall tools
- USB Emulation Control (`setUsbEmulationState`) - Depends on whether USB gadget functionality is retained

These functions should be evaluated on a case-by-case basis as the transition to full Linux continues.
