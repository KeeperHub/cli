package main

import (
	"fmt"
	"os"

	"github.com/keeperhub/cli/internal/config"
	"github.com/keeperhub/cli/internal/version"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
)

func main() {
	streams := iostreams.System()

	f := &cmdutil.Factory{
		AppVersion: version.Version,
		IOStreams:   streams,
		Config: func() (config.Config, error) {
			return config.ReadConfig()
		},
	}

	fmt.Fprintf(f.IOStreams.Out, "kh version %s\n", f.AppVersion)
	os.Exit(0)
}
