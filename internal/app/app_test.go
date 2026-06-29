package app_test

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/app"
	"github.com/gomatic/renderizer/internal/constants"
)

func TestNewLoggerLevels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		logAt     slog.Level
		verbose   app.Verbose
		debug     app.Debugging
		wantEmpty bool
	}{
		{name: "default suppresses info", logAt: slog.LevelInfo, wantEmpty: true},
		{name: "verbose emits info", verbose: true, logAt: slog.LevelInfo, wantEmpty: false},
		{name: "default emits warn", logAt: slog.LevelWarn, wantEmpty: false},
		{name: "debug emits debug", debug: true, logAt: slog.LevelDebug, wantEmpty: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			logger := app.NewLogger(&buf, tt.verbose, tt.debug)
			logger.Log(t.Context(), tt.logAt, "message")
			assert.Equal(t, tt.wantEmpty, buf.Len() == 0)
		})
	}
}

func TestExitCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err  error
		name string
		want app.ExitStatus
	}{
		{name: "success", err: nil, want: 0},
		{name: "parse", err: constants.ErrParseTemplate.With(nil), want: 4},
		{name: "execute", err: constants.ErrExecuteTemplate.With(nil), want: 8},
		{name: "read", err: constants.ErrReadTemplate.With(nil), want: 2},
		{name: "panic", err: constants.ErrRenderPanic.With(nil), want: 15},
		{name: "open is generic", err: constants.ErrOpenTemplate.With(nil), want: 1},
		{name: "missing is generic", err: constants.ErrMissingTemplate.With(nil), want: 1},
		{name: "unknown is generic", err: errors.New("boom"), want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, app.ExitCode(tt.err))
		})
	}
}

func TestExitCodeWrapped(t *testing.T) {
	t.Parallel()
	wrapped := constants.ErrParseTemplate.With(errors.New("syntax"))
	require.Equal(t, app.ExitStatus(4), app.ExitCode(wrapped))
}
