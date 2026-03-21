package enqueuer

import (
	"time"

	"github.com/collect-sound-devices/win-sound-scanner-go/internal/contract"
)

type Request struct {
	Timestamp time.Time
	Event     contract.EventType
	Fields    map[string]string
}

type EnqueueRequest interface {
	EnqueueRequest(request Request) error
}
