package app

import "io"

// Runtime carries the injected IO seams and pre-parsed arguments the command
// actions need, assembled once by the composition root (cmd) and passed to each
// command. Keeping these out of the flag-bound config lets the domain tier stay
// fully testable with fakes.
type Runtime struct {
	Source      io.Reader
	ReadFile    func(name string) ([]byte, error)
	Exists      func(name string) bool
	Getwd       func() (string, error)
	Environ     func() []string
	TimeFormat  string
	Assignments []string
	Capitalize  bool
	Piped       bool
}
