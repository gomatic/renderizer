// Package app holds the composition seams between the CLI framework and the
// render domain: the logger constructor and the error-to-exit-code mapping.
// The CLI command itself is wired in the cmd composition root, since renderizer
// is a single-action tool whose command is its root.
package app

import (
	"io"
	"log/slog"

	"github.com/gomatic/go-log"
)

type (
	// Verbose enables info-level logging.
	Verbose bool
	// Debugging enables debug-level logging.
	Debugging bool
)

// NewLogger builds a logger writing to w via gomatic/go-log. Debug wins over
// verbose; with neither, only warnings and errors are emitted so normal runs
// stay quiet.
func NewLogger(w io.Writer, verbose Verbose, debug Debugging) *slog.Logger {
	return log.LoggerConfig{LogLevel: level(verbose, debug)}.NewLogger(w)
}

// level resolves the go-log textual level from the verbosity flags.
func level(verbose Verbose, debug Debugging) log.Level {
	switch {
	case bool(debug):
		return "debug"
	case bool(verbose):
		return "info"
	default:
		return "warn"
	}
}
