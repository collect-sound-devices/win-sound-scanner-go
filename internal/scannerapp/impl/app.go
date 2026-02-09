package impl

import (
	"strconv"
	"time"

	"github.com/collect-sound-devices/win-sound-dev-go-bridge/pkg/appinfo"

	"github.com/collect-sound-devices/sound-win-scanner/v4/pkg/soundlibwrap"
)

const (
	eventDefaultRenderChanged  = "default_render_changed"
	eventDefaultCaptureChanged = "default_capture_changed"
	eventRenderVolumeChanged   = "render_volume_changed"
	eventCaptureVolumeChanged  = "capture_volume_changed"

	flowRender  = "render"
	flowCapture = "capture"
)

type app struct {
	soundLibHandle soundlibwrap.Handle
	enqueueFunc    func(string, map[string]string)
	logInfo        func(string, ...interface{})
	logError       func(string, ...interface{})
}

func New(enqueue func(string, map[string]string), logInfo func(string, ...interface{}), logError func(string, ...interface{})) (*app, error) {
	a := &app{
		enqueueFunc: enqueue,
		logInfo:     logInfo,
		logError:    logError,
	}
	a.attachHandlers()
	if err := a.init(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *app) init() error {
	h, err := soundlibwrap.Initialize(appinfo.AppName, appinfo.Version)
	if err != nil {
		return err
	}
	a.soundLibHandle = h
	if err := soundlibwrap.RegisterCallbacks(a.soundLibHandle); err != nil {
		_ = soundlibwrap.Uninitialize(a.soundLibHandle)
		a.soundLibHandle = 0
		return err
	}
	return nil
}

func (a *app) Shutdown() {
	if a.soundLibHandle != 0 {
		_ = soundlibwrap.Uninitialize(a.soundLibHandle)
		a.soundLibHandle = 0
	}
}

func (a *app) postDeviceToApi(eventType, flowType, name, pnpID string, renderVolume, captureVolume int) {
	fields := map[string]string{
		"device_message_type": eventType,
		"update_date":         time.Now().UTC().Format(time.RFC3339),
		"flow_type":           flowType,
		"name":                name,
		"pnp_id":              pnpID,
		"render_volume":       strconv.Itoa(renderVolume),
		"capture_volume":      strconv.Itoa(captureVolume),
	}

	a.enqueueFunc("post_device", fields)
}

func (a *app) putVolumeChangeToApi(eventType, pnpID string, volume int) {
	fields := map[string]string{
		"device_message_type": eventType,
		"update_date":         time.Now().UTC().Format(time.RFC3339),
		"volume":              strconv.Itoa(volume),
	}
	if pnpID != "" {
		fields["pnp_id"] = pnpID
	}

	a.enqueueFunc("put_volume_change", fields)
}

func (a *app) attachHandlers() {
	// Device default change notifications.
	soundlibwrap.SetDefaultRenderHandler(func(present bool) {
		if present {
			a.RepostRenderDeviceToApi()
		} else {
			// not yet implemented removeDeviceToApi
			a.logInfo("Render device removed")
		}
	})
	soundlibwrap.SetDefaultCaptureHandler(func(present bool) {
		if present {
			a.RepostCaptureDeviceToApi()
		} else {
			// not yet implemented removeDeviceToApi
			a.logInfo("Capture device removed")
		}
	})

	// Volume change notifications.
	soundlibwrap.SetRenderVolumeChangedHandler(func() {
		if desc, err := soundlibwrap.GetDefaultRender(a.soundLibHandle); err == nil {
			a.putVolumeChangeToApi(eventRenderVolumeChanged, desc.PnpID, int(desc.RenderVolume))
			a.logInfo("Render volume changed: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
		} else {
			a.logError("Render volume changed, can not read it: %v", err)
		}
	})
	soundlibwrap.SetCaptureVolumeChangedHandler(func() {
		if desc, err := soundlibwrap.GetDefaultCapture(a.soundLibHandle); err == nil {
			a.putVolumeChangeToApi(eventCaptureVolumeChanged, desc.PnpID, int(desc.CaptureVolume))
			a.logInfo("Capture volume changed: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.CaptureVolume)
		} else {
			a.logError("Capture volume changed, can not read it: %v", err)
		}
	})
}

func (a *app) RepostRenderDeviceToApi() {
	if desc, err := soundlibwrap.GetDefaultRender(a.soundLibHandle); err == nil {
		renderVolume := int(desc.RenderVolume)
		captureVolume := int(desc.CaptureVolume)
		a.postDeviceToApi(eventDefaultRenderChanged, flowRender, desc.Name, desc.PnpID, renderVolume, captureVolume)
		a.logInfo("Render device identified and updated: name=%q pnpId=%q renderVol=%d captureVol=%d", desc.Name, desc.PnpID, desc.RenderVolume, desc.CaptureVolume)
	} else {
		a.logError("Render device can not be identified: %v", err)
	}
}

func (a *app) RepostCaptureDeviceToApi() {
	if desc, err := soundlibwrap.GetDefaultCapture(a.soundLibHandle); err == nil {
		renderVolume := int(desc.RenderVolume)
		captureVolume := int(desc.CaptureVolume)
		a.postDeviceToApi(eventDefaultCaptureChanged, flowCapture, desc.Name, desc.PnpID, renderVolume, captureVolume)
		a.logInfo("Capture device identified and updated: name=%q pnpId=%q renderVol=%d captureVol=%d", desc.Name, desc.PnpID, desc.RenderVolume, desc.CaptureVolume)
	} else {
		a.logError("Capture device can not be identified: %v", err)
	}
}
