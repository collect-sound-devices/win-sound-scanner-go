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

func New(enqueue func(c.EventType, map[string]string), logInfo, logError logging.Logf) (ScannerApp, error) {
	return NewImpl(enqueue, logInfo, logError)
}

func Run(ctx context.Context, bridgeLogf, logInfo, logError logging.Logf) error {
	if bridgeLogf == nil {
		panic("nil bridge logf")
	}
	if logInfo == nil {
		panic("nil info logger")
	}
	if logError == nil {
		panic("nil error logger")
	}

	reqEnqueuer, cleanupEnqueuer, err := newRequestEnqueuer(ctx, logInfo)
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
			logError("enqueue failed: %v", err)
		}
	}

	logging.AttachSoundlibwrapBridge(bridgeLogf, "cpp backend,")

	logInfo("Initializing...")

	app, err := New(enqueue, logInfo, logError)
	if err != nil {
		return err
	}
	defer app.Shutdown()

	// Keep running until interrupted to receive async logs and change events.
	<-ctx.Done()
	logInfo("Shutting down...")
	return nil
}

func newRequestEnqueuer(ctx context.Context, logf logging.Logf) (enqueuer.EnqueueRequest, func(), error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv(EnvWinSoundEnqueuer)))

	// Return a no-op enqueuer for testing or when RabbitMQ is not available.
	if mode == "empty" {
		return enqueuer.NewEmptyRequestEnqueuer(logf), func() {}, nil
	}

	// Validate that the configured mode is supported.
	if mode != "" && mode != "rabbitmq" {
		return nil, nil, fmt.Errorf("unsupported %s=%q (supported: empty, rabbitmq)", EnvWinSoundEnqueuer, mode)
	}

	cfg, err := rabbitmq.LoadConfigFromEnv()
	if err != nil {
		return nil, nil, err
	}

	publisher, err := rabbitmq.NewRequestPublisher(ctx, cfg, logf)
	if err != nil {
		return nil, nil, err
	}

	reqEnqueuer := rabbitmq.NewRabbitMqEnqueuerWithContext(ctx, publisher, logf)
	cleanup := func() {
		if err := reqEnqueuer.Close(); err != nil {
			logf("[error] rabbitmq enqueuer close failed: %v", err)
		}
	}

	return reqEnqueuer, cleanup, nil
}
