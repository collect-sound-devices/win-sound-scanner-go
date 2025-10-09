package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"win-sound-dev-go-bridge/internal/app"
)

var (
	modOle32           = syscall.NewLazyDLL("ole32.dll")
	procCoInitializeEx = modOle32.NewProc("CoInitializeEx")
	procCoUninitialize = modOle32.NewProc("CoUninitialize")
)

const (
	COINIT_APARTMENTTHREADED = 0x2 // Single-threaded apartment
	COINIT_MULTITHREADED     = 0x0 // Multi-threaded apartment
)

func CoInitializeEx(coInit uintptr) error {
	ret, _, _ := procCoInitializeEx.Call(0, coInit)
	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}

func CoUninitialize() {
	procCoUninitialize.Call()
}

func main() {
	err := CoInitializeEx(COINIT_MULTITHREADED)
	if err != nil {
		log.Fatalf("CoInitializeEx failed: %v", err)
	}
	defer CoUninitialize()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := app.Run(ctx); err != nil {
		log.Fatalf("exit with error: %v", err)
	}
}
