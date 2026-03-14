package version

//go:generate cp ../../.release-please-manifest.json manifest.json

import (
	_ "embed"
	"encoding/json"
)

//go:embed manifest.json
var manifestBytes []byte

// Version is the current build version. Overridden at build time via ldflags:
// go build -ldflags "-X github.com/keeperhub/cli/internal/version.Version=1.0.0"
var Version = ""

func init() {
	if Version != "" {
		return
	}
	var manifest map[string]string
	if err := json.Unmarshal(manifestBytes, &manifest); err == nil {
		if v, ok := manifest["."]; ok {
			Version = v
			return
		}
	}
	Version = "dev"
}
