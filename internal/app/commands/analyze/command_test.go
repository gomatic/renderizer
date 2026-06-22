package analyze_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/app"
	"github.com/gomatic/renderizer/internal/app/commands/analyze"
	"github.com/gomatic/renderizer/internal/constants"
)

func exec(t *testing.T, rt app.Runtime, args ...string) (string, error) {
	t.Helper()
	cmd := analyze.Command(rt)
	var stdout, stderr bytes.Buffer
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr
	err := cmd.Run(context.Background(), append([]string{"analyze"}, args...))
	return stdout.String(), err
}

func TestAnalyzeFile(t *testing.T) {
	rt := app.Runtime{
		ReadFile: func(string) ([]byte, error) { return []byte("{{.Name}}{{range .Items}}{{.Id}}{{end}}"), nil },
	}
	out, err := exec(t, rt, "t.tmpl")
	require.NoError(t, err)
	assert.Contains(t, out, "Name: \"\"")
	assert.Contains(t, out, "Items:")
	assert.Contains(t, out, "Id: \"\"")
}

func TestAnalyzeStdin(t *testing.T) {
	rt := app.Runtime{Source: strings.NewReader("{{.Greeting}}"), Piped: true}
	out, err := exec(t, rt)
	require.NoError(t, err)
	assert.Contains(t, out, "Greeting: \"\"")
}

func TestAnalyzeMissingTemplate(t *testing.T) {
	_, err := exec(t, app.Runtime{})
	require.ErrorIs(t, err, constants.ErrMissingTemplate)
}

func TestAnalyzeFileOpenError(t *testing.T) {
	rt := app.Runtime{ReadFile: func(string) ([]byte, error) { return nil, os.ErrNotExist }}
	_, err := exec(t, rt, "missing.tmpl")
	require.ErrorIs(t, err, constants.ErrOpenTemplate)
}
