package main

import (
	"os"
	"strings"

	"github.com/kardianos/service"
)

func main() {
	stderrLogger := newAppLogger(os.Stderr)

	if len(os.Args) > 1 {
		cmd := strings.ToLower(strings.TrimSpace(os.Args[1]))
		if !isServiceCommand(cmd) {
			fatalLog(stderrLogger, "unsupported command", "command", cmd, "supported", "install, uninstall, start, stop, restart")
		}

		svc, err := newService()
		if err != nil {
			fatalLog(stderrLogger, "service initialization failed", "err", err)
		}
		if err := service.Control(svc, cmd); err != nil {
			fatalLog(stderrLogger, "service command failed", "command", cmd, "err", err)
		}
		return
	}

	if service.Interactive() {
		if err := runConsole(); err != nil {
			fatalLog(stderrLogger, "exit with error", "err", err)
		}
		return
	}

	svc, err := newService()
	if err != nil {
		fatalLog(stderrLogger, "service initialization failed", "err", err)
	}
	if err := svc.Run(); err != nil {
		fatalLog(stderrLogger, "service run failed", "err", err)
	}
}
