# UI Mount Changes - CDROM and URL Mount Removal

## Summary

Removed CDROM/DVD mount mode and URL mount options from the frontend to prevent USB stability issues caused by CDROM mode toggling.

## Changes Made

### 1. Hidden CDROM Mount Option

**File:** `ui/src/routes/devices.$id.mount.tsx`

**Function:** `UsbModeSelector`

- Commented out the CDROM radio button option
- Only "Disk" mode is now visible to users
- Added comment: `/* CDROM option hidden - causes USB stability issues */`

### 2. Hidden URL Mount Option

**File:** `ui/src/routes/devices.$id.mount.tsx`

**Function:** `ModeSelectionView`

- Commented out the URL mount card in the mode selection grid
- Only "JetKVM Storage" option is now visible
- Added comment: `// URL mount hidden - experimental feature not needed`

### 3. Updated Default Mode to "Disk"

**File:** `ui/src/routes/devices.$id.mount.tsx`

**Changes:**
- `UrlView`: Changed default `usbMode` from `"CDROM"` to `"Disk"`
- `DeviceFileView`: Changed default `usbMode` from `"CDROM"` to `"Disk"`
- `handleUrlChange`: Removed logic that auto-selected CDROM for .iso files, now always uses "Disk"
- `handleOnSelectFile`: Removed logic that auto-selected CDROM for .iso files, now always uses "Disk"

## User Impact

### Before
- Users could select between "URL Mount" and "JetKVM Storage"
- Users could toggle between "CDROM" and "Disk" modes
- Toggling to CDROM mode caused USB rebind and HID device issues

### After
- Users only see "JetKVM Storage" option (URL mount hidden)
- Users only see "Disk" mode (CDROM option hidden)
- All images mount as Disk mode, preventing USB reconfiguration issues
- ISO files can still be mounted, but as Disk mode instead of CDROM

## Technical Rationale

The CDROM mode toggle triggers a USB gadget reconfiguration that:
1. Unbinds and rebinds the USB controller
2. Invalidates HID device file handles
3. Causes "transport endpoint shutdown" errors
4. Results in temporary loss of keyboard/mouse functionality

By removing the CDROM option and forcing Disk mode:
- No USB reconfiguration occurs during mount/unmount
- HID devices remain stable
- Better user experience with no input interruptions

## Code Preservation

All code for CDROM and URL mount functionality remains in place, just commented out. This allows for:
- Easy re-enabling if issues are resolved
- Reference for future development
- Minimal code changes (comments only)

## Testing

- UI builds successfully without errors
- Mount dialog now shows only "JetKVM Storage" option
- File selection shows only "Disk" mode
- All images mount as Disk mode by default
