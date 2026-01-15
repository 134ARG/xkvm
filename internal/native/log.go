package native

import (
	"github.com/134ARG/xkvm/internal/logging"
	"github.com/rs/zerolog"
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
