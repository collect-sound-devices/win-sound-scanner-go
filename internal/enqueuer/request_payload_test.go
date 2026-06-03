package enqueuer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/collect-sound-devices/win-sound-scanner-go/internal/contract"
)

func TestBuildRequestPayload_DiscoveredEventProducesPost(t *testing.T) {
	request := Request{
		Timestamp: time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC),
		Event:     contract.EventTypeRenderDeviceDiscovered,
		Fields: map[string]string{
			contract.FieldName:     "Realtek Audio",
			contract.FieldPnpID:    "pnp-1",
			contract.FieldHostName: "host-1",
		},
	}

	result, err := BuildRequestPayload(request)
	if err != nil {
		t.Fatalf("BuildRequestPayload failed: %v", err)
	}
	payload := decodePayload(t, result.Body)

	assertString(t, result.HTTPRequest, "POST")
	assertString(t, result.URLSuffix, "")
	assertString(t, result.DeviceKey, "host-1|pnp-1")
	assertString(t, payload[contract.FieldHTTPRequest], "POST")
	assertNumber(t, payload[contract.FieldDeviceMessageType], float64(contract.MessageTypeDiscovered))
	assertNumber(t, payload[contract.FieldFlowType], float64(contract.FlowTypeRender))
	assertString(t, payload[contract.FieldUpdateDate], "2026-05-26T10:00:00Z")
}

func TestBuildRequestPayload_VolumeEventProducesPutAndUrlSuffix(t *testing.T) {
	request := Request{
		Timestamp: time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC),
		Event:     contract.EventTypeRenderVolumeChanged,
		Fields: map[string]string{
			contract.FieldPnpID:        "pnp-1",
			contract.FieldHostName:     "host-1",
			contract.FieldRenderVolume: " 42 ",
		},
	}

	result, err := BuildRequestPayload(request)
	if err != nil {
		t.Fatalf("BuildRequestPayload failed: %v", err)
	}
	payload := decodePayload(t, result.Body)

	assertString(t, result.HTTPRequest, "PUT")
	assertString(t, result.URLSuffix, "/pnp-1/host-1")
	assertString(t, payload[contract.FieldHTTPRequest], "PUT")
	assertString(t, payload[contract.FieldURLSuffix], "/pnp-1/host-1")
	assertNumber(t, payload[contract.FieldDeviceMessageType], float64(contract.MessageTypeVolumeRenderChanged))
	assertNumber(t, payload[contract.FieldRenderVolume], 42)
	if _, ok := payload[contract.FieldHostName]; ok {
		t.Fatalf("expected %q to be removed from PUT payload", contract.FieldHostName)
	}
}

func TestBuildRequestPayload_ExplicitUpdateDateIsPreserved(t *testing.T) {
	request := Request{
		Timestamp: time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC),
		Event:     contract.EventTypeCaptureDeviceConfirmed,
		Fields: map[string]string{
			contract.FieldUpdateDate: "2026-05-25T09:00:00Z",
		},
	}

	result, err := BuildRequestPayload(request)
	if err != nil {
		t.Fatalf("BuildRequestPayload failed: %v", err)
	}
	payload := decodePayload(t, result.Body)

	assertString(t, payload[contract.FieldUpdateDate], "2026-05-25T09:00:00Z")
	assertNumber(t, payload[contract.FieldDeviceMessageType], float64(contract.MessageTypeConfirmed))
	assertNumber(t, payload[contract.FieldFlowType], float64(contract.FlowTypeCapture))
}

func decodePayload(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	return payload
}

func assertString(t *testing.T, actual any, expected string) {
	t.Helper()
	if actual != expected {
		t.Fatalf("expected %q, got %#v", expected, actual)
	}
}

func assertNumber(t *testing.T, actual any, expected float64) {
	t.Helper()
	if actual != expected {
		t.Fatalf("expected %v, got %#v", expected, actual)
	}
}
