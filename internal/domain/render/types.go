package render

// Named types for every Config field. Flag-bound fields are converted from the
// CLI tier via pointer conversion; injected seams are set by the composition
// root (cmd) so every IO branch is reachable from a test.
type (
	// SettingsFiles are the YAML settings file paths (--settings).
	SettingsFiles []string
	// MissingKeyOption is the text/template missingkey option (--missing).
	MissingKeyOption string
	// EnvironmentName is the context key the environment map is bound under (--environment).
	EnvironmentName string
	// StdinEnabled forces reading the template from stdin (--stdin).
	StdinEnabled bool
	// TestingEnabled makes nondeterministic template functions reproducible (--testing).
	TestingEnabled bool
	// DebuggingEnabled enables debug logging (--debugging).
	DebuggingEnabled bool
	// VerboseEnabled enables verbose logging (--verbose).
	VerboseEnabled bool

	// Capitalization is the initial title-casing state for variable names.
	Capitalization bool
	// TimeFormat is the layout used to recognize a value as a time.
	TimeFormat string
	// AssignmentTokens are the arbitrary --name=value / -C tokens from Tokenize.
	AssignmentTokens []string
	// TemplateFiles are the positional template paths from Tokenize.
	TemplateFiles []string
)

// ReadFileFunc reads a named file. os.ReadFile satisfies it in production.
type ReadFileFunc func(name string) ([]byte, error)

// ExistsFunc reports whether a path exists, used for default-template discovery.
type ExistsFunc func(name string) bool

// GetwdFunc returns the working directory, used to derive default names.
type GetwdFunc func() (string, error)

// EnvironFunc returns the process environment as "KEY=VALUE" strings.
type EnvironFunc func() []string
