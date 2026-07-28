//go:build ignore

package main

import (
	"log"

	"github.com/keeperhub/cli/cmd"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/spf13/cobra/doc"
)

// The //go:generate directive lives in doc.go, not here: this file is excluded
// from the build by the constraint above, so directives in it are never
// scanned.

func main() {
	ios := iostreams.System()
	f := &cmdutil.Factory{AppVersion: "dev", IOStreams: ios}
	root := cmd.NewRootCmd(f)
	root.DisableAutoGenTag = true
	if err := doc.GenMarkdownTree(root, "."); err != nil {
		log.Fatal(err)
	}
}
