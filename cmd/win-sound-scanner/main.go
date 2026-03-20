package main

import (
	"os"
	"strings"

	"github.com/kardianos/service"
)

func main() {
	logger := newStderrLogger()

	if len(os.Args) > 1 {
		cmd := strings.ToLower(strings.TrimSpace(os.Args[1]))
		if !isServiceCommand(cmd) {
			fatalLog(logger, "unsupported command", "command", cmd, "supported", "install, uninstall, start, stop, restart")
		}

		svc, err := newService()
		if err != nil {
			fatalLog(logger, "service initialization failed", "err", err)
		}
		if err := service.Control(svc, cmd); err != nil {
			fatalLog(logger, "service command failed", "command", cmd, "err", err)
		}
		return
	}

	if service.Interactive() {
		if err := runConsole(); err != nil {
			fatalLog(logger, "exit with error", "err", err)
		}
		return
	}

	svc, err := newService()
	if err != nil {
		fatalLog(logger, "service initialization failed", "err", err)
	}
	if err := svc.Run(); err != nil {
		fatalLog(logger, "service run failed", "err", err)
	}
}
