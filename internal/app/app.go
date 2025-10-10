package app

import (
	"context"
	"log"
	"os"
	"strings"

	"win-sound-dev-go-bridge/internal/saawrapper"
	"win-sound-dev-go-bridge/pkg/appinfo"
)

var SaaHandle saawrapper.Handle

func Run(ctx context.Context) error {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	prefix := "log event handler,"
	// Bridge C log messages to Go logger.
	saawrapper.SetLogHandler(func(level, content string) {
		switch strings.ToLower(level) {
		case "trace", "debug":
			logger.Printf("[%s debug] %s", prefix, content)
		case "info":
			logger.Printf("[%s info] %s", prefix, content)
		case "warn", "warning":
			logger.Printf("[%s warn] %s", prefix, content)
		case "error", "critical":
			logger.Printf("[%s error] %s", prefix, content)
		default:
			logger.Printf("[%s log] %s", prefix, content)
		}
	})

	// Device default change notifications.
	saawrapper.SetDefaultRenderHandler(func(present bool) {
		logger.Printf("[default render handler] present=%v", present)
		if present {
			if desc, err := saawrapper.GetDefaultRender(SaaHandle); err == nil {
				logger.Printf("[default render handler] name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
			} else {
				logger.Printf("[default render handler] error: %v", err)
			}
		}
	})
	saawrapper.SetDefaultCaptureHandler(func(present bool) {
		logger.Printf("[default capture handler] present=%v", present)
		if present {
			if desc, err := saawrapper.GetDefaultCapture(SaaHandle); err == nil {
				logger.Printf("[default capture handler] name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
			} else {
				logger.Printf("[default capture handler] error: %v", err)
			}
		}
	})

	logger.Println("Initilizing...")

	// Initialize the C library and register callbacks using the global handle.
	var err error
	SaaHandle, err = saawrapper.Initialize(appinfo.AppName, appinfo.Version)
	if err != nil {
		return err
	}
	defer func() {
		_ = saawrapper.Uninitialize(SaaHandle)
		SaaHandle = 0
	}()

	if err := saawrapper.RegisterCallbacks(SaaHandle); err != nil {
		return err
	}

	// Print the default render and capture devices.
	if desc, err := saawrapper.GetDefaultRender(SaaHandle); err == nil {
		logger.Printf("[initially print default render] name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
	} else {
		logger.Printf("[initially print default render] error: %v", err)
	}
	if desc, err := saawrapper.GetDefaultCapture(SaaHandle); err == nil {
		logger.Printf("[initially print default capture] name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.CaptureVolume)
	} else {
		logger.Printf("[initially print default capture] error: %v", err)
	}

	// Keep running until interrupted to receive async logs and change events.
	<-ctx.Done()
	logger.Println("shutting down...")
	return nil
}
