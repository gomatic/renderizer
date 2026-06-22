package render

import "io"

// Config holds everything Run needs: the flag-bound options the CLI tier binds,
// the arguments Tokenize extracted, and the injected IO seams the composition
// root supplies. It carries no behavior.
type Config struct {
	// Flag-bound options.
	Settings    SettingsFiles
	MissingKey  MissingKeyOption
	Environment EnvironmentName
	Stdin       StdinEnabled
	Testing     TestingEnabled
	Debugging   DebuggingEnabled
	Verbose     VerboseEnabled

	// Parsed arguments and rendering options set by the composition root.
	Capitalize  Capitalization
	TimeFormat  TimeFormat
	Assignments AssignmentTokens
	Templates   TemplateFiles

	// Injected IO seams.
	Source   io.Reader
	ReadFile ReadFileFunc
	Exists   ExistsFunc
	Getwd    GetwdFunc
	Environ  EnvironFunc
}
