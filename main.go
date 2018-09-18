package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kardianos/osext"
	"github.com/urfave/cli"
)

var (
	version  = "2.0.0"
	commit   = "unknown"
	date     = "20060102T150405"
	selfn, _ = osext.Executable()
	selfz    = filepath.Base(selfn)
	semver   = version + "-" + date + "." + commit[:7]
	appver   = selfz + "/" + semver
)

//
type Settings struct {
	// Capitalization is a positional toggles. The following variable names are capitalized (title-case).
	Capitalize bool
	// Set the Missing Key template option. Defaults to "error".
	MissingKey string
	// Configuration yaml
	ConfigFiles cli.StringSlice
	Defaulted   bool
	Config      map[string]interface{}
	// Add the environment map to the variables.
	Environment string
	//
	OutputExtension string
	//
	TimeFormat string
	//
	Stdin bool
	//
	Debugging bool
	//
	Verbose bool
}

//
var settings = Settings{
	Capitalize:  true,
	MissingKey:  "error",
	TimeFormat:  "20060102T150405",
	Environment: "env",
	Config:      map[string]interface{}{},
	ConfigFiles: []string{},
}

//
func main() {
	app := cli.NewApp()
	app.Name = "renderizer"
	app.Usage = "renderizer [options] [--name=value...] template..."
	app.UsageText = "Template renderer"
	app.Version = appver
	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		{
			Name:  "version",
			Usage: "Shows the app version",
			Action: func(ctx *cli.Context) error {
				fmt.Println(ctx.App.Version)
				return nil
			},
		},
	}

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:   "settings, S, s",
			Usage:  `load the settings from the provided YAMLs (default: ".renderizer.yaml")`,
			Value:  &settings.ConfigFiles,
			EnvVar: "RENDERIZER",
		},
		cli.StringFlag{
			Name:        "missing, M, m",
			Usage:       "the 'missingkey' template option (default|zero|error)",
			Value:       "error",
			EnvVar:      "RENDERIZER_MISSINGKEY",
			Destination: &settings.MissingKey,
		},
		cli.StringFlag{
			Name:   "environment, env, E, e",
			Usage:  "load the environment into the variable name instead of as 'env'",
			Value:  settings.Environment,
			EnvVar: "RENDERIZER_ENVIRONMENT",
		},
		cli.BoolFlag{
			Name:        "stdin, c",
			Usage:       "read from stdin",
			Destination: &settings.Stdin,
		},
		cli.BoolFlag{
			Name:        "debugging, debug, D",
			Usage:       "enable debugging server",
			Destination: &settings.Debugging,
		},
		cli.BoolFlag{
			Name:        "verbose, V",
			Usage:       "enable verbose output",
			Destination: &settings.Verbose,
		},
	}

	app.Action = renderizer
	app.Run(os.Args)
}
