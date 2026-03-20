package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kardianos/service"

	"github.com/collect-sound-devices/win-sound-scanner-go/internal/scannerapp"
)

const (
	serviceName        = "WinSoundScanner"
	serviceDisplayName = "Win Sound Scanner"
	serviceDescription = "Collects Windows default sound devices and publishes events."
	serviceLogFileName = "service.log"
)

var serviceEnvKeys = []string{
	scannerapp.EnvWinSoundEnqueuer,
	scannerapp.EnvWinSoundRabbitMQHost,
	scannerapp.EnvWinSoundRabbitMQPort,
	scannerapp.EnvWinSoundRabbitMQVHost,
	scannerapp.EnvWinSoundRabbitMQUser,
	scannerapp.EnvWinSoundRabbitMQPassword,
	scannerapp.EnvWinSoundRabbitMQExchange,
	scannerapp.EnvWinSoundRabbitMQQueue,
	scannerapp.EnvWinSoundRabbitMQRoutingKey,
}

type scannerProgram struct {
	mu      sync.Mutex
	cancel  context.CancelFunc
	done    chan struct{}
	logFile *os.File
}

func (p *scannerProgram) Start(_ service.Service) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.done != nil {
		return nil
	}

	logFile, err := configureServiceFileLogging()
	if err != nil {
		return err
	}
	logger := newAppLogger(logFile)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	p.cancel = cancel
	p.done = done
	p.logFile = logFile

	go func(logger *slog.Logger) {
		defer close(done)
		if err := runScanner(ctx, logger); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("scanner failed", "err", err)
			os.Exit(1)
		}
	}(logger)

	return nil
}

func (p *scannerProgram) Stop(_ service.Service) error {
	p.mu.Lock()
	cancel := p.cancel
	done := p.done
	logFile := p.logFile
	p.cancel = nil
	p.done = nil
	p.logFile = nil
	p.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
	if logFile != nil {
		_ = logFile.Close()
	}
	return nil
}

func newService() (service.Service, error) {
	envVars := collectServiceEnvVars()
	if len(envVars) == 0 {
		envVars = nil
	}

	cfg := &service.Config{
		Name:        serviceName,
		DisplayName: serviceDisplayName,
		Description: serviceDescription,
		Option: service.KeyValue{
			"StartType": "automatic",
			"OnFailure": "restart",
		},
		EnvVars: envVars,
	}

	return service.New(&scannerProgram{}, cfg)
}

func collectServiceEnvVars() map[string]string {
	envVars := make(map[string]string, len(serviceEnvKeys))
	for _, key := range serviceEnvKeys {
		if value, ok := os.LookupEnv(key); ok {
			envVars[key] = value
		}
	}
	return envVars
}

func isServiceCommand(cmd string) bool {
	switch cmd {
	case "install", "uninstall", "start", "stop", "restart":
		return true
	default:
		return false
	}
}

func programDataDir() (string, error) {
	if v, ok := os.LookupEnv("ProgramData"); ok && strings.TrimSpace(v) != "" {
		return v, nil
	}
	if v, ok := os.LookupEnv("ALLUSERSPROFILE"); ok && strings.TrimSpace(v) != "" {
		return v, nil
	}
	return "", errors.New("ProgramData is not available in environment")
}

func configureServiceFileLogging() (*os.File, error) {
	baseDir, err := programDataDir()
	if err != nil {
		return nil, err
	}

	logDir := filepath.Join(baseDir, serviceName)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("create service log directory: %w", err)
	}

	logPath := filepath.Join(logDir, serviceLogFileName)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open service log file: %w", err)
	}

	log.SetOutput(logFile)
	os.Stdout = logFile
	os.Stderr = logFile
	return logFile, nil
}
