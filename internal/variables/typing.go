package variables

import (
	"strconv"
	"time"
)

// TimeFormat is the layout used to recognize a string value as a time.Time.
type TimeFormat string

// typed coerces a raw string into the most specific Go type it parses as,
// trying int64, float64, bool, then time (using format), and falling back to
// the original string. This is the single source of truth for value typing, so
// command-line values and settings values coerce identically.
func typed(format TimeFormat, raw string) any {
	if v, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return v
	}
	if v, err := strconv.ParseFloat(raw, 64); err == nil {
		return v
	}
	if v, err := strconv.ParseBool(raw); err == nil {
		return v
	}
	if v, err := time.Parse(string(format), raw); err == nil {
		return v
	}
	return raw
}

// CollapseSingles, when true, replaces a single-element slice with its lone
// element. Command-line variables are wrapped in slices so repeats append into
// a list; collapsing restores a scalar when a name was given exactly once.
type CollapseSingles bool

// Retype walks a decoded map and coerces every leaf to its specific type:
// untyped strings become typed values, ints widen to int64, and (when
// collapse is set) single-element slices unwrap to their element. It mutates
// and returns source.
func Retype(source map[string]any, format TimeFormat, collapse CollapseSingles) map[string]any {
	for key, value := range source {
		source[key] = retypeValue(value, format, collapse)
	}
	return source
}

// retypeValue coerces a single decoded value, recursing into maps and slices.
func retypeValue(value any, format TimeFormat, collapse CollapseSingles) any {
	switch typedValue := value.(type) {
	case map[string]any:
		return Retype(typedValue, format, collapse)
	case []any:
		return retypeSlice(typedValue, format, collapse)
	case int:
		return int64(typedValue)
	case string:
		return typed(format, typedValue)
	default:
		return value
	}
}

// retypeSlice coerces each element, collapsing a single-element slice to its
// element when collapse is set.
func retypeSlice(slice []any, format TimeFormat, collapse CollapseSingles) any {
	if bool(collapse) && len(slice) == 1 {
		return retypeValue(slice[0], format, collapse)
	}
	for i, element := range slice {
		slice[i] = retypeValue(element, format, collapse)
	}
	return slice
}
