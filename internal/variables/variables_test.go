package variables_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/constants"
	"github.com/gomatic/renderizer/internal/variables"
)

const timeFormat = variables.TimeFormat("20060102T150405")

func TestTokenize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		cliArgs []string
		assigns []string
	}{
		{
			name: "empty",
		},
		{
			name:    "arbitrary long variable is an assignment",
			args:    []string{"--name=World"},
			assigns: []string{"--name=World"},
		},
		{
			name:    "bare arbitrary long variable",
			args:    []string{"--flag"},
			assigns: []string{"--flag"},
		},
		{
			name:    "capitalize toggle is an assignment",
			args:    []string{"-C"},
			assigns: []string{"-C"},
		},
		{
			name:    "known long flag passes through",
			args:    []string{"--verbose"},
			cliArgs: []string{"--verbose"},
		},
		{
			name:    "known long alias passes through",
			args:    []string{"--debug"},
			cliArgs: []string{"--debug"},
		},
		{
			name:    "known value long flag with equals",
			args:    []string{"--settings=a.yaml"},
			cliArgs: []string{"--settings=a.yaml"},
		},
		{
			name:    "known value long flag with space",
			args:    []string{"--settings", "a.yaml"},
			cliArgs: []string{"--settings", "a.yaml"},
		},
		{
			name:    "short flags pass through to cli",
			args:    []string{"-S", "a.yaml", "-V"},
			cliArgs: []string{"-S", "a.yaml", "-V"},
		},
		{
			name:    "templates and subcommands pass through",
			args:    []string{"analyze", "file.tmpl"},
			cliArgs: []string{"analyze", "file.tmpl"},
		},
		{
			name:    "mixed stream",
			args:    []string{"--name=World", "-C", "--verbose", "t.tmpl", "--settings", "s.yaml"},
			cliArgs: []string{"--verbose", "t.tmpl", "--settings", "s.yaml"},
			assigns: []string{"--name=World", "-C"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := variables.Tokenize(tt.args)
			assert.Equal(t, tt.cliArgs, got.Args)
			assert.Equal(t, tt.assigns, got.Assignments)
		})
	}
}

func TestAssignments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		want       variables.Context
		name       string
		tokens     []string
		capitalize variables.Capitalization
	}{
		{
			name:       "single scalar collapses",
			tokens:     []string{"--name=World"},
			capitalize: true,
			want:       variables.Context{"Name": "World"},
		},
		{
			name:       "repeated names append into slice",
			tokens:     []string{"--items=a", "--items=b"},
			capitalize: true,
			want:       variables.Context{"Items": []any{"a", "b"}},
		},
		{
			name:       "dotted notation nests",
			tokens:     []string{"--a.b.c=deep"},
			capitalize: false,
			want:       variables.Context{"a": map[string]any{"b": map[string]any{"c": "deep"}}},
		},
		{
			name:       "bare name is boolean true",
			tokens:     []string{"--flag"},
			capitalize: true,
			want:       variables.Context{"Flag": true},
		},
		{
			name:       "toggle disables capitalization midway",
			tokens:     []string{"--name=first", "-C", "--foo=second"},
			capitalize: true,
			want:       variables.Context{"Name": "first", "foo": "second"},
		},
		{
			name:       "typed values",
			tokens:     []string{"--count=42", "--ratio=3.14", "--on=true"},
			capitalize: true,
			want:       variables.Context{"Count": int64(42), "Ratio": 3.14, "On": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := variables.Assignments(tt.tokens, tt.capitalize, timeFormat)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAssignmentsEmptySegment(t *testing.T) {
	t.Parallel()
	got, err := variables.Assignments([]string{"--a..b=x"}, true, timeFormat)
	require.NoError(t, err)
	want := variables.Context{"A": map[string]any{"": map[string]any{"B": "x"}}}
	assert.Equal(t, want, got)
}

func TestAssignmentsMergeError(t *testing.T) {
	t.Parallel()
	_, err := variables.Assignments([]string{"--a.b=2", "--a=1"}, true, timeFormat)
	require.ErrorIs(t, err, constants.ErrMergeContext)
}

func TestAssignmentsTimeValue(t *testing.T) {
	t.Parallel()
	got, err := variables.Assignments([]string{"--when=20231225T120000"}, true, timeFormat)
	require.NoError(t, err)
	want := time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC)
	assert.Equal(t, want, got["When"])
}

func TestRetype(t *testing.T) {
	t.Parallel()
	source := map[string]any{
		"int":     7,
		"str":     "hello",
		"boolean": true,
		"nested":  map[string]any{"n": 3},
		"multi":   []any{"x", "y"},
		"single":  []any{"only"},
	}
	got := variables.Retype(source, timeFormat, true)
	assert.Equal(t, int64(7), got["int"])
	assert.Equal(t, "hello", got["str"])
	assert.Equal(t, true, got["boolean"])
	assert.Equal(t, map[string]any{"n": int64(3)}, got["nested"])
	assert.Equal(t, []any{"x", "y"}, got["multi"])
	assert.Equal(t, "only", got["single"])
}

func TestRetypeNoCollapse(t *testing.T) {
	t.Parallel()
	got := variables.Retype(map[string]any{"single": []any{"only"}}, timeFormat, false)
	assert.Equal(t, []any{"only"}, got["single"])
}
