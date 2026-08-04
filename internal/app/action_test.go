package app_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/app"
	"github.com/gomatic/renderizer/internal/constants"
)

func TestWrite(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := app.Write(&buf, []byte("hello"), nil)
	require.NoError(t, err)
	assert.Equal(t, "hello", buf.String())
}

func TestWriteWritesPartialOnError(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	sentinel := errors.New("boom")
	err := app.Write(&buf, []byte("partial"), sentinel)
	require.ErrorIs(t, err, sentinel)
	assert.Equal(t, "partial", buf.String(), "output is written even when the command errors")
}

// failingWriter fails every write, standing in for the broken pipe or full
// disk that makes delivering output impossible.
type failingWriter struct{ err error }

func (f failingWriter) Write([]byte) (int, error) { return 0, f.err }

// TestWriteSurfacesAFailedWriteAsErrWriteOutput names the contract that a run
// which could not deliver its output has not succeeded.
//
// Producing output is renderizer's entire purpose, and Write is the seam every
// top-level command action returns through — so a discarded write error meant
// a broken pipe or full disk exited 0 having written nothing, with the shell
// and any caller treating the run as a success.
func TestWriteSurfacesAFailedWriteAsErrWriteOutput(t *testing.T) {
	t.Parallel()

	broken := errors.New("broken pipe")

	err := app.Write(failingWriter{err: broken}, []byte("undeliverable"), nil)

	require.ErrorIs(t, err, constants.ErrWriteOutput, "the failure must be matchable as renderizer's own sentinel")
	assert.ErrorIs(t, err, broken, "and must retain the cause that explains it")
}

// TestWriteKeepsTheCommandErrorWhenTheWriteAlsoFails pins the precedence. The
// command's error is the more specific diagnosis and usually explains the
// short write, so it must not be displaced by the write failure it caused.
func TestWriteKeepsTheCommandErrorWhenTheWriteAlsoFails(t *testing.T) {
	t.Parallel()

	command := errors.New("template failed")

	err := app.Write(failingWriter{err: errors.New("broken pipe")}, []byte("partial"), command)

	require.ErrorIs(t, err, command)
	assert.NotErrorIs(t, err, constants.ErrWriteOutput, "the command's diagnosis wins over the write it caused")
}
