package vfd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type childMetrics struct {
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
}

func RunVFDProcess(devicePath string) {
	if err := Init(devicePath); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize VFD: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stdout, "READY")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 4096), 1024*1024)
	for scanner.Scan() {
		var metrics childMetrics
		if err := json.Unmarshal(scanner.Bytes(), &metrics); err != nil {
			fmt.Fprintf(os.Stderr, "invalid VFD metrics payload: %v\n", err)
			continue
		}
		UpdateHostMetrics(HostMetrics{
			CPUUtil:     float64Value(metrics.CPUUtil),
			RAMUtil:     float64Value(metrics.RAMUtil),
			GPUUtil:     float64Value(metrics.GPUUtil),
			CPUTemp:     float64Value(metrics.CPUTemp),
			GPUTemp:     float64Value(metrics.GPUTemp),
			NetRxBytes:  float64Value(metrics.NetRxBytes),
			NetTxBytes:  float64Value(metrics.NetTxBytes),
			UptimeSec:   intValue(metrics.UptimeSec),
			FailedUnits: intValue(metrics.FailedUnits),
			Connected:   metrics.Connected,
		})
	}
}

func float64Value(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
