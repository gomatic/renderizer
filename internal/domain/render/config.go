package render

import "io"

// Config holds everything Run needs: the flag-bound options the CLI tier binds,
// the arguments Tokenize extracted, and the injected IO seams the composition
// root supplies. It carries no behavior.
type Config struct {
	Source      io.Reader
	Environ     EnvironFunc
	Getwd       GetwdFunc
	Exists      ExistsFunc
	ReadFile    ReadFileFunc
	TimeFormat  TimeFormat
	Environment EnvironmentName
	MissingKey  MissingKeyOption
	Settings    SettingsFiles
	Assignments AssignmentTokens
	Templates   TemplateFiles
	Verbose     VerboseEnabled
	Capitalize  Capitalization
	Debugging   DebuggingEnabled
	Testing     TestingEnabled
	Stdin       StdinEnabled
}
