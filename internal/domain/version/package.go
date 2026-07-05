// Package version orchestrates the version command: it formats the
// "<app> version <build>" line from the application name and the link-time
// build version. It holds no CLI, flag, or output-writing logic. This is the
// domain tier: the seam between the app tier (internal/app/commands/version)
// and the composition root.
package version
