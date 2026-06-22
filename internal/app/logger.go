// Package app holds the composition seams between the CLI framework and the
// render domain: the logger constructor and the error-to-exit-code mapping.
// The CLI command itself is wired in the cmd composition root, since renderizer
// is a single-action tool whose command is its root.
package app

import (
	"io"
	"log/slog"
)

type (
	// Verbose enables info-level logging.
	Verbose bool
	// Debugging enables debug-level logging.
	Debugging bool
)

// NewLogger builds a logger writing to w. Debug wins over verbose; with neither,
// only warnings and errors are emitted so normal runs stay quiet.
func NewLogger(w io.Writer, verbose Verbose, debug Debugging) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level(verbose, debug)}))
}

// level resolves the logging level from the verbosity flags.
func level(verbose Verbose, debug Debugging) slog.Level {
	switch {
	case bool(debug):
		return slog.LevelDebug
	case bool(verbose):
		return slog.LevelInfo
	default:
		return slog.LevelWarn
	}
}
