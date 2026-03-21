package supervisor

import (
	"os"
	"path/filepath"
)

func dataDir() string {
	if v := os.Getenv("XKVM_DATA_DIR"); v != "" {
		return v
	}
	return "/var/lib/xkvm"
}

func logDir() string {
	if v := os.Getenv("XKVM_LOG_DIR"); v != "" {
		return v
	}
	return "/var/log/xkvm"
}

var (
	ErrorDumpDir = filepath.Join(dataDir(), "crashdump")
	AppLogPath   = filepath.Join(logDir(), "last.log")
)

const (
	EnvChildID        = "XKVM_CHILD_ID"
	EnvSubcomponent   = "XKVM_SUBCOMPONENT"
	ErrorDumpLastFile = "last-crash.log"
	ErrorDumpTemplate = "xkvm-%s.log"

	FailsafeReasonVideoMaxRestartAttemptsReached = "failsafe::video.max_restart_attempts_reached"
)
