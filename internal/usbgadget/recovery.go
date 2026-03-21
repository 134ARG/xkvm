package usbgadget

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// SoftReset performs a soft reset by cleaning up and reinitializing the gadget
func (u *UsbGadget) SoftReset() error {
	u.log.Warn().Msg("performing soft reset of USB gadget")

	// Close all HID files first
	u.CloseHidFiles()
	time.Sleep(200 * time.Millisecond)

	// Use existing cleanup function
	if err := u.cleanupStaleGadget(); err != nil {
		u.log.Warn().Err(err).Msg("cleanup failed during soft reset")
		return fmt.Errorf("soft reset cleanup failed: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Reinitialize using existing Init() method
	if err := u.Init(); err != nil {
		return fmt.Errorf("soft reset reinitialization failed: %w", err)
	}

	// Wait for HID devices to appear
	time.Sleep(1 * time.Second)

	// Verify HID devices exist
	if err := u.verifyHidDevices(); err != nil {
		return fmt.Errorf("HID devices not available after soft reset: %w", err)
	}

	u.log.Info().Msg("soft reset completed successfully")
	return nil
}

// HardReset performs a hard reset by reloading kernel modules
func (u *UsbGadget) HardReset() error {
	u.log.Warn().Msg("performing hard reset with kernel module reload")

	// Close all HID files
	u.CloseHidFiles()
	time.Sleep(200 * time.Millisecond)

	// Use existing cleanup function
	if err := u.cleanupStaleGadget(); err != nil {
		u.log.Warn().Err(err).Msg("cleanup failed during hard reset, continuing")
	}

	time.Sleep(500 * time.Millisecond)

	// Unload kernel modules
	u.log.Info().Msg("unloading USB gadget kernel modules")
	modules := []string{"usb_f_hid", "usb_f_mass_storage", "libcomposite"}
	for _, module := range modules {
		cmd := exec.Command("rmmod", module)
		if err := cmd.Run(); err != nil {
			u.log.Warn().Err(err).Str("module", module).Msg("failed to unload module, continuing")
		}
	}

	time.Sleep(1 * time.Second)

	// Reload kernel modules
	u.log.Info().Msg("loading USB gadget kernel modules")
	for i := len(modules) - 1; i >= 0; i-- {
		module := modules[i]
		cmd := exec.Command("modprobe", module)
		if err := cmd.Run(); err != nil {
			u.log.Error().Err(err).Str("module", module).Msg("failed to load module")
			return fmt.Errorf("failed to load module %s: %w", module, err)
		}
	}

	time.Sleep(1 * time.Second)

	// Reinitialize using existing Init() method
	if err := u.Init(); err != nil {
		return fmt.Errorf("hard reset reinitialization failed: %w", err)
	}

	// Wait longer for HID devices after module reload
	time.Sleep(2 * time.Second)

	// Verify HID devices exist
	if err := u.verifyHidDevices(); err != nil {
		return fmt.Errorf("HID devices not available after hard reset: %w", err)
	}

	u.log.Info().Msg("hard reset completed successfully")
	return nil
}

// verifyHidDevices checks if HID device nodes exist and are accessible
func (u *UsbGadget) verifyHidDevices() error {
	devices := []string{"/dev/hidg0", "/dev/hidg1", "/dev/hidg2"}

	for _, device := range devices {
		if _, err := os.Stat(device); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("HID device %s does not exist", device)
			}
			return fmt.Errorf("cannot access HID device %s: %w", device, err)
		}
	}

	u.log.Debug().Msg("all HID devices verified")
	return nil
}

// TryRecovery attempts recovery with escalating strategies
func (u *UsbGadget) TryRecovery() error {
	u.log.Warn().Msg("attempting USB gadget recovery")

	// Strategy 1: Soft reset (cleanup + reinit)
	u.log.Info().Msg("trying soft reset")
	if err := u.SoftReset(); err == nil {
		u.log.Info().Msg("soft reset successful")
		return nil
	} else {
		u.log.Warn().Err(err).Msg("soft reset failed")
	}

	// Strategy 2: Hard reset (module reload)
	u.log.Info().Msg("escalating to hard reset with kernel module reload")
	if err := u.HardReset(); err == nil {
		u.log.Info().Msg("hard reset successful")
		return nil
	} else {
		u.log.Error().Err(err).Msg("hard reset failed")
		return fmt.Errorf("all recovery strategies failed: %w", err)
	}
}
