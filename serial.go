package kvm

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/warthog618/go-gpiocdev"
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
	return pulseGPIO(config.GPIOPwrChip, config.GPIOPwrLine, duration)
}

func pressATXResetButton(duration time.Duration) error {
	return pulseGPIO(config.GPIORstChip, config.GPIORstLine, duration)
}

// pulseGPIO opens a GPIO line, drives it low to clear state, then pulses high
// for the given duration. No-op if chip/line is unconfigured.
func pulseGPIO(chip string, line int, duration time.Duration) error {
	if chip == "" || line < 0 {
		serialLogger.Debug().Msg("GPIO not configured, skipping pulse")
		return nil
	}

	l, err := gpiocdev.RequestLine(chip, line, gpiocdev.AsOutput(0))
	if err != nil {
		return fmt.Errorf("failed to request GPIO %s line %d: %w", chip, line, err)
	}
	defer l.Close()

	// Clear: drive low
	if err := l.SetValue(0); err != nil {
		return fmt.Errorf("failed to set GPIO low: %w", err)
	}
	time.Sleep(10 * time.Millisecond)

	// Pulse: drive high for duration
	if err := l.SetValue(1); err != nil {
		return fmt.Errorf("failed to set GPIO high: %w", err)
	}
	time.Sleep(duration)

	// Release: drive low
	if err := l.SetValue(0); err != nil {
		return fmt.Errorf("failed to set GPIO low after pulse: %w", err)
	}

	serialLogger.Info().Str("chip", chip).Int("line", line).Dur("duration", duration).Msg("GPIO pulse complete")
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

// HardwarePort represents a discovered serial port or GPIO chip.
type HardwarePort struct {
	Path string `json:"path"`
	Type string `json:"type"` // "serial" or "gpio"
	Info string `json:"info"` // human-readable label
}

// discoverSerialPorts returns all serial ports found on the system using
// go.bug.st/serial's built-in enumeration.
func discoverSerialPorts() []HardwarePort {
	ports, err := serial.GetPortsList()
	if err != nil {
		serialLogger.Warn().Err(err).Msg("failed to enumerate serial ports")
		return nil
	}
	var result []HardwarePort
	for _, p := range ports {
		result = append(result, HardwarePort{
			Path: p,
			Type: "serial",
			Info: classifySerialPort(p),
		})
	}
	return result
}

// classifySerialPort returns a human-readable label based on the port path.
func classifySerialPort(path string) string {
	switch {
	case strings.HasPrefix(path, "/dev/ttyUSB"):
		return "USB Serial Adapter"
	case strings.HasPrefix(path, "/dev/ttyACM"):
		return "USB CDC ACM"
	case strings.HasPrefix(path, "/dev/ttyAMA"):
		return "ARM UART"
	case strings.HasPrefix(path, "/dev/ttyS"):
		return "Hardware UART"
	default:
		return "Serial Port"
	}
}

// discoverGPIOChips returns all GPIO character devices found on the system.
// Uses chardev ioctl via go-gpiocdev for chip info (works without sysfs).
func discoverGPIOChips() []HardwarePort {
	chips, err := filepath.Glob("/dev/gpiochip*")
	if err != nil {
		serialLogger.Warn().Err(err).Msg("failed to glob GPIO chips")
		return nil
	}
	var result []HardwarePort
	for _, chipPath := range chips {
		name, label, lines := readGPIOChipInfo(chipPath)
		info := name
		if label != "" {
			info = label
		}
		info += fmt.Sprintf(" (%d lines)", lines)
		result = append(result, HardwarePort{
			Path: chipPath,
			Type: "gpio",
			Info: info,
		})
	}
	return result
}

// readGPIOChipInfo reads chip name, label, and line count via chardev ioctl.
func readGPIOChipInfo(chipPath string) (name string, label string, lines int) {
	c, err := gpiocdev.NewChip(chipPath)
	if err != nil {
		serialLogger.Warn().Err(err).Str("chip", chipPath).Msg("failed to open GPIO chip")
		return filepath.Base(chipPath), "", 0
	}
	defer c.Close()
	return c.Name, c.Label, c.Lines()
}

// getGPIOChipLineCount returns the number of lines for a given chip path.
func getGPIOChipLineCount(chipPath string) (int, error) {
	c, err := gpiocdev.NewChip(chipPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open GPIO chip %s: %w", chipPath, err)
	}
	defer c.Close()
	return c.Lines(), nil
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
	scopedLogger := serialLogger.With().
		Uint16("data_channel_id", *d.ID()).Logger()

	d.OnOpen(func() {
		portPath := config.SerialPortPath
		if portPath == "" {
			scopedLogger.Info().Msg("No serial port configured, channel idle")
			return
		}

		scopedLogger.Info().Str("port", portPath).Msg("Opening serial port")
		var err error
		port, err = serial.Open(portPath, serialPortMode)
		if err != nil {
			scopedLogger.Error().Err(err).Str("port", portPath).Msg("Failed to open serial port")
			return
		}

		// Read from serial → send to WebRTC
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := port.Read(buf)
				if err != nil {
					scopedLogger.Debug().Err(err).Msg("Serial read ended")
					return
				}
				if n > 0 {
					if sendErr := d.Send(buf[:n]); sendErr != nil {
						scopedLogger.Debug().Err(sendErr).Msg("WebRTC send failed")
						return
					}
				}
			}
		}()
	})

	d.OnMessage(func(msg webrtc.DataChannelMessage) {
		if port == nil {
			return
		}
		if _, err := port.Write(msg.Data); err != nil {
			scopedLogger.Warn().Err(err).Msg("Serial write failed")
		}
	})

	d.OnError(func(err error) {
		scopedLogger.Warn().Err(err).Msg("Serial channel error")
	})

	d.OnClose(func() {
		scopedLogger.Info().Msg("Serial channel closed")
		if port != nil {
			port.Close()
			port = nil
		}
	})
}
