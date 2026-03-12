package version

// Version is the current build version. Overridden at build time via ldflags:
// go build -ldflags "-X github.com/keeperhub/cli/internal/version.Version=1.0.0"
var Version = "dev"
