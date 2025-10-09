package version

var (
	// Version
	// These can be overridden at build time via -ldflags.
	// Example:
	//   go build -ldflags "-X 'win-sound-dev-go-bridge/pkg/version.Version=1.0.0' -X 'win-sound-dev-go-bridge/pkg/version.Commit=abcdef' -X 'win-sound-dev-go-bridge/pkg/version.Date=2025-01-01T00:00:00Z'"
	Version = "dev"
	_       = "Unknown CommitDate"
)
