package appinfo

var (
	// AppName Version
	// These can be overridden at build time via -ldflags.
	// Example:
	//   go build -ldflags "-X 'win-sound-dev-go-bridge/pkg/version.Version=1.0.0' -X 'win-sound-dev-go-bridge/pkg/version.AppInfo=abcdef' -X 'win-sound-dev-go-bridge/pkg/version.CommitDate=2025-01-01T00:00:00Z'"
	AppName = "win-sound-dev-go-bridge"
	Version = "dev"
	_       = "Unknown CommitDate"
)
