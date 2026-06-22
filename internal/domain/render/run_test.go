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
		Environment: "env",
		MissingKey:  "error",
		Capitalize:  true,
		TimeFormat:  "20060102T150405",
		ReadFile:    func(string) ([]byte, error) { return nil, os.ErrNotExist },
		Exists:      func(string) bool { return false },
		Getwd:       func() (string, error) { return "/work/dir", nil },
		Environ:     func() []string { return []string{"HOME=/home", "USER=alice"} },
	}
}

func run(t *testing.T, cfg render.Config) (render.Result, error) {
	t.Helper()
	return render.Run(context.Background(), discardLogger(), cfg)
}

func TestRunStdin(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Stdin = true
	cfg.Source = strings.NewReader("Hello, {{.Name}}!")
	cfg.Assignments = render.AssignmentTokens{"--name=World"}

	result, err := run(t, cfg)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!\n", string(result.Output))
}

func TestRunVerboseAndDebugLogging(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Verbose = true
	cfg.Debugging = true
	cfg.Stdin = true
	cfg.Source = strings.NewReader("{{.Name}}")
	cfg.Assignments = render.AssignmentTokens{"--name=X"}

	result, err := run(t, cfg)
	require.NoError(t, err)
	assert.Equal(t, "X\n", string(result.Output))
}

func TestRunStdinEnvironment(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Stdin = true
	cfg.Source = strings.NewReader("{{.env.USER}}")

	result, err := run(t, cfg)
	require.NoError(t, err)
	assert.Equal(t, "alice\n", string(result.Output))
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

func TestRunMultipleTemplates(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Environment = "" // exercise the no-environment branch
	cfg.Templates = render.TemplateFiles{"a.tmpl", "b.tmpl"}
	cfg.ReadFile = mapReadFile(map[string]string{
		"a.tmpl": "First {{.Name}}",
		"b.tmpl": "Second {{.Name}}",
	})
	cfg.Assignments = render.AssignmentTokens{"--name=X"}

	result, err := run(t, cfg)
	require.NoError(t, err)
	assert.Equal(t, "First X\nSecond X\n", string(result.Output))
}

func TestRunSettingsProvidesValue(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Settings = render.SettingsFiles{"s.yaml"}
	cfg.Templates = render.TemplateFiles{"t.tmpl"}
	cfg.ReadFile = mapReadFile(map[string]string{
		"s.yaml": "Name: FromSettings",
		"t.tmpl": "{{.Name}}",
	})

	result, err := run(t, cfg)
	require.NoError(t, err)
	assert.Equal(t, "FromSettings\n", string(result.Output))
}

func TestRunVariablesOverrideSettings(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Settings = render.SettingsFiles{"s.yaml"}
	cfg.Templates = render.TemplateFiles{"t.tmpl"}
	cfg.ReadFile = mapReadFile(map[string]string{
		"s.yaml": "Name: FromSettings",
		"t.tmpl": "{{.Name}}",
	})
	cfg.Assignments = render.AssignmentTokens{"--name=FromCLI"}

	result, err := run(t, cfg)
	require.NoError(t, err)
	assert.Equal(t, "FromCLI\n", string(result.Output))
}

func TestRunDiscoversDefaultTemplate(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Exists = existsIn("renderizer.yaml.tmpl")
	cfg.ReadFile = mapReadFile(map[string]string{"renderizer.yaml.tmpl": "Discovered {{.Value}}"})
	cfg.Assignments = render.AssignmentTokens{"--value=ok"}

	result, err := run(t, cfg)
	require.NoError(t, err)
	assert.Equal(t, "Discovered ok\n", string(result.Output))
}

func TestRunMissingTemplate(t *testing.T) {
	t.Parallel()
	_, err := run(t, baseConfig())
	require.ErrorIs(t, err, constants.ErrMissingTemplate)
}

func TestRunGetwdError(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Getwd = func() (string, error) { return "", errors.New("no cwd") }
	// No templates, no stdin, nothing discoverable: exercises the getwd-error
	// fallback in both mainName and bases before failing.
	_, err := run(t, cfg)
	require.ErrorIs(t, err, constants.ErrMissingTemplate)
}

func TestRunAssignmentsError(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Stdin = true
	cfg.Source = strings.NewReader("x")
	cfg.Assignments = render.AssignmentTokens{"--a.b=2", "--a=1"}

	_, err := run(t, cfg)
	require.ErrorIs(t, err, constants.ErrMergeContext)
}

func TestRunSettingsParseError(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Stdin = true
	cfg.Source = strings.NewReader("x")
	cfg.Settings = render.SettingsFiles{"bad.yaml"}
	cfg.ReadFile = mapReadFile(map[string]string{"bad.yaml": "::: not yaml :::"})

	_, err := run(t, cfg)
	require.ErrorIs(t, err, constants.ErrParseSettings)
}

func TestRunSettingsDeepMerge(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Capitalize = false // so CLI key "a" matches the settings key
	cfg.Settings = render.SettingsFiles{"s.yaml"}
	cfg.Templates = render.TemplateFiles{"t.tmpl"}
	cfg.ReadFile = mapReadFile(map[string]string{
		"s.yaml": "a:\n  fromSettings: settings\n",
		"t.tmpl": "{{.a.fromCLI}}-{{.a.fromSettings}}",
	})
	cfg.Assignments = render.AssignmentTokens{"--a.fromCLI=cli"}

	result, err := run(t, cfg)
	require.NoError(t, err)
	assert.Equal(t, "cli-settings\n", string(result.Output))
}

func TestRunStdinReadError(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Stdin = true
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
	cfg.Stdin = true
	cfg.Source = strings.NewReader("{{.Unclosed")

	_, err := run(t, cfg)
	require.ErrorIs(t, err, constants.ErrParseTemplate)
}

func TestRunExecuteError(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.Stdin = true
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
