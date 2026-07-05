package version

// Config holds everything Run needs: the application name and the link-time
// build version. It carries no behavior.
type Config struct {
	App   AppName
	Build Build
}
