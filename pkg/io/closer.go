package io

import (
	"io"
	"sync"
)

//
type eofCloser struct {
	io.ReadCloser
	closed bool
	mux    sync.RWMutex
}

//
func EOFCloser(r io.ReadCloser) io.ReadCloser {
	if r == nil {
		return nil
	}
	return &eofCloser{ReadCloser: r}
}

// Note the semantic change. An EOFCLoser should not return io.EOF but instead
// return the error returned from Close.
func (e *eofCloser) Read(buf []byte) (n int, err error) {
	e.mux.RLock()
	if e.closed {
		e.mux.RUnlock()
		return 0, io.EOF
	}
	e.mux.RUnlock()
	n, err = e.ReadCloser.Read(buf)
	if err == io.EOF {
		err = e.Close()
		if err != nil {
			return n, err
		}
		return n, io.EOF
	}
	return n, err
}

//
func (e *eofCloser) Close() error {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.closed = true
	return e.ReadCloser.Close()
}
