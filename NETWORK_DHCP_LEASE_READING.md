# DHCP Lease Information Reading

## Overview

This document describes the implementation of system DHCP lease reading for full Linux systems where network management is handled by the OS (NetworkManager, systemd-networkd, etc.) rather than by the application's own DHCP client.

## Problem

On full Linux systems:
- IPv6 addresses are read directly from network interfaces via netlink (works correctly)
- IPv4 DHCP lease information was only populated when the application's own DHCP client obtained a lease
- When NetworkManager or systemd-networkd manages DHCP, the application's DHCP client never runs
- Result: IPv6 info displayed correctly, but IPv4 DHCP lease details were missing

## Solution

Added `DHCPLeaseReader` in `pkg/nmlite/dhcp_lease_reader.go` that reads DHCP lease information from system DHCP clients in read-only mode.

### Supported DHCP Clients

The reader attempts to read lease information from the following DHCP clients in order:

1. **NetworkManager** (`/var/lib/NetworkManager/dhclient-*.lease` or `internal-*.lease`)
2. **dhclient** (`/var/lib/dhcp/dhclient.*.leases` or `/var/lib/dhclient/dhclient-*.leases`)
3. **systemd-networkd** (`/run/systemd/netif/leases/*`)
4. **udhcpc** (`/run/udhcpc.*.info`)

### Integration

The lease reader is integrated into `pkg/nmlite/interface_state.go`:
- Called during interface state updates
- Only reads when `ipv4_mode` is "dhcp" and no lease is currently available
- Populates `DHCPLease4` in the interface state
- Logs which DHCP client the lease was read from

### UI Changes

Updated `ui/src/components/DhcpLeaseCard.tsx`:
- Shows basic IPv4 address info when detailed DHCP lease is unavailable
- Displays full DHCP lease details when available (from system DHCP client)
- Shows helpful message explaining that detailed info requires OS tools when lease data is missing

## Lease Information Provided

When available, the following DHCP lease information is displayed:
- IP Address
- Subnet Mask
- Gateway/Routers
- DNS Servers
- Domain Name
- DHCP Server ID
- Broadcast Address
- Lease Expiry Time (when available)
- NTP Servers (when available)
- Other DHCP options depending on the client

## Read-Only Mode

All network operations remain read-only:
- No network configuration changes
- No DHCP lease renewal
- No DHCP client switching
- Network management is handled by the OS

Users should use OS network management tools:
- `nmcli` / `nmtui` (NetworkManager)
- `networkctl` (systemd-networkd)
- `dhclient`
- Distribution-specific network configuration tools

## Testing

To test the DHCP lease reading:
1. Ensure the system is using DHCP for IPv4
2. Check that the system's DHCP client has obtained a lease
3. View the network settings page in the UI
4. Verify that IPv4 DHCP lease information is displayed

## Future Improvements

Potential enhancements:
- Watch lease files for changes and update in real-time
- Support additional DHCP clients
- Parse more DHCP options
- Display lease renewal/rebinding times
