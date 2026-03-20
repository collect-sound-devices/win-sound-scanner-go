package main

import (
	"io"
	"log/slog"
	"os"
)

// NewLogger builds a structured app logger.
func newAppLogger(writer io.Writer) *slog.Logger {
	if writer == nil {
		panic("nil writer")
	}
	return slog.New(
		slog.NewTextHandler(
			writer,
			&slog.HandlerOptions{Level: slog.LevelInfo})).
		With("app", serviceName)
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
