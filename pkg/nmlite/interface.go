package nmlite

import (
	"context"
	"fmt"
	"net"

	"time"

	"github.com/xkvm/kvm/internal/sync"

	"github.com/rs/zerolog"
	"github.com/vishvananda/netlink"
	"github.com/xkvm/kvm/internal/confparser"
	"github.com/xkvm/kvm/internal/logging"
	"github.com/xkvm/kvm/internal/network/types"
	"github.com/xkvm/kvm/pkg/nmlite/link"
)

type ResolvConfChangeCallback func(family int, resolvConf *types.InterfaceResolvConf) error

// InterfaceManager manages a single network interface
type InterfaceManager struct {
	ctx       context.Context
	ifaceName string
	config    *types.NetworkConfig
	logger    *zerolog.Logger
	state     *types.InterfaceState
	linkState *link.Link
	stateMu   sync.RWMutex

	// Network components - removed in read-only mode
	// staticConfig *StaticConfigManager
	// dhcpClient   *DHCPClient

	// Callbacks
	onStateChange     func(state types.InterfaceState)
	onConfigChange    func(config *types.NetworkConfig)
	onDHCPLeaseChange func(lease *types.DHCPLease)

	// Control
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewInterfaceManager creates a new interface manager
func NewInterfaceManager(ctx context.Context, ifaceName string, config *types.NetworkConfig, logger *zerolog.Logger) (*InterfaceManager, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if logger == nil {
		logger = logging.GetSubsystemLogger("interface")
	}

	scopedLogger := logger.With().Str("interface", ifaceName).Logger()

	// Validate and set defaults
	if err := confparser.SetDefaultsAndValidate(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	im := &InterfaceManager{
		ctx:       ctx,
		ifaceName: ifaceName,
		config:    config,
		logger:    &scopedLogger,
		state: &types.InterfaceState{
			InterfaceName: ifaceName,
		},
		stopCh: make(chan struct{}),
	}

	// Read-only mode: no static config manager or DHCP client initialization

	return im, nil
}

// Start starts managing the interface
func (im *InterfaceManager) Start() error {
	im.stateMu.Lock()
	defer im.stateMu.Unlock()

	im.logger.Info().Msg("starting interface manager (read-only mode)")

	// Start monitoring interface state
	im.wg.Add(1)
	go im.monitorInterfaceState()

	nl := getNetlinkManager()

	// Set the link state
	linkState, err := nl.GetLinkByName(im.ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface: %w", err)
	}
	im.linkState = linkState

	// Read-only mode: no interface up or configuration application

	// Set callback after the interface is up
	nl.AddStateChangeCallback(im.ifaceName, link.StateChangeCallback{
		Async: true,
		Func: func(link *link.Link) {
			im.handleLinkStateChange(link)
		},
	})

	im.logger.Info().Msg("interface manager started (read-only mode)")
	return nil
}

// Stop stops managing the interface
func (im *InterfaceManager) Stop() error {
	im.logger.Info().Msg("stopping interface manager")

	close(im.stopCh)
	im.wg.Wait()

	// Read-only mode: no DHCP client to stop

	im.logger.Info().Msg("interface manager stopped")
	return nil
}

func (im *InterfaceManager) link() (*link.Link, error) {
	nl := getNetlinkManager()
	if nl == nil {
		return nil, fmt.Errorf("netlink manager not initialized")
	}
	return nl.GetLinkByName(im.ifaceName)
}

// IsUp returns true if the interface is up
func (im *InterfaceManager) IsUp() bool {
	im.stateMu.RLock()
	defer im.stateMu.RUnlock()

	if im.state == nil {
		return false
	}

	return im.state.Up
}

// IsOnline returns true if the interface is online
func (im *InterfaceManager) IsOnline() bool {
	im.stateMu.RLock()
	defer im.stateMu.RUnlock()

	if im.state == nil {
		return false
	}

	return im.state.Online
}

// IPv4Ready returns true if the interface has an IPv4 address
func (im *InterfaceManager) IPv4Ready() bool {
	im.stateMu.RLock()
	defer im.stateMu.RUnlock()

	if im.state == nil {
		return false
	}

	return im.state.IPv4Ready
}

// IPv6Ready returns true if the interface has an IPv6 address
func (im *InterfaceManager) IPv6Ready() bool {
	im.stateMu.RLock()
	defer im.stateMu.RUnlock()

	if im.state == nil {
		return false
	}

	return im.state.IPv6Ready
}

// GetIPv4Addresses returns the IPv4 addresses of the interface
func (im *InterfaceManager) GetIPv4Addresses() []string {
	im.stateMu.RLock()
	defer im.stateMu.RUnlock()

	if im.state == nil {
		return []string{}
	}

	return im.state.IPv4Addresses
}

// GetIPv4Address returns the IPv4 address of the interface
func (im *InterfaceManager) GetIPv4Address() string {
	im.stateMu.RLock()
	defer im.stateMu.RUnlock()

	if im.state == nil {
		return ""
	}

	return im.state.IPv4Address
}

// GetIPv6Address returns the IPv6 address of the interface
func (im *InterfaceManager) GetIPv6Address() string {
	im.stateMu.RLock()
	defer im.stateMu.RUnlock()

	if im.state == nil {
		return ""
	}

	return im.state.IPv6Address
}

// GetIPv6Addresses returns the IPv6 addresses of the interface
func (im *InterfaceManager) GetIPv6Addresses() []string {
	im.stateMu.RLock()
	defer im.stateMu.RUnlock()

	addresses := []string{}

	if im.state == nil {
		return addresses
	}

	for _, addr := range im.state.IPv6Addresses {
		addresses = append(addresses, addr.Address.String())
	}

	return addresses
}

// GetMACAddress returns the MAC address of the interface
func (im *InterfaceManager) GetMACAddress() string {
	im.stateMu.RLock()
	defer im.stateMu.RUnlock()

	if im.state == nil {
		return ""
	}

	return im.state.MACAddress
}

// GetState returns the current interface state
func (im *InterfaceManager) GetState() *types.InterfaceState {
	im.stateMu.RLock()
	defer im.stateMu.RUnlock()

	// Return a copy to avoid race conditions
	im.logger.Debug().Interface("state", im.state).Msg("getting interface state")

	state := *im.state
	return &state
}

// NTPServers returns the NTP servers of the interface
func (im *InterfaceManager) NTPServers() []net.IP {
	im.stateMu.RLock()
	defer im.stateMu.RUnlock()

	if im.state == nil {
		return []net.IP{}
	}

	return im.state.NTPServers
}

func (im *InterfaceManager) Domain() string {
	im.stateMu.RLock()
	defer im.stateMu.RUnlock()

	if im.state == nil {
		return ""
	}

	if im.state.DHCPLease4 != nil {
		return im.state.DHCPLease4.Domain
	}

	if im.state.DHCPLease6 != nil {
		return im.state.DHCPLease6.Domain
	}

	return ""
}

// GetConfig returns the current interface configuration
func (im *InterfaceManager) GetConfig() *types.NetworkConfig {
	// Return a copy to avoid race conditions
	config := *im.config
	return &config
}

// SetOnStateChange sets the callback for state changes
func (im *InterfaceManager) SetOnStateChange(callback func(state types.InterfaceState)) {
	im.onStateChange = callback
}

// SetOnConfigChange sets the callback for configuration changes
func (im *InterfaceManager) SetOnConfigChange(callback func(config *types.NetworkConfig)) {
	im.onConfigChange = callback
}

// SetOnDHCPLeaseChange sets the callback for DHCP lease changes
func (im *InterfaceManager) SetOnDHCPLeaseChange(callback func(lease *types.DHCPLease)) {
	im.onDHCPLeaseChange = callback
}

// Read-only mode: all apply/disable methods removed

func (im *InterfaceManager) handleLinkStateChange(link *link.Link) {
	{
		im.stateMu.Lock()
		defer im.stateMu.Unlock()

		if link.IsSame(im.linkState) {
			return
		}

		im.linkState = link
	}

	im.logger.Info().Interface("link", link).Msg("link state changed")

	operState := link.Attrs().OperState
	if operState == netlink.OperUp {
		im.handleLinkUp()
	} else {
		im.handleLinkDown()
	}
}

func (im *InterfaceManager) handleLinkUp() {
	im.logger.Info().Msg("link up (read-only mode)")
	// Read-only mode: just log the event, don't modify network config
}

func (im *InterfaceManager) handleLinkDown() {
	im.logger.Info().Msg("link down (read-only mode)")
	// Read-only mode: just log the event, don't modify network config
}

// monitorInterfaceState monitors the interface state and updates accordingly
func (im *InterfaceManager) monitorInterfaceState() {
	defer im.wg.Done()

	im.logger.Debug().Msg("monitoring interface state")
	// TODO: use netlink subscription instead of polling
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-im.ctx.Done():
			return
		case <-im.stopCh:
			return
		case <-ticker.C:
			if err := im.updateInterfaceState(); err != nil {
				im.logger.Error().Err(err).Msg("failed to update interface state")
			}
		}
	}
}

// updateStateFromDHCPLease updates the state from a DHCP lease
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
