package variables

import (
	"strconv"
	"time"
)

// TimeFormat is the layout used to recognize a string value as a time.Time.
type TimeFormat string

// rawText is the untyped textual form of a variable value — from a
// command-line assignment or a settings file — before coercion.
type rawText string

// typed coerces raw text into the most specific Go type it parses as, trying
// int64, float64, bool, then time (using format), and falling back to the
// original string. This is the single source of truth for value typing, so
// command-line values and settings values coerce identically.
func typed(format TimeFormat, raw rawText) any {
	if v, err := strconv.ParseInt(string(raw), 10, 64); err == nil {
		return v
	}
	if v, err := strconv.ParseFloat(string(raw), 64); err == nil {
		return v
	}
	if v, err := strconv.ParseBool(string(raw)); err == nil {
		return v
	}
	if v, err := time.Parse(string(format), string(raw)); err == nil {
		return v
	}
	return string(raw)
}

// CollapseSingles, when true, replaces a single-element slice with its lone
// element. Command-line variables are wrapped in slices so repeats append into
// a list; collapsing restores a scalar when a name was given exactly once.
type CollapseSingles bool

// Retype walks a decoded map and coerces every leaf to its specific type:
// untyped strings become typed values, ints widen to int64, and (when
// shouldCollapse is set) single-element slices unwrap to their element. It
// mutates and returns source.
func Retype(source map[string]any, format TimeFormat, shouldCollapse CollapseSingles) map[string]any {
	for key, value := range source {
		source[key] = retypeValue(value, format, shouldCollapse)
	}
	return source
}

// retypeValue coerces a single decoded value, recursing into maps and slices.
func retypeValue(value any, format TimeFormat, shouldCollapse CollapseSingles) any {
	switch typedValue := value.(type) {
	case map[string]any:
		return Retype(typedValue, format, shouldCollapse)
	case []any:
		return retypeSlice(typedValue, format, shouldCollapse)
	case int:
		return int64(typedValue)
	case string:
		return typed(format, rawText(typedValue))
	default:
		return value
	}
}

// retypeSlice coerces each element, collapsing a single-element slice to its
// element when shouldCollapse is set.
func retypeSlice(slice []any, format TimeFormat, shouldCollapse CollapseSingles) any {
	if bool(shouldCollapse) && len(slice) == 1 {
		return retypeValue(slice[0], format, shouldCollapse)
	}
	for i, element := range slice {
		slice[i] = retypeValue(element, format, shouldCollapse)
	}
	return slice
}
