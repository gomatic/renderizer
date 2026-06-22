package constants

// Keep these constants sorted alphabetically.
const (
	ErrExecuteTemplate Error = "failed to execute template"
	ErrMergeContext    Error = "failed to merge context"
	ErrMissingTemplate Error = "missing template name"
	ErrOpenTemplate    Error = "failed to open template"
	ErrParseSettings   Error = "failed to parse settings file"
	ErrParseTemplate   Error = "failed to parse template"
	ErrReadSettings    Error = "failed to read settings file"
	ErrReadTemplate    Error = "failed to read template"
	ErrRenderPanic     Error = "template rendering panicked"
)
