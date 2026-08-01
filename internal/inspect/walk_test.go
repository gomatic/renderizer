package inspect_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/inspect"
	"github.com/gomatic/renderizer/internal/template"
)

// analyzed returns the skeleton for source, or fails. The walker is unexported,
// so its contracts are asserted through the skeleton it produces — which is
// also the only thing a user ever sees, so a claim that does not show up here
// is a claim about nothing.
func analyzed(t *testing.T, source string) string {
	t.Helper()
	model, err := inspect.Analyze(template.Funcs(false), "test", []byte(source))
	require.NoError(t, err)
	return string(inspect.Skeleton(model))
}

// TestWalkPipeRecordsEveryFieldInEveryCommand names walkPipe's claim. It
// dereferences the pipe without a nil check, justified by "actions, if
// conditions, and parenthesized arguments never carry a nil pipe" — so that
// sentence is the only thing standing between renderizer and a panic on a
// user's template. Each shape below is one of the three the claim names; if any
// one of them could produce a nil pipe, `renderizer --analyze` would crash on
// ordinary input rather than report a data model.
//
// The second half of the claim is that EVERY command's arguments are recorded,
// not just the first: a pipeline's stages each reference fields, and a walker
// that stopped at the head would silently omit them from the skeleton, telling
// the user their template needs less data than it does.
func TestWalkPipeRecordsEveryFieldInEveryCommand(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "an action pipe carries fields through every stage",
			source: "{{ .First | printf \"%s\" }}{{ .Second }}",
			want:   []string{"First", "Second"},
		},
		{
			name:   "an if condition is a pipe too",
			source: "{{ if .Flag }}{{ .Shown }}{{ end }}",
			want:   []string{"Flag", "Shown"},
		},
		{
			name:   "a parenthesized argument is a nested pipe",
			source: "{{ printf \"%s\" (.Inner) }}",
			want:   []string{"Inner"},
		},
		{
			name:   "later commands in a pipeline are not skipped",
			source: "{{ .Head | printf \"%s%s\" .Tail }}",
			want:   []string{"Head", "Tail"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := analyzed(t, tc.source)
			for _, field := range tc.want {
				assert.Contains(t, got, field, "field %s must reach the skeleton", field)
			}
		})
	}
}

// TestPipeFieldResolvesTheRangeOperand names pipeField's claim. It indexes both
// pipe.Cmds and the command's Args from the end with no bounds check, resting
// on "a range/with pipe always has at least one command, each with at least one
// argument". The shapes below are the ones a template can actually contain, and
// each must resolve to the collection the range or with operates on — the LAST
// argument of the LAST command, which is what a piped operand makes non-obvious.
func TestPipeFieldResolvesTheRangeOperand(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		source  string
		nested  string
		operand string
	}{
		{
			name:    "range over a plain field",
			source:  "{{ range .Items }}{{ .Leaf }}{{ end }}",
			operand: "Items",
			nested:  "Leaf",
		},
		{
			name:    "with over a plain field",
			source:  "{{ with .Section }}{{ .Leaf }}{{ end }}",
			operand: "Section",
			nested:  "Leaf",
		},
		{
			name:    "range with an index and value declaration",
			source:  "{{ range $i, $v := .Items }}{{ $v.Leaf }}{{ end }}",
			operand: "Items",
			nested:  "Leaf",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := analyzed(t, tc.source)

			assert.Contains(t, got, tc.operand, "the operand is the collection the block iterates")
			assert.Contains(t, got, tc.nested, "and its element's fields nest beneath it")
			assert.Less(t, strings.Index(got, tc.operand), strings.Index(got, tc.nested),
				"the element's fields must nest UNDER the operand, not sit beside it")
		})
	}
}
