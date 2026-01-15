# Detailed Edit Guide for pkg/nmlite/link/manager.go

## Methods to DELETE (Write Operations)

### 1. LinkSetUp() - Lines ~115-119
**DELETE ENTIRE METHOD**
```go
func (nm *NetlinkManager) LinkSetUp(link *Link) error {
    nm.mu.RLock()
    defer nm.mu.RUnlock()
    return netlink.LinkSetUp(link)
}
```

### 2. LinkSetDown() - Lines ~121-125
**DELETE ENTIRE METHOD**
```go
func (nm *NetlinkManager) LinkSetDown(link *Link) error {
    nm.mu.RLock()
    defer nm.mu.RUnlock()
    return netlink.LinkSetDown(link)
}
```

### 3. EnsureInterfaceUp() - Lines ~127-132
**DELETE ENTIRE METHOD**
```go
func (nm *NetlinkManager) EnsureInterfaceUp(link *Link) error {
    if link.Attrs().OperState == netlink.OperUp {
        return nil
    }
    return nm.LinkSetUp(link)
}
```

### 4. EnsureInterfaceUpWithTimeout() - Lines ~134-195
**DELETE ENTIRE METHOD** (62 lines)
This method brings interfaces up with retry logic - not needed in read-only mode.

### 5. AddrAdd() - Lines ~201-205
**DELETE ENTIRE METHOD**
```go
func (nm *NetlinkManager) AddrAdd(link *Link, addr *netlink.Addr) error {
    nm.mu.RLock()
    defer nm.mu.RUnlock()
    return netlink.AddrAdd(link, addr)
}
```

### 6. AddrDel() - Lines ~207-211
**DELETE ENTIRE METHOD**
```go
func (nm *NetlinkManager) AddrDel(link *Link, addr *netlink.Addr) error {
    nm.mu.RLock()
    defer nm.mu.RUnlock()
    return netlink.AddrDel(link, addr)
}
```

### 7. RemoveAllAddresses() - Lines ~213-224
**DELETE ENTIRE METHOD**
```go
func (nm *NetlinkManager) RemoveAllAddresses(link *Link, family int) error {
    addrs, err := nm.AddrList(link, family)
    if err != nil {
        return fmt.Errorf("failed to get addresses: %w", err)
    }

    for _, addr := range addrs {
        if err := nm.AddrDel(link, &addr); err != nil {
            nm.logger.Warn().Err(err).Str("address", addr.IP.String()).Msg("failed to remove address")
        }
    }
    return nil
}
```

### 8. RemoveNonLinkLocalIPv6Addresses() - Lines ~226-238
**DELETE ENTIRE METHOD**
```go
func (nm *NetlinkManager) RemoveNonLinkLocalIPv6Addresses(link *Link) error {
    addrs, err := nm.AddrList(link, AfInet6)
    if err != nil {
        return fmt.Errorf("failed to get IPv6 addresses: %w", err)
    }

    for _, addr := range addrs {
        if !addr.IP.IsLinkLocalUnicast() {
            if err := nm.AddrDel(link, &addr); err != nil {
                nm.logger.Warn().Err(err).Str("address", addr.IP.String()).Msg("failed to remove IPv6 address")
            }
        }
    }
    return nil
}
```

### 9. RouteAdd() - Lines ~247-251
**DELETE ENTIRE METHOD**
```go
func (nm *NetlinkManager) RouteAdd(route *netlink.Route) error {
    nm.mu.RLock()
    defer nm.mu.RUnlock()
    return netlink.RouteAdd(route)
}
```

### 10. RouteDel() - Lines ~253-257
**DELETE ENTIRE METHOD**
```go
func (nm *NetlinkManager) RouteDel(route *netlink.Route) error {
    nm.mu.RLock()
    defer nm.mu.RUnlock()
    return netlink.RouteDel(route)
}
```

### 11. RouteReplace() - Lines ~259-263
**DELETE ENTIRE METHOD**
```go
func (nm *NetlinkManager) RouteReplace(route *netlink.Route) error {
    nm.mu.RLock()
    defer nm.mu.RUnlock()
    return netlink.RouteReplace(route)
}
```

### 12. AddDefaultRoute() - Lines ~293-310
**DELETE ENTIRE METHOD**
```go
func (nm *NetlinkManager) AddDefaultRoute(link *Link, gateway net.IP, family int) error {
    var dst *net.IPNet
    switch family {
    case AfInet:
        dst = &ipv4DefaultRoute
    case AfInet6:
        dst = &ipv6DefaultRoute
    default:
        return fmt.Errorf("unsupported address family: %d", family)
    }

    route := &netlink.Route{
        Dst:       dst,
        Gw:        gateway,
        LinkIndex: link.Attrs().Index,
    }
    return nm.RouteReplace(route)
}
```

### 13. RemoveDefaultRoute() - Lines ~312-332
**DELETE ENTIRE METHOD**
```go
func (nm *NetlinkManager) RemoveDefaultRoute(family int) error {
    routes, err := nm.RouteList(nil, family)
    if err != nil {
        return fmt.Errorf("failed to get routes: %w", err)
    }

    for _, route := range routes {
        if route.Dst != nil {
            if family == AfInet && route.Dst.IP.Equal(net.IPv4zero) && route.Dst.Mask.String() == "0.0.0.0/0" {
                if err := nm.RouteDel(&route); err != nil {
                    nm.logger.Warn().Err(err).Msg("failed to remove IPv4 default route")
                }
            }
            if family == AfInet6 && route.Dst.IP.Equal(net.IPv6zero) && route.Dst.Mask.String() == "::/0" {
                if err := nm.RouteDel(&route); err != nil {
                    nm.logger.Warn().Err(err).Msg("failed to remove IPv6 default route")
                }
            }
        }
    }
    return nil
}
```

### 14. reconcileDefaultRoute() - Lines ~334-378
**DELETE ENTIRE METHOD** (45 lines)
This method reconciles default routes by adding/removing them.

### 15. ReconcileLink() - Lines ~380-540
**DELETE ENTIRE METHOD** (160 lines)
This is the main reconciliation method that adds/removes addresses and routes.

## Methods to KEEP (Read Operations)

### ✓ GetLinkByName() - Lines ~107-113
**KEEP** - Only reads link information

### ✓ AddrList() - Lines ~197-201
**KEEP** - Only reads addresses
```go
func (nm *NetlinkManager) AddrList(link *Link, family int) ([]netlink.Addr, error) {
    nm.mu.RLock()
    defer nm.mu.RUnlock()
    return netlink.AddrList(link, family)
}
```

### ✓ RouteList() - Lines ~240-245
**KEEP** - Only reads routes
```go
func (nm *NetlinkManager) RouteList(link *Link, family int) ([]netlink.Route, error) {
    nm.mu.RLock()
    defer nm.mu.RUnlock()
    return netlink.RouteList(link, family)
}
```

### ✓ ListDefaultRoutes() - Lines ~265-278
**KEEP** - Only reads default routes
```go
func (nm *NetlinkManager) ListDefaultRoutes(family int) ([]netlink.Route, error) {
    routes, err := netlink.RouteListFiltered(
        family,
        &netlink.Route{Dst: nil, Table: 254},
        netlink.RT_FILTER_DST|netlink.RT_FILTER_TABLE,
    )
    if err != nil {
        nm.logger.Error().Err(err).Int("family", family).Msg("failed to list default routes")
        return nil, err
    }
    return routes, nil
}
```

### ✓ HasDefaultRoute() - Lines ~280-287
**KEEP** - Only checks if default route exists
```go
func (nm *NetlinkManager) HasDefaultRoute(family int) bool {
    routes, err := nm.ListDefaultRoutes(family)
    if err != nil {
        return false
    }
    return len(routes) > 0
}
```

### ✓ monitorStateChange() - Lines ~50-65
**KEEP** - Only monitors state changes

### ✓ runCallbacks() - Lines ~67-91
**KEEP** - Only runs callbacks on state changes

### ✓ AddStateChangeCallback() - Lines ~47-55
**KEEP** - Only registers callbacks

## Summary

**Lines to DELETE: ~380 lines**
- 15 methods that perform write operations
- All address add/remove operations
- All route add/remove operations
- All interface up/down operations
- All reconciliation logic

**Lines to KEEP: ~150 lines**
- 6 read-only methods
- State monitoring infrastructure
- Callback management

**Result:**
- File size reduced from ~540 lines to ~160 lines
- Only read operations remain
- No network modifications possible
