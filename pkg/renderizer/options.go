package renderizer

//
type Options struct {
	// Capitalization is a positional toggles. The following variable names are capitalized (title-case).
	Capitalize bool
	// Set the Missing Key template option. Defaults to "error".
	MissingKey string
	//
	Config map[string]interface{}
	//
	Arguments []string
	//
	Templates []string
	// Add the environment map to the variables.
	Environment string
	//
	TimeFormat string
	//
	Stdin bool
	//
	Debugging bool
	//
	Verbose bool
	//
	Testing bool
}

//

type Option func(*Options)

//
func WithCapitalization(capitalize bool) Option {
	return func(args *Options) {
		args.Capitalize = capitalize
	}
}

//
func WithMissingKeyString(def string) Option {
	return func(args *Options) {
		args.MissingKey = def
	}
}

//
func WithConfig(config map[string]interface{}) Option {
	if config == nil {
		return nil
	}
	return func(args *Options) {
		args.Config = config
	}
}

//
func WithArguments(arguments []string) Option {
	if arguments == nil {
		return nil
	}
	return func(args *Options) {
		args.Arguments = arguments
	}
}

//
func WithTemplates(templates []string) Option {
	if templates == nil {
		return nil
	}
	return func(args *Options) {
		args.Templates = templates
	}
}

//
func WithEnvironment(environment string) Option {
	return func(args *Options) {
		args.Environment = environment
	}
}

//
func WithTimeFormat(timeFormat string) Option {
	return func(args *Options) {
		args.TimeFormat = timeFormat
	}
}

//
func WithStdin(stdin bool) Option {
	return func(args *Options) {
		args.Stdin = stdin
	}
}

//
func WithDebugging(debugging bool) Option {
	return func(args *Options) {
		args.Debugging = debugging
	}
}

//
func WithVerbose(verbose bool) Option {
	return func(args *Options) {
		args.Verbose = verbose
	}
}

//
func WithTesting(testing bool) Option {
	return func(args *Options) {
		args.Testing = testing
	}
}
