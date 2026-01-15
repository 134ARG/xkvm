# Mass Storage Cleanup on Startup

## Problem

After shutting down XKVM and restarting it, the host computer still sees a CD-ROM or disk device attached from previous runs. This is because the mass storage configuration persists in the USB gadget configfs even after the XKVM application restarts.

## Root Cause

The USB gadget configfs is a kernel-level configuration that persists across application restarts. When XKVM starts up and cleans up the stale gadget, it recreates the mass storage function with the default configuration. The issue was:

1. Default `cdrom` attribute was set to `"1"` (CDROM mode)
2. This caused the host to see a CD-ROM device even with no image mounted
3. The stale configuration from previous runs could persist

## Solution

Changed the default mass storage configuration to Disk mode and ensured proper cleanup.

### Implementation

**File:** `internal/usbgadget/mass_storage.go`

**Changes:**
1. Changed default `cdrom` attribute from `"1"` to `"0"` (Disk mode)
2. The `file` attribute remains `"\n"` (empty) by default

```go
var massStorageLun0Config = gadgetConfigItem{
    order: 3001,
    path:  []string{"functions", "mass_storage.usb0", "lun.0"},
    attrs: gadgetAttributes{
        "cdrom":     "0", // Default to Disk mode (was "1" for CDROM)
        "ro":        "1",
        "removable": "1",
        "file":      "\n", // Empty by default
        "inquiry_string": "XKVM  Virtual Media",
    },
}
```

**File:** `usb_mass_storage.go`

**Function:** `setInitialVirtualMediaState()`

Added attempt to clear mass storage file on startup (logs debug message if it fails, which is expected after cleanup).

## Behavior

### Before Fix
1. Mount a disk image
2. Shut down XKVM without unmounting
3. Restart XKVM
4. Host sees a CD-ROM device (even if no image is mounted)
5. Stale configuration persists

### After Fix
1. Mount a disk image
2. Shut down XKVM without unmounting
3. Restart XKVM
4. **USB gadget is cleaned up and recreated with Disk mode**
5. **No stale devices visible to host**
6. Clean state ready for new mounts

## Technical Details

### Mass Storage Configuration

The mass storage gadget function has these key attributes:
- `file` - Path to backing file (disk image)
- `cdrom` - Whether to present as CD-ROM (1) or Disk (0)
- `ro` - Read-only flag
- `removable` - Removable media flag

### Cleanup Process

1. **Stale Gadget Cleanup** (`internal/usbgadget/cleanup.go`)
   - Unbinds UDC
   - Removes symlinks
   - Removes function directories
   - Removes gadget directory

2. **Gadget Initialization** (`internal/usbgadget/config.go`)
   - Recreates gadget with default configuration
   - Mass storage created with `cdrom="0"` and `file="\n"`

3. **State Initialization** (`usb_mass_storage.go`)
   - Attempts to clear any stale file path
   - Reads current state
   - Sets initial state to null

## Error Handling

The cleanup attempt in `setInitialVirtualMediaState()` may fail with "bad address" error if the gadget was just cleaned up. This is expected and logged at debug level:

```
DBG could not clear mass storage (expected if gadget was just cleaned up)
```

The application continues normally as the gadget is recreated with empty file attribute anyway.

## Logging

On startup, you'll see:
```
INF found stale USB gadget, cleaning up
INF unbinding UDC during cleanup
INF stale USB gadget cleaned up successfully
INF clearing any stale mass storage configuration from previous runs
DBG could not clear mass storage (expected if gadget was just cleaned up)
INF initial virtual media state set initial_virtual_media_state=null
```

This confirms the cleanup happened successfully.

## Related Changes

This fix works in conjunction with:
- **UI Changes** - CDROM mode hidden from frontend (UI_MOUNT_CHANGES.md)
- **USB Stability Fixes** - HID device handling improvements (USB_GADGET_STABILITY_FIXES.md)
- **Cleanup Code** - Stale gadget removal (internal/usbgadget/cleanup.go)

## Testing

To verify the fix:
1. Mount a disk image
2. Shut down XKVM (kill the process or reboot the device)
3. Restart XKVM
4. Check host computer - should see no stale devices
5. Check logs - should see cleanup messages
6. Mount a new image - should work as Disk mode

## Why Disk Mode Default?

Changed from CDROM to Disk mode because:
1. CDROM mode toggle causes USB reconfiguration issues
2. Disk mode is more versatile (works for both .iso and .img files)
3. Modern systems handle disk-mounted ISOs just fine
4. Eliminates USB stability problems from mode switching
