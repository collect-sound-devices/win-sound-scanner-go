package scannerapp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	c "github.com/collect-sound-devices/win-sound-go-bridge/internal/contract"
	"github.com/collect-sound-devices/win-sound-go-bridge/internal/enqueuer"
	"github.com/collect-sound-devices/win-sound-go-bridge/internal/logging"
	"github.com/collect-sound-devices/win-sound-go-bridge/internal/rabbitmq"
)

func NewWithLoggers(enqueue func(c.EventType, map[string]string), infoLogger, errorLogger logging.Logger) (ScannerApp, error) {
	return NewImpl(enqueue, infoLogger, errorLogger)
}

func Run(ctx context.Context) error {
	appLogger := logging.NewLogger("")
	infoLogger := logging.NewLogger("[info] ")
	errorLogger := logging.NewLogger("[error] ")
	reqEnqueuer, cleanupEnqueuer, err := newRequestEnqueuer(ctx, appLogger)
	if err != nil {
		return err
	}
	defer cleanupEnqueuer()

	enqueue := func(event c.EventType, fields map[string]string) {
		if err := reqEnqueuer.EnqueueRequest(enqueuer.Request{
			Timestamp: time.Now(),
			Event:     event,
			Fields:    fields,
		}); err != nil {
			errorLogger.Printf("enqueue failed: %v", err)
		}
	}

	{
		logging.AttachSoundlibwrapBridge(logging.NewPlainLogger(), "cpp backend,")
	}

	infoLogger.Printf("Initializing...")

	app, err := NewWithLoggers(enqueue, infoLogger, errorLogger)
	if err != nil {
		return err
	}
	defer app.Shutdown()

	// Keep running until interrupted to receive async logs and change events.
	<-ctx.Done()
	infoLogger.Printf("Shutting down...")
	return nil
}

func newRequestEnqueuer(ctx context.Context, logger logging.Logger) (enqueuer.EnqueueRequest, func(), error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv(EnvWinSoundEnqueuer)))

	// Return a no-op enqueuer for testing or when RabbitMQ is not available.
	if mode == "empty" {
		return enqueuer.NewEmptyRequestEnqueuer(logger), func() {}, nil
	}

	// Validate that the configured mode is supported.
	if mode != "" && mode != "rabbitmq" {
		return nil, nil, fmt.Errorf("unsupported %s=%q (supported: empty, rabbitmq)", EnvWinSoundEnqueuer, mode)
	}

	cfg, err := rabbitmq.LoadConfigFromEnv()
	if err != nil {
		return nil, nil, err
	}

	publisher, err := rabbitmq.NewRequestPublisher(ctx, cfg, logger)
	if err != nil {
		return nil, nil, err
	}

	reqEnqueuer := rabbitmq.NewRabbitMqEnqueuerWithContext(ctx, publisher, logger)
	cleanup := func() {
		if err := reqEnqueuer.Close(); err != nil {
			logger.Printf("[error] rabbitmq enqueuer close failed: %v", err)
		}
	}

	return reqEnqueuer, cleanup, nil
}
