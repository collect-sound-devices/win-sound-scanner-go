package main

import (
	"io"
	"log/slog"
	"os"
)

// newAppLogger builds a structured app logger.
func newAppLogger(writer io.Writer) *slog.Logger {
	if writer == nil {
		panic("nil writer")
	}
	return slog.New(
		slog.NewTextHandler(
			writer,
			&slog.HandlerOptions{Level: slog.LevelInfo}))
}

func fatalLog(logger *slog.Logger, message string, args ...any) {
	if logger == nil {
		logger = newAppLogger(os.Stderr)
	}
	logger.Error(message, args...)
	os.Exit(1)
}
