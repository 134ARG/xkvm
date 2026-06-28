package types

import (
	"net"
	"slices"
	"time"

	"github.com/vishvananda/netlink"
)

// IPAddress represents a network interface address
type IPAddress struct {
	Family    int
	Address   net.IPNet
	Gateway   net.IP
	MTU       int
	Secondary bool
	Permanent bool
}

func (a *IPAddress) String() string {
	return a.Address.String()
}

func (a *IPAddress) Compare(n netlink.Addr) bool {
	if !a.Address.IP.Equal(n.IP) {
		return false
	}
	if slices.Compare(a.Address.Mask, n.Mask) != 0 {
		return false
	}
	return true
}

func (a *IPAddress) NetlinkAddr() netlink.Addr {
	return netlink.Addr{
		IPNet: &a.Address,
	}
}

func (a *IPAddress) DefaultRoute(linkIndex int) netlink.Route {
	return netlink.Route{
		Dst:       nil,
		Gw:        a.Gateway,
		LinkIndex: linkIndex,
	}
}

// ParsedIPConfig represents the parsed IP configuration
type ParsedIPConfig struct {
	Addresses   []IPAddress
	Nameservers []net.IP
	SearchList  []string
	Domain      string
	MTU         int
	Interface   string
}

// IPv6Address represents an IPv6 address with lifetime information
type IPv6Address struct {
	Address           net.IP     `json:"address"`
	Prefix            net.IPNet  `json:"prefix"`
	ValidLifetime     *time.Time `json:"valid_lifetime"`
	PreferredLifetime *time.Time `json:"preferred_lifetime"`
	Flags             int        `json:"flags"`
	Scope             int        `json:"scope"`
}

// RpcIPv6Address is the RPC representation of an IPv6 address. It carries only the
// fields the UI renders — keeping the getNetworkState payload under the WebRTC SCTP
// path-MTU (~1.2KB), which Firefox cannot reassemble when a message is fragmented.
type RpcIPv6Address struct {
	Address string `json:"address"`
	Prefix  string `json:"prefix"`
	// Lifetimes are seconds remaining (not absolute timestamps): the kernel reports them
	// that way, and an integer is far smaller on the wire than an RFC3339 string — keeping
	// the network-state payload within a single SCTP frame. The client renders them
	// relative to its own clock.
	ValidLifetime     *int64 `json:"valid_lifetime,omitempty"`
	PreferredLifetime *int64 `json:"preferred_lifetime,omitempty"`
	FlagDeprecated    bool   `json:"flag_deprecated"`
	FlagDADFailed     bool   `json:"flag_dad_failed"`
}
