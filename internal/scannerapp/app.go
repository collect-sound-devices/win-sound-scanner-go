package scannerapp

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	c "github.com/collect-sound-devices/win-sound-scanner-go/internal/contract"
	"github.com/collect-sound-devices/win-sound-scanner-go/internal/enqueuer"
	"github.com/collect-sound-devices/win-sound-scanner-go/internal/rabbitmq"
)

// WithComponent adds a component attribute when one is provided.
func WithComponent(logger *slog.Logger, component string) *slog.Logger {
	if logger == nil {
		panic("nil logger")
	}
	component = strings.TrimSpace(component)
	if component == "" {
		return logger
	}
	return logger.With("component", component)
}

func Run(ctx context.Context, logger *slog.Logger) error {
	if ctx == nil {
		panic("nil context")
	}
	if logger == nil {
		panic("nil logger")
	}

	appLogger := WithComponent(logger, "scannerapp")

	reqEnqueuer, cleanupEnqueuer, err := newRequestEnqueuer(ctx, logger)
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
			appLogger.Error("enqueue failed", "event", event, "err", err)
		}
	}

	appLogger.Info("Initializing")

	app, err := NewImpl(enqueue, appLogger)
	if err != nil {
		return err
	}
	defer app.Shutdown()

	// Keep running until interrupted to receive async logs and change events.
	<-ctx.Done()
	appLogger.Info("Shutting down")
	return nil
}

func newRequestEnqueuer(ctx context.Context, logger *slog.Logger) (enqueuer.EnqueueRequest, func(), error) {
	if ctx == nil {
		panic("nil context")
	}
	if logger == nil {
		panic("nil logger")
	}

	mode := strings.ToLower(strings.TrimSpace(os.Getenv(EnvWinSoundEnqueuer)))
	requestLogger := WithComponent(logger, "request_enqueuer")

	// Return a no-op enqueuer for testing or when RabbitMQ is not available.
	if mode == "empty" {
		return enqueuer.NewEmptyRequestEnqueuer(WithComponent(logger, "empty_request_enqueuer")), func() {}, nil
	}

	// Validate that the configured mode is supported.
	if mode != "" && mode != "rabbitmq" {
		return nil, nil, fmt.Errorf("unsupported %s=%q (supported: empty, rabbitmq)", EnvWinSoundEnqueuer, mode)
	}

	cfg, err := rabbitmq.LoadConfigFromEnv()
	if err != nil {
		return nil, nil, err
	}

	publisher, err := rabbitmq.NewRequestPublisher(ctx, cfg, WithComponent(logger, "rabbitmq_publisher"))
	if err != nil {
		return nil, nil, err
	}

	reqEnqueuer := rabbitmq.NewRabbitMqEnqueuerWithContext(ctx, publisher, WithComponent(logger, "rabbitmq_enqueuer"))
	cleanup := func() {
		if err := reqEnqueuer.Close(); err != nil {
			requestLogger.Error("rabbitmq enqueuer close failed", "err", err)
		}
	}

	return reqEnqueuer, cleanup, nil
}
