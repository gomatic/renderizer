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
// stay quiet. It returns a value — slog.Logger is a copyable handle — and the
// callers that feed a domain Run take its address, since *slog.Logger is the
// domain seam's conventional logger type.
func NewLogger(w io.Writer, isVerbose Verbose, isDebug Debugging) slog.Logger {
	return *log.LoggerConfig{Level: level(isVerbose, isDebug)}.NewLogger(w)
}

// level resolves the go-log textual level from the verbosity flags.
func level(isVerbose Verbose, isDebug Debugging) log.Level {
	switch {
	case bool(isDebug):
		return "debug"
	case bool(isVerbose):
		return "info"
	default:
		return "warn"
	}
}
