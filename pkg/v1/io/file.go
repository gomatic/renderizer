package io

import (
	"io"
	"os"
)

var _ io.ReadCloser = &File{}

var stdin io.ReadCloser = os.Stdin

//
type File struct {
	name     string
	rc       io.ReadCloser
}

//
func (f *File) Read(p []byte) (n int, err error) {
	return f.rc.Read(p)
}

//
func (f *File) Close() error {
	return f.rc.Close()
}

// If name is "", this returns stdin.
func NewFile(name string) (f *File, err error) {
	f = &File{
		name: name,
		rc:   stdin,
	}
	if name == "" {
		return f, nil
	}
	fp, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	f.rc = EOFCloser(fp)
	return f, err
}
