// Command renderizer renders Go text/template files from command-line
// variables, YAML settings, and the environment.
//
// This file is the composition root: it tokenizes the arguments, assembles the
// runtime seams, wires the command packages from internal/app/commands, and maps
// the resulting error to a process exit code. All behavior lives in those
// command packages, the domain tier, and the implementation packages.
package main

import (
	"context"
	"io"
	"os"
	"os/signal"

	"github.com/urfave/cli/v3"

	"github.com/gomatic/renderizer/internal/app"
	analyzecmd "github.com/gomatic/renderizer/internal/app/commands/analyze"
	rendercmd "github.com/gomatic/renderizer/internal/app/commands/render"
	versioncmd "github.com/gomatic/renderizer/internal/app/commands/version"
	"github.com/gomatic/renderizer/internal/variables"
)

const timeFormat = "20060102T150405"

// appVersion is overridden at build time via -ldflags "-X main.appVersion=...".
var appVersion = "dev"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	os.Exit(int(run(ctx, os.Args, os.Stdin, os.Stdout, os.Stderr, piped(os.Stdin))))
}

// run tokenizes the arguments, builds the root render command with the analyze
// and version subcommands, runs it, and returns the resulting exit code.
func run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer, isPiped bool) app.ExitStatus {
	// Expose the version to templates as {{.env.RENDERIZER_VERSION}}.
	_ = os.Setenv("RENDERIZER_VERSION", appVersion)
	tokens := variables.Tokenize(args[1:])
	rt := app.Runtime{
		Source:      stdin,
		ReadFile:    os.ReadFile,
		Exists:      exists,
		Getwd:       os.Getwd,
		Environ:     os.Environ,
		Assignments: tokens.Assignments,
		Capitalize:  true,
		TimeFormat:  timeFormat,
		Piped:       isPiped,
	}
	root := rendercmd.Command(rt)
	root.Version = appVersion
	root.EnableShellCompletion = true
	root.Writer = stdout
	root.ErrWriter = stderr
	root.Commands = []*cli.Command{
		analyzecmd.Command(rt),
		versioncmd.Command(root.Name, appVersion),
	}
	err := root.Run(ctx, append([]string{root.Name}, tokens.Args...))
	return app.ExitCode(err)
}

// exists reports whether a path exists, for default-template discovery.
func exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

// piped reports whether f is a pipe rather than a terminal, so stdin is used
// automatically for `… | renderizer`.
func piped(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice == 0
}
