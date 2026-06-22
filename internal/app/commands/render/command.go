// Package render is the app-tier definition of renderizer's default action:
// rendering templates. Because rendering is the tool's root behavior (not a
// subcommand), Command returns the root command carrying the global flags; the
// composition root attaches the analyze and version subcommands to it.
package render

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/gomatic/renderizer/internal/app"
	domain "github.com/gomatic/renderizer/internal/domain/render"
)

const (
	name  = "renderizer"
	usage = "renderizer [options] [--name=value...] template..."
)

// Command returns the root render command, binding the global flags to a config
// and wiring the action that renders and writes the output.
func Command(rt app.Runtime) *cli.Command {
	var cfg domain.Config
	return &cli.Command{
		Name:   name,
		Usage:  usage,
		Flags:  flags(&cfg),
		Action: action(&cfg, rt),
	}
}

// action assembles the full config from flags, runtime seams, and positional
// templates, then runs the render and writes its output.
func action(cfg *domain.Config, rt app.Runtime) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		configure(cfg, rt, cmd)
		logger := app.NewLogger(cmd.Root().ErrWriter, app.Verbose(cfg.Verbose), app.Debugging(cfg.Debugging))
		result, err := domain.Run(ctx, logger, *cfg)
		return app.Write(cmd.Root().Writer, result.Output, err)
	}
}

// configure fills the injected seams and parsed arguments onto cfg.
func configure(cfg *domain.Config, rt app.Runtime, cmd *cli.Command) {
	cfg.Templates = domain.TemplateFiles(cmd.Args().Slice())
	cfg.Assignments = domain.AssignmentTokens(rt.Assignments)
	cfg.Capitalize = domain.Capitalization(rt.Capitalize)
	cfg.TimeFormat = domain.TimeFormat(rt.TimeFormat)
	cfg.Source = rt.Source
	cfg.ReadFile = domain.ReadFileFunc(rt.ReadFile)
	cfg.Exists = domain.ExistsFunc(rt.Exists)
	cfg.Getwd = domain.GetwdFunc(rt.Getwd)
	cfg.Environ = domain.EnvironFunc(rt.Environ)
	cfg.Stdin = cfg.Stdin || domain.StdinEnabled(rt.Piped)
}

// flags binds every renderizer option to its config field.
func flags(cfg *domain.Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:        "settings",
			Aliases:     []string{"S", "s"},
			Usage:       `load settings from the provided YAMLs (default: ".<name>.yaml")`,
			Sources:     cli.EnvVars("RENDERIZER"),
			Destination: (*[]string)(&cfg.Settings),
		},
		&cli.StringFlag{
			Name:        "missing",
			Aliases:     []string{"M", "m"},
			Usage:       "the 'missingkey' template option (default|zero|error|invalid)",
			Value:       "error",
			Sources:     cli.EnvVars("RENDERIZER_MISSINGKEY"),
			Destination: (*string)(&cfg.MissingKey),
		},
		&cli.StringFlag{
			Name:        "environment",
			Aliases:     []string{"env", "E", "e"},
			Usage:       "bind the environment map under this variable name",
			Value:       "env",
			Sources:     cli.EnvVars("RENDERIZER_ENVIRONMENT"),
			Destination: (*string)(&cfg.Environment),
		},
		&cli.BoolFlag{
			Name:        "stdin",
			Aliases:     []string{"c"},
			Usage:       "read the template from stdin",
			Destination: (*bool)(&cfg.Stdin),
		},
		&cli.BoolFlag{
			Name:        "testing",
			Aliases:     []string{"T"},
			Usage:       "make nondeterministic template functions reproducible",
			Sources:     cli.EnvVars("RENDERIZER_TESTING"),
			Destination: (*bool)(&cfg.Testing),
		},
		&cli.BoolFlag{
			Name:        "debugging",
			Aliases:     []string{"debug", "D"},
			Usage:       "enable debug logging",
			Sources:     cli.EnvVars("RENDERIZER_DEBUG"),
			Destination: (*bool)(&cfg.Debugging),
		},
		&cli.BoolFlag{
			Name:        "verbose",
			Aliases:     []string{"V"},
			Usage:       "enable verbose logging",
			Sources:     cli.EnvVars("RENDERIZER_VERBOSE"),
			Destination: (*bool)(&cfg.Verbose),
		},
	}
}
