package render_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/constants"
	"github.com/gomatic/renderizer/internal/domain/render"
)

// TestMergeDefaultsLetsCommandLineVariablesAlwaysWin names mergeDefaults'
// claim. Precedence is the entire reason for having two sources: a settings
// file supplies defaults and `--name=value` overrides them. Reversing it for
// even one key means an explicit argument is silently discarded in favour of a
// file the user may not have written — the tool doing something other than what
// the command said, and reporting success.
//
// The nested case is the one that decays quietly. A merge that replaced a whole
// nested map with the file's version would drop a command-line value set deeper
// in the tree while every top-level key still looked correct.
func TestMergeDefaultsLetsCommandLineVariablesAlwaysWin(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.CapitalizeEnabled = false
	cfg.StdinEnabled = true
	cfg.Source = strings.NewReader("{{.shared}}|{{.onlyInFile}}|{{.nested.deep}}|{{.nested.alsoInFile}}")
	cfg.Settings = render.SettingsFiles{"defaults.yaml"}
	cfg.ReadFile = mapReadFile(map[string]string{
		"defaults.yaml": "shared: from-settings\nonlyInFile: from-settings\n" +
			"nested:\n  deep: from-settings\n  alsoInFile: from-settings\n",
	})
	cfg.Exists = existsIn("defaults.yaml")
	cfg.Assignments = render.AssignmentTokens{"--shared=from-command-line", "--nested.deep=from-command-line"}

	result, err := run(t, cfg)

	require.NoError(t, err)
	assert.Equal(t,
		"from-command-line|from-settings|from-command-line|from-settings\n",
		string(result.Output),
		"command-line values win at every depth; settings fill only the gaps")
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

func TestRunSettingsDeepMerge(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.CapitalizeEnabled = false // so CLI key "a" matches the settings key
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

func TestRunSettingsParseError(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.StdinEnabled = true
	cfg.Source = strings.NewReader("x")
	cfg.Settings = render.SettingsFiles{"bad.yaml"}
	cfg.ReadFile = mapReadFile(map[string]string{"bad.yaml": "::: not yaml :::"})

	_, err := run(t, cfg)
	require.ErrorIs(t, err, constants.ErrParseSettings)
}

func TestRunAssignmentsError(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.StdinEnabled = true
	cfg.Source = strings.NewReader("x")
	cfg.Assignments = render.AssignmentTokens{"--a.b=2", "--a=1"}

	_, err := run(t, cfg)
	require.ErrorIs(t, err, constants.ErrMergeContext)
}

func TestRunStdinEnvironment(t *testing.T) {
	t.Parallel()
	cfg := baseConfig()
	cfg.StdinEnabled = true
	cfg.Source = strings.NewReader("{{.env.USER}}")

	result, err := run(t, cfg)
	require.NoError(t, err)
	assert.Equal(t, "alice\n", string(result.Output))
}
