package logging

import (
	"context"
	"io"
	"log/slog"
	"strings"

	"github.com/collect-sound-devices/sound-win-scanner/v4/pkg/soundlibwrap"
)

// NewLogger builds a structured app logger.
func NewLogger(writer io.Writer, level slog.Leveler) *slog.Logger {
	if writer == nil {
		panic("nil writer")
	}
	if level == nil {
		panic("nil level")
	}

	return slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{Level: level}))
}

// WithComponent adds a component attribute when one is provided.
func WithComponent(logger *slog.Logger, component string) *slog.Logger {
	if logger == nil {
		panic("nil logger")
	}
	component = strings.TrimSpace(component)
	if component == "" {
		return logger
	}
	return logger.With("component", component)
}

// AttachSoundlibwrapBridge forwards soundlibwrap log messages into the provided logger.
func AttachSoundlibwrapBridge(logger *slog.Logger) {
	if logger == nil {
		panic("nil logger")
	}

	soundlibwrap.SetLogHandler(func(timestamp, level, content string) {
		nativeLevel := strings.ToLower(strings.TrimSpace(level))
		args := make([]any, 0, 4)
		if nativeLevel != "" {
			args = append(args, "native_level", nativeLevel)
		}
		if timestamp = strings.TrimSpace(timestamp); timestamp != "" {
			args = append(args, "native_timestamp", timestamp)
		}

		logger.Log(context.Background(), mapSoundlibwrapLevel(nativeLevel), content, args...)
	})
}

func mapSoundlibwrapLevel(level string) slog.Level {
	switch level {
	case "trace", "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "critical":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
