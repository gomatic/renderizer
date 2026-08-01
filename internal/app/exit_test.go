package app_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/app"
	"github.com/gomatic/renderizer/internal/constants"
)

// TestExitCodeMapsEachSentinelToItsHistoricalStatus pins the contract scripts
// depend on: a distinct exit code per failure stage, so a caller can tell a
// template that could not be READ from one that could not be PARSED without
// scraping stderr. Collapsing any two of these into one code is a silent
// break — every renderizer invocation still exits non-zero, and only a script
// branching on the code notices, in production.
//
// Each case matches with errors.Is rather than by message, which is what makes
// the mapping survive a reworded sentinel or an added wrapping layer.
func TestExitCodeMapsEachSentinelToItsHistoricalStatus(t *testing.T) {
	t.Parallel()
	cause := errors.New("boom")

	for _, tc := range []struct {
		wantErr error
		name    string
		want    int
	}{
		{name: "read", wantErr: constants.ErrReadTemplate, want: 2},
		{name: "parse", wantErr: constants.ErrParseTemplate, want: 4},
		{name: "execute", wantErr: constants.ErrExecuteTemplate, want: 8},
		{name: "panic", wantErr: constants.ErrRenderPanic, want: 15},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, app.ExitStatus(tc.want), app.ExitCode(tc.wantErr),
				"the bare sentinel maps to its historical code")

			wrapped := errs(t, tc.wantErr, cause)
			require.ErrorIs(t, wrapped, tc.wantErr,
				"wrapping must preserve the sentinel, or the mapping below is vacuous")
			assert.Equal(t, app.ExitStatus(tc.want), app.ExitCode(wrapped),
				"a sentinel carrying a cause still maps to its own code, not the generic one")
		})
	}
}

// TestExitCodeDistinguishesSuccessFromAnUnrecognizedFailure pins the two ends
// of the mapping. A nil error must not fall through to the generic failure
// code — that would make every successful run look like a failure — and an
// error this package does not recognize must not be silently reported as
// success.
func TestExitCodeDistinguishesSuccessFromAnUnrecognizedFailure(t *testing.T) {
	t.Parallel()

	assert.Equal(t, app.ExitStatus(0), app.ExitCode(nil))
	assert.Equal(t, app.ExitStatus(1), app.ExitCode(errors.New("something else")))
}

// TestExitCodeSentinelsAreMutuallyDistinct guards the mapping's premise. The
// switch returns on the first matching arm, so two sentinels that matched each
// other would make one of them unreachable and its code unreportable — a
// failure the per-sentinel tests above cannot see, because each would still
// pass on the arm that shadowed it.
func TestExitCodeSentinelsAreMutuallyDistinct(t *testing.T) {
	t.Parallel()
	sentinels := []error{
		constants.ErrReadTemplate,
		constants.ErrParseTemplate,
		constants.ErrExecuteTemplate,
		constants.ErrRenderPanic,
	}

	for i, a := range sentinels {
		for j, b := range sentinels {
			if i != j {
				assert.NotErrorIs(t, a, b, "sentinels %d and %d must not match each other", i, j)
			}
		}
	}
}

// errs wraps sentinel around cause the way the packages under test do, so the
// mapping is exercised on the shape it actually receives rather than on a bare
// constant no caller ever returns.
func errs(t *testing.T, sentinel, cause error) error {
	t.Helper()
	type wither interface{ With(error, ...any) error }
	w, ok := sentinel.(wither)
	require.True(t, ok, "every constants sentinel must offer With")
	return w.With(cause)
}

// TestExitCodeReportsUnmappedSentinelsAsGeneric pins the other half of the
// mapping: constants defines more sentinels than ExitCode distinguishes, and
// the ones without a historical code must land on the generic failure rather
// than on some neighbouring stage's code. A sentinel that drifted into, say,
// the parse code would tell a script the template was malformed when it was
// merely missing.
func TestExitCodeReportsUnmappedSentinelsAsGeneric(t *testing.T) {
	t.Parallel()

	for _, sentinel := range []error{constants.ErrOpenTemplate, constants.ErrMissingTemplate} {
		assert.Equal(t, app.ExitStatus(1), app.ExitCode(errs(t, sentinel, nil)),
			"a sentinel with no historical code is a generic failure")
	}
}
