package kvm

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/jetkvm/kvm/internal/mdns"
	"github.com/jetkvm/kvm/internal/network/types"
	"github.com/jetkvm/kvm/pkg/myip"
	"github.com/jetkvm/kvm/pkg/nmlite"
	"github.com/jetkvm/kvm/pkg/nmlite/link"
)

const (
	NetIfName = "wlan0"
)

var (
	networkManager *nmlite.NetworkManager
	publicIPState  *myip.PublicIPState
)

type RpcNetworkSettings struct {
	types.NetworkConfig
}

func (s *RpcNetworkSettings) ToNetworkConfig() *types.NetworkConfig {
	return &s.NetworkConfig
}

type PostRebootAction struct {
	HealthCheck string `json:"healthCheck"`
	RedirectTo  string `json:"redirectTo"`
}

func toRpcNetworkSettings(config *types.NetworkConfig) *RpcNetworkSettings {
	return &RpcNetworkSettings{
		NetworkConfig: *config,
	}
}

func getMdnsOptions() *mdns.MDNSOptions {
	if networkManager == nil {
		return nil
	}

	var ipv4, ipv6 bool
	switch config.NetworkConfig.MDNSMode.String {
	case "auto":
		ipv4 = true
		ipv6 = true
	case "ipv4_only":
		ipv4 = true
	case "ipv6_only":
		ipv6 = true
	}

	return &mdns.MDNSOptions{
		LocalNames: []string{
			networkManager.Hostname(),
			networkManager.FQDN(),
		},
		ListenOptions: &mdns.MDNSListenOptions{
			IPv4: ipv4,
			IPv6: ipv6,
		},
	}
}

func restartMdns() {
	if mDNS == nil {
		return
	}

	options := getMdnsOptions()
	if options == nil {
		return
	}

	if err := mDNS.SetOptions(options); err != nil {
		networkLogger.Error().Err(err).Msg("failed to restart mDNS")
	}
}

func setPublicIPReadyState(ipv4Ready, ipv6Ready bool) {
	if publicIPState == nil {
		return
	}
	publicIPState.SetIPv4AndIPv6(ipv4Ready, ipv6Ready)
}

func networkStateChanged(_ string, state types.InterfaceState) {
	// do not block the main thread

	if currentSession != nil {
		writeJSONRPCEvent("networkState", state.ToRpcInterfaceState(), currentSession)
	}

	if state.Online {
		networkLogger.Info().Msg("network state changed to online")
	}

	setPublicIPReadyState(state.IPv4Ready, state.IPv6Ready)

	// always restart mDNS when the network state changes
	if mDNS != nil {
		restartMdns()
	}
}

func initNetwork() error {
	ensureConfigLoaded()

	// On full Linux systems, we only read network state - no configuration management
	// Network management is handled by the OS (NetworkManager, systemd-networkd, etc.)
	networkLogger.Info().Msg("initializing network manager in read-only mode")

	// Create a minimal network manager for status reading only
	nm := nmlite.NewNetworkManager(context.Background(), networkLogger)
	nm.SetOnInterfaceStateChange(networkStateChanged)

	// Try to add interface for monitoring only - don't fail if it doesn't work
	if err := nm.AddInterface(NetIfName, config.NetworkConfig); err != nil {
		networkLogger.Warn().Err(err).Str("interface", NetIfName).Msg("failed to add interface for monitoring - network status may be limited")
		// Don't return error - we can still function without full network monitoring
	}

	networkManager = nm
	networkLogger.Info().Msg("network manager initialized in read-only mode")
	return nil
}

func initPublicIPState() {
	// the feature will be only enabled if the cloud has been adopted
	// due to privacy reasons

	// but it will be initialized anyway to avoid nil pointer dereferences
	ps := myip.NewPublicIPState(&myip.PublicIPStateConfig{
		Logger:             networkLogger,
		CloudflareEndpoint: config.CloudURL,
		APIEndpoint:        "",
		IPv4:               false,
		IPv6:               false,
		HttpClientGetter: func(family int) *http.Client {
			transport := http.DefaultTransport.(*http.Transport).Clone()
			transport.Proxy = config.NetworkConfig.GetTransportProxyFunc()
			transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				netType := network
				switch family {
				case link.AfInet:
					netType = "tcp4"
				case link.AfInet6:
					netType = "tcp6"
				}
				return (&net.Dialer{}).DialContext(ctx, netType, addr)
			}

			return &http.Client{
				Transport: transport,
				Timeout:   30 * time.Second,
			}
		},
	})
	publicIPState = ps
}

func setHostname(nm *nmlite.NetworkManager, hostname, domain string) error {
	// Hostname setting disabled on full Linux systems
	// Hostname management should be handled by the OS
	networkLogger.Info().Str("hostname", hostname).Str("domain", domain).Msg("hostname setting disabled - use OS hostname management")
	return fmt.Errorf("hostname setting is disabled - use OS hostname management tools (hostnamectl, etc.)")
}

func rpcGetNetworkState() *types.RpcInterfaceState {
	state, _ := networkManager.GetInterfaceState(NetIfName)
	return state.ToRpcInterfaceState()
}

func rpcGetNetworkSettings() *RpcNetworkSettings {
	return toRpcNetworkSettings(config.NetworkConfig)
}

func rpcSetNetworkSettings(settings RpcNetworkSettings) (*RpcNetworkSettings, error) {
	// Network configuration is read-only on full Linux systems
	// Network management should be handled by the OS (NetworkManager, systemd-networkd, etc.)
	networkLogger.Warn().Msg("Network configuration changes are disabled - use OS network management tools")
	return nil, fmt.Errorf("network configuration is read-only - use OS network management tools (NetworkManager, systemd-networkd, etc.)")
}

func rpcRenewDHCPLease() error {
	// DHCP lease renewal is read-only on full Linux systems
	// DHCP management should be handled by the OS
	networkLogger.Warn().Msg("DHCP lease renewal is disabled - use OS DHCP client management")
	return fmt.Errorf("DHCP lease renewal is read-only - use OS DHCP client (dhclient, NetworkManager, etc.)")
}

func rpcToggleDHCPClient() error {
	// DHCP client switching is read-only on full Linux systems
	// DHCP client management should be handled by the OS
	networkLogger.Warn().Msg("DHCP client switching is disabled - use OS network management")
	return fmt.Errorf("DHCP client switching is read-only - use OS network management tools")
}

func rpcGetPublicIPAddresses(refresh bool) ([]myip.PublicIP, error) {
	if publicIPState == nil {
		return nil, fmt.Errorf("public IP state not initialized")
	}

	if refresh {
		if err := publicIPState.ForceUpdate(); err != nil {
			return nil, err
		}
	}

	return publicIPState.GetAddresses(), nil
}

func rpcCheckPublicIPAddresses() error {
	if publicIPState == nil {
		return fmt.Errorf("public IP state not initialized")
	}

	return publicIPState.ForceUpdate()
}
