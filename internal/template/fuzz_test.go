package template_test

import (
	"errors"
	"testing"
	texttemplate "text/template"

	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/constants"
	"github.com/gomatic/renderizer/internal/template"
)

// FuzzRenderFailsOnlyWithItsSentinels pins Render's whole error contract on
// arbitrary template source: it never crashes the process — a parse failure is
// ErrParseTemplate, an execute failure is ErrExecuteTemplate, and a panic in a
// template function is recovered as ErrRenderPanic — and any failure returns
// nil output while a success is deterministic.
func FuzzRenderFailsOnlyWithItsSentinels(f *testing.F) {
	for _, seed := range []string{
		``,
		`plain text`,
		`{{.Name}}`,
		`{{range .Items}}{{.}}{{end}}`,
		`{{boom}}`,
		`{{if .A}}{{.B.C}}{{end}}`,
		`{{template "missing"}}`,
		`{{.A | printf "%d"}}`,
		"{{`unterminated",
		`{{define "x"}}x{{end}}{{template "x"}}`,
	} {
		f.Add([]byte(seed))
	}
	funcs := texttemplate.FuncMap{"boom": func() string { panic("boom") }}
	data := map[string]any{"Name": "n", "Items": []int{1}, "A": true}
	f.Fuzz(func(t *testing.T, source []byte) {
		out, err := template.Render(funcs, template.MissingKey("zero"), "fuzz", source, data)
		if err != nil {
			require.True(t,
				errors.Is(err, constants.ErrParseTemplate) ||
					errors.Is(err, constants.ErrExecuteTemplate) ||
					errors.Is(err, constants.ErrRenderPanic),
				"error must be one of the three documented sentinels, got %v", err)
			require.Nil(t, out, "a failed render must return no output")
			return
		}
		again, err := template.Render(funcs, template.MissingKey("zero"), "fuzz", source, data)
		require.NoError(t, err, "a successful render must succeed again")
		require.Equal(t, out, again, "rendering is deterministic for the same source and data")
	})
}
