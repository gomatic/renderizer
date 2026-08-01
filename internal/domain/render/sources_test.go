package render_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/constants"
	"github.com/gomatic/renderizer/internal/domain/render"
)

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
