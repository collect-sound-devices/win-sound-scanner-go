package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows"
)

const appLogTimeLayout = "2006/01/02 15:04:05.000000-07:00"
const appLogUnknownComponent = "unknown"

type appLogHandler struct {
	mu     *sync.Mutex
	writer io.Writer
	level  slog.Leveler
	attrs  []slog.Attr
}

// newAppLogger builds a structured app logger.
func newAppLogger(writer io.Writer) *slog.Logger {
	if writer == nil {
		panic("nil writer")
	}
	return slog.New(&appLogHandler{
		mu:     &sync.Mutex{},
		writer: writer,
		level:  slog.LevelInfo,
	})
}

func fatalLog(logger *slog.Logger, message string, args ...any) {
	if logger == nil {
		logger = newAppLogger(os.Stderr)
	}
	logger.Error(message, args...)
	os.Exit(1)
}

func (h *appLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.level != nil {
		minLevel = h.level.Level()
	}
	return level >= minLevel
}

func (h *appLogHandler) Handle(_ context.Context, record slog.Record) error {
	if !h.Enabled(context.Background(), record.Level) {
		return nil
	}

	timestampText := h.findNativeTimestamp("", h.attrs)
	record.Attrs(func(attr slog.Attr) bool {
		timestampText = h.findNativeTimestamp(timestampText, []slog.Attr{attr})
		return true
	})
	if timestampText == "" {
		timestamp := record.Time
		if timestamp.IsZero() {
			timestamp = time.Now()
		}
		timestampText = timestamp.Local().Format(appLogTimeLayout)
	}

	var builder strings.Builder
	builder.Grow(128)
	component := appLogUnknownComponent
	component = h.findComponent(component, h.attrs)
	record.Attrs(func(attr slog.Attr) bool {
		component = h.findComponent(component, []slog.Attr{attr})
		return true
	})
	builder.WriteString(timestampText)
	builder.WriteByte(' ')
	builder.WriteString(appLogLevel(record.Level))
	builder.WriteString(" [")
	builder.WriteString(component)
	builder.WriteString("] [")
	builder.WriteString(strconv.FormatUint(uint64(windows.GetCurrentThreadId()), 10))
	builder.WriteString("] ")
	builder.WriteString(record.Message)

	h.appendAttrs(&builder, h.attrs)
	record.Attrs(func(attr slog.Attr) bool {
		h.appendAttr(&builder, attr)
		return true
	})
	builder.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.writer, builder.String())
	return err
}

func (h *appLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := h.clone()
	clone.attrs = append(clone.attrs, attrs...)
	return clone
}

func (h *appLogHandler) WithGroup(name string) slog.Handler {
	return h.clone()
}

func (h *appLogHandler) clone() *appLogHandler {
	attrs := make([]slog.Attr, len(h.attrs))
	copy(attrs, h.attrs)
	return &appLogHandler{
		mu:     h.mu,
		writer: h.writer,
		level:  h.level,
		attrs:  attrs,
	}
}

func (h *appLogHandler) appendAttrs(builder *strings.Builder, attrs []slog.Attr) {
	for _, attr := range attrs {
		h.appendAttr(builder, attr)
	}
}

func (h *appLogHandler) appendAttr(builder *strings.Builder, attr slog.Attr) {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) || attr.Key == "" {
		return
	}
	if attr.Key == "component" {
		return
	}
	if attr.Key == "native_timestamp" {
		return
	}
	if attr.Key == "native_level" {
		return
	}
	if attr.Value.Kind() == slog.KindGroup {
		for _, child := range attr.Value.Group() {
			h.appendAttr(builder, child)
		}
		return
	}
	builder.WriteByte(' ')
	builder.WriteString(attr.Key)
	builder.WriteByte('=')
	appendAttrValue(builder, attr.Value)
}

func (h *appLogHandler) findComponent(current string, attrs []slog.Attr) string {
	for _, attr := range attrs {
		current = findComponentAttr(current, attr)
	}
	return current
}

func (h *appLogHandler) findNativeTimestamp(current string, attrs []slog.Attr) string {
	for _, attr := range attrs {
		current = findNativeTimestampAttr(current, attr)
	}
	return current
}

func findComponentAttr(current string, attr slog.Attr) string {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return current
	}
	if attr.Value.Kind() == slog.KindGroup {
		for _, child := range attr.Value.Group() {
			current = findComponentAttr(current, child)
		}
		return current
	}
	if attr.Key != "component" {
		return current
	}
	component := strings.TrimSpace(attrValueString(attr.Value))
	if component == "" {
		return appLogUnknownComponent
	}
	return component
}

func findNativeTimestampAttr(current string, attr slog.Attr) string {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return current
	}
	if attr.Value.Kind() == slog.KindGroup {
		for _, child := range attr.Value.Group() {
			current = findNativeTimestampAttr(current, child)
		}
		return current
	}
	if attr.Key != "native_timestamp" {
		return current
	}
	timestamp := strings.TrimSpace(attrValueString(attr.Value))
	if timestamp == "" {
		return current
	}
	return timestamp
}

func appendAttrValue(builder *strings.Builder, value slog.Value) {
	switch value.Kind() {
	case slog.KindString:
		builder.WriteString(strconv.Quote(value.String()))
	case slog.KindInt64:
		builder.WriteString(strconv.FormatInt(value.Int64(), 10))
	case slog.KindUint64:
		builder.WriteString(strconv.FormatUint(value.Uint64(), 10))
	case slog.KindFloat64:
		builder.WriteString(strconv.FormatFloat(value.Float64(), 'f', -1, 64))
	case slog.KindBool:
		builder.WriteString(strconv.FormatBool(value.Bool()))
	case slog.KindDuration:
		builder.WriteString(strconv.Quote(value.Duration().String()))
	case slog.KindTime:
		builder.WriteString(strconv.Quote(value.Time().Format(time.RFC3339Nano)))
	case slog.KindAny:
		appendAnyValue(builder, value.Any())
	default:
		builder.WriteString(strconv.Quote(value.String()))
	}
}

func appendAnyValue(builder *strings.Builder, value any) {
	switch v := value.(type) {
	case nil:
		builder.WriteString("<nil>")
	case error:
		builder.WriteString(strconv.Quote(v.Error()))
	case fmt.Stringer:
		builder.WriteString(strconv.Quote(v.String()))
	case string:
		builder.WriteString(strconv.Quote(v))
	default:
		builder.WriteString(fmt.Sprint(v))
	}
}

func attrValueString(value slog.Value) string {
	switch value.Kind() {
	case slog.KindString:
		return value.String()
	case slog.KindInt64:
		return strconv.FormatInt(value.Int64(), 10)
	case slog.KindUint64:
		return strconv.FormatUint(value.Uint64(), 10)
	case slog.KindFloat64:
		return strconv.FormatFloat(value.Float64(), 'f', -1, 64)
	case slog.KindBool:
		return strconv.FormatBool(value.Bool())
	case slog.KindDuration:
		return value.Duration().String()
	case slog.KindTime:
		return value.Time().Format(time.RFC3339Nano)
	case slog.KindAny:
		return anyValueString(value.Any())
	default:
		return value.String()
	}
}

func anyValueString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	case string:
		return v
	default:
		return fmt.Sprint(v)
	}
}

func appLogLevel(level slog.Level) string {
	switch {
	case level <= slog.LevelDebug:
		return "D"
	case level >= slog.LevelError:
		return "E"
	case level >= slog.LevelWarn:
		return "W"
	default:
		return "I"
	}
}
