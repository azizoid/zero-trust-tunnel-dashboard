package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the version of the application
	// This should be set via build flags: -ldflags "-X github.com/azizoid/zero-trust-tunnel-dashboard/pkg/version.Version=v0.1.0"
	Version = "v1.3.3"

	// Commit is the git commit hash
	// This should be set via build flags: -ldflags "-X github.com/azizoid/zero-trust-tunnel-dashboard/pkg/version.Commit=$(git rev-parse --short HEAD)"
	Commit = "unknown"

	// BuildDate is the build date
	// This should be set via build flags: -ldflags "-X github.com/azizoid/zero-trust-tunnel-dashboard/pkg/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	BuildDate = "unknown"
)

// Info returns version information
func Info() string {
	return fmt.Sprintf("tunnel-dash %s (commit: %s, built: %s, go: %s)",
		Version, Commit, BuildDate, runtime.Version())
}

// Short returns a short version string.
func Short() string {
	return Version
}
