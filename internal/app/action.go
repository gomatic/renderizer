package app

import (
	"io"

	"github.com/gomatic/renderizer/internal/constants"
)

// Write renders a command's output to w and returns the command's error. The
// output is written even when err is non-nil, so a partially-rendered result
// (e.g. several templates where a later one failed) still reaches the writer.
// This is the shared seam between every command action and its writer, mirroring
// template.cli's app.Default but for raw byte output rather than JSON.
//
// A failed WRITE is a failure of the whole program, not a detail: delivering
// rendered output is what renderizer is for, so a run that could not deliver it
// has not succeeded. Discarding that error let a broken pipe or a full disk
// exit 0 having written nothing — every caller is a top-level command action
// whose return value becomes the process exit status.
//
// The command's own error still wins when both fail: it is the more specific
// diagnosis, and it explains a short write rather than competing with it.
func Write(w io.Writer, output []byte, err error) error {
	_, writeErr := w.Write(output)
	if err != nil {
		return err
	}
	if writeErr != nil {
		return constants.ErrWriteOutput.With(writeErr)
	}
	return nil
}
