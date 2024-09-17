package pkg

// These values will be injected into these variables at the build time.
var (
	// GitCommit Git Commit SHA
	GitCommit string
	// Version version of the CLI
	Version string
)

// GetVersion get latest version
func GetVersion() string {
	if len(Version) == 0 {
		return "dev"
	}
	return Version
}
