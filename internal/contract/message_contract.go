package contract

type SoundDeviceEventType uint8

const (
	EventTypeVolumeRenderChanged   SoundDeviceEventType = 3
	EventTypeVolumeCaptureChanged  SoundDeviceEventType = 4
	EventTypeDefaultRenderChanged  SoundDeviceEventType = 5
	EventTypeDefaultCaptureChanged SoundDeviceEventType = 6
)

const (
	RequestPostDevice      = "post_device"
	RequestPutVolumeChange = "put_volume_change"
)

const (
	EventDefaultRenderChanged  = "default_render_changed"
	EventDefaultCaptureChanged = "default_capture_changed"
	EventRenderVolumeChanged   = "render_volume_changed"
	EventCaptureVolumeChanged  = "capture_volume_changed"
)

const (
	FlowRender  = "render"
	FlowCapture = "capture"
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
