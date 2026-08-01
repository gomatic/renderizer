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
