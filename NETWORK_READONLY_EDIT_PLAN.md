# Network Read-Only Mode - Edit Plan

## Problem Scope

The `pkg/nmlite` package currently performs network interface configuration writes to the OS. To make network information read-only, we need to remove all operations that modify network interface state, addresses, routes, and system configuration.

## Files Requiring Changes

### 1. **pkg/nmlite/interface.go** - MAJOR CHANGES
**Current Issues:**
- `applyIPv4Config()` - Applies IPv4 configuration (static/DHCP)
- `applyIPv6Config()` - Applies IPv6 configuration (static/DHCP/SLAAC)
- `applyIPv4Static()` - Configures static IPv4 addresses
- `applyIPv4DHCP()` - Starts DHCP client
- `disableIPv4()` - Removes IPv4 addresses
- `applyIPv6Static()` - Configures static IPv6 addresses
- `applyIPv6DHCP()` - Starts DHCPv6 client
- `applyIPv6SLAAC()` - Enables SLAAC and sends router solicitation
- `applyIPv6SLAACAndDHCP()` - Enables both SLAAC and DHCPv6
- `applyIPv6LinkLocal()` - Configures link-local only
- `disableIPv6()` - Disables IPv6
- `handleLinkUp()` - Applies configuration when link comes up
- `handleLinkDown()` - Removes addresses when link goes down
- `applyDHCPLease()` - Applies DHCP lease to interface
- `ReconcileLinkAddrs()` - Reconciles addresses on interface
- `SendRouterSolicitation()` - Sends IPv6 router solicitation

**Actions:**
- Remove all `apply*()` methods (lines ~400-650)
- Remove all `disable*()` methods
- Simplify `handleLinkUp()` - remove all apply/renew calls, keep only logging
- Simplify `handleLinkDown()` - remove address removal, keep only logging
- Remove `applyDHCPLease()` method (lines ~759-781)
- Remove `ReconcileLinkAddrs()` method (lines ~747-758)
- Remove `SendRouterSolicitation()` method (lines ~620-648)
- Remove `convertDHCPLeaseToIPv4Config()` helper (lines ~783-803)
- In `NewInterfaceManager()`:
  - Remove `staticConfig` initialization
  - Remove `dhcpClient` initialization
  - Remove DHCP callback setup
- In `Start()`:
  - Remove commented `EnsureInterfaceUpWithTimeout()` call
  - Remove commented `applyConfiguration()` call
  - Keep only state monitoring
- In `Stop()`:
  - Remove DHCP client stop call
- Remove `RenewDHCPLease()` method (lines ~350-357)
- Remove `SetOnResolvConfChange()` callback setter

### 2. **pkg/nmlite/manager.go** - MODERATE CHANGES
**Current Issues:**
- `SetInterfaceConfig()` - Already commented out (good!)
- `RenewDHCPLease()` - Renews DHCP lease
- `CleanUpLegacyDHCPClients()` - Kills legacy DHCP clients

**Actions:**
- Remove `RenewDHCPLease()` method entirely
- Remove `CleanUpLegacyDHCPClients()` method entirely
- Remove `shouldKillLegacyDHCPClients()` helper method

### 3. **pkg/nmlite/static.go** - REMOVE ENTIRE FILE
**Current Issues:**
- `ToIPv4Static()` - Converts config to static IPv4 (used for applying)
- `ToIPv6Static()` - Converts config to static IPv6 (used for applying)
- `DisableIPv4()` - Removes IPv4 addresses
- `DisableIPv6()` - Disables IPv6 via sysctl
- `EnableIPv6SLAAC()` - Enables SLAAC via sysctl
- `EnableIPv6LinkLocal()` - Enables link-local via sysctl
- `removeIPv4DefaultRoute()` - Removes default route

**Actions:**
- **DELETE THIS ENTIRE FILE** - All methods modify network configuration

### 4. **pkg/nmlite/dhcp.go** - REMOVE ENTIRE FILE
**Current Issues:**
- `NewDHCPClient()` - Creates DHCP client
- `Start()` - Starts DHCP client (modifies network)
- `Stop()` - Stops DHCP client
- `Renew()` - Renews DHCP lease
- `Release()` - Releases DHCP lease
- `SetIPv4()` / `SetIPv6()` - Enable/disable DHCP

**Actions:**
- **DELETE THIS ENTIRE FILE** - DHCP client actively modifies network configuration

### 5. **pkg/nmlite/link/manager.go** - MAJOR CHANGES
**Current Issues:**
- `LinkSetUp()` - Brings interface up
- `LinkSetDown()` - Brings interface down
- `EnsureInterfaceUp()` - Ensures interface is up
- `EnsureInterfaceUpWithTimeout()` - Brings interface up with retry
- `AddrAdd()` - Adds IP address to interface
- `AddrDel()` - Removes IP address from interface
- `RemoveAllAddresses()` - Removes all addresses
- `RemoveNonLinkLocalIPv6Addresses()` - Removes non-link-local IPv6 addresses
- `RouteAdd()` - Adds route
- `RouteDel()` - Removes route
- `RouteReplace()` - Replaces route
- `AddDefaultRoute()` - Adds default route
- `RemoveDefaultRoute()` - Removes default route
- `reconcileDefaultRoute()` - Reconciles default routes
- `ReconcileLink()` - Reconciles addresses and routes on interface

**Actions:**
- Remove `LinkSetUp()` method
- Remove `LinkSetDown()` method
- Remove `EnsureInterfaceUp()` method
- Remove `EnsureInterfaceUpWithTimeout()` method
- Remove `AddrAdd()` method
- Remove `AddrDel()` method
- Remove `RemoveAllAddresses()` method
- Remove `RemoveNonLinkLocalIPv6Addresses()` method
- Remove `RouteAdd()` method
- Remove `RouteDel()` method
- Remove `RouteReplace()` method
- Remove `AddDefaultRoute()` method
- Remove `RemoveDefaultRoute()` method
- Remove `reconcileDefaultRoute()` method
- Remove `ReconcileLink()` method
- Keep only read operations: `GetLinkByName()`, `AddrList()`, `RouteList()`, `ListDefaultRoutes()`, `HasDefaultRoute()`

### 6. **pkg/nmlite/link/sysctl.go** - REMOVE ENTIRE FILE
**Current Issues:**
- `setSysctlValues()` - Writes to /proc/sys
- `EnableIPv6()` - Enables IPv6 via sysctl
- `DisableIPv6()` - Disables IPv6 via sysctl
- `EnableIPv6SLAAC()` - Enables SLAAC via sysctl
- `EnableIPv6LinkLocal()` - Enables link-local via sysctl

**Actions:**
- **DELETE THIS ENTIRE FILE** - All methods write to sysctl

### 7. **pkg/nmlite/link/netlink.go** - MINOR CHANGES
**Current Issues:**
- `Link.SetMTU()` - Sets MTU on interface

**Actions:**
- Remove `SetMTU()` method from Link struct

### 8. **pkg/nmlite/resolvconf.go** - ALREADY COMMENTED OUT
**Status:** All code is already commented out ✓

### 9. **pkg/nmlite/hostname.go** - ALREADY COMMENTED OUT
**Status:** All code is already commented out ✓

### 10. **pkg/nmlite/interface_state.go** - NO CHANGES NEEDED
**Status:** Only reads interface state ✓

### 11. **pkg/nmlite/state.go** - NO CHANGES NEEDED
**Status:** Only reads state information ✓

### 12. **pkg/nmlite/utils.go** - NO CHANGES NEEDED
**Status:** Only utility functions for comparison ✓

### 13. **pkg/nmlite/dhcp_lease_reader.go** - NO CHANGES NEEDED
**Status:** Only reads DHCP lease files (read-only) ✓

## Summary of Changes

### Architecture Change

**BEFORE (Current - Read/Write):**
```
NetworkManager
├── InterfaceManager
│   ├── StaticConfigManager (WRITES to netlink)
│   ├── DHCPClient (STARTS/STOPS dhcp, WRITES addresses)
│   ├── applyIPv4Config() → WRITES addresses
│   ├── applyIPv6Config() → WRITES addresses
│   ├── ReconcileLinkAddrs() → WRITES addresses
│   └── handleLinkUp/Down() → WRITES addresses
└── NetlinkManager
    ├── AddrAdd/AddrDel (WRITES addresses)
    ├── RouteAdd/RouteDel (WRITES routes)
    ├── LinkSetUp/Down (CHANGES interface state)
    ├── ReconcileLink() (WRITES addresses/routes)
    └── sysctl operations (WRITES kernel params)
```

**AFTER (Target - Read-Only):**
```
NetworkManager
├── InterfaceManager
│   ├── GetState() → reads state
│   ├── GetConfig() → reads config
│   ├── updateInterfaceState() → reads from netlink
│   └── monitorInterfaceState() → monitors changes
└── NetlinkManager
    ├── GetLinkByName() (reads link info)
    ├── AddrList() (reads addresses)
    ├── RouteList() (reads routes)
    ├── ListDefaultRoutes() (reads routes)
    └── HasDefaultRoute() (checks routes)
```

### Files to DELETE:
1. `pkg/nmlite/static.go` - All methods modify network config
2. `pkg/nmlite/dhcp.go` - DHCP client modifies network config
3. `pkg/nmlite/link/sysctl.go` - All methods write to sysctl

### Files to MODIFY:
1. `pkg/nmlite/interface.go` - Remove all apply/disable/reconcile methods
2. `pkg/nmlite/manager.go` - Remove DHCP renewal and cleanup methods
3. `pkg/nmlite/link/manager.go` - Remove all write operations, keep only reads
4. `pkg/nmlite/link/netlink.go` - Remove SetMTU method

### Files UNCHANGED (already read-only):
1. `pkg/nmlite/interface_state.go` ✓
2. `pkg/nmlite/state.go` ✓
3. `pkg/nmlite/utils.go` ✓
4. `pkg/nmlite/dhcp_lease_reader.go` ✓
5. `pkg/nmlite/resolvconf.go` ✓ (already commented out)
6. `pkg/nmlite/hostname.go` ✓ (already commented out)
7. `pkg/nmlite/link/utils.go` ✓ (parsing utilities only)
8. `pkg/nmlite/link/types.go` ✓ (type definitions only)
9. `pkg/nmlite/link/consts.go` ✓ (constants only)

## Expected Behavior After Changes

### What WILL Work (Read Operations):
- ✓ Read current network interface state (up/down/online)
- ✓ Read IP addresses (IPv4 and IPv6)
- ✓ Read MAC addresses
- ✓ Read routes and gateways
- ✓ Read DHCP lease information from system files (udhcpc, dhclient, jetdhcpc)
- ✓ Monitor interface state changes via netlink
- ✓ Detect when interfaces go up/down
- ✓ Detect when IP addresses change
- ✓ Read MTU, link speed, and other interface attributes
- ✓ Read DNS servers from DHCP leases
- ✓ Read NTP servers from DHCP leases
- ✓ Display network configuration in UI

### What WILL NOT Work (Write Operations):
- ✗ Modify IP addresses (static or DHCP)
- ✗ Modify routes or gateways
- ✗ Start/stop DHCP clients
- ✗ Renew DHCP leases
- ✗ Modify sysctl settings (IPv6 enable/disable, SLAAC, etc.)
- ✗ Bring interfaces up/down
- ✗ Modify MTU
- ✗ Send router solicitations
- ✗ Apply network configuration changes
- ✗ Remove or add addresses

### RPC Methods Impact:
- `rpcRenewDHCPLease()` - Already returns error (no change needed) ✓
- `rpcSetNetworkSettings()` - Already returns error (no change needed) ✓
- `rpcToggleDHCPClient()` - Already returns error (no change needed) ✓
- `rpcGetNetworkSettings()` - Will continue to work (read-only) ✓
- `rpcGetInterfaceState()` - Will continue to work (read-only) ✓

## Compilation Impact

After removing these methods, the following will need to be addressed:

### Direct Callers Found:
1. **network.go** - `rpcRenewDHCPLease()` - Already returns error (read-only mode) ✓
2. **jsonrpc.go** - RPC method registration for `renewDHCPLease` - Can stay (already returns error)
3. **pkg/nmlite/interface.go** - Internal calls to removed methods - MUST FIX
4. **pkg/nmlite/manager.go** - Internal calls to removed methods - MUST FIX

### Struct Fields to Remove:
From `InterfaceManager`:
- `staticConfig *StaticConfigManager` - No longer needed
- `dhcpClient *DHCPClient` - No longer needed
- `onResolvConfChange ResolvConfChangeCallback` - No longer needed

### Methods to Keep (Read-Only):
- All `Get*()` methods
- All `Is*()` state check methods
- `GetState()`, `GetConfig()`
- `updateInterfaceState()` - reads state only
- `monitorInterfaceState()` - monitors state only

## Next Steps

1. Review this plan
2. Get approval
3. Execute changes in order:
   - Delete files first
   - Modify link/manager.go
   - Modify interface.go
   - Modify manager.go
   - Fix compilation errors
   - Update tests
