// Package constants declares renderizer's sentinel error values. The error
// mechanism — the errors.Is-matchable string type and its cause-wrapping With —
// lives in the shared gomatic/go-error library; these values are renderizer's
// own.
package constants

// Imported bare (the package is named errs); this file declares only sentinels,
// so each declaration reads errs.Const.
import errs "github.com/gomatic/go-error"

// Keep these constants sorted alphabetically.
const (
	ErrExecuteTemplate errs.Const = "failed to execute template"
	ErrMergeContext    errs.Const = "failed to merge context"
	ErrMissingTemplate errs.Const = "missing template name"
	ErrOpenTemplate    errs.Const = "failed to open template"
	ErrParseSettings   errs.Const = "failed to parse settings file"
	ErrParseTemplate   errs.Const = "failed to parse template"
	ErrReadSettings    errs.Const = "failed to read settings file"
	ErrReadTemplate    errs.Const = "failed to read template"
	ErrRenderPanic     errs.Const = "template rendering panicked"
)
