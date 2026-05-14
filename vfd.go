package kvm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type VFDConfig struct {
	Enabled    bool   `json:"enabled"`
	DevicePath string `json:"devicePath"`
	ListenPort int    `json:"listenPort"`
}

type VFDHostMetrics struct {
	CPUUtil     *float64 `json:"cpuUtil"`
	RAMUtil     *float64 `json:"ramUtil"`
	GPUUtil     *float64 `json:"gpuUtil"`
	CPUTemp     *float64 `json:"cpuTemp"`
	GPUTemp     *float64 `json:"gpuTemp"`
	NetRxBytes  *float64 `json:"netRxBytes"`
	NetTxBytes  *float64 `json:"netTxBytes"`
	UptimeSec   *int     `json:"uptimeSec"`
	FailedUnits *int     `json:"failedUnits"`
	Connected   bool     `json:"connected"`
	UpdatedAt   int64    `json:"updatedAt"`
}

type hostMetricsPayload struct {
	Metrics struct {
		CPU struct {
			Util float64 `json:"util"`
		} `json:"cpu"`
		RAM struct {
			Util float64 `json:"util"`
		} `json:"ram"`
		GPU struct {
			Util float64  `json:"util"`
			Temp *float64 `json:"temp"`
		} `json:"gpu"`
		Temp struct {
			CPU *float64 `json:"cpu"`
			GPU *float64 `json:"gpu"`
		} `json:"temp"`
		Net struct {
			RxBytes float64 `json:"rx_bytes"`
			TxBytes float64 `json:"tx_bytes"`
		} `json:"net"`
		Sys struct {
			Uptime      int `json:"uptime"`
			FailedUnits int `json:"failed_units"`
		} `json:"sys"`
	} `json:"metrics"`
}

var vfdState = struct {
	sync.Mutex
	hostMetrics VFDHostMetrics
	childStdin  io.WriteCloser
}{}

const (
	vfdChildReadyMessage = "READY"
	vfdChildMaxAttempts  = 3
	vfdChildReadyTimeout = 10 * time.Second
)

func initVFD() {
	if !config.VFDEnabled {
		return
	}

	listenPort := config.VFDListenPort
	if listenPort <= 0 {
		listenPort = 9101
	}

	go runVFDHostMetricsServer(listenPort)
	startVFDChild(config.VFDDevicePath)
	logger.Info().Int("port", listenPort).Msg("VFD initialized")
}

func startVFDChild(devicePath string) {
	for attempt := 1; attempt <= vfdChildMaxAttempts; attempt++ {
		if startVFDChildAttempt(devicePath, attempt) {
			return
		}
		time.Sleep(time.Second)
	}
	logger.Warn().Int("attempts", vfdChildMaxAttempts).Msg("VFD child failed to start; disabling VFD runtime until next xKVM restart")
}

func startVFDChildAttempt(devicePath string, attempt int) bool {
	binaryPath, err := os.Executable()
	if err != nil {
		logger.Warn().Err(err).Msg("failed to resolve executable for VFD child")
		return false
	}

	cmd := exec.Command(binaryPath, "-subcomponent=vfd", "-vfd-device", devicePath)
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:   true,
		Pdeathsig: syscall.SIGTERM,
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Warn().Err(err).Int("attempt", attempt).Msg("failed to open VFD child stdout")
		return false
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		logger.Warn().Err(err).Int("attempt", attempt).Msg("failed to open VFD child stdin")
		return false
	}
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		logger.Warn().Err(err).Int("attempt", attempt).Msg("failed to start VFD child")
		return false
	}

	ready := make(chan bool, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		readySeen := false
		for scanner.Scan() {
			line := scanner.Text()
			if !readySeen && line == vfdChildReadyMessage {
				readySeen = true
				ready <- true
				continue
			}
			fmt.Fprintln(os.Stdout, line)
		}
		if !readySeen {
			ready <- false
		}
	}()

	select {
	case ok := <-ready:
		if !ok {
			_ = stdin.Close()
			_ = cmd.Wait()
			logger.Warn().Int("attempt", attempt).Msg("VFD child exited before ready")
			return false
		}
	case <-time.After(vfdChildReadyTimeout):
		_ = stdin.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		logger.Warn().Int("attempt", attempt).Dur("timeout", vfdChildReadyTimeout).Msg("VFD child ready timeout")
		return false
	}

	vfdState.Lock()
	vfdState.childStdin = stdin
	vfdState.Unlock()

	go func() {
		if err := cmd.Wait(); err != nil {
			logger.Warn().Err(err).Msg("VFD child exited")
		} else {
			logger.Info().Msg("VFD child exited")
		}
		vfdState.Lock()
		if vfdState.childStdin == stdin {
			vfdState.childStdin = nil
		}
		vfdState.Unlock()
	}()
	logger.Info().Int("pid", cmd.Process.Pid).Int("attempt", attempt).Msg("VFD child started")
	return true
}

func runVFDHostMetricsServer(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Warn().Err(err).Int("port", port).Msg("failed to listen for VFD host metrics")
		return
	}
	defer listener.Close()

	logger.Info().Int("port", port).Msg("VFD host metrics server started")

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Warn().Err(err).Msg("VFD metrics accept failed")
			time.Sleep(time.Second)
			continue
		}
		go handleVFDHostMetricsConn(conn)
	}
}

func handleVFDHostMetricsConn(conn net.Conn) {
	defer conn.Close()
	logger.Info().Str("remote", conn.RemoteAddr().String()).Msg("VFD host metrics connected")

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 4096), 1024*1024)
	for scanner.Scan() {
		var payload hostMetricsPayload
		if err := json.Unmarshal(scanner.Bytes(), &payload); err != nil {
			logger.Debug().Err(err).Msg("invalid VFD host metrics payload")
			continue
		}
		hostMetrics := hostMetricsFromPayload(payload, true)
		setVFDHostMetrics(hostMetrics)
		writeVFDChildMetrics(hostMetrics)
	}

	hostMetrics := VFDHostMetrics{UpdatedAt: time.Now().UnixMilli()}
	setVFDHostMetrics(hostMetrics)
	writeVFDChildMetrics(hostMetrics)
	logger.Info().Str("remote", conn.RemoteAddr().String()).Msg("VFD host metrics disconnected")
}

func hostMetricsFromPayload(payload hostMetricsPayload, connected bool) VFDHostMetrics {
	return VFDHostMetrics{
		CPUUtil:     ptr(payload.Metrics.CPU.Util),
		RAMUtil:     ptr(payload.Metrics.RAM.Util),
		GPUUtil:     ptr(payload.Metrics.GPU.Util),
		CPUTemp:     payload.Metrics.Temp.CPU,
		GPUTemp:     gpuTempFromPayload(payload),
		NetRxBytes:  ptr(payload.Metrics.Net.RxBytes),
		NetTxBytes:  ptr(payload.Metrics.Net.TxBytes),
		UptimeSec:   ptr(payload.Metrics.Sys.Uptime),
		FailedUnits: ptr(payload.Metrics.Sys.FailedUnits),
		Connected:   connected,
		UpdatedAt:   time.Now().UnixMilli(),
	}
}

func gpuTempFromPayload(payload hostMetricsPayload) *float64 {
	if payload.Metrics.Temp.GPU != nil {
		return payload.Metrics.Temp.GPU
	}
	return payload.Metrics.GPU.Temp
}

func setVFDHostMetrics(metrics VFDHostMetrics) {
	vfdState.Lock()
	defer vfdState.Unlock()
	vfdState.hostMetrics = metrics
}

func writeVFDChildMetrics(metrics VFDHostMetrics) {
	payload, err := json.Marshal(metrics)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to marshal VFD child metrics")
		return
	}
	payload = append(payload, '\n')

	vfdState.Lock()
	defer vfdState.Unlock()
	if vfdState.childStdin == nil {
		return
	}
	if _, err := vfdState.childStdin.Write(payload); err != nil {
		logger.Warn().Err(err).Msg("failed to write VFD child metrics")
		_ = vfdState.childStdin.Close()
		vfdState.childStdin = nil
	}
}

func ptr[T any](value T) *T {
	return &value
}

func rpcGetVFDConfig() (VFDConfig, error) {
	return VFDConfig{
		Enabled:    config.VFDEnabled,
		DevicePath: config.VFDDevicePath,
		ListenPort: config.VFDListenPort,
	}, nil
}

func rpcSetVFDConfig(vfdConfig VFDConfig) error {
	if vfdConfig.ListenPort <= 0 || vfdConfig.ListenPort > 65535 {
		return fmt.Errorf("invalid VFD listen port: %d", vfdConfig.ListenPort)
	}

	config.VFDEnabled = vfdConfig.Enabled
	config.VFDDevicePath = vfdConfig.DevicePath
	config.VFDListenPort = vfdConfig.ListenPort
	return SaveConfig()
}

func rpcGetVFDHostMetrics() (VFDHostMetrics, error) {
	vfdState.Lock()
	defer vfdState.Unlock()
	return vfdState.hostMetrics, nil
}
