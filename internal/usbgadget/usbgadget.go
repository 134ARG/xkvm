// Package usbgadget provides a high-level interface to manage USB gadgets
// THIS PACKAGE IS FOR INTERNAL USE ONLY AND ITS API MAY CHANGE WITHOUT NOTICE
package usbgadget

import (
	"context"
	"os"
	"path"
	"sync"
	"time"

	"github.com/134ARG/xkvm/internal/logging"
	"github.com/rs/zerolog"
)

// Devices is a struct that represents the USB devices that can be enabled on a USB gadget.
type Devices struct {
	AbsoluteMouse bool `json:"absolute_mouse"`
	RelativeMouse bool `json:"relative_mouse"`
	Keyboard      bool `json:"keyboard"`
	MassStorage   bool `json:"mass_storage"`
}

// Config is a struct that represents the customizations for a USB gadget.
// TODO: rename to something else that won't confuse with the USB gadget configuration
type Config struct {
	VendorId     string `json:"vendor_id"`
	ProductId    string `json:"product_id"`
	SerialNumber string `json:"serial_number"`
	Manufacturer string `json:"manufacturer"`
	Product      string `json:"product"`

	strictMode bool // when it's enabled, all warnings will be converted to errors
	isEmpty    bool
}

var defaultUsbGadgetDevices = Devices{
	AbsoluteMouse: true,
	RelativeMouse: true,
	Keyboard:      true,
	MassStorage:   true,
}

type KeysDownState struct {
	Modifier byte      `json:"modifier"`
	Keys     ByteSlice `json:"keys"`
}

// UsbGadget is a struct that represents a USB gadget.
type UsbGadget struct {
	name          string
	udc           string
	kvmGadgetPath string
	configC1Path  string

	configMap    map[string]gadgetConfigItem
	customConfig Config

	configLock    sync.Mutex
	lifecycleLock sync.Mutex

	keyboardHidFile *os.File
	keyboardLock    sync.Mutex
	absMouseHidFile *os.File
	absMouseLock    sync.Mutex
	relMouseHidFile *os.File
	relMouseLock    sync.Mutex

	keyboardState byte          // keyboard latched state (NumLock, CapsLock, ScrollLock, Compose, Kana)
	keysDownState KeysDownState // keyboard dynamic state (modifier keys and pressed keys)

	kbdAutoReleaseLock   sync.Mutex
	kbdAutoReleaseTimers map[byte]*time.Timer

	keyboardStateLock   sync.Mutex
	keyboardStateCtx    context.Context
	keyboardStateCancel context.CancelFunc

	enabledDevices Devices

	strictMode bool // only intended for testing for now

	absMouseAccumulatedWheelY float64

	lastUserInput time.Time

	tx     *UsbGadgetTransaction
	txLock sync.Mutex

	onKeyboardStateChange *func(state KeyboardState)
	onKeysDownChange      *func(state KeysDownState)
	onKeepAliveReset      *func()

	log *zerolog.Logger

	logSuppressionCounter map[string]int
	logSuppressionLock    sync.Mutex

	hidSuspended     bool
	hidSuspendedLock sync.RWMutex
}

const configFSPath = "/sys/kernel/config"
const gadgetPath = "/sys/kernel/config/usb_gadget"

var defaultLogger = logging.GetSubsystemLogger("usbgadget")

// NewUsbGadget creates a new UsbGadget.
func NewUsbGadget(name string, enabledDevices *Devices, config *Config, logger *zerolog.Logger) (*UsbGadget, error) {
	return newUsbGadget(name, defaultGadgetConfig, enabledDevices, config, logger)
}

func newUsbGadget(name string, configMap map[string]gadgetConfigItem, enabledDevices *Devices, config *Config, logger *zerolog.Logger) (*UsbGadget, error) {
	if logger == nil {
		logger = defaultLogger
	}

	if enabledDevices == nil {
		enabledDevices = &defaultUsbGadgetDevices
	}

	if config == nil {
		config = &Config{isEmpty: true}
	}

	keyboardCtx, keyboardCancel := context.WithCancel(context.Background())

	g := &UsbGadget{
		name:                 name,
		kvmGadgetPath:        path.Join(gadgetPath, name),
		configC1Path:         path.Join(gadgetPath, name, "configs/c.1"),
		configMap:            configMap,
		customConfig:         *config,
		configLock:           sync.Mutex{},
		lifecycleLock:        sync.Mutex{},
		keyboardLock:         sync.Mutex{},
		absMouseLock:         sync.Mutex{},
		relMouseLock:         sync.Mutex{},
		txLock:               sync.Mutex{},
		keyboardStateCtx:     keyboardCtx,
		keyboardStateCancel:  keyboardCancel,
		keyboardState:        0,
		keysDownState:        KeysDownState{Modifier: 0, Keys: []byte{0, 0, 0, 0, 0, 0}}, // must be initialized to hidKeyBufferSize (6) zero bytes
		kbdAutoReleaseTimers: make(map[byte]*time.Timer),
		enabledDevices:       *enabledDevices,
		lastUserInput:        time.Now(),
		log:                  logger,

		strictMode: config.strictMode,

		logSuppressionCounter: make(map[string]int),

		absMouseAccumulatedWheelY: 0,
	}
	if err := g.Init(); err != nil {
		logger.Error().Err(err).Msg("failed to init USB gadget")
		return nil, err
	}

	return g, nil
}

// Close cleans up resources used by the USB gadget
func (u *UsbGadget) Close() error {
	// Cancel keyboard state context
	if u.keyboardStateCancel != nil {
		u.keyboardStateCancel()
	}

	// Stop auto-release timer
	u.kbdAutoReleaseLock.Lock()
	for _, timer := range u.kbdAutoReleaseTimers {
		if timer != nil {
			timer.Stop()
		}
	}
	u.kbdAutoReleaseTimers = make(map[byte]*time.Timer)
	u.kbdAutoReleaseLock.Unlock()

	// Close HID files
	u.CloseHidFiles()

	return nil
}

// CloseHidFiles closes all open HID device files
func (u *UsbGadget) CloseHidFiles() {
	if u.keyboardStateCancel != nil {
		u.keyboardStateCancel()
	}

	u.keyboardLock.Lock()
	if u.keyboardHidFile != nil {
		u.keyboardHidFile.Close()
		u.keyboardHidFile = nil
		u.log.Debug().Msg("closed keyboard HID file")
	}
	u.keyboardLock.Unlock()

	u.absMouseLock.Lock()
	if u.absMouseHidFile != nil {
		u.absMouseHidFile.Close()
		u.absMouseHidFile = nil
		u.log.Debug().Msg("closed absolute mouse HID file")
	}
	u.absMouseLock.Unlock()

	u.relMouseLock.Lock()
	if u.relMouseHidFile != nil {
		u.relMouseHidFile.Close()
		u.relMouseHidFile = nil
		u.log.Debug().Msg("closed relative mouse HID file")
	}
	u.relMouseLock.Unlock()
}

func (u *UsbGadget) cancelAutoReleaseTimers() {
	u.kbdAutoReleaseLock.Lock()
	for key, timer := range u.kbdAutoReleaseTimers {
		if timer != nil {
			timer.Stop()
		}
		delete(u.kbdAutoReleaseTimers, key)
	}
	u.kbdAutoReleaseLock.Unlock()
}

func (u *UsbGadget) clearKeysDownState() {
	clearKeys := make([]byte, hidKeyBufferSize)

	u.keyboardStateLock.Lock()
	changed := u.keysDownState.Modifier != 0
	if !changed {
		for _, key := range u.keysDownState.Keys {
			if key != 0 {
				changed = true
				break
			}
		}
	}
	u.keysDownState = KeysDownState{Modifier: 0, Keys: clearKeys}
	u.keyboardStateLock.Unlock()

	if changed && u.onKeysDownChange != nil {
		(*u.onKeysDownChange)(u.GetKeysDownState())
	}
}

func (u *UsbGadget) releaseHidStateBeforeClose() {
	if u.enabledDevices.Keyboard {
		u.keyboardLock.Lock()
		if u.keyboardHidFile != nil {
			clearKeys := make([]byte, hidKeyBufferSize)
			if _, err := u.writeWithTimeout(u.keyboardHidFile, append([]byte{0, 0}, clearKeys...)); err != nil {
				u.log.Warn().Err(err).Msg("failed to release keyboard state before HID close")
			}
		}
		u.keyboardLock.Unlock()
	}

	if u.enabledDevices.AbsoluteMouse {
		u.absMouseLock.Lock()
		if u.absMouseHidFile != nil {
			if _, err := u.writeWithTimeout(u.absMouseHidFile, []byte{1, 0, 0, 0, 0, 0}); err != nil {
				u.log.Warn().Err(err).Msg("failed to release absolute mouse buttons before HID close")
			}
		}
		u.absMouseLock.Unlock()
	}

	if u.enabledDevices.RelativeMouse {
		u.relMouseLock.Lock()
		if u.relMouseHidFile != nil {
			if _, err := u.writeWithTimeout(u.relMouseHidFile, []byte{0, 0, 0, 0}); err != nil {
				u.log.Warn().Err(err).Msg("failed to release relative mouse buttons before HID close")
			}
		}
		u.relMouseLock.Unlock()
	}
}

func (u *UsbGadget) prepareHidForReconfigure() {
	u.SuspendHidOperations()
	u.cancelAutoReleaseTimers()
	u.releaseHidStateBeforeClose()
	u.clearKeysDownState()
	u.CloseHidFiles()
}

// SuspendHidOperations suspends HID operations during USB reconfiguration
func (u *UsbGadget) SuspendHidOperations() {
	u.hidSuspendedLock.Lock()
	defer u.hidSuspendedLock.Unlock()
	u.hidSuspended = true
	u.log.Debug().Msg("HID operations suspended")
}

// ResumeHidOperations resumes HID operations after USB reconfiguration
func (u *UsbGadget) ResumeHidOperations() {
	u.hidSuspendedLock.Lock()
	defer u.hidSuspendedLock.Unlock()
	u.hidSuspended = false
	u.log.Debug().Msg("HID operations resumed")
}

// IsHidSuspended checks if HID operations are currently suspended
func (u *UsbGadget) IsHidSuspended() bool {
	u.hidSuspendedLock.RLock()
	defer u.hidSuspendedLock.RUnlock()
	return u.hidSuspended
}
