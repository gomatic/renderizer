package version

// Named types for every Config field, converted from the composition root's
// raw values so the domain signature carries domain names.
type (
	// AppName is the application name printed in the version line.
	AppName string
	// Build is the build version stamped at link time via -X main.version.
	Build string
)
