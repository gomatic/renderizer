package analyze

// Named types for the analyze config and its injected seam.
type (
	// TemplateFile is the template path to analyze; empty means read stdin.
	TemplateFile string
)

// ReadFileFunc reads a named file. os.ReadFile satisfies it in production.
type ReadFileFunc func(name string) ([]byte, error)
