package render_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/gomatic/renderizer/internal/app"
	"github.com/gomatic/renderizer/internal/app/commands/render"
)

func mapReadFile(files map[string]string) func(string) ([]byte, error) {
	return func(name string) ([]byte, error) {
		content, ok := files[name]
		if !ok {
			return nil, os.ErrNotExist
		}
		return []byte(content), nil
	}
}

func baseRuntime(source string) app.Runtime {
	return app.Runtime{
		Source:            strings.NewReader(source),
		ReadFile:          func(string) ([]byte, error) { return nil, os.ErrNotExist },
		Exists:            func(string) bool { return false },
		Getwd:             func() (string, error) { return "/work", nil },
		Environ:           func() []string { return []string{"HOME=/home"} },
		CapitalizeEnabled: true,
		TimeFormat:        "20060102T150405",
	}
}

func exec(t *testing.T, rt app.Runtime, args ...string) (string, error) {
	t.Helper()
	cmd := render.Command(rt)
	var stdout, stderr bytes.Buffer
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr
	err := cmd.Run(context.Background(), append([]string{"renderizer"}, args...))
	return stdout.String(), err
}

func TestRenderStdin(t *testing.T) {
	// Arbitrary --name=value is extracted by the tokenizer into rt.Assignments;
	// cli only receives the known flags.
	rt := baseRuntime("Hello {{.Name}}")
	rt.Assignments = []string{"--name=World"}
	out, err := exec(t, rt, "--stdin")
	require.NoError(t, err)
	assert.Contains(t, out, "Hello World")
}

func TestRenderTemplateFile(t *testing.T) {
	rt := baseRuntime("")
	rt.ReadFile = mapReadFile(map[string]string{"t.tmpl": "Hi {{.Name}}"})
	rt.Assignments = []string{"--name=Bob"}
	out, err := exec(t, rt, "t.tmpl")
	require.NoError(t, err)
	assert.Contains(t, out, "Hi Bob")
}

func TestRenderPipedStdinWithoutFlag(t *testing.T) {
	rt := baseRuntime("Hi {{.Name}}")
	rt.IsPiped = true
	rt.Assignments = []string{"--name=Zed"}
	out, err := exec(t, rt)
	require.NoError(t, err)
	assert.Contains(t, out, "Hi Zed")
}

func TestRenderVerboseAndDebug(t *testing.T) {
	rt := baseRuntime("{{.Name}}")
	rt.Assignments = []string{"--name=X"}
	out, err := exec(t, rt, "--stdin", "--verbose", "--debugging")
	require.NoError(t, err)
	assert.Contains(t, out, "X")
}

func TestRenderError(t *testing.T) {
	_, err := exec(t, baseRuntime("{{.Missing}}"), "--stdin")
	require.Error(t, err)
}

// TestConfiguredCopiesTheParsedConfigPerRun names configured's claim: "value
// in, value out — the parsed config is copied, never mutated through a
// pointer." The claim is invisible in a single run and decisive across two.
// Command() closes over one Config that the flag bindings write into; if
// configured mutated it through a pointer instead of returning a copy, the
// first run's positional templates and derived StdinEnabled would still be
// sitting in that Config when the second run began, and the second render
// would silently include the first invocation's inputs.
//
// StdinEnabled is the sharpest case, because configured ORs the piped-stdin
// detection into it. Under mutation that OR is cumulative: once any run is
// piped, every later run believes stdin is available and reads a source that
// belongs to a previous invocation.
func TestConfiguredCopiesTheParsedConfigPerRun(t *testing.T) {
	t.Parallel()
	rt := baseRuntime("")
	rt.ReadFile = mapReadFile(map[string]string{
		"first.tmpl":  "first={{.Name}}",
		"second.tmpl": "second={{.Name}}",
	})
	rt.Assignments = []string{"--name=Once"}
	cmd := render.Command(rt)

	first := runOnce(t, cmd, "first.tmpl")
	second := runOnce(t, cmd, "second.tmpl")

	assert.Contains(t, first, "first=Once")
	assert.NotContains(t, first, "second=", "the first run cannot see a later run's template")
	assert.Contains(t, second, "second=Once")
	assert.NotContains(t, second, "first=",
		"the second run must not inherit the first run's positional templates")
}

// TestConfiguredDoesNotAccumulateStdinAcrossRuns is the StdinEnabled half of
// the same claim, isolated so a failure names the field. A piped run must not
// leave stdin enabled for a later run that was given a template file.
func TestConfiguredDoesNotAccumulateStdinAcrossRuns(t *testing.T) {
	t.Parallel()
	rt := baseRuntime("piped={{.Name}}")
	rt.IsPiped = true
	rt.ReadFile = mapReadFile(map[string]string{"file.tmpl": "file={{.Name}}"})
	rt.Assignments = []string{"--name=Once"}
	cmd := render.Command(rt)

	piped := runOnce(t, cmd)
	fromFile := runOnce(t, cmd, "file.tmpl")

	assert.Contains(t, piped, "piped=Once")
	assert.Contains(t, fromFile, "file=Once")
	assert.NotContains(t, fromFile, "piped=",
		"a run reading a template file must not also drain the previous run's stdin")
}

// runOnce executes an already-constructed command, which is what makes the two
// tests above able to observe state carried between runs.
func runOnce(t *testing.T, cmd *cli.Command, args ...string) string {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr
	require.NoError(t, cmd.Run(context.Background(), append([]string{"renderizer"}, args...)))
	return stdout.String()
}
