package link

import (
	"github.com/134ARG/xkvm/internal/sync"

	"github.com/rs/zerolog"
	"github.com/vishvananda/netlink"
)

// StateChangeHandler is the function type for link state callbacks
type StateChangeHandler func(link *Link)

// StateChangeCallback is the struct for link state callbacks
type StateChangeCallback struct {
	Async bool
	Func  StateChangeHandler
}

// NetlinkManager provides centralized netlink operations
type NetlinkManager struct {
	logger               *zerolog.Logger
	mu                   sync.RWMutex
	stateChangeCallbacks map[string][]StateChangeCallback
}

func newNetlinkManager(logger *zerolog.Logger) *NetlinkManager {
	if logger == nil {
		logger = &zerolog.Logger{} // Default no-op logger
	}
	n := &NetlinkManager{
		logger:               logger,
		stateChangeCallbacks: make(map[string][]StateChangeCallback),
	}
	n.monitorStateChange()
	return n
}

// GetNetlinkManager returns the singleton NetlinkManager instance
func GetNetlinkManager() *NetlinkManager {
	netlinkManagerOnce.Do(func() {
		netlinkManagerInstance = newNetlinkManager(nil)
	})
	return netlinkManagerInstance
}

// InitializeNetlinkManager initializes the singleton NetlinkManager with a logger
func InitializeNetlinkManager(logger *zerolog.Logger) *NetlinkManager {
	netlinkManagerOnce.Do(func() {
		netlinkManagerInstance = newNetlinkManager(logger)
	})
	return netlinkManagerInstance
}

// AddStateChangeCallback adds a callback for link state changes
func (nm *NetlinkManager) AddStateChangeCallback(ifname string, callback StateChangeCallback) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if _, ok := nm.stateChangeCallbacks[ifname]; !ok {
		nm.stateChangeCallbacks[ifname] = make([]StateChangeCallback, 0)
	}

	nm.stateChangeCallbacks[ifname] = append(nm.stateChangeCallbacks[ifname], callback)
}

// Interface operations
func (nm *NetlinkManager) monitorStateChange() {
	updateCh := make(chan netlink.LinkUpdate)
	// we don't need to stop the subscription, as it will be closed when the program exits
	stopCh := make(chan struct{}) //nolint:unused
	if err := netlink.LinkSubscribe(updateCh, stopCh); err != nil {
		nm.logger.Error().Err(err).Msg("failed to subscribe to link state changes")
	}

	nm.logger.Info().Msg("state change monitoring started")

	go func() {
		for update := range updateCh {
			nm.runCallbacks(update)
		}
	}()
}

func (nm *NetlinkManager) runCallbacks(update netlink.LinkUpdate) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	ifname := update.Link.Attrs().Name
	callbacks, ok := nm.stateChangeCallbacks[ifname]

	l := nm.logger.With().Str("interface", ifname).Logger()
	if !ok {
		l.Trace().Msg("no state change callbacks for interface")
		return
	}

	for _, callback := range callbacks {
		l.Trace().
			Interface("callback", callback).
			Bool("async", callback.Async).
			Msg("calling callback")

		if callback.Async {
			go callback.Func(&Link{Link: update.Link})
		} else {
			callback.Func(&Link{Link: update.Link})
		}
	}
}

// GetLinkByName gets a network link by name
func (nm *NetlinkManager) GetLinkByName(name string) (*Link, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	link, err := netlink.LinkByName(name)
	if err != nil {
		return nil, err
	}
	return &Link{Link: link}, nil
}

// Address operations (read-only)

// AddrList gets all addresses for a link
func (nm *NetlinkManager) AddrList(link *Link, family int) ([]netlink.Addr, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	return netlink.AddrList(link, family)
}

// Route operations (read-only)
func (nm *NetlinkManager) RouteList(link *Link, family int) ([]netlink.Route, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	return netlink.RouteList(link, family)
}

// ListDefaultRoutes lists the default routes for the given family
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

// HasDefaultRoute checks if a default route exists for the given family
func (nm *NetlinkManager) HasDefaultRoute(family int) bool {
	routes, err := nm.ListDefaultRoutes(family)
	if err != nil {
		return false
	}
	return len(routes) > 0
}
