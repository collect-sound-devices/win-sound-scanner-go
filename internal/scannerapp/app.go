package scannerapp

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/enqueuer"
	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/scannerapp/impl"

	"github.com/collect-sound-devices/sound-win-scanner/v4/pkg/soundlibwrap"
)

var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)

const (
	eventDefaultRenderChanged  = "default_render_changed"
	eventDefaultCaptureChanged = "default_capture_changed"
	eventRenderVolumeChanged   = "render_volume_changed"
	eventCaptureVolumeChanged  = "capture_volume_changed"

	flowRender  = "render"
	flowCapture = "capture"
)

func logf(level, format string, v ...interface{}) {
	if level == "" {
		level = "info"
	}
	logger.Printf("["+level+"] "+format, v...)
}

func logInfo(format string, v ...interface{}) {
	logf("info", format, v...)
}

func logError(format string, v ...interface{}) {
	logf("error", format, v...)
}

func New(enqueue func(string, map[string]string)) (ScannerApp, error) {
	return impl.New(enqueue, logInfo, logError)
}

func Run(ctx context.Context) error {
	reqEnqueuer := enqueuer.NewEmptyRequestEnqueuer(logger)
	enqueue := func(name string, fields map[string]string) {
		if err := reqEnqueuer.EnqueueRequest(enqueuer.Request{
			Name:      name,
			Timestamp: time.Now(),
			Fields:    fields,
		}); err != nil {
			logError("enqueue failed: %v", err)
		}
	}

	{
		logHandlerLogger := log.New(os.Stdout, "", 0)
		prefix := "cpp backend,"
		// Bridge C soundlibwrap messages to Go logHandlerLogger.
		soundlibwrap.SetLogHandler(func(timestamp, level, content string) {
			switch strings.ToLower(level) {
			case "trace", "debug":
				logHandlerLogger.Printf("%s [%s debug] %s", timestamp, prefix, content)
			case "info":
				logHandlerLogger.Printf("%s [%s info] %s", timestamp, prefix, content)
			case "warn", "warning":
				logHandlerLogger.Printf("%s [%s warn] %s", timestamp, prefix, content)
			case "error", "critical":
				logHandlerLogger.Printf("%s [%s error] %s", timestamp, prefix, content)
			default:
				logHandlerLogger.Printf("%s [%s info] %s", timestamp, prefix, content)
			}
		})
	}

	logInfo("Initializing...")

	app, err := New(enqueue)
	if err != nil {
		return err
	}
	defer app.Shutdown()

	// Post the default render and capture devices.
	app.RepostRenderDeviceToApi()
	app.RepostCaptureDeviceToApi()

	// Keep running until interrupted to receive async logs and change events.
	<-ctx.Done()
	logInfo("Shutting down...")
	return nil
}
