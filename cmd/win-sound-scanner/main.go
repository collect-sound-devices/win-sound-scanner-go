package main

import (
	"os"
	"strings"

	"github.com/kardianos/service"
)

func main() {
	if len(os.Args) > 1 {
		cmd := strings.ToLower(strings.TrimSpace(os.Args[1]))
		if !isServiceCommand(cmd) {
			errorLogger.Fatalf("unsupported command %q (supported: install, uninstall, start, stop, restart)", cmd)
		}

		svc, err := newService()
		if err != nil {
			errorLogger.Fatalf("service initialization failed: %v", err)
		}
		if err := service.Control(svc, cmd); err != nil {
			errorLogger.Fatalf("service command %q failed: %v", cmd, err)
		}
		return
	}

	if service.Interactive() {
		if err := runConsole(); err != nil {
			errorLogger.Fatalf("exit with error: %v", err)
		}
		return
	}

	svc, err := newService()
	if err != nil {
		errorLogger.Fatalf("service initialization failed: %v", err)
	}
	if err := svc.Run(); err != nil {
		errorLogger.Fatalf("service run failed: %v", err)
	}
}
