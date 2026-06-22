// Package version is the app-tier definition of the `version` subcommand, which
// prints the version (alongside the built-in --version/-v flag).
package version

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

const (
	name  = "version"
	usage = "show the version"
)

// Command returns the version subcommand, printing "<app> version <version>".
func Command(app, version string) *cli.Command {
	return &cli.Command{
		Name:  name,
		Usage: usage,
		Action: func(_ context.Context, cmd *cli.Command) error {
			_, err := fmt.Fprintf(cmd.Root().Writer, "%s version %s\n", app, version)
			return err
		},
	}
}
