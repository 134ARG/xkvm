package usbgadget

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

func getUdcs() []string {
	var udcs []string

	files, err := os.ReadDir("/sys/devices/platform/usbdrd")
	if err != nil {
		return nil
	}

	for _, file := range files {
		if !file.IsDir() || !strings.HasSuffix(file.Name(), ".usb") {
			continue
		}
		udcs = append(udcs, file.Name())
	}

	return udcs
}

func rebindUsb(udc string, ignoreUnbindError bool) error {
	// Check if already bound before unbinding
	udcPath := path.Join(dwc3Path, udc)
	if _, err := os.Stat(udcPath); err == nil {
		// UDC is bound, unbind it
		err = os.WriteFile(path.Join(dwc3Path, "unbind"), []byte(udc), 0644)
		if err != nil && !ignoreUnbindError {
			return fmt.Errorf("failed to unbind UDC: %w", err)
		}
		// Give it a moment to unbind
		time.Sleep(100 * time.Millisecond)
	}

	// Now bind
	err := os.WriteFile(path.Join(dwc3Path, "bind"), []byte(udc), 0644)
	if err != nil {
		return fmt.Errorf("failed to bind UDC: %w", err)
	}

	// Wait for HID devices to become available after bind
	time.Sleep(500 * time.Millisecond)

	return nil
}

func (u *UsbGadget) rebindUsb(ignoreUnbindError bool) error {
	u.log.Info().Str("udc", u.udc).Msg("rebinding USB gadget to UDC")
	return rebindUsb(u.udc, ignoreUnbindError)
}

// RebindUsb rebinds the USB gadget to the UDC.
func (u *UsbGadget) RebindUsb(ignoreUnbindError bool) error {
	u.configLock.Lock()
	defer u.configLock.Unlock()

	return u.rebindUsb(ignoreUnbindError)
}

// GetUsbState returns the current state of the USB gadget
func (u *UsbGadget) GetUsbState() (state string) {
	stateFile := path.Join("/sys/class/udc", u.udc, "state")
	stateBytes, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "not attached"
		} else {
			u.log.Trace().Err(err).Msg("failed to read usb state")
		}
		return "unknown"
	}
	return strings.TrimSpace(string(stateBytes))
}

// IsUDCBound checks if the UDC state is bound.
func (u *UsbGadget) IsUDCBound() (bool, error) {
	udcFilePath := path.Join(dwc3Path, u.udc)
	_, err := os.Stat(udcFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("error checking USB emulation state: %w", err)
	}
	return true, nil
}

// BindUDC binds the gadget to the UDC.
func (u *UsbGadget) BindUDC() error {
	// Check if already bound
	bound, err := u.IsUDCBound()
	if err != nil {
		return fmt.Errorf("error checking UDC bind state: %w", err)
	}
	if bound {
		u.log.Debug().Msg("UDC already bound, skipping")
		return nil
	}

	err = os.WriteFile(path.Join(dwc3Path, "bind"), []byte(u.udc), 0644)
	if err != nil {
		return fmt.Errorf("error binding UDC: %w", err)
	}
	return nil
}

// UnbindUDC unbinds the gadget from the UDC.
func (u *UsbGadget) UnbindUDC() error {
	// Check if bound before unbinding
	bound, err := u.IsUDCBound()
	if err != nil {
		return fmt.Errorf("error checking UDC bind state: %w", err)
	}
	if !bound {
		u.log.Debug().Msg("UDC not bound, skipping unbind")
		return nil
	}

	err = os.WriteFile(path.Join(dwc3Path, "unbind"), []byte(u.udc), 0644)
	if err != nil {
		return fmt.Errorf("error unbinding UDC: %w", err)
	}
	return nil
}
