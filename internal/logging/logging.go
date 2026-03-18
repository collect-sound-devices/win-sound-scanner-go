package logging

import (
	"log"
	"os"
	"strings"

	"github.com/collect-sound-devices/sound-win-scanner/v4/pkg/soundlibwrap"
)

type Logf func(format string, v ...interface{})

const appLogFlags = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lmsgprefix

// NewLogger builds a timestamped logger for app messages.
func NewLogger(prefix string) *log.Logger {
	return log.New(os.Stdout, prefix, appLogFlags)
}

// NewPlainLogger builds a logger without any prefix/flags, useful for bridging messages that already include timestamps.
func NewPlainLogger() *log.Logger {
	return log.New(os.Stdout, "", 0)
}

// AttachSoundlibwrapBridge forwards soundlibwrap log messages into the provided logger.
// The logger should typically have no flags/prefix so the embedded timestamp is preserved.
func AttachSoundlibwrapBridge(logf Logf, prefix string) {
	if logf == nil {
		panic("nil logf")
	}
	if prefix == "" {
		prefix = "cpp backend"
	}

	soundlibwrap.SetLogHandler(func(timestamp, level, content string) {
		lvl := strings.ToLower(level)
		levelTag := "info"
		switch lvl {
		case "trace", "debug":
			levelTag = "debug"
		case "warn", "warning":
			levelTag = "warn"
		case "error", "critical":
			levelTag = "error"
		}

		logf("%s [%s %s] %s", timestamp, prefix, levelTag, content)
	})
}
