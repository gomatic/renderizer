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

type (
	// assignmentToken is one raw `--name=value` (or bare `--name`) token.
	assignmentToken string
	// variableName is an assignment's dotted variable name, dashes stripped.
	variableName string
	// nameSegment is one dot-separated segment of a variable name.
	nameSegment string
)

// Assignments builds a Context from the ordered assignment tokens produced by
// Tokenize. Each `--name=value` becomes a (possibly nested) entry; a bare
// `--name` becomes boolean true; and each -C flips capitalization of the names
// that follow. Repeated names append into a slice, then single-element slices
// collapse back to scalars so a name given once is a scalar.
func Assignments(tokens []string, shouldCapitalize Capitalization, format TimeFormat) (Context, error) {
	global := map[string]any{}
	for _, token := range tokens {
		if token == capitalizeToggle {
			shouldCapitalize = !shouldCapitalize
			continue
		}
		entry := assignment(assignmentToken(token), shouldCapitalize, format)
		if err := mergo.Merge(&global, entry, mergo.WithAppendSlice); err != nil {
			return nil, constants.ErrMergeContext.With(err)
		}
	}
	return Retype(global, format, true), nil
}

// assignment parses one `--name=value` (or bare `--name`) token into a single
// nested map whose leaf is a one-element slice, so repeats append on merge.
func assignment(token assignmentToken, shouldCapitalize Capitalization, format TimeFormat) map[string]any {
	name, value := splitAssignment(token)
	path := casedPath(name, shouldCapitalize)
	return nest(path, []any{leaf(value, format)})
}

// rawValue is one assignment's raw value: its text and whether the token
// carried a value at all — a bare `--name` has none, signaling a boolean.
type rawValue struct {
	text  string
	isSet bool
}

// splitAssignment strips leading dashes and separates name from an optional
// value. A token without `=` yields an unset value, signaling a boolean.
func splitAssignment(token assignmentToken) (variableName, rawValue) {
	body := strings.TrimLeft(string(token), "-")
	name, value, found := strings.Cut(body, "=")
	return variableName(name), rawValue{text: value, isSet: found}
}

// leaf types a present value or yields boolean true for a bare name.
func leaf(value rawValue, format TimeFormat) any {
	if !value.isSet {
		return true
	}
	return typed(format, rawText(value.text))
}

// casedPath splits a dotted name into segments, title-casing each when
// capitalization is enabled.
func casedPath(name variableName, shouldCapitalize Capitalization) []string {
	segments := strings.Split(string(name), ".")
	if !bool(shouldCapitalize) {
		return segments
	}
	for i, segment := range segments {
		segments[i] = title(nameSegment(segment))
	}
	return segments
}

// title upper-cases the first rune and lower-cases the rest, matching the
// historical capitalization of variable names.
func title(segment nameSegment) string {
	if segment == "" {
		return ""
	}
	return strings.ToUpper(string(segment[:1])) + strings.ToLower(string(segment[1:]))
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
