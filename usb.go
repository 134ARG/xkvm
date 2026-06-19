package kvm

import (
	"fmt"
	"sync"
	"time"

	"github.com/134ARG/xkvm/internal/usbgadget"
)

var gadget *usbgadget.UsbGadget

// initUsbGadget initializes the USB gadget.
// call it only after the config is loaded.
func initUsbGadget() {
	var err error
	gadget, err = usbgadget.NewUsbGadget(
		"xkvm",
		config.UsbDevices,
		config.UsbConfig,
		usbLogger,
	)
	if err != nil {
		usbLogger.Error().Err(err).Msg("failed to initialize USB gadget")
		gadget = nil
		return
	}

	// Start health check monitoring

	go func() {
		for {
			checkUSBState()
			time.Sleep(500 * time.Millisecond)
		}
	}()

	gadget.SetOnKeyboardStateChange(func(state usbgadget.KeyboardState) {
		if cs := getCurrentSession(); cs != nil {
			cs.reportHidRPCKeyboardLedState(state)
		}
	})

	gadget.SetOnKeysDownChange(func(state usbgadget.KeysDownState) {
		if cs := getCurrentSession(); cs != nil {
			cs.enqueueKeysDownState(state)
		}
	})

	gadget.SetOnKeepAliveReset(func() {
		if cs := getCurrentSession(); cs != nil {
			cs.resetKeepAliveTime()
		}
	})

	gadget.StartKeyboardLedListener()
}

func rpcKeyboardReport(modifier byte, keys []byte) error {
	if gadget == nil || !gadget.IsInitialized() {
		return fmt.Errorf("USB gadget not initialized")
	}
	return gadget.KeyboardReport(modifier, keys)
}

func rpcKeypressReport(key byte, press bool) error {
	if gadget == nil || !gadget.IsInitialized() {
		return fmt.Errorf("USB gadget not initialized")
	}
	return gadget.KeypressReport(key, press)
}

func rpcAbsMouseReport(x int, y int, buttons uint8) error {
	if gadget == nil || !gadget.IsInitialized() {
		return fmt.Errorf("USB gadget not initialized")
	}
	return gadget.AbsMouseReport(x, y, buttons)
}

func rpcRelMouseReport(dx int8, dy int8, buttons uint8) error {
	if gadget == nil || !gadget.IsInitialized() {
		return fmt.Errorf("USB gadget not initialized")
	}
	return gadget.RelMouseReport(dx, dy, buttons)
}

func rpcWheelReport(wheelY int8) error {
	if gadget == nil || !gadget.IsInitialized() {
		return fmt.Errorf("USB gadget not initialized")
	}
	return gadget.AbsMouseWheelReport(wheelY)
}

func rpcGetKeyboardLedState() (state usbgadget.KeyboardState) {
	if gadget == nil || !gadget.IsInitialized() {
		return usbgadget.KeyboardState{}
	}
	return gadget.GetKeyboardState()
}

func rpcGetKeysDownState() (state usbgadget.KeysDownState) {
	if gadget == nil || !gadget.IsInitialized() {
		return usbgadget.KeysDownState{}
	}
	return gadget.GetKeysDownState()
}

var (
	usbState     = "unknown"
	usbStateLock sync.Mutex
)

func rpcGetUSBState() (state string) {
	if gadget == nil || !gadget.IsInitialized() {
		return "not initialized"
	}
	return gadget.GetUsbState()
}

// currentUSBState returns the last observed USB state under the lock.
func currentUSBState() string {
	usbStateLock.Lock()
	defer usbStateLock.Unlock()
	return usbState
}

func triggerUSBStateUpdate(state string) {
	go func() {
		cs := getCurrentSession()
		if cs == nil {
			usbLogger.Info().Msg("No active RPC session, skipping USB state update")
			return
		}
		writeJSONRPCEvent("usbState", state, cs)
	}()
}

func checkUSBState() {
	if gadget == nil || !gadget.IsInitialized() {
		return
	}

	usbStateLock.Lock()
	defer usbStateLock.Unlock()

	newState := gadget.GetUsbState()

	usbLogger.Trace().Str("old", usbState).Str("new", newState).Msg("Checking USB state")

	if newState == usbState {
		return
	}

	usbLogger.Info().Str("from", usbState).Str("to", newState).Msg("USB state changed")
	usbState = newState

	triggerUSBStateUpdate(newState)
}
