package types

import (
	"encoding/json"
	"net"
	"time"

	"golang.org/x/sys/unix"
)

// InterfaceState represents the current state of a network interface
type InterfaceState struct {
	InterfaceName string        `json:"interface_name"`
	Hostname      string        `json:"hostname"`
	MACAddress    string        `json:"mac_address"`
	Up            bool          `json:"up"`
	Online        bool          `json:"online"`
	IPv4Ready     bool          `json:"ipv4_ready"`
	IPv6Ready     bool          `json:"ipv6_ready"`
	IPv4Address   string        `json:"ipv4_address,omitempty"`
	IPv6Address   string        `json:"ipv6_address,omitempty"`
	IPv6LinkLocal string        `json:"ipv6_link_local,omitempty"`
	IPv6Gateway   string        `json:"ipv6_gateway,omitempty"`
	IPv4Addresses []string      `json:"ipv4_addresses,omitempty"`
	IPv6Addresses []IPv6Address `json:"ipv6_addresses,omitempty"`
	NTPServers    []net.IP      `json:"ntp_servers,omitempty"`
	DHCPLease4    *DHCPLease    `json:"dhcp_lease,omitempty"`
	DHCPLease6    *DHCPLease    `json:"dhcp_lease6,omitempty"`
	LastUpdated   time.Time     `json:"last_updated"`
}

// RpcDHCPLease is the trimmed wire representation of a DHCP lease — only the fields the UI
// renders, keeping the network-state payload small enough for a single SCTP frame.
type RpcDHCPLease struct {
	IPAddress   net.IP     `json:"ip,omitempty"`
	Netmask     net.IP     `json:"netmask,omitempty"`
	Routers     []net.IP   `json:"routers,omitempty"`
	DNS         []net.IP   `json:"dns_servers,omitempty"`
	ServerID    string     `json:"server_id,omitempty"`
	LeaseExpiry *time.Time `json:"lease_expiry,omitempty"`
	DHCPClient  string     `json:"dhcp_client,omitempty"`
}

func toRpcDHCPLease(l *DHCPLease) *RpcDHCPLease {
	if l == nil {
		return nil
	}
	return &RpcDHCPLease{
		IPAddress:   l.IPAddress,
		Netmask:     l.Netmask,
		Routers:     l.Routers,
		DNS:         l.DNS,
		ServerID:    l.ServerID,
		LeaseExpiry: truncateToSecond(l.LeaseExpiry),
		DHCPClient:  l.DHCPClient,
	}
}

// truncateToSecond drops sub-second precision from a timestamp. Lifetimes are derived from
// time.Now() so they carry useless nanoseconds that bloat the JSON (~10 bytes each); the UI
// only renders minute precision.
func truncateToSecond(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	s := t.Truncate(time.Second)
	return &s
}

// secondsRemaining converts an absolute expiry into whole seconds from now — a compact
// integer instead of an RFC3339 string. Returns nil for an unset time so the field is
// omitted; past times clamp to 0.
func secondsRemaining(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	s := int64(time.Until(*t).Seconds())
	if s < 0 {
		s = 0
	}
	return &s
}

// prefixNetwork returns the masked network of an address prefix (e.g. "fd00:9178:9178::/64")
// — the actual prefix, rather than the full host address repeated with a prefix length.
func prefixNetwork(p net.IPNet) string {
	if len(p.IP) == 0 || len(p.Mask) == 0 {
		return ""
	}
	return (&net.IPNet{IP: p.IP.Mask(p.Mask), Mask: p.Mask}).String()
}

// RpcInterfaceState is the RPC representation of an interface state. It lists only the
// fields the UI consumes (not the full internal InterfaceState), so the payload stays
// within a single SCTP frame that Firefox can reassemble.
type RpcInterfaceState struct {
	InterfaceName string           `json:"interface_name"`
	Hostname      string           `json:"hostname"`
	MACAddress    string           `json:"mac_address"`
	Up            bool             `json:"up"`
	Online        bool             `json:"online"`
	IPv4Address   string           `json:"ipv4_address,omitempty"`
	IPv6Address   string           `json:"ipv6_address,omitempty"`
	IPv6LinkLocal string           `json:"ipv6_link_local,omitempty"`
	IPv6Gateway   string           `json:"ipv6_gateway,omitempty"`
	IPv4Addresses []string         `json:"ipv4_addresses,omitempty"`
	IPv6Addresses []RpcIPv6Address `json:"ipv6_addresses"`
	DHCPLease     *RpcDHCPLease    `json:"dhcp_lease,omitempty"`
}

// maxRpcIPv6Addresses bounds how many IPv6 addresses we put on the wire. We only expose
// routable addresses (link-local is already in the IPv6LinkLocal scalar), so the cap only
// ever drops surplus global/ULA addresses.
const maxRpcIPv6Addresses = 4

// rpcStateMaxBytes is the serialized-size budget for a network-state message. pion's
// multi-fragment SCTP messages are not reassembled by Firefox, so getNetworkState must fit
// in a single SCTP DATA chunk (~1095 B usable over a 1200-byte path MTU). We budget below
// that — leaving room for the JSON-RPC envelope — by dropping IPv6 addresses until it fits.
const rpcStateMaxBytes = 1024

// ToRpcInterfaceState converts an InterfaceState to a RpcInterfaceState. It exposes only
// routable IPv6 addresses (capped at maxRpcIPv6Addresses), then adaptively drops trailing
// addresses until the marshalled message fits rpcStateMaxBytes — so the richest payload
// that fits is sent, regardless of DHCP-lease size or address lengths.
func (s *InterfaceState) ToRpcInterfaceState() *RpcInterfaceState {
	addrs := make([]RpcIPv6Address, 0, maxRpcIPv6Addresses)
	for _, addr := range s.IPv6Addresses {
		// Skip link-local/host/nowhere scopes — they're noise here and link-local is
		// surfaced separately. RT_SCOPE_LINK (253) and above are non-routable.
		if addr.Scope >= unix.RT_SCOPE_LINK {
			continue
		}
		if len(addrs) >= maxRpcIPv6Addresses {
			break
		}
		addrs = append(addrs, RpcIPv6Address{
			Address:           addr.Address.String(),
			Prefix:            prefixNetwork(addr.Prefix),
			ValidLifetime:     secondsRemaining(addr.ValidLifetime),
			PreferredLifetime: secondsRemaining(addr.PreferredLifetime),
			FlagDeprecated:    addr.Flags&unix.IFA_F_DEPRECATED != 0,
			FlagDADFailed:     addr.Flags&unix.IFA_F_DADFAILED != 0,
		})
	}

	rpc := &RpcInterfaceState{
		InterfaceName: s.InterfaceName,
		Hostname:      s.Hostname,
		MACAddress:    s.MACAddress,
		Up:            s.Up,
		Online:        s.Online,
		IPv4Address:   s.IPv4Address,
		IPv6Address:   s.IPv6Address,
		IPv6LinkLocal: s.IPv6LinkLocal,
		IPv6Gateway:   s.IPv6Gateway,
		IPv4Addresses: s.IPv4Addresses,
		IPv6Addresses: addrs,
		DHCPLease:     toRpcDHCPLease(s.DHCPLease4),
	}

	// Drop trailing IPv6 addresses until the message fits the single-frame budget.
	for len(rpc.IPv6Addresses) > 0 {
		encoded, err := json.Marshal(rpc)
		if err != nil || len(encoded) <= rpcStateMaxBytes {
			break
		}
		rpc.IPv6Addresses = rpc.IPv6Addresses[:len(rpc.IPv6Addresses)-1]
	}

	return rpc
}
