//go:build ignore

package main

import (
	"log"

	"github.com/keeperhub/cli/cmd"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/spf13/cobra/doc"
)

//go:generate go run generate.go

func main() {
	ios := iostreams.System()
	f := &cmdutil.Factory{AppVersion: "dev", IOStreams: ios}
	root := cmd.NewRootCmd(f)
	root.DisableAutoGenTag = true
	if err := doc.GenMarkdownTree(root, "."); err != nil {
		log.Fatal(err)
	}
}
