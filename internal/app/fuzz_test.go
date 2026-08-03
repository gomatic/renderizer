package app_test

import (
	"bytes"
	"testing"

	goerr "github.com/gomatic/go-error"
	"github.com/stretchr/testify/require"

	"github.com/gomatic/renderizer/internal/app"
)

// FuzzWritePassesBytesAndErrorThrough pins Write's contract on arbitrary
// output bytes: the bytes reach the writer verbatim — even when the command
// erred, so a partially-rendered result still lands — and the command's error
// comes back unchanged.
func FuzzWritePassesBytesAndErrorThrough(f *testing.F) {
	f.Add([]byte(nil), false)
	f.Add([]byte(""), true)
	f.Add([]byte("rendered output\n"), false)
	f.Add([]byte{0x00, 0xff, 0xfe}, true)
	const boom goerr.Const = "boom"
	f.Fuzz(func(t *testing.T, output []byte, failed bool) {
		var in error
		if failed {
			in = boom
		}
		var w bytes.Buffer
		out := app.Write(&w, output, in)
		require.True(t, bytes.Equal(output, w.Bytes()), "the output bytes must reach the writer verbatim")
		require.ErrorIs(t, out, in, "the command's error must come back unchanged")
	})
}
