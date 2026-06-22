package app_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/app"
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
