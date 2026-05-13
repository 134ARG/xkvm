//go:build linux

package vfd

/*
#cgo linux,arm64 CFLAGS: -I${SRCDIR}/lib/aarch64/static
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/lib/aarch64/static -l:libch347.a -lpthread
#cgo linux,amd64 CFLAGS: -I${SRCDIR}/lib/x64/static
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/lib/x64/static -l:libch347.a -lpthread
#include "xkvm_vfd.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

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
	cPath := C.CString(devicePath)
	defer C.free(unsafe.Pointer(cPath))

	if ret := C.xkvm_vfd_init(cPath); ret != 0 {
		return fmt.Errorf("vfd init failed: %d", int(ret))
	}
	return nil
}

func Shutdown() {
	C.xkvm_vfd_shutdown()
}

func UpdateHostMetrics(metrics HostMetrics) {
	C.xkvm_vfd_update_host_metrics(&C.xkvm_vfd_host_metrics_t{
		cpu_util:     C.double(metrics.CPUUtil),
		ram_util:     C.double(metrics.RAMUtil),
		gpu_util:     C.double(metrics.GPUUtil),
		cpu_temp:     C.double(metrics.CPUTemp),
		gpu_temp:     C.double(metrics.GPUTemp),
		net_rx_bytes: C.double(metrics.NetRxBytes),
		net_tx_bytes: C.double(metrics.NetTxBytes),
		uptime_sec:   C.int(metrics.UptimeSec),
		failed_units: C.int(metrics.FailedUnits),
		connected:    C.bool(metrics.Connected),
	})
}
