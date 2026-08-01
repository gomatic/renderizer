package variables_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gomatic/renderizer/internal/variables"
)

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

// TestCollapseSinglesRestoresAScalarGivenExactlyOnce names CollapseSingles'
// claim. Every command-line variable is accumulated into a slice so that
// repeating a name appends rather than overwrites; collapsing is what undoes
// that representation choice for the common case. "Exactly once" is the whole
// contract — collapsing a two-element slice would destroy a value the user
// supplied, and failing to collapse a one-element slice hands the template a
// []any where it expects a scalar, so `{{.name}}` renders as `[value]`.
func TestCollapseSinglesRestoresAScalarGivenExactlyOnce(t *testing.T) {
	t.Parallel()
	// Retype mutates and returns its argument, so each case gets its own map.
	fresh := func() map[string]any {
		return map[string]any{
			"once":  []any{"solo"},
			"twice": []any{"first", "second"},
			"none":  []any{},
			"plain": "already scalar",
		}
	}

	collapsed := variables.Retype(fresh(), "20060102", variables.CollapseSingles(true))

	assert.Equal(t, "solo", collapsed["once"], "a name given exactly once becomes its lone element")
	assert.Equal(t, []any{"first", "second"}, collapsed["twice"], "a repeated name stays a list")
	assert.Equal(t, []any{}, collapsed["none"], "an empty list has no lone element to become")
	assert.Equal(t, "already scalar", collapsed["plain"])

	kept := variables.Retype(fresh(), "20060102", variables.CollapseSingles(false))
	assert.Equal(t, []any{"solo"}, kept["once"], "with collapsing off the slice is preserved as given")
}
