package main

import (
	"io"
	"log/slog"
	"os"

	"github.com/collect-sound-devices/win-sound-go-bridge/internal/logging"
)

func newAppLogger(writer io.Writer) *slog.Logger {
	return logging.NewLogger(writer, slog.LevelInfo).With("app", serviceName)
}

func newBootstrapLogger() *slog.Logger {
	return newAppLogger(os.Stderr)
}

func fatalLog(logger *slog.Logger, message string, args ...any) {
	if logger == nil {
		logger = newBootstrapLogger()
	}
	logger.Error(message, args...)
	os.Exit(1)
}
