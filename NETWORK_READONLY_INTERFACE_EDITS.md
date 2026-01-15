# Detailed Edit Guide for pkg/nmlite/interface.go

## Struct Changes

### InterfaceManager struct (lines ~20-40)
**REMOVE these fields:**
```go
staticConfig *StaticConfigManager
dhcpClient   *DHCPClient
onResolvConfChange ResolvConfChangeCallback
```

**KEEP these fields:**
```go
ctx       context.Context
ifaceName string
config    *types.NetworkConfig
logger    *zerolog.Logger
state     *types.InterfaceState
linkState *link.Link
stateMu   sync.RWMutex
onStateChange      func(state types.InterfaceState)
onConfigChange     func(config *types.NetworkConfig)
onDHCPLeaseChange  func(lease *types.DHCPLease)
stopCh chan struct{}
wg     sync.WaitGroup
```

## Method Changes

### NewInterfaceManager() - Lines ~43-110
**REMOVE:**
- Lines ~75-79: staticConfig initialization
```go
im.staticConfig, err = NewStaticConfigManager(ifaceName, &scopedLogger)
if err != nil {
    return nil, fmt.Errorf("failed to create static config manager: %w", err)
}
```

- Lines ~81-85: dhcpClient initialization
```go
im.dhcpClient, err = NewDHCPClient(ctx, ifaceName, &scopedLogger, config.DHCPClient.String)
if err != nil {
    return nil, fmt.Errorf("failed to create DHCP client: %w", err)
}
```

- Lines ~87-107: DHCP callback setup (entire block)
```go
im.dhcpClient.SetOnLeaseChange(func(lease *types.DHCPLease) {
    // ... entire callback
})
```

### Start() - Lines ~113-170
**REMOVE:**
- Lines ~143-153: Commented out interface up and config application
```go
// _, linkUpErr := nl.EnsureInterfaceUpWithTimeout(...)
// if linkUpErr != nil {
//     ...
// } else {
//     if err := im.applyConfiguration(); err != nil {
//         ...
//     }
// }
```

**KEEP:**
- Interface monitoring setup
- Link state retrieval
- State change callback registration

### Stop() - Lines ~172-188
**REMOVE:**
- Lines ~179-185: DHCP client stop
```go
if im.dhcpClient != nil {
    if err := im.dhcpClient.Stop(); err != nil {
        return fmt.Errorf("failed to stop DHCP client: %w", err)
    }
}
```

### RenewDHCPLease() - Lines ~350-357
**DELETE ENTIRE METHOD**

### SetOnResolvConfChange() - Lines ~370-372
**DELETE ENTIRE METHOD**

### applyIPv4Config() - Lines ~374-390
**DELETE ENTIRE METHOD**

### applyIPv6Config() - Lines ~392-408
**DELETE ENTIRE METHOD**

### applyIPv4Static() - Lines ~410-449
**DELETE ENTIRE METHOD**

### applyIPv4DHCP() - Lines ~451-459
**DELETE ENTIRE METHOD**

### disableIPv4() - Lines ~461-471
**DELETE ENTIRE METHOD**

### applyIPv6Static() - Lines ~473-500
**DELETE ENTIRE METHOD**

### applyIPv6DHCP() - Lines ~502-510
**DELETE ENTIRE METHOD**

### applyIPv6SLAAC() - Lines ~512-542
**DELETE ENTIRE METHOD**

### applyIPv6SLAACAndDHCP() - Lines ~544-556
**DELETE ENTIRE METHOD**

### applyIPv6LinkLocal() - Lines ~558-566
**DELETE ENTIRE METHOD**

### disableIPv6() - Lines ~568-576
**DELETE ENTIRE METHOD**

### handleLinkStateChange() - Lines ~578-596
**KEEP but simplify** - Remove apply calls, keep only logging

### SendRouterSolicitation() - Lines ~598-648
**DELETE ENTIRE METHOD**

### handleLinkUp() - Lines ~650-670
**SIMPLIFY to:**
```go
func (im *InterfaceManager) handleLinkUp() {
    im.logger.Info().Msg("link up")
    // Read-only mode: just log the event, don't modify network config
}
```

### handleLinkDown() - Lines ~672-688
**SIMPLIFY to:**
```go
func (im *InterfaceManager) handleLinkDown() {
    im.logger.Info().Msg("link down")
    // Read-only mode: just log the event, don't modify network config
}
```

### monitorInterfaceState() - Lines ~690-710
**KEEP UNCHANGED** - Only reads state

### updateStateFromDHCPLease() - Lines ~712-743
**KEEP but REMOVE the onResolvConfChange call:**
- Lines ~715-743: Keep DHCP lease state update
- Remove lines ~726-742: resolv.conf update callback

**SIMPLIFIED VERSION:**
```go
func (im *InterfaceManager) updateStateFromDHCPLease(lease *types.DHCPLease) {
    family := link.AfInet

    im.stateMu.Lock()
    if lease.IsIPv6() {
        im.state.DHCPLease6 = lease
        family = link.AfInet6
    } else {
        im.state.DHCPLease4 = lease
    }
    im.stateMu.Unlock()

    // Read-only mode: don't update resolv.conf
    im.logger.Debug().
        Int("family", family).
        Str("ip", lease.IPAddress.String()).
        Msg("DHCP lease updated in state (read-only mode)")
}
```

### ReconcileLinkAddrs() - Lines ~745-758
**DELETE ENTIRE METHOD**

### applyDHCPLease() - Lines ~759-781
**DELETE ENTIRE METHOD**

### convertDHCPLeaseToIPv4Config() - Lines ~783-803
**DELETE ENTIRE METHOD**

## Summary of Line Deletions

**Total lines to delete: ~450 lines**
- Struct fields: 3 fields
- Complete methods deleted: 20 methods
- Simplified methods: 3 methods
- Initialization code removed: ~50 lines

**Remaining functionality:**
- State reading and monitoring
- Configuration retrieval
- Interface state queries (IsUp, IsOnline, IPv4Ready, etc.)
- Callback management for state changes
