package main

import (
	"bytes"
	"context"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewAppLoggerFormatsInfoLine(t *testing.T) {
	var buffer bytes.Buffer
	logger := newAppLogger(&buffer).With("component", "scanner")

	logger.Info("message", "count", 2)

	line := strings.TrimSpace(buffer.String())
	pattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{6} I \[scanner\] \[\d+\] message count=2$`)
	if !pattern.MatchString(line) {
		t.Fatalf("unexpected log line: %q", line)
	}
}

func TestAppLogHandlerUsesLocalTimeAndMicroseconds(t *testing.T) {
	var buffer bytes.Buffer
	handler := &appLogHandler{
		mu:     new(sync.Mutex),
		writer: &buffer,
		level:  slog.LevelDebug,
	}
	stamp := time.Date(2026, time.March, 11, 11, 13, 49, 469899000, time.UTC)
	record := slog.NewRecord(stamp, slog.LevelInfo, "message", 0)

	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatalf("handle failed: %v", err)
	}

	line := strings.TrimSpace(buffer.String())
	expectedPrefix := stamp.Local().Format(appLogTimeLayout) + " I ["
	if !strings.HasPrefix(line, expectedPrefix) {
		t.Fatalf("expected prefix %q, got %q", expectedPrefix, line)
	}
	if !strings.Contains(line, " [unknown] [") {
		t.Fatalf("expected unknown component in %q", line)
	}
	if !strings.Contains(line, "] message") {
		t.Fatalf("expected message in %q", line)
	}
}

func TestAppLogLevelMapping(t *testing.T) {
	tests := []struct {
		level slog.Level
		want  string
	}{
		{level: slog.LevelDebug, want: "D"},
		{level: slog.LevelInfo, want: "I"},
		{level: slog.LevelWarn, want: "W"},
		{level: slog.LevelError, want: "E"},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			if got := appLogLevel(test.level); got != test.want {
				t.Fatalf("appLogLevel(%v) = %q, want %q", test.level, got, test.want)
			}
		})
	}
}

func TestWithGroupIsNoOp(t *testing.T) {
	var buffer bytes.Buffer
	logger := slog.New(newAppLogger(&buffer).Handler().WithGroup("ignored").WithAttrs([]slog.Attr{slog.String("component", "scanner")}))

	logger.Info("message")

	line := strings.TrimSpace(buffer.String())
	if !strings.Contains(line, ` [scanner] [`) {
		t.Fatalf("expected component prefix in %q", line)
	}
	if strings.Contains(line, ` component="scanner"`) {
		t.Fatalf("expected component attr to be removed from tail in %q", line)
	}
}

func TestNewAppLoggerUsesUnknownComponentWhenMissing(t *testing.T) {
	var buffer bytes.Buffer

	newAppLogger(&buffer).Info("message", "count", 2)

	line := strings.TrimSpace(buffer.String())
	pattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{6} I \[unknown\] \[\d+\] message count=2$`)
	if !pattern.MatchString(line) {
		t.Fatalf("unexpected log line: %q", line)
	}
}
