package app_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gomatic/renderizer/internal/app"
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
