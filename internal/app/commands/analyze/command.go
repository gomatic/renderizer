// Package analyze is the app-tier definition of the `analyze` subcommand, which
// infers the input data model a template requires.
package analyze

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/gomatic/renderizer/internal/app"
	"github.com/gomatic/renderizer/internal/constants"
	domain "github.com/gomatic/renderizer/internal/domain/analyze"
)

const (
	name     = "analyze"
	usage    = "infer the input data model a template requires"
	argUsage = "[template-file]"
)

// Command returns the analyze subcommand.
func Command(rt app.Runtime) *cli.Command {
	return &cli.Command{
		Name:      name,
		Usage:     usage,
		ArgsUsage: argUsage,
		Action:    action(rt),
	}
}

// action reads the template named by the first argument (or stdin) and writes
// its inferred data-model skeleton.
func action(rt app.Runtime) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		file := firstArg(cmd.Args().Slice())
		if file == "" && !rt.Piped {
			return constants.ErrMissingTemplate
		}
		logger := app.NewLogger(cmd.Root().ErrWriter, false, false)
		result, err := domain.Run(ctx, logger, domain.Config{
			Template: domain.TemplateFile(file),
			Source:   rt.Source,
			ReadFile: domain.ReadFileFunc(rt.ReadFile),
		})
		return app.Write(cmd.Root().Writer, result.Output, err)
	}
}

// firstArg returns the first argument, or empty when there are none.
func firstArg(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return ""
}
