package variables

import (
	"strings"

	"dario.cat/mergo"

	"github.com/gomatic/renderizer/internal/constants"
)

// Capitalization is the initial state of the title-casing toggle applied to
// variable names. It defaults to true (names are title-cased) and flips at each
// -C encountered.
type Capitalization bool

// Context is the typed, possibly nested variable map handed to a template.
type Context map[string]any

// Assignments builds a Context from the ordered assignment tokens produced by
// Tokenize. Each `--name=value` becomes a (possibly nested) entry; a bare
// `--name` becomes boolean true; and each -C flips capitalization of the names
// that follow. Repeated names append into a slice, then single-element slices
// collapse back to scalars so a name given once is a scalar.
func Assignments(tokens []string, capitalize Capitalization, format TimeFormat) (Context, error) {
	global := map[string]any{}
	for _, token := range tokens {
		if token == capitalizeToggle {
			capitalize = !capitalize
			continue
		}
		entry := assignment(token, capitalize, format)
		if err := mergo.Merge(&global, entry, mergo.WithAppendSlice); err != nil {
			return nil, constants.ErrMergeContext.With(err)
		}
	}
	return Retype(global, format, true), nil
}

// assignment parses one `--name=value` (or bare `--name`) token into a single
// nested map whose leaf is a one-element slice, so repeats append on merge.
func assignment(token string, capitalize Capitalization, format TimeFormat) map[string]any {
	name, value := splitAssignment(token)
	path := casedPath(name, capitalize)
	return nest(path, []any{leaf(value, format)})
}

// splitAssignment strips leading dashes and separates name from an optional
// value. A token without `=` yields an empty value, signaling a boolean.
func splitAssignment(token string) (string, *string) {
	body := strings.TrimLeft(token, "-")
	parts := strings.SplitN(body, "=", 2)
	if len(parts) == 1 {
		return parts[0], nil
	}
	return parts[0], &parts[1]
}

// leaf types a present value or yields boolean true for a bare name.
func leaf(value *string, format TimeFormat) any {
	if value == nil {
		return true
	}
	return typed(format, *value)
}

// casedPath splits a dotted name into segments, title-casing each when
// capitalization is enabled.
func casedPath(name string, capitalize Capitalization) []string {
	segments := strings.Split(name, ".")
	if !bool(capitalize) {
		return segments
	}
	for i, segment := range segments {
		segments[i] = title(segment)
	}
	return segments
}

// title upper-cases the first rune and lower-cases the rest, matching the
// historical capitalization of variable names.
func title(segment string) string {
	if segment == "" {
		return segment
	}
	return strings.ToUpper(segment[:1]) + strings.ToLower(segment[1:])
}

// nest builds a map nesting leaf under the given path of keys.
func nest(path []string, value any) map[string]any {
	root := map[string]any{}
	current := root
	for i, key := range path {
		if i == len(path)-1 {
			current[key] = value
			break
		}
		next := map[string]any{}
		current[key] = next
		current = next
	}
	return root
}
