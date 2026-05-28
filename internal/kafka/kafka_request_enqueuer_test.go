package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/collect-sound-devices/win-sound-scanner-go/internal/contract"
	"github.com/collect-sound-devices/win-sound-scanner-go/internal/enqueuer"
)

type fakePublisher struct {
	key  []byte
	body []byte
	err  error
}

func (p *fakePublisher) Publish(_ context.Context, key []byte, body []byte) error {
	p.key = append([]byte(nil), key...)
	p.body = append([]byte(nil), body...)
	return p.err
}

func (p *fakePublisher) Close() error {
	return nil
}

func TestEnqueueRequest_PublishesPayloadWithDeviceKey(t *testing.T) {
	publisher := &fakePublisher{}
	sut := NewEnqueuerWithContext(context.Background(), publisher, slog.Default(), time.Second)

	err := sut.EnqueueRequest(enqueuer.Request{
		Timestamp: time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC),
		Event:     contract.EventTypeRenderDeviceDiscovered,
		Fields: map[string]string{
			contract.FieldPnpID:    "pnp-1",
			contract.FieldHostName: "host-1",
		},
	})
	if err != nil {
		t.Fatalf("EnqueueRequest failed: %v", err)
	}

	if string(publisher.key) != "host-1|pnp-1" {
		t.Fatalf("unexpected key: %q", string(publisher.key))
	}

	var payload map[string]any
	if err := json.Unmarshal(publisher.body, &payload); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	if payload[contract.FieldHTTPRequest] != "POST" {
		t.Fatalf("expected POST payload, got %#v", payload[contract.FieldHTTPRequest])
	}
}

func TestEnqueueRequest_PropagatesPublisherError(t *testing.T) {
	publisher := &fakePublisher{err: errors.New("boom")}
	sut := NewEnqueuerWithContext(context.Background(), publisher, slog.Default(), time.Second)

	err := sut.EnqueueRequest(enqueuer.Request{
		Event:  contract.EventTypeRenderDeviceDiscovered,
		Fields: map[string]string{},
	})
	if err == nil {
		t.Fatal("expected publisher error")
	}
}
