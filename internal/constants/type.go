package constants

import (
	"fmt"
	"strings"
)

// Error is the package's sentinel-error type. Declare every error the program
// can emit as a const of this type, so each path is matchable with errors.Is
// instead of by string comparison.
type Error string

func (e Error) Error() string { return string(e) }

// With wraps a cause and appends contextual args. A non-nil cause is joined
// with %w so errors.Is still matches both the sentinel and the cause. The args
// are rendered space-separated, so callers pass clean key/value pairs —
// .With(err, "file", name) — without baking separators into the key.
func (e Error) With(err error, args ...any) error {
	out := error(e)
	if err != nil {
		out = fmt.Errorf("%w: %w", e, err)
	}
	if len(args) > 0 {
		out = fmt.Errorf("%w: %s", out, strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
	}
	return out
}

var _ error = Error("")
