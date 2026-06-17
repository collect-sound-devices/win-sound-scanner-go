package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/collect-sound-devices/win-sound-scanner-go/internal/enqueuer"
)

type MessagePublisher interface {
	Publish(ctx context.Context, key []byte, body []byte) error
	Close() error
}

type Enqueuer struct {
	baseCtx        context.Context
	publisher      MessagePublisher
	logger         *slog.Logger
	publishTimeout time.Duration
}

func NewEnqueuerWithContext(baseCtx context.Context, publisher MessagePublisher, logger *slog.Logger, publishTimeout time.Duration) *Enqueuer {
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
		publishTimeout = defaultWriteTimeout
	}

	return &Enqueuer{
		baseCtx:        baseCtx,
		publisher:      publisher,
		logger:         logger,
		publishTimeout: publishTimeout,
	}
}

func (e *Enqueuer) EnqueueRequest(request enqueuer.Request) error {
	e.logger.Info("Preparing request in Kafka enqueuer", "event", request.Event, "fields", request.Fields)
	payload, err := enqueuer.BuildRequestPayload(request)
	if err != nil {
		return fmt.Errorf("marshal kafka payload: %w", err)
	}

	e.logger.Info("publishing event", "method", payload.HTTPRequest, "urlSuffix", payload.URLSuffix, "key", payload.DeviceKey, "updated", payload.UpdateDateUtc)

	ctx, cancel := context.WithTimeout(e.baseCtx, e.publishTimeout)
	defer cancel()
	if err := e.publisher.Publish(ctx, []byte(payload.DeviceKey), payload.Body); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}

	return nil
}

func (e *Enqueuer) Close() error {
	return e.publisher.Close()
}
