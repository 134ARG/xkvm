# USB Gadget Troubleshooting Guide

## Common Issues and Solutions

### Issue 1: USB Gadget Fails to Initialize After Crash

**Symptoms:**
- Error: "failed to init USB gadget"
- Error: "directory already exists"
- USB devices not working after restart

**Cause:**
Stale USB gadget directories left from previous crash.

**Solution:**
The system now automatically cleans up stale state on startup. Check logs for:
```
found stale USB gadget, cleaning up
stale USB gadget cleaned up successfully
```

**Manual Recovery:**
If automatic cleanup fails:
```bash
# Unbind UDC
echo "" > /sys/kernel/config/usb_gadget/xkvm/UDC

# Remove gadget
rm -rf /sys/kernel/config/usb_gadget/xkvm
```

### Issue 2: "USB gadget not initialized" Errors

**Symptoms:**
- RPC calls return "USB gadget not initialized"
- Keyboard/mouse not working
- Virtual media not mounting

**Cause:**
USB gadget failed to initialize or crashed.

**Solution:**
1. Check initialization logs:
```bash
journalctl -u xkvm | grep "USB gadget"
```

2. Check health status via RPC:
```javascript
rpc.call("getUsbGadgetHealth")
```

3. Restart XKVM service:
```bash
systemctl restart xkvm
```

### Issue 3: Lock Acquisition Timeout

**Symptoms:**
- Error: "timeout acquiring USB gadget lock after 10s"
- Slow initialization

**Cause:**
Another process is holding the lock or stale lock file.

**Solution:**
1. Check for other XKVM processes:
```bash
ps aux | grep xkvm
```

2. Remove stale lock file:
```bash
rm -f /var/lock/xkvm-usb.lock
```

3. Restart service

### Issue 4: UDC Not Bound

**Symptoms:**
- USB state shows "not attached"
- Host computer doesn't detect USB devices
- Health check shows `udc_bound: false`

**Cause:**
UDC failed to bind or was unbound.

**Solution:**
1. Check UDC availability:
```bash
ls /sys/devices/platform/usbdrd/*.usb
```

2. Check current binding:
```bash
cat /sys/kernel/config/usb_gadget/xkvm/UDC
```

3. Manual rebind:
```bash
# Find UDC name
UDC=$(ls /sys/devices/platform/usbdrd/*.usb | head -1 | xargs basename)

# Bind
echo $UDC > /sys/kernel/config/usb_gadget/xkvm/UDC
```

### Issue 5: Transient Errors During Init

**Symptoms:**
- Errors like "device or resource busy"
- "resource temporarily unavailable"
- Initialization succeeds after retry

**Cause:**
Temporary resource contention.

**Solution:**
The system now automatically retries with exponential backoff. Check logs for:
```
retrying operation
operation succeeded after retry
```

If retries fail, check for:
- Other processes accessing USB gadget
- Kernel module issues
- Hardware problems

### Issue 6: Health Check Failures

**Symptoms:**
- Logs show "USB gadget health check failed"
- Automatic recovery attempts
- Intermittent USB issues

**Cause:**
USB gadget in unhealthy state.

**Solution:**
1. Check health status:
```javascript
rpc.call("getUsbGadgetHealth")
```

2. Review health check logs:
```bash
journalctl -u xkvm | grep "health check"
```

3. If recovery fails repeatedly:
   - Check kernel logs: `dmesg | grep usb`
   - Verify hardware connections
   - Check for kernel module issues

## Diagnostic Commands

### Check USB Gadget State
```bash
# Check if gadget exists
ls -la /sys/kernel/config/usb_gadget/xkvm

# Check UDC binding
cat /sys/kernel/config/usb_gadget/xkvm/UDC

# Check USB state
cat /sys/class/udc/*/state

# List functions
ls -la /sys/kernel/config/usb_gadget/xkvm/functions/

# List config symlinks
ls -la /sys/kernel/config/usb_gadget/xkvm/configs/c.1/
```

### Check Logs
```bash
# Recent USB gadget logs
journalctl -u xkvm --since "10 minutes ago" | grep -i usb

# Initialization logs
journalctl -u xkvm -b | grep "initUsbGadget"

# Health check logs
journalctl -u xkvm | grep "health check"

# Cleanup logs
journalctl -u xkvm | grep "cleanup"
```

### Check Lock Status
```bash
# Check if lock file exists
ls -la /var/lock/xkvm-usb.lock

# Check which process holds lock (if any)
lsof /var/lock/xkvm-usb.lock
```

### Check Kernel Modules
```bash
# List USB gadget modules
lsmod | grep usb

# Check for errors
dmesg | grep -i "usb\|gadget" | tail -20
```

## RPC Debugging

### Get Health Status
```javascript
// Check USB gadget health
const health = await rpc.call("getUsbGadgetHealth");
console.log(health);
// {
//   healthy: true,
//   last_check: "2024-01-14T10:30:00Z",
//   udc_bound: true,
//   gadget_exists: true,
//   config_exists: true
// }
```

### Get USB State
```javascript
const state = await rpc.call("getUSBState");
console.log(state); // "configured", "not attached", etc.
```

### Test USB Operations
```javascript
// Test keyboard
await rpc.call("keyboardReport", { modifier: 0, keys: [0x04] }); // 'a'

// Test mouse
await rpc.call("absMouseReport", { x: 100, y: 100, buttons: 0 });
```

## Prevention Best Practices

### 1. Monitor Health
Enable health monitoring in production:
```go
gadget.StartHealthCheck(ctx)
```

### 2. Graceful Shutdown
Ensure proper cleanup on shutdown:
```go
defer gadget.Cleanup()
```

### 3. Error Handling
Always check for errors:
```go
if err := gadget.KeyboardReport(mod, keys); err != nil {
    log.Error().Err(err).Msg("keyboard report failed")
}
```

### 4. Regular Monitoring
Monitor logs for warnings:
```bash
journalctl -u xkvm -f | grep -i "warn\|error"
```

## Advanced Debugging

### Enable Trace Logging
Set log level to TRACE for detailed debugging:
```json
{
  "default_log_level": "TRACE"
}
```

### Kernel USB Debugging
Enable USB debugging in kernel:
```bash
echo 'module usbcore =p' > /sys/kernel/debug/dynamic_debug/control
echo 'module usb_f_hid =p' > /sys/kernel/debug/dynamic_debug/control
```

### Capture USB Traffic
Use usbmon to capture USB traffic:
```bash
modprobe usbmon
cat /sys/kernel/debug/usb/usbmon/0u
```

## Getting Help

When reporting issues, include:

1. **System Information:**
   - XKVM version
   - Kernel version: `uname -r`
   - Hardware model

2. **Logs:**
   ```bash
   journalctl -u xkvm --since "1 hour ago" > xkvm.log
   dmesg > kernel.log
   ```

3. **Health Status:**
   ```bash
   # Via RPC or check logs for health check output
   ```

4. **USB Gadget State:**
   ```bash
   ls -laR /sys/kernel/config/usb_gadget/xkvm/ > gadget-state.txt
   ```

5. **Steps to Reproduce:**
   - What were you doing when the issue occurred?
   - Can you reproduce it consistently?
   - What error messages did you see?

## Emergency Recovery

If USB gadget is completely broken:

```bash
# 1. Stop service
systemctl stop xkvm

# 2. Clean up everything
echo "" > /sys/kernel/config/usb_gadget/xkvm/UDC 2>/dev/null
rm -rf /sys/kernel/config/usb_gadget/xkvm 2>/dev/null
rm -f /var/lock/xkvm-usb.lock

# 3. Restart service
systemctl start xkvm

# 4. Check logs
journalctl -u xkvm -f
```

If that doesn't work:
```bash
# Reboot the system
reboot
```
