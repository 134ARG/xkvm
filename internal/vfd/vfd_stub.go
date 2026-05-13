//go:build !linux || !cgo

package vfd

import "fmt"

type HostMetrics struct {
	CPUUtil     float64
	RAMUtil     float64
	GPUUtil     float64
	CPUTemp     float64
	GPUTemp     float64
	NetRxBytes  float64
	NetTxBytes  float64
	UptimeSec   int
	FailedUnits int
	Connected   bool
}

func Init(devicePath string) error {
	return fmt.Errorf("VFD is only supported on Linux builds with cgo enabled")
}

func Shutdown() {}

func UpdateHostMetrics(metrics HostMetrics) {}
