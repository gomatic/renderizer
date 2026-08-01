package constants_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/constants"
)

// TestErrorWithKeepsBothTheSentinelAndTheCauseMatchable pins what this package
// owns about its sentinels: after With, errors.Is must find the sentinel AND,
// when one was supplied, the cause. Losing either is a silent break — losing
// the sentinel makes every ExitCode arm fall through to the generic failure,
// and losing the cause puts the underlying os or template error out of reach of
// a caller that wants to inspect it.
//
// How the two are RENDERED into a message is go-error's contract, not this
// package's, and is asserted in go-error's own tests. Matching on the message
// here would pin a format this package does not define, and would keep passing
// whether or not the sentinel survived — which is precisely why the rule bans
// it.
func TestErrorWithKeepsBothTheSentinelAndTheCauseMatchable(t *testing.T) {
	t.Parallel()
	cause := errors.New("boom")

	for _, tc := range []struct {
		wantErr     error
		name        string
		args        []any
		wantIsCause bool
	}{
		{name: "bare sentinel", wantErr: nil, args: nil},
		{name: "with cause", wantErr: cause, args: nil, wantIsCause: true},
		{name: "with args only", wantErr: nil, args: []any{"file", "x.tmpl"}},
		{name: "with cause and args", wantErr: cause, args: []any{"x.tmpl"}, wantIsCause: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := constants.ErrOpenTemplate.With(tc.wantErr, tc.args...)

			require.ErrorIs(t, got, constants.ErrOpenTemplate,
				"the sentinel must survive wrapping, whatever else is attached")
			assert.Equal(t, tc.wantIsCause, errors.Is(got, cause),
				"a cause is matchable exactly when one was supplied")
		})
	}
}

// TestErrorIsDistinct guards the premise every other error test rests on. If two
// sentinels matched each other, a test asserting the wrong one would still pass
// and the ExitCode switch would return whichever arm it reached first.
func TestErrorIsDistinct(t *testing.T) {
	t.Parallel()

	assert.NotErrorIs(t, constants.ErrParseTemplate, constants.ErrExecuteTemplate)
	assert.NotErrorIs(t, fmt.Errorf("x"), constants.ErrReadTemplate)
}

// TestErrorTextIsStable pins the one thing about the prose this package does
// own: the text it declares. This is not error MATCHING — nothing branches on
// it — and it is checked on the constant itself rather than on a composed
// error, so it cannot stand in for an identity assertion.
func TestErrorTextIsStable(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "missing template name", constants.ErrMissingTemplate.Error())
	assert.Equal(t, "failed to open template", constants.ErrOpenTemplate.Error())
}
