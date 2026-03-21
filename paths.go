package kvm

import (
	"os"
	"path/filepath"
)

var (
	ConfigDir = getEnvOr("XKVM_CONFIG_DIR", "/etc/xkvm")
	DataDir   = getEnvOr("XKVM_DATA_DIR", "/var/lib/xkvm")
	LogDir    = getEnvOr("XKVM_LOG_DIR", "/var/log/xkvm")
)

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func ConfigPath(elem ...string) string {
	return filepath.Join(append([]string{ConfigDir}, elem...)...)
}

func DataPath(elem ...string) string {
	return filepath.Join(append([]string{DataDir}, elem...)...)
}

func LogPath(elem ...string) string {
	return filepath.Join(append([]string{LogDir}, elem...)...)
}
