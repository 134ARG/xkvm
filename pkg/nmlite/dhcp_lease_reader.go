package nmlite

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/134ARG/xkvm/internal/network/types"
)

// DHCPLeaseReader reads DHCP lease information from system DHCP clients
type DHCPLeaseReader struct {
	interfaceName string
}

// NewDHCPLeaseReader creates a new DHCP lease reader
func NewDHCPLeaseReader(interfaceName string) *DHCPLeaseReader {
	return &DHCPLeaseReader{
		interfaceName: interfaceName,
	}
}

// ReadSystemDHCPLease attempts to read DHCP lease from various system DHCP clients
func (r *DHCPLeaseReader) ReadSystemDHCPLease() *types.DHCPLease {
	// Try different DHCP client lease file locations in order of preference
	readers := []func() *types.DHCPLease{
		r.readNetworkManagerViaNmcli, // Try nmcli first (most reliable for NetworkManager)
		r.readNetworkManagerLease,
		r.readDhclientLease,
		r.readSystemdNetworkdLease,
		r.readUdhcpcLease,
	}

	for _, reader := range readers {
		if lease := reader(); lease != nil {
			return lease
		}
	}

	return nil
}

// readNetworkManagerViaNmcli reads lease from NetworkManager using nmcli
func (r *DHCPLeaseReader) readNetworkManagerViaNmcli() *types.DHCPLease {
	// Check if nmcli is available
	if _, err := exec.LookPath("nmcli"); err != nil {
		return nil
	}

	// Run nmcli to get DHCP information
	cmd := exec.Command("nmcli", "-t", "-f", "IP4.ADDRESS,IP4.GATEWAY,IP4.DNS,DHCP4.OPTION", "device", "show", r.interfaceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	return r.parseNetworkManagerNmcliOutput(output)
}

// parseNetworkManagerNmcliOutput parses nmcli output
func (r *DHCPLeaseReader) parseNetworkManagerNmcliOutput(output []byte) *types.DHCPLease {
	lines := strings.Split(string(output), "\n")
	lease := &types.DHCPLease{
		InterfaceName: r.interfaceName,
		DHCPClient:    "NetworkManager",
	}

	dhcpOptions := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Parse IP4 fields
		if strings.HasPrefix(key, "IP4.ADDRESS") {
			// Format: 192.168.1.109/24
			if strings.Contains(value, "/") {
				ipParts := strings.Split(value, "/")
				if ip := net.ParseIP(ipParts[0]); ip != nil {
					lease.IPAddress = ip
					// Convert CIDR to netmask
					if len(ipParts) > 1 {
						if prefixLen, err := strconv.Atoi(ipParts[1]); err == nil {
							mask := net.CIDRMask(prefixLen, 32)
							lease.Netmask = net.IPv4(mask[0], mask[1], mask[2], mask[3])
						}
					}
				}
			}
		} else if key == "IP4.GATEWAY" {
			if ip := net.ParseIP(value); ip != nil {
				lease.Routers = append(lease.Routers, ip)
			}
		} else if strings.HasPrefix(key, "IP4.DNS") {
			if ip := net.ParseIP(value); ip != nil {
				lease.DNS = append(lease.DNS, ip)
			}
		} else if strings.HasPrefix(key, "DHCP4.OPTION") {
			// Parse DHCP option: key = value
			if strings.Contains(value, "=") {
				optParts := strings.SplitN(value, "=", 2)
				if len(optParts) == 2 {
					optKey := strings.TrimSpace(optParts[0])
					optValue := strings.TrimSpace(optParts[1])
					dhcpOptions[optKey] = optValue
				}
			}
		}
	}

	// Parse DHCP options
	if serverID, ok := dhcpOptions["dhcp_server_identifier"]; ok {
		lease.ServerID = serverID
	}
	if domain, ok := dhcpOptions["domain_name"]; ok {
		lease.Domain = strings.Trim(domain, "\"")
	}
	if expiry, ok := dhcpOptions["expiry"]; ok {
		// expiry is a Unix timestamp
		if timestamp, err := strconv.ParseInt(expiry, 10, 64); err == nil {
			expiryTime := time.Unix(timestamp, 0)
			lease.LeaseExpiry = &expiryTime
		}
	}
	if leaseTime, ok := dhcpOptions["dhcp_lease_time"]; ok {
		if seconds, err := strconv.ParseInt(leaseTime, 10, 64); err == nil {
			lease.LeaseTime = time.Duration(seconds) * time.Second
		}
	}
	if broadcast, ok := dhcpOptions["broadcast_address"]; ok {
		if ip := net.ParseIP(broadcast); ip != nil {
			lease.Broadcast = ip
		}
	}

	// Only return if we have at least an IP address
	if lease.IPAddress != nil {
		return lease
	}

	return nil
}

// readNetworkManagerLease reads lease from NetworkManager
func (r *DHCPLeaseReader) readNetworkManagerLease() *types.DHCPLease {
	// NetworkManager stores leases in /var/lib/NetworkManager/
	leasePath := fmt.Sprintf("/var/lib/NetworkManager/dhclient-%s.lease", r.interfaceName)
	if _, err := os.Stat(leasePath); err == nil {
		return r.parseDhclientLeaseFile(leasePath)
	}

	// Also try internal-* format
	leasePath = fmt.Sprintf("/var/lib/NetworkManager/internal-%s.lease", r.interfaceName)
	if _, err := os.Stat(leasePath); err == nil {
		return r.parseDhclientLeaseFile(leasePath)
	}

	return nil
}

// readDhclientLease reads lease from dhclient
func (r *DHCPLeaseReader) readDhclientLease() *types.DHCPLease {
	// Common dhclient lease file locations
	paths := []string{
		fmt.Sprintf("/var/lib/dhcp/dhclient.%s.leases", r.interfaceName),
		fmt.Sprintf("/var/lib/dhclient/dhclient-%s.leases", r.interfaceName),
		fmt.Sprintf("/var/lib/dhcp/dhclient-%s.leases", r.interfaceName),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return r.parseDhclientLeaseFile(path)
		}
	}

	return nil
}

// readSystemdNetworkdLease reads lease from systemd-networkd
func (r *DHCPLeaseReader) readSystemdNetworkdLease() *types.DHCPLease {
	// systemd-networkd stores leases in /run/systemd/netif/leases/
	leaseDir := "/run/systemd/netif/leases"

	// Find the lease file for this interface
	entries, err := os.ReadDir(leaseDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Read the file to check if it's for our interface
		content, err := os.ReadFile(filepath.Join(leaseDir, entry.Name()))
		if err != nil {
			continue
		}

		// Check if this lease is for our interface
		if strings.Contains(string(content), fmt.Sprintf("IFNAME=%s", r.interfaceName)) {
			return r.parseSystemdNetworkdLease(content)
		}
	}

	return nil
}

// readUdhcpcLease reads lease from udhcpc
func (r *DHCPLeaseReader) readUdhcpcLease() *types.DHCPLease {
	leasePath := fmt.Sprintf("/run/udhcpc.%s.info", r.interfaceName)
	if _, err := os.Stat(leasePath); err != nil {
		return nil
	}

	content, err := os.ReadFile(leasePath)
	if err != nil {
		return nil
	}

	return r.parseUdhcpcLease(content)
}

// parseDhclientLeaseFile parses a dhclient-style lease file
func (r *DHCPLeaseReader) parseDhclientLeaseFile(path string) *types.DHCPLease {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	// Parse dhclient lease format (ISC DHCP format)
	// This is a simplified parser - dhclient format is complex
	lines := strings.Split(string(content), "\n")
	lease := &types.DHCPLease{
		InterfaceName: r.interfaceName,
		DHCPClient:    "dhclient",
	}

	var inLease bool
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "lease {") {
			inLease = true
			continue
		}
		if strings.HasPrefix(line, "}") {
			inLease = false
			continue
		}
		if !inLease {
			continue
		}

		// Parse lease fields
		if strings.HasPrefix(line, "fixed-address ") {
			addr := strings.TrimSuffix(strings.TrimPrefix(line, "fixed-address "), ";")
			lease.IPAddress = net.ParseIP(addr)
		} else if strings.HasPrefix(line, "option subnet-mask ") {
			mask := strings.TrimSuffix(strings.TrimPrefix(line, "option subnet-mask "), ";")
			lease.Netmask = net.ParseIP(mask)
		} else if strings.HasPrefix(line, "option routers ") {
			routers := strings.TrimSuffix(strings.TrimPrefix(line, "option routers "), ";")
			for _, router := range strings.Split(routers, ",") {
				if ip := net.ParseIP(strings.TrimSpace(router)); ip != nil {
					lease.Routers = append(lease.Routers, ip)
				}
			}
		} else if strings.HasPrefix(line, "option domain-name-servers ") {
			dns := strings.TrimSuffix(strings.TrimPrefix(line, "option domain-name-servers "), ";")
			for _, server := range strings.Split(dns, ",") {
				if ip := net.ParseIP(strings.TrimSpace(server)); ip != nil {
					lease.DNS = append(lease.DNS, ip)
				}
			}
		} else if strings.HasPrefix(line, "option domain-name ") {
			domain := strings.Trim(strings.TrimSuffix(strings.TrimPrefix(line, "option domain-name "), ";"), "\"")
			lease.Domain = domain
		} else if strings.HasPrefix(line, "option dhcp-server-identifier ") {
			serverID := strings.TrimSuffix(strings.TrimPrefix(line, "option dhcp-server-identifier "), ";")
			lease.ServerID = serverID
		} else if strings.HasPrefix(line, "option broadcast-address ") {
			broadcast := strings.TrimSuffix(strings.TrimPrefix(line, "option broadcast-address "), ";")
			lease.Broadcast = net.ParseIP(broadcast)
		} else if strings.HasPrefix(line, "expire ") {
			// Parse expire time: expire 3 2026/01/15 12:34:56;
			expireStr := strings.TrimSuffix(strings.TrimPrefix(line, "expire "), ";")
			parts := strings.Fields(expireStr)
			if len(parts) >= 3 {
				// Format: weekday YYYY/MM/DD HH:MM:SS
				timeStr := parts[1] + " " + parts[2]
				if t, err := time.Parse("2006/01/02 15:04:05", timeStr); err == nil {
					lease.LeaseExpiry = &t
				}
			}
		}
	}

	// Only return if we have at least an IP address
	if lease.IPAddress != nil {
		return lease
	}

	return nil
}

// parseSystemdNetworkdLease parses a systemd-networkd lease file
func (r *DHCPLeaseReader) parseSystemdNetworkdLease(content []byte) *types.DHCPLease {
	lines := strings.Split(string(content), "\n")
	lease := &types.DHCPLease{
		InterfaceName: r.interfaceName,
		DHCPClient:    "systemd-networkd",
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "ADDRESS":
			if ip := net.ParseIP(value); ip != nil {
				lease.IPAddress = ip
			}
		case "NETMASK":
			if mask := net.ParseIP(value); mask != nil {
				lease.Netmask = mask
			}
		case "ROUTER":
			if ip := net.ParseIP(value); ip != nil {
				lease.Routers = append(lease.Routers, ip)
			}
		case "DNS":
			for _, dns := range strings.Fields(value) {
				if ip := net.ParseIP(dns); ip != nil {
					lease.DNS = append(lease.DNS, ip)
				}
			}
		case "DOMAIN":
			lease.Domain = value
		case "SERVER_ADDRESS":
			lease.ServerID = value
		case "BROADCAST":
			if ip := net.ParseIP(value); ip != nil {
				lease.Broadcast = ip
			}
		case "T1":
			// Renewal time in seconds
			// Not directly used in our structure
		case "T2":
			// Rebinding time in seconds
			// Not directly used in our structure
		}
	}

	// Only return if we have at least an IP address
	if lease.IPAddress != nil {
		return lease
	}

	return nil
}

// parseUdhcpcLease parses a udhcpc lease file
func (r *DHCPLeaseReader) parseUdhcpcLease(content []byte) *types.DHCPLease {
	// udhcpc stores lease as key=value pairs
	var leaseData map[string]string
	if err := json.Unmarshal(content, &leaseData); err != nil {
		// Try parsing as simple key=value format
		leaseData = make(map[string]string)
		for _, line := range strings.Split(string(content), "\n") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				leaseData[parts[0]] = parts[1]
			}
		}
	}

	lease := &types.DHCPLease{
		InterfaceName: r.interfaceName,
		DHCPClient:    "udhcpc",
	}

	if ip, ok := leaseData["ip"]; ok {
		lease.IPAddress = net.ParseIP(ip)
	}
	if subnet, ok := leaseData["subnet"]; ok {
		lease.Netmask = net.ParseIP(subnet)
	}
	if router, ok := leaseData["router"]; ok {
		if ip := net.ParseIP(router); ip != nil {
			lease.Routers = append(lease.Routers, ip)
		}
	}
	if dns, ok := leaseData["dns"]; ok {
		for _, server := range strings.Fields(dns) {
			if ip := net.ParseIP(server); ip != nil {
				lease.DNS = append(lease.DNS, ip)
			}
		}
	}
	if domain, ok := leaseData["domain"]; ok {
		lease.Domain = domain
	}
	if serverid, ok := leaseData["serverid"]; ok {
		lease.ServerID = serverid
	}
	if broadcast, ok := leaseData["broadcast"]; ok {
		lease.Broadcast = net.ParseIP(broadcast)
	}

	// Only return if we have at least an IP address
	if lease.IPAddress != nil {
		return lease
	}

	return nil
}
