package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/collect-sound-devices/win-sound-scanner-go/internal/enqueuer"
)

// RabbitMessagePublisher is the publishing contract expected from a RabbitMQ publisher.
type RabbitMessagePublisher interface {
	Publish(ctx context.Context, body []byte) error
	Close() error
}

// RabbitMqEnqueuer writes requests to RabbitMQ using the shared message-shaping.
type RabbitMqEnqueuer struct {
	baseCtx        context.Context
	publisher      RabbitMessagePublisher
	logger         *slog.Logger
	publishTimeout time.Duration
}

func NewRabbitMqEnqueuerWithContext(baseCtx context.Context, publisher RabbitMessagePublisher, logger *slog.Logger) *RabbitMqEnqueuer {
	if baseCtx == nil {
		panic("nil context")
	}
	if publisher == nil {
		panic("nil publisher")
	}
	if logger == nil {
		panic("nil logger")
	}

	return newRabbitMqEnqueuer(
		baseCtx,
		publisher,
		logger,
		10*time.Second,
	)
}

func newRabbitMqEnqueuer(
	baseCtx context.Context,
	publisher RabbitMessagePublisher,
	logger *slog.Logger,
	publishTimeout time.Duration,
) *RabbitMqEnqueuer {
	if baseCtx == nil {
		panic("nil context")
	}
	if publisher == nil {
		panic("nil publisher")
	}
	if logger == nil {
		panic("nil logger")
	}
	if publishTimeout <= 0 {
		publishTimeout = 10 * time.Second
	}

	return &RabbitMqEnqueuer{
		baseCtx:        baseCtx,
		publisher:      publisher,
		logger:         logger,
		publishTimeout: publishTimeout,
	}
}

func (e *RabbitMqEnqueuer) EnqueueRequest(request enqueuer.Request) error {
	e.logger.Info("Preparing request in RabbitMQ enqueuer", "event", request.Event, "fields", request.Fields)
	payload, err := enqueuer.BuildRequestPayload(request)
	if err != nil {
		return fmt.Errorf("marshal rabbitmq payload: %w", err)
	}

	e.logger.Info("publishing request", "method", payload.HTTPRequest, "urlSuffix", payload.URLSuffix, "updated", payload.UpdateDateUtc)

	ctx, cancel := context.WithTimeout(e.baseCtx, e.publishTimeout)
	defer cancel()
	if err := e.publisher.Publish(ctx, payload.Body); err != nil {
		return fmt.Errorf("publish request: %w", err)
	}

	return nil
}

func (e *RabbitMqEnqueuer) Close() error {
	return e.publisher.Close()
}
