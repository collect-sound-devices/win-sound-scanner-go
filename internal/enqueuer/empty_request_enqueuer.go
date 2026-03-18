package enqueuer

import (
	"github.com/collect-sound-devices/win-sound-go-bridge/internal/logging"
)

type EmptyRequestEnqueuer struct {
	logf logging.Logf
}

func NewEmptyRequestEnqueuer(logf logging.Logf) *EmptyRequestEnqueuer {
	if logf == nil {
		panic("nil logf")
	}
	return &EmptyRequestEnqueuer{logf: logf}
}

func (e *EmptyRequestEnqueuer) EnqueueRequest(request Request) error {
	e.logf(
		"[info, empty enqueuer] event=%d fields=%v",
		request.Event,
		request.Fields,
	)
	return nil
}
