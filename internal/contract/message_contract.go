package contract

type EventType uint8

const (
	EventTypeNothing EventType = iota
	EventTypeRenderDeviceConfirmed
	EventTypeCaptureDeviceConfirmed
	EventTypeRenderDeviceDiscovered
	EventTypeCaptureDeviceDiscovered
	EventTypeRenderVolumeChanged
	EventTypeCaptureVolumeChanged
)

type MessageType uint8

const (
	MessageTypeConfirmed  = 0
	MessageTypeDiscovered = 1
	//	MessageTypeDetached                          = 2
	MessageTypeVolumeRenderChanged   MessageType = 3
	MessageTypeVolumeCaptureChanged  MessageType = 4
	MessageTypeDefaultRenderChanged  MessageType = 5
	MessageTypeDefaultCaptureChanged MessageType = 6
)

type FlowType uint8

const (
	FlowTypeRender  FlowType = 1
	FlowTypeCapture FlowType = 2
)

const (
	FieldDeviceMessageType   = "deviceMessageType"
	FieldUpdateDate          = "updateDate"
	FieldFlowType            = "flowType"
	FieldName                = "name"
	FieldPnpID               = "pnpId"
	FieldRenderVolume        = "renderVolume"
	FieldCaptureVolume       = "captureVolume"
	FieldVolume              = "volume"
	FieldHostName            = "hostName"
	FieldOperationSystemName = "operationSystemName"
	FieldHTTPRequest         = "httpRequest"
	FieldURLSuffix           = "urlSuffix"
)
