package update

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/keeperhub/cli/internal/version"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// DefaultHomebrewChecker is the real Homebrew detection using the current executable path.
func DefaultHomebrewChecker() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		resolved = exe
	}
	return isHomebrewPath(resolved)
}

// isHomebrewPath checks whether a resolved executable path is inside a Homebrew prefix.
func isHomebrewPath(path string) bool {
	prefixes := []string{
		"/opt/homebrew/",
		"/usr/local/Cellar/",
		"/usr/local/opt/",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// TestableIsHomebrewInstall exposes the path-check logic for unit tests without
// touching the real filesystem. It accepts an already-resolved path.
func TestableIsHomebrewInstall(resolvedPath string) bool {
	return isHomebrewPath(resolvedPath)
}

// HomebrewCheckerFunc is the injectable function used to detect Homebrew installs.
// Replace in tests to avoid touching the real filesystem.
var HomebrewCheckerFunc = DefaultHomebrewChecker

// DefaultDetectLatest calls the real go-selfupdate API.
func DefaultDetectLatest(slug string) (*selfupdate.Release, bool, error) {
	return selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug(slug))
}

// DetectLatestFunc is the injectable function used to check for a new release.
// Replace in tests to avoid hitting the GitHub API.
var DetectLatestFunc = DefaultDetectLatest

// NewUpdateCmd creates the `kh update` cobra command.
func NewUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update kh to the latest version",
		Long: `Update kh to the latest version by downloading the newest release from GitHub.

If kh was installed via Homebrew, this command will print the appropriate
brew command to use instead of replacing the binary directly. Homebrew manages
its own binary lifecycle and must be used to keep the installation consistent.`,
		Example: `  # Check for and install the latest version
  kh update`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(f)
		},
	}

	return cmd
}

func runUpdate(f *cmdutil.Factory) error {
	if HomebrewCheckerFunc() {
		fmt.Fprintf(f.IOStreams.Out, "kh was installed via Homebrew. Run: brew upgrade kh\n")
		return nil
	}

	latest, found, err := DetectLatestFunc("keeperhub/cli")
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}

	if !found {
		fmt.Fprintf(f.IOStreams.Out, "No release found.\n")
		return nil
	}

	// "dev" means no version was injected via ldflags — always update.
	if version.Version != "dev" && latest.LessOrEqual(version.Version) {
		fmt.Fprintf(f.IOStreams.Out, "Already on latest version: %s\n", version.Version)
		return nil
	}

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return fmt.Errorf("getting executable path: %w", err)
	}

	fmt.Fprintf(f.IOStreams.Out, "Updating from %s to %s...\n", version.Version, latest.Version())

	if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetName, exe); err != nil {
		return fmt.Errorf("updating binary: %w", err)
	}

	fmt.Fprintf(f.IOStreams.Out, "Updated to %s\n", latest.Version())
	return nil
}
