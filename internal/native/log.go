package native

import (
	"github.com/rs/zerolog"
	"github.com/xkvm/kvm/internal/logging"
)

var nativeLogger = logging.GetSubsystemLogger("native")
var displayLogger = logging.GetSubsystemLogger("display")

type nativeLogMessage struct {
	Level    zerolog.Level
	Message  string
	File     string
	FuncName string
	Line     int
}
