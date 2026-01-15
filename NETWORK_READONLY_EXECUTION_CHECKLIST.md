# Network Read-Only Mode - Execution Checklist

## Pre-Execution Verification

- [ ] Review all edit plan documents:
  - [ ] NETWORK_READONLY_EDIT_PLAN.md (main plan)
  - [ ] NETWORK_READONLY_INTERFACE_EDITS.md (interface.go details)
  - [ ] NETWORK_READONLY_LINK_MANAGER_EDITS.md (link/manager.go details)
- [ ] Confirm approval from team/stakeholders
- [ ] Create backup branch: `git checkout -b backup-before-readonly`
- [ ] Create working branch: `git checkout -b feature/network-readonly-mode`

## Execution Order

### Phase 1: Delete Complete Files (Safest First)
These files are entirely about writing network config, so delete them first.

- [ ] **Step 1.1:** Delete `pkg/nmlite/link/sysctl.go`
  ```bash
  git rm pkg/nmlite/link/sysctl.go
  ```
  - Impact: Removes all sysctl write operations
  - Callers: interface.go (will be fixed in Phase 3)

- [ ] **Step 1.2:** Delete `pkg/nmlite/static.go`
  ```bash
  git rm pkg/nmlite/static.go
  ```
  - Impact: Removes StaticConfigManager
  - Callers: interface.go (will be fixed in Phase 3)

- [ ] **Step 1.3:** Delete `pkg/nmlite/dhcp.go`
  ```bash
  git rm pkg/nmlite/dhcp.go
  ```
  - Impact: Removes DHCPClient
  - Callers: interface.go (will be fixed in Phase 3)

- [ ] **Step 1.4:** Verify deletion
  ```bash
  git status
  # Should show 3 deleted files
  ```

### Phase 2: Modify link/manager.go
Remove all write operations from the netlink manager.

- [ ] **Step 2.1:** Open `pkg/nmlite/link/manager.go`

- [ ] **Step 2.2:** Delete write methods (15 methods, ~380 lines):
  - [ ] Delete `LinkSetUp()` (lines ~115-119)
  - [ ] Delete `LinkSetDown()` (lines ~121-125)
  - [ ] Delete `EnsureInterfaceUp()` (lines ~127-132)
  - [ ] Delete `EnsureInterfaceUpWithTimeout()` (lines ~134-195)
  - [ ] Delete `AddrAdd()` (lines ~201-205)
  - [ ] Delete `AddrDel()` (lines ~207-211)
  - [ ] Delete `RemoveAllAddresses()` (lines ~213-224)
  - [ ] Delete `RemoveNonLinkLocalIPv6Addresses()` (lines ~226-238)
  - [ ] Delete `RouteAdd()` (lines ~247-251)
  - [ ] Delete `RouteDel()` (lines ~253-257)
  - [ ] Delete `RouteReplace()` (lines ~259-263)
  - [ ] Delete `AddDefaultRoute()` (lines ~293-310)
  - [ ] Delete `RemoveDefaultRoute()` (lines ~312-332)
  - [ ] Delete `reconcileDefaultRoute()` (lines ~334-378)
  - [ ] Delete `ReconcileLink()` (lines ~380-540)

- [ ] **Step 2.3:** Verify remaining methods are read-only:
  - [ ] `GetLinkByName()` ✓
  - [ ] `AddrList()` ✓
  - [ ] `RouteList()` ✓
  - [ ] `ListDefaultRoutes()` ✓
  - [ ] `HasDefaultRoute()` ✓
  - [ ] `monitorStateChange()` ✓
  - [ ] `runCallbacks()` ✓
  - [ ] `AddStateChangeCallback()` ✓

- [ ] **Step 2.4:** Save file

### Phase 3: Modify link/netlink.go
Remove MTU write operation.

- [ ] **Step 3.1:** Open `pkg/nmlite/link/netlink.go`

- [ ] **Step 3.2:** Delete `SetMTU()` method from Link struct
  ```go
  // DELETE THIS:
  func (l *Link) SetMTU(mtu int) error {
      l.mu.Lock()
      defer l.mu.Unlock()
      return netlink.LinkSetMTU(l.Link, mtu)
  }
  ```

- [ ] **Step 3.3:** Save file

### Phase 4: Modify interface.go
This is the most complex file - remove all apply/disable methods.

- [ ] **Step 4.1:** Open `pkg/nmlite/interface.go`

- [ ] **Step 4.2:** Modify InterfaceManager struct:
  - [ ] Remove `staticConfig *StaticConfigManager` field
  - [ ] Remove `dhcpClient *DHCPClient` field
  - [ ] Remove `onResolvConfChange ResolvConfChangeCallback` field

- [ ] **Step 4.3:** Modify `NewInterfaceManager()`:
  - [ ] Remove staticConfig initialization (~5 lines)
  - [ ] Remove dhcpClient initialization (~5 lines)
  - [ ] Remove DHCP callback setup (~20 lines)

- [ ] **Step 4.4:** Modify `Start()`:
  - [ ] Remove commented `EnsureInterfaceUpWithTimeout()` call
  - [ ] Remove commented `applyConfiguration()` call

- [ ] **Step 4.5:** Modify `Stop()`:
  - [ ] Remove DHCP client stop call (~7 lines)

- [ ] **Step 4.6:** Delete methods (20 methods, ~450 lines):
  - [ ] Delete `RenewDHCPLease()` (lines ~350-357)
  - [ ] Delete `SetOnResolvConfChange()` (lines ~370-372)
  - [ ] Delete `applyIPv4Config()` (lines ~374-390)
  - [ ] Delete `applyIPv6Config()` (lines ~392-408)
  - [ ] Delete `applyIPv4Static()` (lines ~410-449)
  - [ ] Delete `applyIPv4DHCP()` (lines ~451-459)
  - [ ] Delete `disableIPv4()` (lines ~461-471)
  - [ ] Delete `applyIPv6Static()` (lines ~473-500)
  - [ ] Delete `applyIPv6DHCP()` (lines ~502-510)
  - [ ] Delete `applyIPv6SLAAC()` (lines ~512-542)
  - [ ] Delete `applyIPv6SLAACAndDHCP()` (lines ~544-556)
  - [ ] Delete `applyIPv6LinkLocal()` (lines ~558-566)
  - [ ] Delete `disableIPv6()` (lines ~568-576)
  - [ ] Delete `SendRouterSolicitation()` (lines ~598-648)
  - [ ] Delete `ReconcileLinkAddrs()` (lines ~745-758)
  - [ ] Delete `applyDHCPLease()` (lines ~759-781)
  - [ ] Delete `convertDHCPLeaseToIPv4Config()` (lines ~783-803)

- [ ] **Step 4.7:** Simplify `handleLinkUp()`:
  ```go
  func (im *InterfaceManager) handleLinkUp() {
      im.logger.Info().Msg("link up")
      // Read-only mode: just log the event
  }
  ```

- [ ] **Step 4.8:** Simplify `handleLinkDown()`:
  ```go
  func (im *InterfaceManager) handleLinkDown() {
      im.logger.Info().Msg("link down")
      // Read-only mode: just log the event
  }
  ```

- [ ] **Step 4.9:** Simplify `updateStateFromDHCPLease()`:
  - [ ] Keep DHCP lease state update
  - [ ] Remove resolv.conf callback (~15 lines)
  - [ ] Add comment about read-only mode

- [ ] **Step 4.10:** Save file

### Phase 5: Modify manager.go
Remove DHCP renewal and cleanup methods.

- [ ] **Step 5.1:** Open `pkg/nmlite/manager.go`

- [ ] **Step 5.2:** Delete methods:
  - [ ] Delete `RenewDHCPLease()` (lines ~180-188)
  - [ ] Delete `shouldKillLegacyDHCPClients()` (lines ~207-220)
  - [ ] Delete `CleanUpLegacyDHCPClients()` (lines ~222-228)

- [ ] **Step 5.3:** Verify `SetInterfaceConfig()` is still commented out ✓

- [ ] **Step 5.4:** Save file

### Phase 6: Compilation Check

- [ ] **Step 6.1:** Attempt to build
  ```bash
  go build ./...
  ```

- [ ] **Step 6.2:** Fix any compilation errors:
  - [ ] Check for calls to deleted methods
  - [ ] Check for references to deleted types
  - [ ] Check for missing imports

- [ ] **Step 6.3:** Run tests
  ```bash
  go test ./pkg/nmlite/...
  ```

- [ ] **Step 6.4:** Fix or remove tests that verify write operations

### Phase 7: Integration Testing

- [ ] **Step 7.1:** Test network state reading:
  - [ ] Can read interface state (up/down)
  - [ ] Can read IP addresses
  - [ ] Can read routes
  - [ ] Can read DHCP leases from system

- [ ] **Step 7.2:** Verify write operations fail gracefully:
  - [ ] RPC methods return appropriate errors
  - [ ] No panics or crashes
  - [ ] Proper error messages

- [ ] **Step 7.3:** Test UI:
  - [ ] Network settings page displays correctly
  - [ ] Shows current network configuration
  - [ ] Shows appropriate read-only messages

### Phase 8: Documentation

- [ ] **Step 8.1:** Update README or docs:
  - [ ] Document read-only mode
  - [ ] Explain how to configure network (use OS tools)
  - [ ] List what operations are no longer supported

- [ ] **Step 8.2:** Update comments in code:
  - [ ] Add "read-only mode" comments where appropriate
  - [ ] Update package documentation

### Phase 9: Commit and Review

- [ ] **Step 9.1:** Review all changes:
  ```bash
  git diff
  ```

- [ ] **Step 9.2:** Commit changes:
  ```bash
  git add -A
  git commit -m "refactor: convert pkg/nmlite to read-only mode

  - Remove all network interface write operations
  - Delete static config manager, DHCP client, and sysctl operations
  - Keep only read operations for monitoring network state
  - Network configuration must now be done via OS tools
  
  Deleted files:
  - pkg/nmlite/static.go
  - pkg/nmlite/dhcp.go
  - pkg/nmlite/link/sysctl.go
  
  Modified files:
  - pkg/nmlite/interface.go (~450 lines removed)
  - pkg/nmlite/manager.go (~50 lines removed)
  - pkg/nmlite/link/manager.go (~380 lines removed)
  - pkg/nmlite/link/netlink.go (removed SetMTU)
  "
  ```

- [ ] **Step 9.3:** Push for review:
  ```bash
  git push origin feature/network-readonly-mode
  ```

## Rollback Plan

If issues are discovered:

```bash
# Rollback to backup branch
git checkout backup-before-readonly

# Or revert the commit
git revert <commit-hash>
```

## Success Criteria

- [ ] Code compiles without errors
- [ ] All tests pass
- [ ] Network state can be read
- [ ] No network modifications occur
- [ ] RPC methods return appropriate errors
- [ ] UI displays network info correctly
- [ ] No crashes or panics
- [ ] Documentation updated

## Estimated Time

- Phase 1 (Delete files): 5 minutes
- Phase 2 (link/manager.go): 20 minutes
- Phase 3 (link/netlink.go): 5 minutes
- Phase 4 (interface.go): 45 minutes
- Phase 5 (manager.go): 10 minutes
- Phase 6 (Compilation): 30 minutes
- Phase 7 (Integration testing): 60 minutes
- Phase 8 (Documentation): 20 minutes
- Phase 9 (Commit/Review): 15 minutes

**Total: ~3.5 hours**
