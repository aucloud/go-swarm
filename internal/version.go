package internal

import (
	"fmt"
)

var (
	// Version release version
	Version = "0.1.1"

	// Build will be overwritten automatically by the build system
	Build = "dev"

	// Commit will be overwritten automatically by the build system
	Commit = "HEAD"
)

// FullVersion returns the full version, build and commit hash
func FullVersion() string {
	return fmt.Sprintf("%s-%s@%s", Version, Build, Commit)
}
