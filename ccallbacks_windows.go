//go:build windows && cgo

package main

/*
#include <stdint.h>
#include <stdbool.h>

// Keep this in sync with the declaration in internal/saa.
typedef struct {
    char Level[12];
    char Content[256];
} SaaLogMessage;
*/
import "C"

import (
	"unsafe"

	"win-sound-dev-go-bridge/internal/saawrapper"
)

//export cgoSaaDefaultRenderChanged
func cgoSaaDefaultRenderChanged(present C.int) {
	saawrapper.NotifyDefaultRenderChanged(present != 0)
}

//export cgoSaaDefaultCaptureChanged
func cgoSaaDefaultCaptureChanged(present C.int) {
	saawrapper.NotifyDefaultCaptureChanged(present != 0)
}

//export cgoSaaGotLogMessage
func cgoSaaGotLogMessage(msg C.SaaLogMessage) {
	level := C.GoString(&msg.Level[0])
	content := C.GoString(&msg.Content[0])
	_ = unsafe.Pointer(nil) // silence potential unused import on some toolchains
	saawrapper.NotifyGotLogMessage(level, content)
}
