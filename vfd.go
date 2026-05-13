package kvm

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/134ARG/xkvm/internal/vfd"
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
	started     bool
	listener    net.Listener
	serverDone  chan struct{}
	conns       map[net.Conn]struct{}
	hostMetrics VFDHostMetrics
}{conns: make(map[net.Conn]struct{})}

var vfdApplyLock sync.Mutex

func initVFD() {
	if err := applyVFDConfig(VFDConfig{
		Enabled:    config.VFDEnabled,
		DevicePath: config.VFDDevicePath,
		ListenPort: config.VFDListenPort,
	}); err != nil {
		logger.Warn().Err(err).Msg("failed to apply VFD config")
	}
}

func applyVFDConfig(vfdConfig VFDConfig) error {
	if vfdConfig.ListenPort <= 0 {
		vfdConfig.ListenPort = 9101
	}

	done := stopVFD()
	waitVFDServer(done)

	if !vfdConfig.Enabled {
		return nil
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", vfdConfig.ListenPort))
	if err != nil {
		return fmt.Errorf("failed to listen for VFD host metrics on port %d: %w", vfdConfig.ListenPort, err)
	}

	done = make(chan struct{})
	vfdState.Lock()
	vfdState.listener = listener
	vfdState.serverDone = done
	vfdState.Unlock()

	go runVFDHostMetricsServer(listener, done)
	logger.Info().Int("port", vfdConfig.ListenPort).Msg("VFD host metrics server started")

	if err := vfd.Init(vfdConfig.DevicePath); err != nil {
		done := stopVFD()
		waitVFDServer(done)
		logger.Warn().Err(err).Msg("failed to initialize VFD")
		return err
	}

	vfdState.Lock()
	vfdState.started = true
	vfdState.Unlock()

	logger.Info().Int("port", vfdConfig.ListenPort).Msg("VFD initialized")
	return nil
}

func runVFDHostMetricsServer(listener net.Listener, done chan<- struct{}) {
	defer close(done)
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			logger.Warn().Err(err).Msg("VFD metrics accept failed")
			time.Sleep(time.Second)
			continue
		}
		go handleVFDHostMetricsConn(conn)
	}
}

func stopVFD() chan struct{} {
	vfdState.Lock()
	defer vfdState.Unlock()

	done := vfdState.serverDone
	if vfdState.listener != nil {
		_ = vfdState.listener.Close()
		vfdState.listener = nil
		vfdState.serverDone = nil
	}
	for conn := range vfdState.conns {
		_ = conn.Close()
	}
	if vfdState.started {
		vfd.Shutdown()
		vfdState.started = false
	}
	vfdState.hostMetrics = VFDHostMetrics{UpdatedAt: time.Now().UnixMilli()}
	vfd.UpdateHostMetrics(vfd.HostMetrics{})
	return done
}

func waitVFDServer(done <-chan struct{}) {
	if done != nil {
		<-done
	}
}

func handleVFDHostMetricsConn(conn net.Conn) {
	registerVFDHostMetricsConn(conn)
	defer unregisterVFDHostMetricsConn(conn)
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
		vfd.UpdateHostMetrics(vfd.HostMetrics{
			CPUUtil:     payload.Metrics.CPU.Util,
			RAMUtil:     payload.Metrics.RAM.Util,
			GPUUtil:     payload.Metrics.GPU.Util,
			CPUTemp:     floatValue(payload.Metrics.Temp.CPU),
			GPUTemp:     floatValue(gpuTempFromPayload(payload)),
			NetRxBytes:  payload.Metrics.Net.RxBytes,
			NetTxBytes:  payload.Metrics.Net.TxBytes,
			UptimeSec:   payload.Metrics.Sys.Uptime,
			FailedUnits: payload.Metrics.Sys.FailedUnits,
			Connected:   true,
		})
	}

	setVFDHostMetrics(VFDHostMetrics{UpdatedAt: time.Now().UnixMilli()})
	vfd.UpdateHostMetrics(vfd.HostMetrics{})
	logger.Info().Str("remote", conn.RemoteAddr().String()).Msg("VFD host metrics disconnected")
}

func registerVFDHostMetricsConn(conn net.Conn) {
	vfdState.Lock()
	defer vfdState.Unlock()
	vfdState.conns[conn] = struct{}{}
}

func unregisterVFDHostMetricsConn(conn net.Conn) {
	_ = conn.Close()
	vfdState.Lock()
	defer vfdState.Unlock()
	delete(vfdState.conns, conn)
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

func floatValue(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func setVFDHostMetrics(metrics VFDHostMetrics) {
	vfdState.Lock()
	defer vfdState.Unlock()
	vfdState.hostMetrics = metrics
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

	vfdApplyLock.Lock()
	defer vfdApplyLock.Unlock()

	oldConfig := VFDConfig{
		Enabled:    config.VFDEnabled,
		DevicePath: config.VFDDevicePath,
		ListenPort: config.VFDListenPort,
	}

	if err := applyVFDConfig(vfdConfig); err != nil {
		if rollbackErr := applyVFDConfig(oldConfig); rollbackErr != nil {
			logger.Warn().Err(rollbackErr).Msg("failed to restore previous VFD config after apply failure")
		}
		return err
	}

	config.VFDEnabled = vfdConfig.Enabled
	config.VFDDevicePath = vfdConfig.DevicePath
	config.VFDListenPort = vfdConfig.ListenPort
	if err := SaveConfig(); err != nil {
		config.VFDEnabled = oldConfig.Enabled
		config.VFDDevicePath = oldConfig.DevicePath
		config.VFDListenPort = oldConfig.ListenPort
		if rollbackErr := applyVFDConfig(oldConfig); rollbackErr != nil {
			logger.Warn().Err(rollbackErr).Msg("failed to restore previous VFD config after save failure")
		}
		return err
	}
	return nil
}

func rpcGetVFDHostMetrics() (VFDHostMetrics, error) {
	vfdState.Lock()
	defer vfdState.Unlock()
	return vfdState.hostMetrics, nil
}
