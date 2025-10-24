package network

import (
	"fmt"
	"net"
	"sync"

	"github.com/jetkvm/kvm/internal/confparser"
	"github.com/rs/zerolog"
	"github.com/vishvananda/netlink"
)

// NetworkInterfaceState now acts as a read-only observer
// Renamed fields removed: dhcpClient, cbConfigChange, onInitialCheck, checked, defaultHostname
type NetworkInterfaceState struct {
	interfaceName string
	interfaceUp   bool
	ipv4Addr      *net.IP
	ipv4Addresses []string
	ipv6Addr      *net.IP
	ipv6Addresses []IPv6Address
	ipv6LinkLocal *net.IP
	ntpAddresses  []*net.IP
	macAddr       *net.HardwareAddr

	l         *zerolog.Logger
	stateLock sync.RWMutex // Changed to RWMutex for better concurrency

	config *NetworkConfig

	currentHostname string
	currentFqdn     string

	onStateChange func(state *NetworkInterfaceState)

	// Channel to stop monitoring
	stopChan chan struct{}
	stopped  bool
}

type NetworkInterfaceOptions struct {
	InterfaceName string
	Logger        *zerolog.Logger
	OnStateChange func(state *NetworkInterfaceState)
	NetworkConfig *NetworkConfig
}

func NewNetworkInterfaceState(opts *NetworkInterfaceOptions) (*NetworkInterfaceState, error) {
	if opts.NetworkConfig == nil {
		return nil, fmt.Errorf("NetworkConfig cannot be nil")
	}

	err := confparser.SetDefaultsAndValidate(opts.NetworkConfig)
	if err != nil {
		return nil, err
	}

	s := &NetworkInterfaceState{
		interfaceName: opts.InterfaceName,
		l:             opts.Logger,
		onStateChange: opts.OnStateChange,
		config:        opts.NetworkConfig,
		ntpAddresses:  make([]*net.IP, 0),
		stopChan:      make(chan struct{}),
	}

	return s, nil
}

// Getter methods (unchanged)
func (s *NetworkInterfaceState) IsUp() bool {
	// s.stateLock.RLock()
	// defer s.stateLock.RUnlock()
	return s.interfaceUp
}

func (s *NetworkInterfaceState) HasIPAssigned() bool {
	// s.stateLock.RLock()
	// defer s.stateLock.RUnlock()
	return s.ipv4Addr != nil || s.ipv6Addr != nil
}

func (s *NetworkInterfaceState) IsOnline() bool {
	return s.IsUp() && s.HasIPAssigned()
}

func (s *NetworkInterfaceState) IPv4() *net.IP {
	// s.stateLock.RLock()
	// defer s.stateLock.RUnlock()
	return s.ipv4Addr
}

func (s *NetworkInterfaceState) IPv4String() string {
	if ip := s.IPv4(); ip != nil {
		return ip.String()
	}
	return "..."
}

func (s *NetworkInterfaceState) IPv6() *net.IP {
	// s.stateLock.RLock()
	// defer s.stateLock.RUnlock()
	return s.ipv6Addr
}

func (s *NetworkInterfaceState) IPv6String() string {
	if ip := s.IPv6(); ip != nil {
		return ip.String()
	}
	return "..."
}

func (s *NetworkInterfaceState) NtpAddresses() []*net.IP {
	// s.stateLock.RLock()
	// defer s.stateLock.RUnlock()
	return s.ntpAddresses
}

func (s *NetworkInterfaceState) NtpAddressesString() []string {
	// s.stateLock.RLock()
	// defer s.stateLock.RUnlock()

	ntpServers := []string{}
	if len(s.ntpAddresses) > 0 {
		for _, server := range s.ntpAddresses {
			s.l.Debug().IPAddr("server", *server).Msg("converting NTP address")
			ntpServers = append(ntpServers, server.String())
		}
	}
	return ntpServers
}

func (s *NetworkInterfaceState) MAC() *net.HardwareAddr {
	// s.stateLock.RLock()
	// defer s.stateLock.RUnlock()
	return s.macAddr
}

func (s *NetworkInterfaceState) MACString() string {
	if mac := s.MAC(); mac != nil {
		return mac.String()
	}
	return ""
}

func (s *NetworkInterfaceState) GetHostname() string {
	// s.stateLock.RLock()
	// defer s.stateLock.RUnlock()
	return s.currentHostname
}

func (s *NetworkInterfaceState) GetFQDN() string {
	// s.stateLock.RLock()
	// defer s.stateLock.RUnlock()
	return s.currentFqdn
}

// update reads current state from system (read-only, no writes)
func (s *NetworkInterfaceState) update() error {
	// s.stateLock.Lock()
	// defer s.stateLock.Unlock()

	iface, err := netlink.LinkByName(s.interfaceName)
	if err != nil {
		s.l.Error().Err(err).Msg("failed to get interface")
		return err
	}

	// Detect if the interface status changed
	var changed bool
	attrs := iface.Attrs()
	state := attrs.OperState
	newInterfaceUp := state == netlink.OperUp

	// Check if the interface is coming up/down
	interfaceGoingUp := !s.interfaceUp && newInterfaceUp
	interfaceGoingDown := s.interfaceUp && !newInterfaceUp

	if s.interfaceUp != newInterfaceUp {
		s.interfaceUp = newInterfaceUp
		changed = true
	}

	if interfaceGoingUp {
		s.l.Info().Msg("interface state transitioned to up")
	} else if interfaceGoingDown {
		s.l.Info().Msg("interface state transitioned to down")
	}

	// Set the MAC address
	s.macAddr = &attrs.HardwareAddr

	// Get the IP addresses
	addrs, err := netlinkAddrs(iface)
	if err != nil {
		s.l.Error().Err(err).Msg("failed to get ip addresses")
		return err
	}

	var (
		ipv4Addresses       = make([]net.IP, 0)
		ipv4AddressesString = make([]string, 0)
		ipv6Addresses       = make([]IPv6Address, 0)
		ipv6LinkLocal       *net.IP
	)

	for _, addr := range addrs {
		if addr.IP.To4() != nil {
			// IPv4 - just read, don't delete
			ipv4Addresses = append(ipv4Addresses, addr.IP)
			ipv4AddressesString = append(ipv4AddressesString, addr.IPNet.String())
		} else if addr.IP.To16() != nil {
			if s.config.IPv6Mode.String == "disabled" {
				continue
			}

			// Check if it's a link local address
			if addr.IP.IsLinkLocalUnicast() {
				ipv6LinkLocal = &addr.IP
				continue
			}

			if !addr.IP.IsGlobalUnicast() {
				s.l.Trace().Str("ipv6", addr.IP.String()).Msg("not a global unicast address, skipping")
				continue
			}

			ipv6Addresses = append(ipv6Addresses, IPv6Address{
				Address:           addr.IP,
				Prefix:            *addr.IPNet,
				ValidLifetime:     lifetimeToTime(addr.ValidLft),
				PreferredLifetime: lifetimeToTime(addr.PreferedLft),
				Scope:             addr.Scope,
			})
		}
	}

	// Update IPv4
	if len(ipv4Addresses) > 0 {
		if s.ipv4Addr == nil || s.ipv4Addr.String() != ipv4Addresses[0].String() {
			scopedLogger := s.l.With().Str("ipv4", ipv4Addresses[0].String()).Logger()
			if s.ipv4Addr != nil {
				scopedLogger.Info().
					Str("old_ipv4", s.ipv4Addr.String()).
					Msg("IPv4 address changed")
			} else {
				scopedLogger.Info().Msg("IPv4 address found")
			}
			s.ipv4Addr = &ipv4Addresses[0]
			changed = true
		}
	} else if s.ipv4Addr != nil {
		// IP was removed
		s.l.Info().Str("old_ipv4", s.ipv4Addr.String()).Msg("IPv4 address removed")
		s.ipv4Addr = nil
		changed = true
	}
	s.ipv4Addresses = ipv4AddressesString

	// Update IPv6
	if s.config.IPv6Mode.String != "disabled" {
		if ipv6LinkLocal != nil {
			if s.ipv6LinkLocal == nil || s.ipv6LinkLocal.String() != ipv6LinkLocal.String() {
				scopedLogger := s.l.With().Str("ipv6", ipv6LinkLocal.String()).Logger()
				if s.ipv6LinkLocal != nil {
					scopedLogger.Info().
						Str("old_ipv6", s.ipv6LinkLocal.String()).
						Msg("IPv6 link local address changed")
				} else {
					scopedLogger.Info().Msg("IPv6 link local address found")
				}
				s.ipv6LinkLocal = ipv6LinkLocal
				changed = true
			}
		}
		s.ipv6Addresses = ipv6Addresses

		if len(ipv6Addresses) > 0 {
			if s.ipv6Addr == nil || s.ipv6Addr.String() != ipv6Addresses[0].Address.String() {
				scopedLogger := s.l.With().Str("ipv6", ipv6Addresses[0].Address.String()).Logger()
				if s.ipv6Addr != nil {
					scopedLogger.Info().
						Str("old_ipv6", s.ipv6Addr.String()).
						Msg("IPv6 address changed")
				} else {
					scopedLogger.Info().Msg("IPv6 address found")
				}
				s.ipv6Addr = &ipv6Addresses[0].Address
				changed = true
			}
		} else if s.ipv6Addr != nil {
			// IPv6 was removed
			s.l.Info().Str("old_ipv6", s.ipv6Addr.String()).Msg("IPv6 address removed")
			s.ipv6Addr = nil
			changed = true
		}
	}

	// Update hostname/FQDN from system
	s.updateHostnameFromSystem()
	// Trigger callback if changed
	if changed && s.onStateChange != nil {
		s.onStateChange(s)
	}

	return nil
}

// updateHostnameFromSystem reads hostname from system
func (s *NetworkInterfaceState) updateHostnameFromSystem() {
	// Read from /etc/hostname or use net.LookupHost
	hostname, err := readSystemHostname()
	if err != nil {
		s.l.Warn().Err(err).Msg("failed to read hostname")
		hostname = "jetkvm"
	}
	s.currentHostname = hostname

	// Try to resolve FQDN
	if s.ipv4Addr != nil {
		if names, err := net.LookupAddr(s.ipv4Addr.String()); err == nil && len(names) > 0 {
			s.currentFqdn = names[0]
		} else {
			s.currentFqdn = hostname
		}
	} else {
		s.currentFqdn = hostname
	}
}
