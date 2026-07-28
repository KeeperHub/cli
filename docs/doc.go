// Package docs holds the generated command reference.
//
// The generator itself lives in generate.go, which carries a `//go:build
// ignore` constraint so it is not compiled into the module. That constraint is
// also why the `//go:generate` directive cannot live there: a build-excluded
// file is never scanned, so `go generate ./docs/` silently did nothing, exited
// 0, and left the CI docs check comparing an unchanged tree against itself.
//
// Keeping the directive in this compiled file is what makes `go generate
// ./docs/` regenerate the reference, and therefore what makes the docs check
// and the downstream sync to keeperhub meaningful.
package docs

//go:generate go run generate.go
