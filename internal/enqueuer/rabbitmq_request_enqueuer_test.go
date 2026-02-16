package enqueuer

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/contract"
)

type testPublisher struct {
	lastBody []byte
	calls    int
	pubErr   error
}

type testLogger struct{}

func (l testLogger) Printf(string, ...interface{}) {}

func (p *testPublisher) Publish(_ context.Context, body []byte) error {
	p.calls++
	p.lastBody = append([]byte(nil), body...)
	return p.pubErr
}

func (p *testPublisher) Close() error {
	return nil
}

func TestRabbitMqEnqueuerPostDevice(t *testing.T) {
	pub := &testPublisher{}
	enq := newRabbitMqEnqueuer(context.Background(), pub, testLogger{}, "hostA", "windows", time.Second)

	err := enq.EnqueueRequest(Request{
		Name:      contract.RequestPostDevice,
		Timestamp: time.Date(2026, 2, 10, 9, 8, 7, 0, time.UTC),
		Fields: map[string]string{
			contract.FieldDeviceMessageType: contract.EventDefaultRenderChanged,
			contract.FieldUpdateDate:        "2026-02-10T09:08:07Z",
			contract.FieldFlowType:          contract.FlowRender,
			contract.FieldName:              "Speaker",
			contract.FieldPnpID:             "PNP-1",
			contract.FieldRenderVolume:      "32",
			contract.FieldCaptureVolume:     "0",
		},
	})
	if err != nil {
		t.Fatalf("EnqueueRequest error: %v", err)
	}

	if pub.calls != 1 {
		t.Fatalf("expected 1 publish call, got %d", pub.calls)
	}

	var got map[string]any
	if err := json.Unmarshal(pub.lastBody, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if got[contract.FieldHTTPRequestCamel] != "POST" {
		t.Fatalf("expected httpRequest=POST, got %#v", got[contract.FieldHTTPRequestCamel])
	}
	if got[contract.FieldURLSuffixCamel] != "" {
		t.Fatalf("expected empty urlSuffix, got %#v", got[contract.FieldURLSuffixCamel])
	}
	if got[contract.FieldHostNameCamel] != "hostA" {
		t.Fatalf("expected hostName=hostA, got %#v", got[contract.FieldHostNameCamel])
	}
	if got[contract.FieldOperationSystemNameCamel] != "windows" {
		t.Fatalf("expected operationSystemName=windows, got %#v", got[contract.FieldOperationSystemNameCamel])
	}
	if got[contract.FieldRenderVolumeCamel] != float64(32) {
		t.Fatalf("expected renderVolume=32, got %#v", got[contract.FieldRenderVolumeCamel])
	}
	if got[contract.FieldCaptureVolumeCamel] != float64(0) {
		t.Fatalf("expected captureVolume=0, got %#v", got[contract.FieldCaptureVolumeCamel])
	}
}

func TestRabbitMqEnqueuerPutVolumeChange(t *testing.T) {
	pub := &testPublisher{}
	enq := newRabbitMqEnqueuer(context.Background(), pub, testLogger{}, "myHost", "windows", time.Second)

	err := enq.EnqueueRequest(Request{
		Name:      contract.RequestPutVolumeChange,
		Timestamp: time.Date(2026, 2, 10, 9, 8, 7, 0, time.UTC),
		Fields: map[string]string{
			contract.FieldDeviceMessageType: contract.EventRenderVolumeChanged,
			contract.FieldVolume:            "44",
			contract.FieldPnpID:             "DEV-22",
		},
	})
	if err != nil {
		t.Fatalf("EnqueueRequest error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(pub.lastBody, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if got[contract.FieldHTTPRequestCamel] != "PUT" {
		t.Fatalf("expected httpRequest=PUT, got %#v", got[contract.FieldHTTPRequestCamel])
	}
	if got[contract.FieldURLSuffixCamel] != "/DEV-22/myHost" {
		t.Fatalf("expected urlSuffix=/DEV-22/myHost, got %#v", got[contract.FieldURLSuffixCamel])
	}
	if got[contract.FieldUpdateDateCamel] != "2026-02-10T09:08:07Z" {
		t.Fatalf("expected fallback updateDate from timestamp, got %#v", got[contract.FieldUpdateDateCamel])
	}
	if got[contract.FieldVolume] != float64(44) {
		t.Fatalf("expected volume=44, got %#v", got[contract.FieldVolume])
	}
}
