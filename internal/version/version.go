package version

// Set via ldflags at build time. Defaults to "dev" for local builds.
var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
)
