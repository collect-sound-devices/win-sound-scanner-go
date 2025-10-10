//go:build windows && cgo

package saawrapper

/*
#include <stdint.h>
#include <stdbool.h>
#include "SoundAgentApi.h"
*/
import "C"

//export cgoSaaDefaultRenderChanged
func cgoSaaDefaultRenderChanged(present C.int) {
	NotifyDefaultRenderChanged(present != 0)
}

//export cgoSaaDefaultCaptureChanged
func cgoSaaDefaultCaptureChanged(present C.int) {
	NotifyDefaultCaptureChanged(present != 0)
}

//export cgoSaaGotLogMessage
func cgoSaaGotLogMessage(msg C.SaaLogMessage) {
	level := C.GoString(&msg.Level[0])
	content := C.GoString(&msg.Content[0])
	NotifyGotLogMessage(level, content)
}
