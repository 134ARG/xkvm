//go:build linux

package network

import (
	"net"
	"time"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
)

func (s *NetworkInterfaceState) HandleLinkUpdate(update netlink.LinkUpdate) {
	if update.Link.Attrs().Name == s.interfaceName {
		s.l.Info().Msg("interface link update received")
		_ = s.update()
	}
}

// Run starts monitoring the network interface (read-only observer)
func (s *NetworkInterfaceState) Run() error {
	updates := make(chan netlink.LinkUpdate)
	addrUpdates := make(chan netlink.AddrUpdate)
	done := make(chan struct{})

	// Subscribe to link updates
	if err := netlink.LinkSubscribe(updates, done); err != nil {
		s.l.Warn().Err(err).Msg("failed to subscribe to link updates")
		return err
	}

	// Subscribe to address updates
	if err := netlink.AddrSubscribe(addrUpdates, done); err != nil {
		s.l.Warn().Err(err).Msg("failed to subscribe to address updates")
		close(done)
		return err
	}

	// Initial state read
	if err := s.update(); err != nil {
		return err
	}

	s.l.Info().Str("interface", s.interfaceName).Msg("Started monitoring network interface")

	go func() {
		ticker := time.NewTicker(5 * time.Second) // Poll every 5s as fallback
		defer ticker.Stop()

		for {
			select {
			case update := <-updates:
				s.HandleLinkUpdate(update)
			case addrUpdate := <-addrUpdates:
				// Check if this update is for our interface
				link, _ := netlink.LinkByIndex(addrUpdate.LinkIndex)
				if link != nil && link.Attrs().Name == s.interfaceName {
					s.l.Debug().Msg("interface address update received")
					_ = s.update()
				}
			case <-ticker.C:
				// Periodic refresh as fallback
				_ = s.update()
			case <-s.stopChan:
				close(done)
				s.l.Info().Msg("Stopped monitoring network interface")
				return
			}
		}
	}()

	s.l.Log().Msg("exiting Run()...")
	return nil
}

// Stop stops monitoring the network interface
func (s *NetworkInterfaceState) Stop() {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()
	
	if !s.stopped {
		close(s.stopChan)
		s.stopped = true
	}
}

func netlinkAddrs(iface netlink.Link) ([]netlink.Addr, error) {
	return netlink.AddrList(iface, nl.FAMILY_ALL)
}

func readSystemHostname() (string, error) {
	hostname, err := net.LookupHost("localhost")
	if err != nil {
		// Fallback to reading /etc/hostname
		return "jetkvm", nil
	}
	if len(hostname) > 0 {
		return hostname[0], nil
	}
	return "jetkvm", nil
}
