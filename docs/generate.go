//go:build ignore

package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/keeperhub/cli/cmd"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/spf13/cobra/doc"
)

// The //go:generate directive lives in doc.go, not here: this file is excluded
// from the build by the constraint above, so directives in it are never
// scanned.

func main() {
	// GenMarkdownTree only writes; it never removes. Without this sweep a page
	// survives its own command: renaming `protocol` to `plugin` left three
	// kh_protocol*.md behind, and deprecating `serve` left kh_serve.md, all
	// still published while cobra had correctly dropped them from the index.
	// Their content never changes, so the docs check sees no diff and cannot
	// notice. Clearing first makes the generated tree authoritative, so a
	// removal shows up as a deletion the check can catch.
	if err := pruneGeneratedPages("."); err != nil {
		log.Fatal(err)
	}

	ios := iostreams.System()
	f := &cmdutil.Factory{AppVersion: "dev", IOStreams: ios}
	root := cmd.NewRootCmd(f)
	root.DisableAutoGenTag = true
	if err := doc.GenMarkdownTree(root, "."); err != nil {
		log.Fatal(err)
	}
}

// pruneGeneratedPages removes the generated command reference from dir.
//
// Only `kh*.md` is touched: the hand-written guides (quickstart.md,
// concepts.md, execution-recovery.md) and the generator's own sources live
// alongside it and must survive.
func pruneGeneratedPages(dir string) error {
	matches, err := filepath.Glob(filepath.Join(dir, "kh*.md"))
	if err != nil {
		return err
	}
	for _, path := range matches {
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}
