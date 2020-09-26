package io

import (
	"fmt"
	"io"
)

//
func Readers(ss ...string) (os []io.Reader, err error) {
	stdin := 0
	os = make([]io.Reader, len(ss))
	for i, s := range ss {
		if s == "" {
			stdin++
		}
		if stdin > 1 {
			continue
		}
		f, err_ := NewFile(s)
		if err_ != nil {
			err = chain(err, err_)
		}
		os[i] = f
	}
	if stdin > 1 {
		err = chain(err, fmt.Errorf("too many stdin: %d", stdin))
	}
	return os, err
}

func chain(err1, err2 error) error {
	if err1 == nil {
		return err2
	}
	return fmt.Errorf("%w: %s", err2, err1)
}
