package variables

import "strings"

// capitalizeToggle is the positional flag that flips capitalization of the
// variable names that follow it.
const capitalizeToggle = "-C"

// Tokens is the result of splitting raw arguments into the two streams the CLI
// needs: arbitrary variable assignments parsed by this package, and everything
// else (known flags, positional templates, and subcommands) handed to
// urfave/cli. Arbitrary `--name=value` flags must be extracted here because
// urfave/cli rejects undefined flags.
type Tokens struct {
	Args        []string
	Assignments []string
}

type (
	// argument is one raw command-line argument under classification.
	argument string
	// flagName is a long flag's name, dashes and value stripped.
	flagName string
)

// Tokenize splits args (excluding the program name) into urfave/cli arguments
// and arbitrary variable assignments. Only long `--name[=value]` flags that are
// not renderizer's own, plus the positional `-C` toggle, are treated as
// assignments; everything else — known flags (long or short), templates, and
// subcommands — passes through to urfave/cli.
func Tokenize(args []string) Tokens {
	tokens := Tokens{}
	for _, arg := range args {
		if isAssignment(argument(arg)) {
			tokens.Assignments = append(tokens.Assignments, arg)
			continue
		}
		tokens.Args = append(tokens.Args, arg)
	}
	return tokens
}

// isAssignment reports whether arg is an arbitrary variable assignment: the -C
// toggle, or a long flag whose name is not one of renderizer's own.
func isAssignment(arg argument) bool {
	if string(arg) == capitalizeToggle {
		return true
	}
	if !strings.HasPrefix(string(arg), "--") {
		return false
	}
	key, _, _ := strings.Cut(strings.TrimPrefix(string(arg), "--"), "=")
	return !knownLongFlag(flagName(key))
}

// knownLongFlag reports whether key is one of renderizer's own long flags (or
// long aliases), which must reach urfave/cli rather than become a variable.
func knownLongFlag(key flagName) bool {
	switch key {
	case "settings", "missing", "environment", "env",
		"stdin", "testing", "debugging", "debug", "verbose", "help", "version":
		return true
	}
	return false
}
