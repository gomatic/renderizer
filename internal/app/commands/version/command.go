// Package version is the app-tier definition of the `version` subcommand, which
// prints the version (alongside the built-in --version/-v flag). It wires the
// version domain (internal/domain/version) to the CLI framework and the shared
// output seam.
package version

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/gomatic/renderizer/internal/app"
	domain "github.com/gomatic/renderizer/internal/domain/version"
)

const (
	name  = "version"
	usage = "show the version"
)

// Command returns the version subcommand, printing "<app> version <build>".
func Command(appName domain.AppName, build domain.Build) *cli.Command {
	return &cli.Command{
		Name:  name,
		Usage: usage,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			logger := app.NewLogger(cmd.Root().ErrWriter, false, false)
			result, err := domain.Run(ctx, &logger, domain.Config{App: appName, Build: build})
			return app.Write(cmd.Root().Writer, result.Output, err)
		},
	}
}
