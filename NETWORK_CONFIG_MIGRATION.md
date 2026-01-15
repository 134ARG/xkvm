# Network Configuration Migration

## Overview

After porting to a full Linux OS, network configuration is now **read-only** and managed by the operating system. The `network_config` section in the configuration file is no longer used for network management.

## What Changed

### Before (Embedded System)
- Network configuration was stored in `/userdata/kvm_config.json`
- Settings like `ipv4_mode`, `ipv6_mode`, `dhcp_client`, `hostname`, etc. were managed by the application
- Users could modify network settings through the web UI

### After (Full Linux OS)
- Network configuration is managed by the OS (NetworkManager, systemd-networkd, etc.)
- The web UI displays current network state in **read-only mode**
- Network settings are read from the actual Linux network interfaces
- The `network_config` section is **ignored** during config load and **excluded** during config save

## Migration Strategy

The migration is designed to be **non-breaking** and **automatic**:

1. **Config Loading**: When loading config, any persisted `network_config` is ignored and replaced with defaults
2. **Config Saving**: When saving config, `network_config` is excluded from the JSON output
3. **In-Memory**: NetworkConfig still exists in memory for backward compatibility with existing code
4. **RPC Methods**: Network RPC methods return errors explaining the read-only nature

## Automatic Cleanup

The system automatically handles the migration:

- **On Load**: Ignores persisted `network_config` and uses defaults
- **On Save**: Excludes `network_config` from being written to disk
- **Gradual Cleanup**: Old `network_config` sections will be removed on the next config save

## Manual Cleanup (Optional)

If you want to immediately clean up existing config files, run:

```bash
sudo ./scripts/migrate_network_config.sh
```

This script will:
1. Create a backup of your current config
2. Remove the `network_config` section
3. Preserve all other settings

## Network Management

To configure network settings on the full Linux system, use:

### NetworkManager (most common)
```bash
# Command line
nmcli connection modify <connection-name> ipv4.addresses 192.168.1.100/24
nmcli connection modify <connection-name> ipv4.gateway 192.168.1.1
nmcli connection up <connection-name>

# Text UI
nmtui
```

### systemd-networkd
Edit files in `/etc/systemd/network/`

### Manual Configuration
Edit `/etc/network/interfaces` or `/etc/netplan/` depending on your distribution

### Hostname
```bash
hostnamectl set-hostname <new-hostname>
```

## Web UI Behavior

The network settings page now:
- Displays current network state (read-only)
- Shows IPv4/IPv6 addresses, DHCP leases, DNS servers
- Displays a notice that configuration must be done through OS tools
- Prevents modification attempts with helpful error messages

## Backward Compatibility

The following are preserved for backward compatibility:

1. **NetworkConfig struct**: Still exists in memory with default values
2. **RPC Methods**: `getNetworkSettings()` returns defaults, `setNetworkSettings()` returns an error
3. **Config Field**: `network_config` field exists but is marked as deprecated with `omitempty`

## Technical Details

### Code Changes

**config.go**:
- `LoadConfig()`: Always uses default NetworkConfig, ignores persisted values
- `saveConfig()`: Creates a copy of config with `NetworkConfig = nil` before saving
- `Config.NetworkConfig`: Marked as deprecated with `omitempty` JSON tag

**network.go**:
- `rpcSetNetworkSettings()`: Returns error explaining read-only mode
- `rpcRenewDHCPLease()`: Returns error explaining read-only mode
- `setHostname()`: Returns error explaining read-only mode
- `initNetwork()`: Creates minimal network manager for status reading only

**UI**:
- Network settings page displays read-only information
- Shows helpful notice about using OS network tools
- Prevents modification attempts

## Benefits

1. **No Breaking Changes**: Existing code continues to work
2. **Automatic Migration**: No manual intervention required
3. **Clean Separation**: Network management is properly delegated to the OS
4. **User Clarity**: UI clearly indicates read-only status
5. **Gradual Cleanup**: Old config sections are removed naturally over time

## Troubleshooting

### Config still has network_config after upgrade
This is normal. It will be removed on the next config save. To remove it immediately, run the migration script.

### Network settings not updating in UI
The UI reads from actual network interfaces. If changes aren't reflected:
1. Check your OS network configuration
2. Restart the network service
3. Refresh the web UI

### Cannot change network settings
This is expected. Use your OS network management tools as described above.

## Related Files

- `config.go` - Config loading/saving with network_config exclusion
- `network.go` - Network state reading and deprecated RPC methods
- `ui/src/routes/devices.$id.settings.network.tsx` - Read-only network UI
- `scripts/migrate_network_config.sh` - Manual migration script
- `internal/network/types/config.go` - NetworkConfig struct definition
