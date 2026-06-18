package usbgadget

import "time"

func (u *UsbGadget) resetUserInputTime() {
	u.lastUserInputLock.Lock()
	u.lastUserInput = time.Now()
	u.lastUserInputLock.Unlock()
}

func (u *UsbGadget) GetLastUserInputTime() time.Time {
	u.lastUserInputLock.Lock()
	defer u.lastUserInputLock.Unlock()
	return u.lastUserInput
}
