package kvm

import (
	"time"

	"github.com/pion/webrtc/v4"
	"go.bug.st/serial"
)

const serialPortPath = "/dev/ttyS3"

var port serial.Port

func mountATXControl() error {
	// ATX control disabled - stub implementation for future custom logic
	serialLogger.Info().Msg("ATX control mounted (stub implementation)")
	return nil
}

func unmountATXControl() error {
	// ATX control disabled - stub implementation
	serialLogger.Info().Msg("ATX control unmounted (stub implementation)")
	return nil
}

var (
	ledHDDState bool
	ledPWRState bool
	btnRSTState bool
	btnPWRState bool
)

func runATXControl() {
	// ATX control disabled - stub implementation for future custom logic
	scopedLogger := serialLogger.With().Str("service", "atx_control").Logger()
	scopedLogger.Info().Msg("ATX control service started (stub implementation)")

	// Future custom ATX control logic can be implemented here
	// For now, just maintain default states
	ledHDDState = false
	ledPWRState = false
	btnRSTState = false
	btnPWRState = false
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

func mountDCControl() error {
	// DC control disabled - stub implementation for future custom logic
	serialLogger.Info().Msg("DC control mounted (stub implementation)")
	return nil
}

func unmountDCControl() error {
	// DC control disabled - stub implementation
	serialLogger.Info().Msg("DC control unmounted (stub implementation)")
	return nil
}

var dcState DCPowerState

func runDCControl() {
	// DC control disabled - stub implementation for future custom logic
	scopedLogger := serialLogger.With().Str("service", "dc_control").Logger()
	scopedLogger.Info().Msg("DC control service started (stub implementation)")

	// Future custom DC control logic can be implemented here
	// For now, initialize with default state
	dcState = DCPowerState{
		IsOn:         false,
		Voltage:      0.0,
		Current:      0.0,
		Power:        0.0,
		RestoreState: -1, // Not supported
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

func reopenSerialPort() error {
	// Serial port control disabled - stub implementation
	serialLogger.Info().Msg("Serial port reopen requested (stub implementation)")
	return nil
}

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
