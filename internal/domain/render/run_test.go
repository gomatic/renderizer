package render_test

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
	"github.com/gomatic/renderizer/internal/domain/render"
)

// discardLogger returns a logger that writes nowhere.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// mapReadFile builds a ReadFile that serves content from files and returns
// os.ErrNotExist for anything else.
func mapReadFile(files map[string]string) render.ReadFileFunc {
	return func(name string) ([]byte, error) {
		content, ok := files[name]
		if !ok {
			return nil, os.ErrNotExist
		}
		return []byte(content), nil
	}
}

// existsIn builds an Exists from a set of present names.
func existsIn(present ...string) render.ExistsFunc {
	set := map[string]bool{}
	for _, name := range present {
		set[name] = true
	}
	return func(name string) bool { return set[name] }
}

// errReader fails on Read, exercising the stdin read-error path.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read boom") }

// baseConfig is a complete config with every IO seam injected; tests override
// the fields they exercise.
func baseConfig() render.Config {
	return render.Config{
		Environment:       "env",
		MissingKey:        "error",
		CapitalizeEnabled: true,
		TimeFormat:        "20060102T150405",
		ReadFile:          func(string) ([]byte, error) { return nil, os.ErrNotExist },
		Exists:            func(string) bool { return false },
		Getwd:             func() (string, error) { return "/work/dir", nil },
		Environ:           func() []string { return []string{"HOME=/home", "USER=alice"} },
	}
}

func run(t *testing.T, cfg render.Config) (render.Result, error) {
	t.Helper()
	return render.Run(context.Background(), discardLogger(), cfg)
}

func TestRunStdin(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.StdinEnabled = true
	cfg.Source = strings.NewReader("Hello, {{.Name}}!")
	cfg.Assignments = render.AssignmentTokens{"--name=World"}

	result, err := run(t, cfg)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!\n", string(result.Output))
}

func TestRunVerboseAndDebugLogging(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.VerboseEnabled = true
	cfg.DebuggingEnabled = true
	cfg.StdinEnabled = true
	cfg.Source = strings.NewReader("{{.Name}}")
	cfg.Assignments = render.AssignmentTokens{"--name=X"}

	result, err := run(t, cfg)
	require.NoError(t, err)
	assert.Equal(t, "X\n", string(result.Output))
}

func TestRunTemplateFile(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Templates = render.TemplateFiles{"t.tmpl"}
	cfg.ReadFile = mapReadFile(map[string]string{"t.tmpl": "Hi {{.Name}}"})
	cfg.Assignments = render.AssignmentTokens{"--name=Bob"}

	result, err := run(t, cfg)
	require.NoError(t, err)
	assert.Equal(t, "Hi Bob\n", string(result.Output))
}

func TestRunStdinReadError(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.StdinEnabled = true
	cfg.Source = errReader{}

	_, err := run(t, cfg)
	require.ErrorIs(t, err, constants.ErrReadTemplate)
}

func TestRunFileOpenError(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Templates = render.TemplateFiles{"missing.tmpl"}

	_, err := run(t, cfg)
	require.ErrorIs(t, err, constants.ErrOpenTemplate)
}

func TestRunParseError(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.StdinEnabled = true
	cfg.Source = strings.NewReader("{{.Unclosed")

	_, err := run(t, cfg)
	require.ErrorIs(t, err, constants.ErrParseTemplate)
}

func TestRunExecuteError(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.StdinEnabled = true
	cfg.Source = strings.NewReader("{{.Missing}}")

	_, err := run(t, cfg)
	require.ErrorIs(t, err, constants.ErrExecuteTemplate)
}

func TestRunRenderErrorReturnsPartialOutput(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Templates = render.TemplateFiles{"ok.tmpl", "bad.tmpl"}
	cfg.ReadFile = mapReadFile(map[string]string{
		"ok.tmpl":  "good",
		"bad.tmpl": "{{.Unclosed",
	})

	result, err := run(t, cfg)
	require.ErrorIs(t, err, constants.ErrParseTemplate)
	assert.Equal(t, "good\n", string(result.Output))
}
