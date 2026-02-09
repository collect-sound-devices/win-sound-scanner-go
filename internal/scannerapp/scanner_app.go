package scannerapp

type ScannerApp interface {
	RepostRenderDeviceToApi()
	RepostCaptureDeviceToApi()
	Shutdown()
}
