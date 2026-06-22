package analyze_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/constants"
	"github.com/gomatic/renderizer/internal/domain/analyze"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read boom") }

func run(t *testing.T, cfg analyze.Config) (analyze.Result, error) {
	t.Helper()
	return analyze.Run(context.Background(), discardLogger(), cfg)
}

func TestRunFile(t *testing.T) {
	t.Parallel()
	cfg := analyze.Config{
		Template: "t.tmpl",
		ReadFile: func(string) ([]byte, error) { return []byte("{{.Name}}{{range .Items}}{{.Id}}{{end}}"), nil },
	}
	result, err := run(t, cfg)
	require.NoError(t, err)
	output := string(result.Output)
	assert.Contains(t, output, "Name: \"\"")
	assert.Contains(t, output, "Items:")
	assert.Contains(t, output, "Id: \"\"")
}

func TestRunStdin(t *testing.T) {
	t.Parallel()
	cfg := analyze.Config{Source: strings.NewReader("{{.Greeting}}")}
	result, err := run(t, cfg)
	require.NoError(t, err)
	assert.Contains(t, string(result.Output), "Greeting: \"\"")
}

func TestRunFileOpenError(t *testing.T) {
	t.Parallel()
	cfg := analyze.Config{
		Template: "missing.tmpl",
		ReadFile: func(string) ([]byte, error) { return nil, os.ErrNotExist },
	}
	_, err := run(t, cfg)
	require.ErrorIs(t, err, constants.ErrOpenTemplate)
}

func TestRunStdinReadError(t *testing.T) {
	t.Parallel()
	cfg := analyze.Config{Source: errReader{}}
	_, err := run(t, cfg)
	require.ErrorIs(t, err, constants.ErrReadTemplate)
}

func TestRunParseError(t *testing.T) {
	t.Parallel()
	cfg := analyze.Config{Source: strings.NewReader("{{.Unclosed")}
	_, err := run(t, cfg)
	require.ErrorIs(t, err, constants.ErrParseTemplate)
}
