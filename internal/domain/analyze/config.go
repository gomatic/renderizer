package analyze

import "io"

// Config holds the analyze inputs: the template to analyze and the injected IO
// seams. It carries no behavior.
type Config struct {
	Template TemplateFile
	Source   io.Reader
	ReadFile ReadFileFunc
}
