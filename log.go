package kvm

import (
	"github.com/134ARG/xkvm/internal/logging"
	"github.com/rs/zerolog"
)

func ErrorfL(l *zerolog.Logger, format string, err error, args ...any) error {
	return logging.ErrorfL(l, format, err, args...)
}

var (
	logger          = logging.GetSubsystemLogger("xkvm")
	failsafeLogger  = logging.GetSubsystemLogger("failsafe")
	networkLogger   = logging.GetSubsystemLogger("network")
	cloudLogger     = logging.GetSubsystemLogger("cloud")
	websocketLogger = logging.GetSubsystemLogger("websocket")
	webrtcLogger    = logging.GetSubsystemLogger("webrtc")
	nativeLogger    = logging.GetSubsystemLogger("native")
	nbdLogger       = logging.GetSubsystemLogger("nbd")
	jsonRpcLogger   = logging.GetSubsystemLogger("jsonrpc")
	hidRPCLogger    = logging.GetSubsystemLogger("hidrpc")
	websecureLogger = logging.GetSubsystemLogger("websecure")
	otaLogger       = logging.GetSubsystemLogger("ota")
	serialLogger    = logging.GetSubsystemLogger("serial")
	terminalLogger  = logging.GetSubsystemLogger("terminal")
	displayLogger   = logging.GetSubsystemLogger("display")
	wolLogger       = logging.GetSubsystemLogger("wol")
	usbLogger       = logging.GetSubsystemLogger("usb")
	// external components
	ginLogger = logging.GetSubsystemLogger("gin")
)
