package inspect_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/constants"
	"github.com/gomatic/renderizer/internal/inspect"
	"github.com/gomatic/renderizer/internal/template"
)

// skeleton analyzes source and returns its YAML data-model skeleton.
func skeleton(t *testing.T, source string) string {
	t.Helper()
	model, err := inspect.Analyze(template.Funcs(false), "test", []byte(source))
	require.NoError(t, err)
	return string(inspect.Skeleton(model))
}

func TestAnalyzeSkeleton(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		source   string
		contains []string
	}{
		{
			name:     "scalar field",
			source:   "{{.Name}}",
			contains: []string{"Name: \"\""},
		},
		{
			name:     "nested fields",
			source:   "{{.A.B.C}}",
			contains: []string{"A:", "B:", "C: \"\""},
		},
		{
			name:     "range list of scalars with index and value",
			source:   "{{range $i, $c := .Items}}{{$i}}{{$c}}{{end}}",
			contains: []string{"Items:", "- \"\""},
		},
		{
			name:     "range list of structs",
			source:   "{{range .Users}}{{.Email}}{{end}}",
			contains: []string{"Users:", "Email: \"\""},
		},
		{
			name:     "range value variable carries element fields",
			source:   "{{range $u := .People}}{{$u.Name}}{{end}}",
			contains: []string{"People:", "Name: \"\""},
		},
		{
			name:     "range else uses outer scope",
			source:   "{{range .Items}}{{.}}{{else}}{{.Empty}}{{end}}",
			contains: []string{"Items:", "Empty: \"\""},
		},
		{
			name:     "with shifts dot",
			source:   "{{with .Obj}}{{.Field}}{{end}}",
			contains: []string{"Obj:", "Field: \"\""},
		},
		{
			name:     "with declaration binds variable",
			source:   "{{with $o := .Obj}}{{$o.Field}}{{end}}",
			contains: []string{"Obj:", "Field: \"\""},
		},
		{
			name:     "with over literal keeps outer dot",
			source:   `{{with "x"}}{{.Top}}{{end}}`,
			contains: []string{"Top: \"\""},
		},
		{
			name:     "if branches and condition",
			source:   "{{if .Cond}}{{.Affirm}}{{else}}{{.Deny}}{{end}}",
			contains: []string{"Cond: \"\"", "Affirm: \"\"", "Deny: \"\""},
		},
		{
			name:     "range over a range variable's field",
			source:   "{{range $x := .Items}}{{range $x.Subs}}{{.Leaf}}{{end}}{{end}}",
			contains: []string{"Items:", "Subs:", "Leaf: \"\""},
		},
		{
			name:     "root variable resolves to top",
			source:   "{{range .Items}}{{$.Root}}{{end}}",
			contains: []string{"Items:", "Root: \"\""},
		},
		{
			name:     "bare dollar is a no-op",
			source:   "{{range .Items}}{{$}}{{end}}",
			contains: []string{"Items:"},
		},
		{
			name:     "chain trailing fields are ignored, base captured",
			source:   "{{(.A).B}}",
			contains: []string{"A: \"\""},
		},
		{
			name:     "field piped to function",
			source:   `{{.A | printf "%s"}}`,
			contains: []string{"A: \"\""},
		},
		{
			name:     "parenthesized field argument",
			source:   `{{printf "%s" (.B)}}`,
			contains: []string{"B: \"\""},
		},
		{
			name:     "range over non-field leaves an anonymous element",
			source:   `{{.Top}}{{range (printf "x")}}{{.}}{{end}}`,
			contains: []string{"Top: \"\""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := skeleton(t, tt.source)
			for _, want := range tt.contains {
				assert.Contains(t, got, want)
			}
		})
	}
}

func TestAnalyzeUnknownVariableIsIgnored(t *testing.T) {
	t.Parallel()
	// $x is bound by an assignment action, which this analyzer does not track,
	// so $x.B is not attributed; only .A from the assignment is captured.
	got := skeleton(t, "{{$x := .A}}{{$x.B}}")
	assert.Contains(t, got, "A: \"\"")
	assert.NotContains(t, got, "B")
}

func TestAnalyzeEmptyModel(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "# template requires no input data\n", skeleton(t, "just text {{/* comment */}}"))
}

func TestAnalyzeListFlag(t *testing.T) {
	t.Parallel()
	model, err := inspect.Analyze(template.Funcs(false), "test", []byte("{{range .Items}}{{.Name}}{{end}}"))
	require.NoError(t, err)
	items := model.Fields["Items"]
	require.NotNil(t, items)
	assert.True(t, items.List)
	assert.Contains(t, items.Fields, "Name")
}

func TestAnalyzeParseError(t *testing.T) {
	t.Parallel()
	_, err := inspect.Analyze(template.Funcs(false), "test", []byte("{{.Unclosed"))
	require.ErrorIs(t, err, constants.ErrParseTemplate)
}
