package main

import (
	"log"
	"os"
	"strings"

	"github.com/kardianos/service"
)

func main() {
	if len(os.Args) > 1 {
		cmd := strings.ToLower(strings.TrimSpace(os.Args[1]))
		if !isServiceCommand(cmd) {
			log.Fatalf("[error] unsupported command %q (supported: install, uninstall, start, stop, restart)", cmd)
		}

		svc, err := newService()
		if err != nil {
			log.Fatalf("[error] service initialization failed: %v", err)
		}
		if err := service.Control(svc, cmd); err != nil {
			log.Fatalf("[error] service command %q failed: %v", cmd, err)
		}
		return
	}

	if service.Interactive() {
		if err := runConsole(); err != nil {
			log.Fatalf("[error] exit with error: %v", err)
		}
		return
	}

	svc, err := newService()
	if err != nil {
		log.Fatalf("[error] service initialization failed: %v", err)
	}
	if err := svc.Run(); err != nil {
		log.Fatalf("[error] service run failed: %v", err)
	}
}
