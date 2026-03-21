package enqueuer

import (
	"log/slog"
)

type EmptyRequestEnqueuer struct {
	logger *slog.Logger
}

func NewEmptyRequestEnqueuer(logger *slog.Logger) *EmptyRequestEnqueuer {
	if logger == nil {
		panic("nil logger")
	}
	return &EmptyRequestEnqueuer{logger: logger}
}

func (e *EmptyRequestEnqueuer) EnqueueRequest(request Request) error {
	e.logger.Info("Dropping request in empty enqueuer", "event", request.Event, "fields", request.Fields)
	return nil
}
