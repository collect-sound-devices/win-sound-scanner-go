package scannerapp

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/collect-sound-devices/sound-win-scanner/v4/pkg/soundlibwrap"
	c "github.com/collect-sound-devices/win-sound-scanner-go/internal/contract"
	"github.com/collect-sound-devices/win-sound-scanner-go/pkg/appinfo"
)

type ScannerApp interface {
	RepostRenderDeviceToApi(c.EventType)
	RepostCaptureDeviceToApi(c.EventType)
	Shutdown()
}

type scannerAppImpl struct {
	soundLibHandle soundlibwrap.Handle
	enqueueFunc    func(c.EventType, map[string]string)
	logger         *slog.Logger
	osName         string
	hostName       string
}

func NewImpl(enqueue func(c.EventType, map[string]string), logger *slog.Logger) (*scannerAppImpl, error) {
	if enqueue == nil {
		panic("nil enqueue")
	}
	if logger == nil {
		panic("nil logger")
	}

	app := &scannerAppImpl{
		enqueueFunc: enqueue,
		logger:      logger,
	}
	app.attachHandlers()
	if err := app.init(); err != nil {
		return nil, err
	}

	// Post the default render and capture devices.
	app.RepostRenderDeviceToApi(c.EventTypeRenderDeviceConfirmed)
	app.RepostCaptureDeviceToApi(c.EventTypeCaptureDeviceConfirmed)

	return app, nil
}

func (app *scannerAppImpl) init() error {
	h, err := soundlibwrap.Initialize(appinfo.AppName, appinfo.Version)
	if err != nil {
		return err
	}

	app.soundLibHandle = h
	if err := soundlibwrap.RegisterCallbacks(app.soundLibHandle); err != nil {
		_ = soundlibwrap.Uninitialize(app.soundLibHandle)
		app.soundLibHandle = 0
		return err
	}

	if osName, err := soundlibwrap.GetExtendedOperatingSystemName(app.soundLibHandle); err != nil || strings.TrimSpace(osName) == "" {
		app.logger.Warn("Cannot get OS name", "err", err)
		app.osName = "Unknown OS"
	} else {
		app.osName = osName
	}

	if hostName, err := os.Hostname(); err != nil || strings.TrimSpace(hostName) == "" {
		app.logger.Warn("Cannot get host name", "err", err)
		app.hostName = "unknown-host"
	} else {
		app.hostName = hostName
	}

	return nil
}

func (app *scannerAppImpl) attachHandlers() {

	soundlibwrap.SetLogHandler(func(timestamp, level, content string) {
		nativeLevel := strings.ToLower(strings.TrimSpace(level))
		args := make([]any, 0, 4)
		if nativeLevel != "" {
			args = append(args, "native_level", nativeLevel)
		}
		if timestamp = strings.TrimSpace(timestamp); timestamp != "" {
			args = append(args, "native_timestamp", timestamp)
		}

		getLogLevel := func(level string) slog.Level {
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

		app.logger.Log(context.Background(), getLogLevel(nativeLevel),
			content, args...)
	})

	// Device default change notifications.
	soundlibwrap.SetDefaultRenderHandler(func(present bool) {
		if present {
			app.RepostRenderDeviceToApi(c.EventTypeRenderDeviceDiscovered)
		} else {
			// not yet implemented removeDeviceToApi
			app.logger.Info("Render device removed")
		}
	})
	soundlibwrap.SetDefaultCaptureHandler(func(present bool) {
		if present {
			app.RepostCaptureDeviceToApi(c.EventTypeCaptureDeviceDiscovered)
		} else {
			// not yet implemented removeDeviceToApi
			app.logger.Info("Capture device removed")
		}
	})

	// Volume change notifications.
	soundlibwrap.SetRenderVolumeChangedHandler(func() {
		if desc, err := soundlibwrap.GetDefaultRender(app.soundLibHandle); err == nil {
			app.putVolumeChangeToApi(c.EventTypeRenderVolumeChanged, desc.PnpID, int(desc.RenderVolume))
			app.logger.Info("Render volume changed", "name", desc.Name, "pnpId", desc.PnpID, "volume", desc.RenderVolume)
		} else {
			app.logger.Error("Render volume changed, cannot read it", "err", err)
		}
	})
	soundlibwrap.SetCaptureVolumeChangedHandler(func() {
		if desc, err := soundlibwrap.GetDefaultCapture(app.soundLibHandle); err == nil {
			app.putVolumeChangeToApi(c.EventTypeCaptureVolumeChanged, desc.PnpID, int(desc.CaptureVolume))
			app.logger.Info("Capture volume changed", "name", desc.Name, "pnpId", desc.PnpID, "volume", desc.CaptureVolume)
		} else {
			app.logger.Error("Capture volume changed, cannot read it", "err", err)
		}
	})
}

func (app *scannerAppImpl) Shutdown() {
	if app.soundLibHandle != 0 {
		_ = soundlibwrap.Uninitialize(app.soundLibHandle)
		app.soundLibHandle = 0
	}
}

func (app *scannerAppImpl) putVolumeChangeToApi(event c.EventType, pnpID string, volume int) {
	fields := map[string]string{
		c.FieldUpdateDate: time.Now().UTC().Format(time.RFC3339),
		c.FieldVolume:     strconv.Itoa(volume),
		c.FieldHostName:   app.hostName,
	}
	if pnpID != "" {
		fields[c.FieldPnpID] = pnpID
	}

	app.enqueueFunc(event, fields)
}

func (app *scannerAppImpl) RepostRenderDeviceToApi(event c.EventType) {
	if desc, err := soundlibwrap.GetDefaultRender(app.soundLibHandle); err == nil {
		renderVolume := int(desc.RenderVolume)
		captureVolume := int(desc.CaptureVolume)
		app.postDeviceToApi(event, desc.Name, desc.PnpID, renderVolume, captureVolume)
	} else {
		app.logger.Error("Render device cannot be identified", "err", err)
	}
}

func (app *scannerAppImpl) RepostCaptureDeviceToApi(event c.EventType) {
	if desc, err := soundlibwrap.GetDefaultCapture(app.soundLibHandle); err == nil {
		renderVolume := int(desc.RenderVolume)
		captureVolume := int(desc.CaptureVolume)
		app.postDeviceToApi(event, desc.Name, desc.PnpID, renderVolume, captureVolume)
	} else {
		app.logger.Error("Capture device cannot be identified", "err", err)
	}
}

func (app *scannerAppImpl) postDeviceToApi(event c.EventType, name, pnpID string, renderVolume, captureVolume int) {
	fields := map[string]string{
		c.FieldUpdateDate:          time.Now().UTC().Format(time.RFC3339),
		c.FieldName:                name,
		c.FieldPnpID:               pnpID,
		c.FieldRenderVolume:        strconv.Itoa(renderVolume),
		c.FieldCaptureVolume:       strconv.Itoa(captureVolume),
		c.FieldOperationSystemName: app.osName,
		c.FieldHostName:            app.hostName,
	}

	app.enqueueFunc(event, fields)
}
