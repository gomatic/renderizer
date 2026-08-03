package inspect_test

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/gomatic/renderizer/internal/constants"
	"github.com/gomatic/renderizer/internal/inspect"
)

// FuzzAnalyzeInfersASoundModel pins Analyze's contract on arbitrary template
// source: it never crashes — a source that won't parse fails with
// ErrParseTemplate and an empty model, and any source that parses yields a
// deterministic model whose Skeleton renders as valid YAML.
func FuzzAnalyzeInfersASoundModel(f *testing.F) {
	for _, seed := range []string{
		``,
		`plain`,
		`{{.A.B}}`,
		`{{range .Items}}{{.Name}}{{end}}`,
		`{{with .Sub}}{{.Leaf}}{{end}}`,
		`{{if .Flag}}{{.Then}}{{else}}{{.Else}}{{end}}`,
		`{{range $i, $v := .Rows}}{{$v.Cell}}{{end}}`,
		`{{$.Top}}{{.Top}}`,
		"{{`unterminated",
		`{{noSuchFunc .X}}`,
	} {
		f.Add([]byte(seed))
	}
	funcs := template.FuncMap{"noSuchFunc": func(any) string { return "" }}
	f.Fuzz(func(t *testing.T, source []byte) {
		model, err := inspect.Analyze(funcs, "fuzz", source)
		if err != nil {
			require.ErrorIs(t, err, constants.ErrParseTemplate, "the only failure is the parse sentinel")
			require.Nil(t, model.Fields, "a failed analysis must return the zero model")
			return
		}
		again, err := inspect.Analyze(funcs, "fuzz", source)
		require.NoError(t, err, "a successful analysis must succeed again")
		require.Equal(t, model, again, "analysis is deterministic for the same source")

		var decoded any
		require.NoError(t, yaml.Unmarshal(inspect.Skeleton(model), &decoded),
			"the model's skeleton must be valid YAML")
	})
}
