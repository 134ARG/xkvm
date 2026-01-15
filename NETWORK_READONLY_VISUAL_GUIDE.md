# Network Read-Only Mode - Visual Guide

## File Structure Changes

```
pkg/nmlite/
├── manager.go              [MODIFY] Remove 3 methods (~50 lines)
├── interface.go            [MODIFY] Remove 20 methods (~450 lines)
├── interface_state.go      [KEEP] Read-only ✓
├── state.go                [KEEP] Read-only ✓
├── utils.go                [KEEP] Utilities ✓
├── netlink.go              [KEEP] Wrapper ✓
├── dhcp_lease_reader.go    [KEEP] Read-only ✓
├── resolvconf.go           [KEEP] Already commented out ✓
├── hostname.go             [KEEP] Already commented out ✓
├── static.go               [DELETE] ✗
├── dhcp.go                 [DELETE] ✗
├── link/
│   ├── manager.go          [MODIFY] Remove 15 methods (~380 lines)
│   ├── netlink.go          [MODIFY] Remove SetMTU method
│   ├── types.go            [KEEP] Type definitions ✓
│   ├── consts.go           [KEEP] Constants ✓
│   ├── utils.go            [KEEP] Parsing utilities ✓
│   └── sysctl.go           [DELETE] ✗
├── jetdhcpc/               [KEEP] Used for reading leases ✓
├── udhcpc/                 [KEEP] Used for reading leases ✓
└── dhcplease/              [KEEP] Used for reading leases ✓
```

## Method Removal Map

### pkg/nmlite/interface.go

```
InterfaceManager
├── [KEEP] NewInterfaceManager()      - Simplified
├── [KEEP] Start()                    - Simplified
├── [KEEP] Stop()                     - Simplified
├── [KEEP] GetState()                 ✓
├── [KEEP] GetConfig()                ✓
├── [KEEP] IsUp()                     ✓
├── [KEEP] IsOnline()                 ✓
├── [KEEP] IPv4Ready()                ✓
├── [KEEP] IPv6Ready()                ✓
├── [KEEP] GetIPv4Address()           ✓
├── [KEEP] GetIPv6Address()           ✓
├── [KEEP] GetMACAddress()            ✓
├── [KEEP] SetOnStateChange()         ✓
├── [KEEP] SetOnConfigChange()        ✓
├── [KEEP] SetOnDHCPLeaseChange()     ✓
├── [KEEP] updateInterfaceState()     ✓
├── [KEEP] monitorInterfaceState()    ✓
├── [KEEP] handleLinkStateChange()    - Simplified
├── [DELETE] RenewDHCPLease()         ✗
├── [DELETE] SetOnResolvConfChange()  ✗
├── [DELETE] applyIPv4Config()        ✗
├── [DELETE] applyIPv6Config()        ✗
├── [DELETE] applyIPv4Static()        ✗
├── [DELETE] applyIPv4DHCP()          ✗
├── [DELETE] disableIPv4()            ✗
├── [DELETE] applyIPv6Static()        ✗
├── [DELETE] applyIPv6DHCP()          ✗
├── [DELETE] applyIPv6SLAAC()         ✗
├── [DELETE] applyIPv6SLAACAndDHCP()  ✗
├── [DELETE] applyIPv6LinkLocal()     ✗
├── [DELETE] disableIPv6()            ✗
├── [DELETE] SendRouterSolicitation() ✗
├── [DELETE] handleLinkUp()           ✗ (replaced with simple version)
├── [DELETE] handleLinkDown()         ✗ (replaced with simple version)
├── [DELETE] ReconcileLinkAddrs()     ✗
├── [DELETE] applyDHCPLease()         ✗
└── [DELETE] convertDHCPLeaseToIPv4Config() ✗
```

### pkg/nmlite/manager.go

```
NetworkManager
├── [KEEP] NewNetworkManager()        ✓
├── [KEEP] AddInterface()             ✓
├── [KEEP] RemoveInterface()          ✓
├── [KEEP] GetInterface()             ✓
├── [KEEP] ListInterfaces()           ✓
├── [KEEP] GetInterfaceState()        ✓
├── [KEEP] GetInterfaceConfig()       ✓
├── [KEEP] SetOnInterfaceStateChange() ✓
├── [KEEP] SetOnConfigChange()        ✓
├── [KEEP] SetOnDHCPLeaseChange()     ✓
├── [KEEP] Stop()                     ✓
├── [KEEP] IsOnline()                 ✓
├── [KEEP] IsUp()                     ✓
├── [KEEP] Hostname()                 ✓
├── [KEEP] Domain()                   ✓
├── [KEEP] NTPServers()               ✓
├── [DELETE] RenewDHCPLease()         ✗
├── [DELETE] CleanUpLegacyDHCPClients() ✗
└── [DELETE] shouldKillLegacyDHCPClients() ✗
```

### pkg/nmlite/link/manager.go

```
NetlinkManager
├── [KEEP] GetLinkByName()            ✓
├── [KEEP] AddrList()                 ✓
├── [KEEP] RouteList()                ✓
├── [KEEP] ListDefaultRoutes()        ✓
├── [KEEP] HasDefaultRoute()          ✓
├── [KEEP] monitorStateChange()       ✓
├── [KEEP] runCallbacks()             ✓
├── [KEEP] AddStateChangeCallback()   ✓
├── [DELETE] LinkSetUp()              ✗
├── [DELETE] LinkSetDown()            ✗
├── [DELETE] EnsureInterfaceUp()      ✗
├── [DELETE] EnsureInterfaceUpWithTimeout() ✗
├── [DELETE] AddrAdd()                ✗
├── [DELETE] AddrDel()                ✗
├── [DELETE] RemoveAllAddresses()     ✗
├── [DELETE] RemoveNonLinkLocalIPv6Addresses() ✗
├── [DELETE] RouteAdd()               ✗
├── [DELETE] RouteDel()               ✗
├── [DELETE] RouteReplace()           ✗
├── [DELETE] AddDefaultRoute()        ✗
├── [DELETE] RemoveDefaultRoute()     ✗
├── [DELETE] reconcileDefaultRoute()  ✗
└── [DELETE] ReconcileLink()          ✗
```

## Data Flow Changes

### BEFORE (Read-Write Mode)

```
┌─────────────────────────────────────────────────────────────┐
│                         User/RPC                            │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    NetworkManager                           │
│  • SetInterfaceConfig() ──────────────────┐                │
│  • RenewDHCPLease() ──────────────────┐   │                │
└───────────────────────────────────────┼───┼────────────────┘
                                        │   │
                                        ▼   ▼
┌─────────────────────────────────────────────────────────────┐
│                  InterfaceManager                           │
│  • applyIPv4Config() ────────────┐                         │
│  • applyIPv6Config() ────────┐   │                         │
│  • ReconcileLinkAddrs() ──┐  │   │                         │
└───────────────────────────┼──┼───┼─────────────────────────┘
                            │  │   │
                            ▼  ▼   ▼
┌─────────────────────────────────────────────────────────────┐
│              StaticConfigManager / DHCPClient               │
│  • ToIPv4Static() ──────────┐                              │
│  • Start/Stop DHCP ──────┐  │                              │
└──────────────────────────┼──┼──────────────────────────────┘
                           │  │
                           ▼  ▼
┌─────────────────────────────────────────────────────────────┐
│                    NetlinkManager                           │
│  • AddrAdd/AddrDel ──────────────┐                         │
│  • RouteAdd/RouteDel ────────┐   │                         │
│  • LinkSetUp/Down ───────┐   │   │                         │
│  • ReconcileLink() ───┐  │   │   │                         │
└───────────────────────┼──┼───┼───┼─────────────────────────┘
                        │  │   │   │
                        ▼  ▼   ▼   ▼
┌─────────────────────────────────────────────────────────────┐
│                    Linux Kernel                             │
│  • netlink: modify addresses, routes, link state           │
│  • sysctl: modify IPv6 settings                            │
│  • DHCP: start/stop clients                                │
└─────────────────────────────────────────────────────────────┘
```

### AFTER (Read-Only Mode)

```
┌─────────────────────────────────────────────────────────────┐
│                         User/RPC                            │
│  • SetInterfaceConfig() → ERROR ✗                          │
│  • RenewDHCPLease() → ERROR ✗                              │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    NetworkManager                           │
│  • GetInterfaceState() ✓                                   │
│  • GetInterfaceConfig() ✓                                  │
│  • Monitor state changes ✓                                 │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  InterfaceManager                           │
│  • GetState() ✓                                            │
│  • updateInterfaceState() ✓                                │
│  • monitorInterfaceState() ✓                               │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    NetlinkManager                           │
│  • GetLinkByName() ✓                                       │
│  • AddrList() ✓                                            │
│  • RouteList() ✓                                           │
│  • Monitor link changes ✓                                  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Linux Kernel                             │
│  • netlink: READ addresses, routes, link state             │
│  • /var/lib/dhcp: READ DHCP leases                         │
│  • /proc/net: READ network info                            │
└─────────────────────────────────────────────────────────────┘
```

## Size Comparison

```
BEFORE:
┌────────────────────────────────────────────────────────┐
│ interface.go          ████████████████████  800 lines │
│ manager.go            ████████  240 lines             │
│ link/manager.go       ██████████████████  540 lines   │
│ static.go             █████  150 lines                │
│ dhcp.go               ██████  180 lines               │
│ link/sysctl.go        ██  50 lines                    │
│                                                        │
│ TOTAL: 1960 lines                                     │
└────────────────────────────────────────────────────────┘

AFTER:
┌────────────────────────────────────────────────────────┐
│ interface.go          ███████████  350 lines          │
│ manager.go            ██████  190 lines               │
│ link/manager.go       █████  160 lines                │
│ static.go             [DELETED]                        │
│ dhcp.go               [DELETED]                        │
│ link/sysctl.go        [DELETED]                        │
│                                                        │
│ TOTAL: 700 lines (-64%)                               │
└────────────────────────────────────────────────────────┘
```

## Operation Categories

### ✓ KEEP (Read Operations)
```
┌─────────────────────────────────────────┐
│ • Get interface state                   │
│ • Get IP addresses                      │
│ • Get MAC address                       │
│ • Get routes                            │
│ • Get DHCP leases (from files)          │
│ • Monitor state changes                 │
│ • Check if interface is up/online       │
│ • List interfaces                       │
│ • Get configuration                     │
└─────────────────────────────────────────┘
```

### ✗ DELETE (Write Operations)
```
┌─────────────────────────────────────────┐
│ • Add/remove IP addresses               │
│ • Add/remove routes                     │
│ • Bring interface up/down               │
│ • Start/stop DHCP clients               │
│ • Renew DHCP leases                     │
│ • Modify sysctl parameters              │
│ • Set MTU                               │
│ • Send router solicitations             │
│ • Apply network configuration           │
│ • Reconcile addresses/routes            │
└─────────────────────────────────────────┘
```

## Quick Reference

| Operation | Before | After |
|-----------|--------|-------|
| Read interface state | ✓ | ✓ |
| Read IP addresses | ✓ | ✓ |
| Read routes | ✓ | ✓ |
| Read DHCP leases | ✓ | ✓ |
| Monitor changes | ✓ | ✓ |
| Set IP addresses | ✓ | ✗ |
| Set routes | ✓ | ✗ |
| Start DHCP | ✓ | ✗ |
| Renew DHCP | ✓ | ✗ |
| Modify sysctl | ✓ | ✗ |
| Set MTU | ✓ | ✗ |

---

**Legend:**
- ✓ = Supported / Keep
- ✗ = Not supported / Delete
- [MODIFY] = File needs changes
- [DELETE] = File will be deleted
- [KEEP] = File unchanged
