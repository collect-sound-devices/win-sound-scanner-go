//go:build windows && cgo

package saawrapper

/*
#cgo windows CFLAGS: -DUNICODE -D_UNICODE
#cgo windows LDFLAGS: -L${SRCDIR} -lSoundAgentApiDll

#include <stdint.h>
#include <stdbool.h>
#include <stdlib.h>

#include "SoundAgentApi.h"

// Forward declarations of Go-exported callbacks (implemented in package main)
void __stdcall cgoSaaDefaultRenderChanged(BOOL presentOrAbsent);
void __stdcall cgoSaaDefaultCaptureChanged(BOOL presentOrAbsent);
void __stdcall cgoSaaGotLogMessage(SaaLogMessage message);
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

type Handle uint64

type Description struct {
	PnpID         string
	Name          string
	IsRender      bool
	IsCapture     bool
	RenderVolume  uint16
	CaptureVolume uint16
}

type DefaultChangedCallback func(present bool)
type GotLogMessageCallback func(level, content string)

// Handlers set by app.
var (
	logHandler     GotLogMessageCallback
	renderHandler  DefaultChangedCallback
	captureHandler DefaultChangedCallback
)

func SetLogHandler(h GotLogMessageCallback)             { logHandler = h }
func SetDefaultRenderHandler(h DefaultChangedCallback)  { renderHandler = h }
func SetDefaultCaptureHandler(h DefaultChangedCallback) { captureHandler = h }

func NotifyDefaultRenderChanged(present bool) {
	if renderHandler != nil {
		renderHandler(present)
	}
}
func NotifyDefaultCaptureChanged(present bool) {
	if captureHandler != nil {
		captureHandler(present)
	}
}
func NotifyGotLogMessage(level, content string) {
	if logHandler != nil {
		logHandler(level, content)
	}
}

func Initialize(appName, appVersion string) (Handle, error) {
	var h C.SaaHandle
	var cAppName, cAppVersion *C.char
	if appName != "" {
		cAppName = C.CString(appName)
		defer C.free(unsafe.Pointer(cAppName))
	}
	if appVersion != "" {
		cAppVersion = C.CString(appVersion)
		defer C.free(unsafe.Pointer(cAppVersion))
	}
	rc := C.SaaInitialize(
		&h,
		(C.TSaaGotLogMessageCallback)(C.cgoSaaGotLogMessage),
		cAppName,
		cAppVersion,
	)
	if rc != 0 {
		return 0, fmt.Errorf("SaaInitialize failed: rc=%d", int32(rc))
	}
	return Handle(h), nil
}

func RegisterCallbacks(h Handle, onRender, onCapture bool) error {
	var renderCb C.TSaaDefaultChangedCallback
	var captureCb C.TSaaDefaultChangedCallback
	if onRender {
		renderCb = (C.TSaaDefaultChangedCallback)(C.cgoSaaDefaultRenderChanged)
	}
	if onCapture {
		captureCb = (C.TSaaDefaultChangedCallback)(C.cgoSaaDefaultCaptureChanged)
	}
	rc := C.SaaRegisterCallbacks(C.SaaHandle(h), renderCb, captureCb)
	if rc != 0 {
		return fmt.Errorf("SaaRegisterCallbacks failed: rc=%d", int32(rc))
	}
	return nil
}

func GetDefaultRender(h Handle) (Description, error) {
	var cd C.SaaDescription
	rc := C.SaaGetDefaultRender(C.SaaHandle(h), &cd)
	if rc != 0 {
		return Description{}, fmt.Errorf("SaaGetDefaultRender failed: rc=%d", int32(rc))
	}
	return fromCDesc(&cd), nil
}

func GetDefaultCapture(h Handle) (Description, error) {
	var cd C.SaaDescription
	rc := C.SaaGetDefaultCapture(C.SaaHandle(h), &cd)
	if rc != 0 {
		return Description{}, fmt.Errorf("SaaGetDefaultCapture failed: rc=%d", int32(rc))
	}
	return fromCDesc(&cd), nil
}

func Uninitialize(h Handle) error {
	if h == 0 {
		return errors.New("invalid handle")
	}
	rc := C.SaaUnInitialize(C.SaaHandle(h))
	if rc != 0 {
		return fmt.Errorf("SaaUnInitialize failed: rc=%d", int32(rc))
	}
	return nil
}

func fromCDesc(cd *C.SaaDescription) Description {
	return Description{
		PnpID:         C.GoString(&cd.PnpId[0]),
		Name:          C.GoString(&cd.Name[0]),
		IsRender:      cd.IsRender != 0,
		IsCapture:     cd.IsCapture != 0,
		RenderVolume:  uint16(cd.RenderVolume),
		CaptureVolume: uint16(cd.CaptureVolume),
	}
}
