package kvm

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
)

func extractSerialNumber() (string, error) {
	content, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "", err
	}

	r, err := regexp.Compile(`Serial\s*:\s*(\S+)`)
	if err != nil {
		return "", fmt.Errorf("failed to compile regex: %w", err)
	}

	matches := r.FindStringSubmatch(string(content))
	if len(matches) < 2 {
		return "", fmt.Errorf("no serial found")
	}

	return matches[1], nil
}

var deviceID string
var deviceIDOnce sync.Once

func GetDeviceID() string {
	deviceIDOnce.Do(func() {
		serial, err := extractSerialNumber()
		if err != nil {
			logger.Warn().Msg("unknown serial number, the program likely not running on RV1106")
			deviceID = "unknown_device_id"
		} else {
			deviceID = serial
		}
	})
	return deviceID
}

func GetDefaultHostname() string {
	deviceId := GetDeviceID()
	if deviceId == "unknown_device_id" {
		return "xkvm"
	}

	return fmt.Sprintf("xkvm-%s", strings.ToLower(deviceId))
}
