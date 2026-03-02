package kvm

import (
	"time"

	"github.com/pion/webrtc/v4"
	"go.bug.st/serial"
)

// const serialPortPath = "/dev/ttyS3"

var port serial.Port

var (
	ledHDDState bool
	ledPWRState bool
	// btnRSTState bool
	// btnPWRState bool
	atxStopChan chan struct{}
)

func mountATXControl() error {
	serialLogger.Info().Msg("ATX control mounting")
	atxStopChan = make(chan struct{})
	go runATXControl()
	return nil
}

func unmountATXControl() error {
	serialLogger.Info().Msg("ATX control unmounting")
	if atxStopChan != nil {
		close(atxStopChan)
		atxStopChan = nil
	}
	return nil
}

func runATXControl() {
	scopedLogger := serialLogger.With().Str("service", "atx_control").Logger()
	scopedLogger.Info().Msg("ATX control polling started")

	// Initialize default states
	ledHDDState = false
	ledPWRState = false
	// btnRSTState = false
	// btnPWRState = false

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-atxStopChan:
			scopedLogger.Info().Msg("ATX control polling stopped")
			return
		case <-ticker.C:
			// TODO: Read actual LED states from hardware
			// ledPWRState = readPWRLed()
			// ledHDDState = readHDDLed()
		}
	}
}

func pressATXPowerButton(duration time.Duration) error {
	// ATX power button control disabled - stub implementation for future custom logic
	serialLogger.Info().Dur("duration", duration).Msg("ATX power button press requested (stub implementation)")
	return nil
}

func pressATXResetButton(duration time.Duration) error {
	// ATX reset button control disabled - stub implementation for future custom logic
	serialLogger.Info().Dur("duration", duration).Msg("ATX reset button press requested (stub implementation)")
	return nil
}

var dcStopChan chan struct{}
var dcState DCPowerState

func mountDCControl() error {
	serialLogger.Info().Msg("DC control mounting")
	dcStopChan = make(chan struct{})
	go runDCControl()
	return nil
}

func unmountDCControl() error {
	serialLogger.Info().Msg("DC control unmounting")
	if dcStopChan != nil {
		close(dcStopChan)
		dcStopChan = nil
	}
	return nil
}

func runDCControl() {
	scopedLogger := serialLogger.With().Str("service", "dc_control").Logger()
	scopedLogger.Info().Msg("DC control polling started")

	// Initialize default state
	dcState = DCPowerState{
		IsOn:         false,
		Voltage:      0.0,
		Current:      0.0,
		Power:        0.0,
		RestoreState: -1, // Not supported
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-dcStopChan:
			scopedLogger.Info().Msg("DC control polling stopped")
			return
		case <-ticker.C:
			// TODO: Read actual DC state from hardware
			// dcState.IsOn = readDCPowerState()
			// dcState.Voltage = readDCVoltage()
			// dcState.Current = readDCCurrent()
			// dcState.Power = dcState.Voltage * dcState.Current
		}
	}
}

func setDCPowerState(on bool) error {
	// DC power control disabled - stub implementation for future custom logic
	serialLogger.Info().Bool("on", on).Msg("DC power state change requested (stub implementation)")
	return nil
}

func setDCRestoreState(state int) error {
	// DC restore state control disabled - stub implementation for future custom logic
	serialLogger.Info().Int("state", state).Msg("DC restore state change requested (stub implementation)")
	return nil
}

var defaultMode = &serial.Mode{
	BaudRate: 115200,
	DataBits: 8,
	Parity:   serial.NoParity,
	StopBits: serial.OneStopBit,
}

func initSerialPort() {
	// Serial port ATX/DC control disabled on this platform
	// Custom control logic can be added here in the future
	serialLogger.Info().Msg("Serial port control disabled - using stub implementation")

	switch config.ActiveExtension {
	case "atx-power":
		_ = mountATXControl()
	case "dc-power":
		_ = mountDCControl()
	}
}

// func reopenSerialPort() error {
// 	// Serial port control disabled - stub implementation
// 	serialLogger.Info().Msg("Serial port reopen requested (stub implementation)")
// 	return nil
// }

func handleSerialChannel(d *webrtc.DataChannel) {
	// Serial channel handling disabled - stub implementation for future custom logic
	scopedLogger := serialLogger.With().
		Uint16("data_channel_id", *d.ID()).Logger()

	scopedLogger.Info().Msg("Serial channel handling disabled (stub implementation)")

	d.OnOpen(func() {
		scopedLogger.Info().Msg("Serial channel opened (stub - no actual serial communication)")
	})

	d.OnMessage(func(msg webrtc.DataChannelMessage) {
		scopedLogger.Debug().Int("bytes", len(msg.Data)).Msg("Serial message received (stub - discarded)")
	})

	d.OnError(func(err error) {
		scopedLogger.Warn().Err(err).Msg("Serial channel error")
	})

	d.OnClose(func() {
		scopedLogger.Info().Msg("Serial channel closed")
	})
}
