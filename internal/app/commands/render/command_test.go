package render_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		Source:     strings.NewReader(source),
		ReadFile:   func(string) ([]byte, error) { return nil, os.ErrNotExist },
		Exists:     func(string) bool { return false },
		Getwd:      func() (string, error) { return "/work", nil },
		Environ:    func() []string { return []string{"HOME=/home"} },
		Capitalize: true,
		TimeFormat: "20060102T150405",
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
	rt.Piped = true
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
