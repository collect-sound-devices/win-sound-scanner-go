package enqueuer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/collect-sound-devices/win-sound-scanner-go/internal/contract"
)

type RequestPayload struct {
	Body          []byte
	HTTPRequest   string
	URLSuffix     string
	DeviceKey     string
	UpdateDateUtc string
}

func BuildRequestPayload(request Request) (RequestPayload, error) {
	payload := make(map[string]any, len(request.Fields)+4)
	for key, value := range request.Fields {
		payload[key] = normalizeValue(key, value)
	}

	deviceKey := buildDeviceKey(request.Fields)
	flowType, messageType := calculateFlowAndMessageType(request.Event)
	payload[contract.FieldDeviceMessageType] = messageType

	httpRequest, urlSuffix := resolveHttpRequest(request, payload)
	payload[contract.FieldHTTPRequest] = httpRequest
	payload[contract.FieldURLSuffix] = urlSuffix

	if httpRequest == "POST" && flowType != 0 {
		payload[contract.FieldFlowType] = flowType
	}
	var updateDateUtc = request.Timestamp.UTC().Format(time.RFC3339)
	if _, ok := payload[contract.FieldUpdateDate]; ok {
		updateDateUtc = payload[contract.FieldUpdateDate].(string)
	} else {
		payload[contract.FieldUpdateDate] = updateDateUtc
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return RequestPayload{}, fmt.Errorf("marshal request payload: %w", err)
	}

	return RequestPayload{
		Body:          body,
		HTTPRequest:   httpRequest,
		URLSuffix:     urlSuffix,
		DeviceKey:     deviceKey,
		UpdateDateUtc: updateDateUtc,
	}, nil
}

func resolveHttpRequest(request Request, payload map[string]any) (string, string) {
	var httpRequest string

	switch request.Event {
	case contract.EventTypeRenderDeviceDiscovered,
		contract.EventTypeCaptureDeviceDiscovered,
		contract.EventTypeRenderDeviceConfirmed,
		contract.EventTypeCaptureDeviceConfirmed:
		httpRequest = "POST"
	default:
		httpRequest = "PUT"
	}

	urlSuffix := readStringField(payload, contract.FieldURLSuffix)
	if urlSuffix == "" && httpRequest == "PUT" {
		pnpID := readStringField(payload, contract.FieldPnpID)
		hostName := readStringField(payload, contract.FieldHostName)

		urlSuffix = fmt.Sprintf("/%s/%s", pnpID, hostName)
		delete(payload, contract.FieldHostName)
	}

	return httpRequest, urlSuffix
}

func readStringField(payload map[string]any, key string) string {
	if v, ok := payload[key]; ok {
		s, okString := v.(string)
		if okString {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func calculateFlowAndMessageType(event contract.EventType) (contract.FlowType, contract.MessageType) {
	var flow contract.FlowType
	var message contract.MessageType

	switch event {
	case contract.EventTypeRenderDeviceConfirmed, contract.EventTypeRenderDeviceDiscovered, contract.EventTypeRenderVolumeChanged:
		flow = contract.FlowTypeRender
	case contract.EventTypeCaptureDeviceConfirmed, contract.EventTypeCaptureDeviceDiscovered, contract.EventTypeCaptureVolumeChanged:
		flow = contract.FlowTypeCapture
	default:
		flow = 0
	}

	switch event {
	case contract.EventTypeRenderDeviceConfirmed, contract.EventTypeCaptureDeviceConfirmed:
		message = contract.MessageTypeConfirmed
	case contract.EventTypeRenderDeviceDiscovered, contract.EventTypeCaptureDeviceDiscovered:
		message = contract.MessageTypeDiscovered
	case contract.EventTypeRenderVolumeChanged:
		message = contract.MessageTypeVolumeRenderChanged
	case contract.EventTypeCaptureVolumeChanged:
		message = contract.MessageTypeVolumeCaptureChanged
	default:
		message = 0
	}

	return flow, message
}

func normalizeValue(key string, value string) any {
	trimmed := strings.TrimSpace(value)
	switch key {
	case contract.FieldRenderVolume, contract.FieldCaptureVolume, contract.FieldVolume:
		if n, err := strconv.Atoi(trimmed); err == nil {
			return n
		}
	}
	return value
}

func buildDeviceKey(fields map[string]string) string {
	pnpID := strings.TrimSpace(fields[contract.FieldPnpID])
	hostName := strings.TrimSpace(fields[contract.FieldHostName])
	if pnpID != "" && hostName != "" {
		return hostName + "|" + pnpID
	}
	if pnpID != "" {
		return pnpID
	}
	return hostName
}
