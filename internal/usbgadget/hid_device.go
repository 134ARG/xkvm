package usbgadget

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// hidWriteLocked is the shared low-level HID report write path used by the mouse
// devices. The caller must hold the relevant device lock.
//
// It lazily opens the device at devicePath (O_RDWR) into *file, writes data with
// the given timeout, and manages the handle on failure. A write deadline timeout
// is treated as a dropped report (the host endpoint is full) and returns nil; on
// any other error the handle is closed so the next call reopens it.
//
// HID operations suspended during USB reconfiguration are skipped here, under the
// device lock, so a late report cannot reopen a HID file that the reconfigure
// path just closed.
func (u *UsbGadget) hidWriteLocked(file **os.File, devicePath string, data []byte, timeout time.Duration, counterName string) error {
	if u.IsHidSuspended() {
		return nil
	}

	if *file == nil {
		f, err := os.OpenFile(devicePath, os.O_RDWR, 0666)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", devicePath, err)
		}
		*file = f
	}

	if _, err := u.writeWithTimeoutDuration(*file, data, timeout); err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return nil
		}
		(*file).Close()
		*file = nil
		u.logWithSuppression(counterName, 100, u.log, err, "failed to write to %s", devicePath)
		return err
	}
	u.resetLogSuppressionCounter(counterName)
	return nil
}
