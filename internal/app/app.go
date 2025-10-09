package app

import (
	"context"
	"log"
	"os"
	"strings"

	"win-sound-dev-go-bridge/internal/saawrapper"
	"win-sound-dev-go-bridge/pkg/version"
)

func Run(ctx context.Context) error {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Bridge C log messages to Go logger.
	saawrapper.SetLogHandler(func(level, content string) {
		switch strings.ToLower(level) {
		case "trace", "debug":
			logger.Printf("[debug] %s", content)
		case "info":
			logger.Printf("[info] %s", content)
		case "warn", "warning":
			logger.Printf("[warn] %s", content)
		case "error", "critical":
			logger.Printf("[error] %s", content)
		default:
			logger.Printf("[log] %s", content)
		}
	})

	// Device default change notifications.
	saawrapper.SetDefaultRenderHandler(func(present bool) {
		logger.Printf("default render present=%v", present)
	})
	saawrapper.SetDefaultCaptureHandler(func(present bool) {
		logger.Printf("default capture present=%v", present)
	})

	// Initialize the C library and register callbacks.
	h, err := saawrapper.Initialize("win-sound-dev-go-bridge", version.Version)
	if err != nil {
		return err
	}
	defer func() {
		_ = saawrapper.Uninitialize(h)
	}()

	if err := saawrapper.RegisterCallbacks(h, true, true); err != nil {
		return err
	}

	// Usage example: query and log current defaults.
	if desc, err := saawrapper.GetDefaultRender(h); err == nil {
		logger.Printf("default render: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
	} else {
		logger.Printf("default render: error: %v", err)
	}
	if desc, err := saawrapper.GetDefaultCapture(h); err == nil {
		logger.Printf("default capture: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.CaptureVolume)
	} else {
		logger.Printf("default capture: error: %v", err)
	}

	// Keep running until interrupted to receive async logs and change events.
	<-ctx.Done()
	logger.Println("shutting down...")
	return nil
}
