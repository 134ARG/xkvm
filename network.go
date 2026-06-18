package kvm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/134ARG/xkvm/internal/mdns"
	"github.com/134ARG/xkvm/internal/network/types"
	"github.com/134ARG/xkvm/pkg/nmlite"
)

const (
	NetIfName             = "wlan0"
	publicIPLookupTimeout = 5 * time.Second
)

var (
	networkManager *nmlite.NetworkManager
)

type RpcNetworkSettings struct {
	types.NetworkConfig
}

type RpcPublicIP struct {
	Family      string     `json:"family"`
	IPAddress   string     `json:"ip,omitempty"`
	LastUpdated *time.Time `json:"last_updated,omitempty"`
	Error       string     `json:"error,omitempty"`
}

func (s *RpcNetworkSettings) ToNetworkConfig() *types.NetworkConfig {
	return &s.NetworkConfig
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

func networkStateChanged(_ string, state types.InterfaceState) {
	// do not block the main thread

	if cs := getCurrentSession(); cs != nil {
		writeJSONRPCEvent("networkState", state.ToRpcInterfaceState(), cs)
	}

	if state.Online {
		networkLogger.Info().Msg("network state changed to online")
	}

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

// func setHostname(nm *nmlite.NetworkManager, hostname, domain string) error {
// 	// Hostname setting disabled on full Linux systems
// 	// Hostname management should be handled by the OS
// 	networkLogger.Info().Str("hostname", hostname).Str("domain", domain).Msg("hostname setting disabled - use OS hostname management")
// 	return fmt.Errorf("hostname setting is disabled - use OS hostname management tools (hostnamectl, etc.)")
// }

func rpcGetNetworkState() *types.RpcInterfaceState {
	state, _ := networkManager.GetInterfaceState(NetIfName)
	return state.ToRpcInterfaceState()
}

func rpcGetNetworkSettings() *RpcNetworkSettings {
	// Return minimal default settings for backward compatibility
	// Actual network configuration should be read from rpcGetNetworkState()
	return toRpcNetworkSettings(config.NetworkConfig)
}

// rpcGetNetworkSettingsDeprecated is a deprecated wrapper that logs a warning
// Deprecated: Use rpcGetNetworkState() instead. Network config is read-only.
func rpcGetNetworkSettingsDeprecated() *RpcNetworkSettings {
	networkLogger.Warn().Msg("getNetworkSettings is deprecated - use getNetworkState instead")
	return rpcGetNetworkSettings()
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

// func rpcToggleDHCPClient() error {
// 	// DHCP client switching is read-only on full Linux systems
// 	// DHCP client management should be handled by the OS
// 	networkLogger.Warn().Msg("DHCP client switching is disabled - use OS network management")
// 	return fmt.Errorf("DHCP client switching is read-only - use OS network management tools")
// }

func rpcGetPublicIPAddresses(_ bool) ([]RpcPublicIP, error) {
	now := time.Now()

	return []RpcPublicIP{
		queryPublicIPRow("ipv4", config.PublicIPv4Endpoint, "tcp4", true, now),
		queryPublicIPRow("ipv6", config.PublicIPv6Endpoint, "tcp6", false, now),
	}, nil
}

func rpcCheckPublicIPAddresses() error {
	_, _ = rpcGetPublicIPAddresses(true)
	return nil
}

func queryPublicIPRow(family, endpoint, network string, ipv4 bool, now time.Time) RpcPublicIP {
	result := RpcPublicIP{Family: family}

	ip, code, err := queryPublicIP(endpoint, network, ipv4)
	if err == nil {
		networkLogger.Debug().Str("family", family).Str("ip", ip.String()).Msg("public IP query succeeded")
		result.IPAddress = ip.String()
		result.LastUpdated = &now
		return result
	}

	result.Error = code
	if strings.TrimSpace(endpoint) == "" {
		networkLogger.Debug().Str("family", family).Str("error", code).Msg("public IP endpoint is not configured")
	} else {
		networkLogger.Warn().Err(err).Str("family", family).Str("endpoint", endpoint).Str("error", code).Msg("public IP query failed")
	}
	return result
}

func queryPublicIP(endpoint, network string, ipv4 bool) (net.IP, string, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return nil, "not configured", fmt.Errorf("public IP endpoint is not configured")
	}
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, _, addr string) (net.Conn, error) {
		return (&net.Dialer{Timeout: publicIPLookupTimeout}).DialContext(ctx, network, addr)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   publicIPLookupTimeout,
	}

	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, publicIPRequestErrorCode(err), err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, strconv.Itoa(resp.StatusCode), fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 128))
	if err != nil {
		return nil, "read failed", err
	}

	ip := net.ParseIP(strings.TrimSpace(string(body)))
	if ip == nil {
		return nil, "parse failed", fmt.Errorf("invalid IP address")
	}
	if ipv4 {
		ip = ip.To4()
		if ip == nil {
			return nil, "wrong family", fmt.Errorf("expected IPv4 address")
		}
		return ip, "", nil
	}
	if ip.To4() != nil {
		return nil, "wrong family", fmt.Errorf("expected IPv6 address")
	}
	return ip, "", nil
}

func publicIPRequestErrorCode(err error) string {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}
	return "request failed"
}
