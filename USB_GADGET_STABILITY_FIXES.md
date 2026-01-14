# USB Gadget Stability Fixes

## Problem Summary

After mounting USB mass storage as disk and then detaching it, the system experienced stability issues with HID devices (keyboard and mouse). The logs showed:

1. HID device files (`/dev/hidg0`, `/dev/hidg1`) becoming unavailable after USB rebind
2. Multiple "failed to open hidg1: no such device or address" errors
3. "write /dev/hidg1: cannot send after transport endpoint shutdown" errors
4. Race conditions between USB state changes and HID operations

## Root Causes

1. **HID files not closed before USB rebind**: When changing mass storage mode (CDROM toggle), the USB gadget is unbound and rebound, but HID file handles remained open
2. **Insufficient delay after USB rebind**: Only 100ms delay after unbind, no delay after bind for HID devices to initialize
3. **No suspension mechanism**: HID operations continued during USB reconfiguration, causing write failures
4. **No retry logic**: Failed to reopen HID devices after USB reconfiguration

## Solutions Implemented

### 1. Added HID Suspension Mechanism

**Files Modified:**
- `internal/usbgadget/usbgadget.go`

**Changes:**
- Added `hidSuspended` flag and `hidSuspendedLock` to UsbGadget struct
- Implemented `SuspendHidOperations()`, `ResumeHidOperations()`, and `IsHidSuspended()` methods
- HID operations are now suspended during USB reconfiguration to prevent write failures

### 2. Proper HID File Lifecycle Management

**Files Modified:**
- `internal/usbgadget/usbgadget.go`

**Changes:**
- Extracted `CloseHidFiles()` method to properly close all HID device files
- Updated `Close()` method to use `CloseHidFiles()`
- HID files are now closed before USB rebind and reopened after

### 3. Increased Delay After USB Rebind

**Files Modified:**
- `internal/usbgadget/udc.go`

**Changes:**
- Added 500ms delay after USB bind operation
- This allows HID devices to fully initialize before operations resume

### 4. Added Retry Logic for HID Device Opening

**Files Modified:**
- `internal/usbgadget/hid_keyboard.go`
- `internal/usbgadget/hid_mouse_absolute.go`
- `internal/usbgadget/hid_mouse_relative.go`

**Changes:**
- Implemented `openKeyboardHidFileWithRetry()` with exponential backoff
- Implemented `openAbsMouseHidFileWithRetry()` with exponential backoff
- Implemented `openRelMouseHidFileWithRetry()` with exponential backoff
- Retries up to 5 times with increasing delays (50ms, 100ms, 200ms, 400ms, 800ms)
- Updated all HID file opening to use retry logic
- Ensures all HID devices (keyboard and both mice) are properly reopened after USB reconfiguration

### 5. HID Suspension Checks in Write Operations

**Files Modified:**
- `internal/usbgadget/hid_keyboard.go`
- `internal/usbgadget/hid_mouse_absolute.go`
- `internal/usbgadget/hid_mouse_relative.go`

**Changes:**
- Added suspension checks in all HID write operations
- Operations return early with descriptive error if HID is suspended
- Prevents "no such device or address" errors during reconfiguration

### 6. Updated USB Configuration Flow

**Files Modified:**
- `internal/usbgadget/config.go`
- `internal/usbgadget/config_tx.go`

**Changes:**
- `configureUsbGadget()` now suspends HID operations and closes HID files before rebind
- `WithTransaction()` resumes HID operations and reopens all HID files after transaction
- Added `ReopenKeyboardHidFile()` method for keyboard recovery
- Added `ReopenAllHidFiles()` method to reopen keyboard and both mouse HID files
- Ensures complete HID device recovery after any USB reconfiguration

### 7. HID RPC Message Handling Protection

**Files Modified:**
- `hidrpc.go`

**Changes:**
- Added suspension check in `handleHidRPCMessage()`
- HID RPC messages are silently dropped during USB reconfiguration
- Prevents cascading errors from client-side input during reconfiguration

## Testing Recommendations

1. **Mount and unmount mass storage multiple times**
   - Verify no HID errors appear in logs
   - Confirm keyboard and mouse continue working after each operation

2. **Toggle CDROM mode while using HID devices**
   - Send keyboard/mouse input during mode change
   - Verify operations resume smoothly after reconfiguration

3. **Stress test with rapid configuration changes**
   - Quickly mount/unmount/toggle CDROM mode
   - Monitor for any race conditions or deadlocks

4. **Long-running stability test**
   - Leave system running with periodic mass storage operations
   - Verify no HID device handle leaks or degradation over time

## Expected Behavior After Fixes

1. No "no such device or address" errors after USB reconfiguration
2. No "transport endpoint shutdown" errors
3. Smooth HID operation resumption after mass storage changes
4. Clean log output with only informational messages about HID suspension/resumption
5. Keyboard and mouse work reliably after any USB gadget reconfiguration

## Monitoring

Watch for these log messages to confirm proper operation:

- `HID operations suspended` - Before USB rebind
- `closed keyboard HID file` - Keyboard HID file being closed
- `closed absolute mouse HID file` - Absolute mouse HID file being closed
- `closed relative mouse HID file` - Relative mouse HID file being closed
- `HID operations resumed` - After USB rebind
- `reopening keyboard HID file after USB reconfiguration` - Keyboard recovery process
- `keyboard HID file reopened successfully` - Successful keyboard recovery
- `reopening absolute mouse HID file after USB reconfiguration` - Mouse recovery process
- `absolute mouse HID file reopened successfully` - Successful absolute mouse recovery
- `reopening relative mouse HID file after USB reconfiguration` - Mouse recovery process
- `relative mouse HID file reopened successfully` - Successful relative mouse recovery

## Root Cause of "transport endpoint shutdown" Error

The error you saw was caused by stale file descriptors. When USB rebind happens:
1. The kernel invalidates all open file descriptors to `/dev/hidg0`, `/dev/hidg1`, `/dev/hidg2`
2. Any attempt to write to these stale file descriptors results in "transport endpoint shutdown"
3. The fix ensures all HID files are closed before rebind and reopened after with retry logic
4. This prevents any stale file descriptors from being used
