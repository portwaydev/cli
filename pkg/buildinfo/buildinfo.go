package buildinfo

// These variables are intended to be set at build time via -ldflags.
// Defaults are for local development builds.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
