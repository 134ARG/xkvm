package kvm

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/warthog618/go-gpiocdev"
	"go.bug.st/serial"
)

// const serialPortPath = "/dev/ttyS3"

// port is the active serial port (or nil). It is opened/closed from WebRTC
// datachannel callbacks that run on different goroutines, so all access goes
// through getPort/setPort. Blocking reads/writes use a locally captured
// reference rather than holding portLock.
var (
	port     serial.Port
	portLock sync.Mutex
)

func getPort() serial.Port {
	portLock.Lock()
	defer portLock.Unlock()
	return port
}

func setPort(p serial.Port) {
	portLock.Lock()
	defer portLock.Unlock()
	port = p
}

var (
	// ATX LED state is written by the single runATXControl poller and read from
	// RPC/event goroutines; atomics keep those reads/writes race-free.
	ledHDDState       atomic.Bool
	ledPWRState       atomic.Bool
	atxStateAvailable atomic.Bool
	// btnRSTState bool
	// btnPWRState bool
	atxStopChan chan struct{}
)

func mountATXControl() error {
	if atxStopChan != nil {
		return nil
	}
	serialLogger.Info().Msg("ATX control mounting")
	atxStopChan = make(chan struct{})
	go runATXControl()
	return nil
}

// func unmountATXControl() error {
// 	serialLogger.Info().Msg("ATX control unmounting")
// 	if atxStopChan != nil {
// 		close(atxStopChan)
// 		atxStopChan = nil
// 	}
// 	return nil
// }

func runATXControl() {
	scopedLogger := serialLogger.With().Str("service", "atx_control").Logger()
	scopedLogger.Info().Msg("ATX control polling started")

	// Initialize default states
	ledHDDState.Store(false)
	ledPWRState.Store(false)
	atxStateAvailable.Store(false)
	// btnRSTState = false
	// btnPWRState = false

	prevPWR := false
	prevHDD := false
	prevAvailable := false

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-atxStopChan:
			scopedLogger.Info().Msg("ATX control polling stopped")
			return
		case <-ticker.C:
			pwrValue, pwrAvailable := readGPIOInput(config.GPIOPwrLedChip, config.GPIOPwrLedLine)
			hddValue, hddAvailable := readGPIOInput(config.GPIOHddLedChip, config.GPIOHddLedLine)

			pwr := pwrAvailable && pwrValue == config.GPIOPwrLedActiveHigh
			hdd := hddAvailable && hddValue == config.GPIOHddLedActiveHigh
			hdd = pwr && hdd

			atxStateAvailable.Store(pwrAvailable)
			ledPWRState.Store(pwr)
			ledHDDState.Store(hdd)

			if pwr != prevPWR || hdd != prevHDD || pwrAvailable != prevAvailable {
				prevPWR = pwr
				prevHDD = hdd
				prevAvailable = pwrAvailable
				triggerATXStateUpdate()
			}
		}
	}
}

func triggerATXStateUpdate() {
	go func() {
		cs := getCurrentSession()
		if cs == nil {
			return
		}
		writeJSONRPCEvent("atxState", currentATXState(), cs)
	}()
}

func pressATXPowerButton(duration time.Duration) error {
	return pulseGPIO(config.GPIOPwrChip, config.GPIOPwrLine, duration, config.GPIOPwrActiveHigh)
}

func pressATXResetButton(duration time.Duration) error {
	return pulseGPIO(config.GPIORstChip, config.GPIORstLine, duration, config.GPIORstActiveHigh)
}

func currentATXState() ATXState {
	return ATXState{
		Power:             ledPWRState.Load(),
		HDD:               ledHDDState.Load(),
		ATXStateAvailable: atxStateAvailable.Load(),
	}
}

// readGPIOInput reads a single GPIO line as input.
// The second return value is false when the line is unconfigured or unreadable.
func readGPIOInput(chip string, line int) (bool, bool) {
	if chip == "" || line < 0 {
		return false, false
	}

	l, err := gpiocdev.RequestLine(chip, line, gpiocdev.AsInput)
	if err != nil {
		serialLogger.Trace().Err(err).Str("chip", chip).Int("line", line).Msg("failed to request GPIO input line")
		return false, false
	}
	defer l.Close()

	val, err := l.Value()
	if err != nil {
		serialLogger.Trace().Err(err).Str("chip", chip).Int("line", line).Msg("failed to read GPIO input value")
		return false, false
	}
	return val != 0, true
}

// pulseGPIO opens a GPIO line, drives it to idle state, then pulses to active
// state for the given duration. Polarity is determined by activeHigh.
func pulseGPIO(chip string, line int, duration time.Duration, activeHigh bool) error {
	if chip == "" || line < 0 {
		return fmt.Errorf("ATX GPIO is not configured")
	}

	activeVal := 1
	idleVal := 0
	if !activeHigh {
		activeVal = 0
		idleVal = 1
	}

	l, err := gpiocdev.RequestLine(chip, line, gpiocdev.AsOutput(idleVal))
	if err != nil {
		return fmt.Errorf("failed to request GPIO %s line %d: %w", chip, line, err)
	}
	defer l.Close()

	// Clear: drive to idle
	if err := l.SetValue(idleVal); err != nil {
		return fmt.Errorf("failed to set GPIO idle: %w", err)
	}
	time.Sleep(10 * time.Millisecond)

	// Pulse: drive to active for duration
	if err := l.SetValue(activeVal); err != nil {
		return fmt.Errorf("failed to set GPIO active: %w", err)
	}
	time.Sleep(duration)

	// Release: drive to idle
	if err := l.SetValue(idleVal); err != nil {
		return fmt.Errorf("failed to set GPIO idle after pulse: %w", err)
	}

	serialLogger.Info().Str("chip", chip).Int("line", line).Dur("duration", duration).Bool("activeHigh", activeHigh).Msg("GPIO pulse complete")
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

	_ = mountATXControl()

	switch config.ActiveExtension {
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
		p, err := serial.Open(portPath, serialPortMode)
		if err != nil {
			scopedLogger.Error().Err(err).Str("port", portPath).Msg("Failed to open serial port")
			return
		}
		setPort(p)

		// Read from serial → send to WebRTC. Use the locally captured port
		// reference so the blocking Read doesn't hold portLock.
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := p.Read(buf)
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
		p := getPort()
		if p == nil {
			return
		}
		if _, err := p.Write(msg.Data); err != nil {
			scopedLogger.Warn().Err(err).Msg("Serial write failed")
		}
	})

	d.OnError(func(err error) {
		scopedLogger.Warn().Err(err).Msg("Serial channel error")
	})

	d.OnClose(func() {
		scopedLogger.Info().Msg("Serial channel closed")
		if p := getPort(); p != nil {
			p.Close()
			setPort(nil)
		}
	})
}
