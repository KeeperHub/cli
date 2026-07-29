//go:build tools

package docs

// generate.go carries `//go:build ignore`, so `go mod tidy` cannot see its
// import of cobra/doc and prunes go-md2man and blackfriday from go.sum. Doc
// generation then fails with a missing go.sum entry, which is what broke the
// docs check on every dependabot module bump.
//
// This blank import is what stops that. `go mod tidy` evaluates imports under
// all build tag combinations, unlike the compiler, so the `tools` tag keeps the
// file out of every real build while still holding those hashes pinned.

import _ "github.com/spf13/cobra/doc"
