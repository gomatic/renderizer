package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/app"
)

// exec runs the CLI end to end with the given stdin and arguments, returning
// captured stdout, stderr, and the exit status.
func exec(t *testing.T, stdin string, isPiped bool, args ...string) (string, string, app.ExitStatus) {
	t.Helper()
	t.Setenv("RENDERIZER_TESTING", "true")
	var stdout, stderr bytes.Buffer
	code := run(
		context.Background(),
		append([]string{"renderizer"}, args...),
		strings.NewReader(stdin),
		&stdout, &stderr, isPiped,
	)
	return stdout.String(), stderr.String(), code
}

func TestVersion(t *testing.T) {
	for _, flag := range []string{"--version", "-v"} {
		t.Run(flag, func(t *testing.T) {
			out, _, code := exec(t, "", false, flag)
			assert.Equal(t, app.ExitStatus(0), code)
			assert.Contains(t, out, "renderizer")
		})
	}
}

func TestVersionSubcommand(t *testing.T) {
	out, _, code := exec(t, "", false, "version")
	require.Equal(t, app.ExitStatus(0), code)
	assert.Contains(t, out, "renderizer version")
}

func TestRenderizerVersionEnv(t *testing.T) {
	out, _, code := exec(t, "{{.env.RENDERIZER_VERSION}}", false, "--stdin")
	require.Equal(t, app.ExitStatus(0), code)
	assert.NotEmpty(t, strings.TrimSpace(out))
}

func TestHelp(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		t.Run(flag, func(t *testing.T) {
			out, _, code := exec(t, "", false, flag)
			assert.Equal(t, app.ExitStatus(0), code)
			assert.Contains(t, out, "USAGE:")
		})
	}
}

func TestStdinRendering(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     string
		args     []string
	}{
		{"simple variable", "Hello, {{.Name}}!", "Hello, World!", []string{"--stdin", "--name=World"}},
		{"integer value", "Count: {{.Count}}", "Count: 42", []string{"--stdin", "--count=42"}},
		{"boolean value", "{{if .Flag}}YES{{else}}NO{{end}}", "YES", []string{"--stdin", "--flag=true"}},
		{"bare boolean", "{{if .Flag}}YES{{else}}NO{{end}}", "YES", []string{"--stdin", "--flag"}},
		{"dotted notation", "{{.A.B.C}}", "nested", []string{"--stdin", "--a.b.c=nested"}},
		{"typed equality", "{{if eq .Count 42}}ok{{end}}", "ok", []string{"--stdin", "--count=42"}},
		{"upper function", "{{upper .Text}}", "HI", []string{"--stdin", "--text=hi"}},
		{"command_line in testing mode", "{{command_line}}", "testing", []string{"--stdin"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, stderr, code := exec(t, tt.template, false, tt.args...)
			require.Equal(t, app.ExitStatus(0), code, "stderr: %s", stderr)
			assert.Contains(t, out, tt.want)
		})
	}
}

func TestSprigFunctions(t *testing.T) {
	out, _, code := exec(t, `{{ "hi" | b64enc }}`, false, "--stdin")
	require.Equal(t, app.ExitStatus(0), code)
	assert.Contains(t, out, "aGk=")
}

func TestCapitalizationToggle(t *testing.T) {
	out, _, code := exec(t, "{{.Name}} {{.foo}}", false, "--stdin", "--name=first", "-C", "--foo=second")
	require.Equal(t, app.ExitStatus(0), code)
	assert.Contains(t, out, "first second")
}

func TestRepeatedValues(t *testing.T) {
	out, _, code := exec(t, "{{range .Items}}{{.}},{{end}}", false,
		"--stdin", "--items=one", "--items=two", "--items=three")
	require.Equal(t, app.ExitStatus(0), code)
	assert.Contains(t, out, "one,two,three,")
}

func TestEnvironment(t *testing.T) {
	t.Setenv("RENDERIZER_DEMO", "value")
	out, _, code := exec(t, "{{.env.RENDERIZER_DEMO}}", false, "--stdin")
	require.Equal(t, app.ExitStatus(0), code)
	assert.Contains(t, out, "value")
}

func TestCustomEnvironmentName(t *testing.T) {
	t.Setenv("RENDERIZER_DEMO", "value")
	out, _, code := exec(t, "{{.vars.RENDERIZER_DEMO}}", false, "--stdin", "--environment=vars")
	require.Equal(t, app.ExitStatus(0), code)
	assert.Contains(t, out, "value")
}

func TestPipedStdinWithoutFlag(t *testing.T) {
	out, _, code := exec(t, "Hi {{.Name}}", true, "--name=Bob")
	require.Equal(t, app.ExitStatus(0), code)
	assert.Contains(t, out, "Hi Bob")
}

func TestMissingKey(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want app.ExitStatus
	}{
		{"error by default", []string{"--stdin"}, 8},
		{"zero suppresses", []string{"--stdin", "--missing=zero"}, 0},
		{"default suppresses", []string{"--stdin", "--missing=default"}, 0},
		{"invalid suppresses", []string{"--stdin", "--missing=invalid"}, 0},
		{"unknown resets to error", []string{"--stdin", "--missing=bogus"}, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, code := exec(t, "{{.Missing}}", false, tt.args...)
			assert.Equal(t, tt.want, code)
		})
	}
}

func TestTemplateFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "t.tmpl")
	require.NoError(t, os.WriteFile(path, []byte("Hello, {{.Name}}!"), 0o644))

	out, _, code := exec(t, "", false, path, "--name=World")
	require.Equal(t, app.ExitStatus(0), code)
	assert.Contains(t, out, "Hello, World!")
}

func TestSettingsFile(t *testing.T) {
	dir := t.TempDir()
	tmpl := filepath.Join(dir, "t.tmpl")
	settings := filepath.Join(dir, "s.yaml")
	require.NoError(t, os.WriteFile(tmpl, []byte("{{.Name}}:{{range .Items}}{{.}},{{end}}"), 0o644))
	require.NoError(t, os.WriteFile(settings, []byte("Name: FromSettings\nItems:\n  - a\n  - b\n"), 0o644))

	out, _, code := exec(t, "", false, tmpl, "--settings="+settings)
	require.Equal(t, app.ExitStatus(0), code)
	assert.Contains(t, out, "FromSettings:a,b,")

	out, _, code = exec(t, "", false, tmpl, "--settings="+settings, "--name=Overridden")
	require.Equal(t, app.ExitStatus(0), code)
	assert.Contains(t, out, "Overridden:a,b,")
}

func TestMultipleTemplates(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.tmpl")
	b := filepath.Join(dir, "b.tmpl")
	require.NoError(t, os.WriteFile(a, []byte("First {{.Name}}"), 0o644))
	require.NoError(t, os.WriteFile(b, []byte("Second {{.Name}}"), 0o644))

	out, _, code := exec(t, "", false, a, b, "--name=X")
	require.Equal(t, app.ExitStatus(0), code)
	assert.Contains(t, out, "First X")
	assert.Contains(t, out, "Second X")
}

func TestDefaultTemplateDiscovery(t *testing.T) {
	dir := t.TempDir()
	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(dir))
	require.NoError(t, os.WriteFile("renderizer.yaml.tmpl", []byte("Test: {{.Value}}"), 0o644))

	out, _, code := exec(t, "", false, "--value=success")
	require.Equal(t, app.ExitStatus(0), code)
	assert.Contains(t, out, "Test: success")
}

func TestAnalyzeFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "t.tmpl")
	require.NoError(t, os.WriteFile(path, []byte("{{.Name}}{{range .Items}}{{.Id}}{{end}}"), 0o644))

	out, _, code := exec(t, "", false, "analyze", path)
	require.Equal(t, app.ExitStatus(0), code)
	assert.Contains(t, out, "Name: \"\"")
	assert.Contains(t, out, "Items:")
	assert.Contains(t, out, "Id: \"\"")
}

func TestAnalyzeStdin(t *testing.T) {
	out, _, code := exec(t, "{{.Greeting}}", true, "analyze")
	require.Equal(t, app.ExitStatus(0), code)
	assert.Contains(t, out, "Greeting: \"\"")
}

func TestAnalyzeMissingTemplate(t *testing.T) {
	_, _, code := exec(t, "", false, "analyze")
	assert.Equal(t, app.ExitStatus(1), code)
}

func TestErrorCases(t *testing.T) {
	tests := []struct {
		name     string
		template string
		args     []string
		want     app.ExitStatus
	}{
		{"missing template file", "", []string{filepath.Join(t.TempDir(), "nope.tmpl")}, 1},
		{"parse error", "{{.Unclosed", []string{"--stdin"}, 4},
		{"execute error", "{{.Missing}}", []string{"--stdin"}, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, code := exec(t, tt.template, false, tt.args...)
			assert.Equal(t, tt.want, code)
		})
	}
}

func TestPiped(t *testing.T) {
	regular, err := os.CreateTemp(t.TempDir(), "stdin")
	require.NoError(t, err)
	t.Cleanup(func() { _ = regular.Close() })
	assert.True(t, piped(regular), "a regular file carries no char-device bit, so stdin reads as piped")

	device, err := os.Open(os.DevNull)
	require.NoError(t, err)
	t.Cleanup(func() { _ = device.Close() })
	assert.False(t, piped(device), "a character device is a terminal, not a pipe")
}

func TestExecute(t *testing.T) {
	t.Setenv("RENDERIZER_TESTING", "true")
	device, err := os.Open(os.DevNull)
	require.NoError(t, err)
	t.Cleanup(func() { _ = device.Close() })

	var stdout, stderr bytes.Buffer
	code := execute([]string{"renderizer", "--version"}, device, &stdout, &stderr)
	require.Equal(t, app.ExitStatus(0), code, "stderr: %s", stderr.String())
	assert.Contains(t, stdout.String(), "renderizer")
}

func TestMainEntry(t *testing.T) {
	originalExit, originalArgs := osExit, os.Args
	t.Cleanup(func() { osExit, os.Args = originalExit, originalArgs })

	var code int
	osExit = func(c int) { code = c }
	os.Args = []string{"renderizer", "--version"}

	main()
	assert.Equal(t, 0, code)
}
