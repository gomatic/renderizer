package app

import "io"

// Write renders a command's output to w and returns the command's error. The
// output is written even when err is non-nil, so a partially-rendered result
// (e.g. several templates where a later one failed) still reaches the writer.
// This is the shared seam between every command action and its writer, mirroring
// template.cli's app.Default but for raw byte output rather than JSON.
func Write(w io.Writer, output []byte, err error) error {
	_, _ = w.Write(output)
	return err
}
