package template_test

import (
	"testing"
	texttemplate "text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/constants"
	"github.com/gomatic/renderizer/internal/template"
)

func TestNormalizeMissingKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   template.MissingKey
		want template.MissingKey
	}{
		{"zero", "zero"},
		{"error", "error"},
		{"default", "default"},
		{"invalid", "invalid"},
		{"bogus", "error"},
		{"", "error"},
	}
	for _, tt := range tests {
		t.Run(string(tt.in), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, template.NormalizeMissingKey(tt.in))
		})
	}
}

func TestFuncsTesting(t *testing.T) {
	t.Parallel()
	funcs := template.Funcs(true)
	commandLine, ok := funcs["command_line"].(func() string)
	require.True(t, ok)
	assert.Equal(t, "testing", commandLine())
}

func TestFuncsProduction(t *testing.T) {
	t.Parallel()
	funcs := template.Funcs(false)
	assert.Contains(t, funcs, "upper")
	assert.Contains(t, funcs, "command_line")
	commandLine, ok := funcs["command_line"].(func() string)
	require.True(t, ok)
	assert.NotEqual(t, "testing", commandLine(), "production command_line is the real command line, not the testing stub")
}

func TestFuncsTestingRandIsDeterministic(t *testing.T) {
	t.Parallel()
	first := template.Funcs(true)["rand"].(func() int64)()
	second := template.Funcs(true)["rand"].(func() int64)()
	assert.Equal(t, first, second, "testing-mode rand must be reproducible across runs")
}

func TestRenderRecoversPanic(t *testing.T) {
	t.Parallel()
	// A malformed function map makes text/template's Funcs panic; Render must
	// recover it as ErrRenderPanic rather than crash the process.
	funcs := texttemplate.FuncMap{"bad": 42}
	_, err := template.Render(funcs, "error", "test", []byte("hi"), nil)
	require.ErrorIs(t, err, constants.ErrRenderPanic)
}

func TestFuncsIncludeSprig(t *testing.T) {
	t.Parallel()
	got, err := template.Render(template.Funcs(false), "error", "test", []byte(`{{ b64enc "x" }}`), nil)
	require.NoError(t, err)
	assert.Equal(t, "eA==", string(got))
}

func TestFuncsIncludeSprigV3(t *testing.T) {
	t.Parallel()
	// toRawJson exists only in Sprig v3, proving the v3 library is wired in.
	got, err := template.Render(template.Funcs(false), "error", "test", []byte(`{{ toRawJson (list 1 2 3) }}`), nil)
	require.NoError(t, err)
	assert.Equal(t, "[1,2,3]", string(got))
}

func TestFuncsClashPrefersFuncmap(t *testing.T) {
	t.Parallel()
	// funcmap's trim is strings.Trim (s, cutset); sprig's trim is one-argument
	// TrimSpace. funcmap wins, so the two-argument form is the one that resolves.
	got, err := template.Render(template.Funcs(false), "error", "test", []byte(`{{ trim "xhix" "x" }}`), nil)
	require.NoError(t, err)
	assert.Equal(t, "hi", string(got))
}

func TestRender(t *testing.T) {
	t.Parallel()
	funcs := template.Funcs(true)
	tests := []struct {
		name    string
		source  string
		data    any
		want    string
		wantErr error
	}{
		{
			name:   "renders variable",
			source: "Hello, {{.Name}}!",
			data:   map[string]any{"Name": "World"},
			want:   "Hello, World!",
		},
		{
			name:   "applies funcmap function",
			source: "{{upper .Text}}",
			data:   map[string]any{"Text": "hello"},
			want:   "HELLO",
		},
		{
			name:    "parse error",
			source:  "{{.Unclosed",
			data:    nil,
			wantErr: constants.ErrParseTemplate,
		},
		{
			name:    "execute error on missing key",
			source:  "{{.Missing}}",
			data:    map[string]any{},
			wantErr: constants.ErrExecuteTemplate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := template.Render(funcs, "error", "test", []byte(tt.source), tt.data)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}
