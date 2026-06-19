package usbgadget

import "time"

const dwc3Path = "/sys/bus/platform/drivers/dwc3"

const hidWriteTimeout = 10 * time.Millisecond
const keyboardHidWriteTimeout = 100 * time.Millisecond

const (
	absMouseHidPath = "/dev/hidg1"
	relMouseHidPath = "/dev/hidg2"
)
