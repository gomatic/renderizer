package app

import (
	"errors"

	"github.com/gomatic/renderizer/internal/constants"
)

// Exit codes preserve the historical renderizer semantics: distinct codes per
// failure stage so scripts can distinguish a read failure from a template
// parse or execute failure.
type ExitStatus int

const (
	exitSuccess ExitStatus = 0
	exitGeneric ExitStatus = 1
	exitRead    ExitStatus = 2
	exitParse   ExitStatus = 4
	exitExecute ExitStatus = 8
	exitPanic   ExitStatus = 15
)

// ExitCode maps a Run error to a process exit code. A nil error is success; a
// recognized sentinel maps to its historical code; anything else is a generic
// failure.
func ExitCode(err error) ExitStatus {
	switch {
	case err == nil:
		return exitSuccess
	case errors.Is(err, constants.ErrParseTemplate):
		return exitParse
	case errors.Is(err, constants.ErrExecuteTemplate):
		return exitExecute
	case errors.Is(err, constants.ErrReadTemplate):
		return exitRead
	case errors.Is(err, constants.ErrRenderPanic):
		return exitPanic
	default:
		return exitGeneric
	}
}
