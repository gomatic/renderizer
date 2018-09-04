package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kardianos/osext"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

var (
	version  = "2.0.0"
	commit   = "unknown"
	date     = "none"
	selfn, _ = osext.Executable()
	selfd, _ = osext.ExecutableFolder()
	selfz    = filepath.Base(selfn)
	semver   = version + "-" + commit[:7] + "." + date
	appver   = selfz + "/" + semver
	started  = time.Now().Format("20060102T150405Z0700")
)

//
type Settings struct {
	// Capitalization is a positional toggles. The following variable names are capitalized (title-case).
	Capitalize bool
	// Set the Missing Key template option. Defaults to "error".
	MissingKey string
	// Configuration yaml
	ConfigFiles []string
	Defaulted   bool
	Config      map[string]interface{}
	//
	Arguments []string
	// Add the environment map to the variables.
	Environment string
	//
	TimeFormat string
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
	Arguments:   []string{},
}

//
func main() {
	app := cli.NewApp()
	app.Name = "renderizer"
	app.Usage = "renderizer [options] [--name=value...] template..."
	app.UsageText = "Template renderer"
	app.Version = appver
	app.EnableBashCompletion = true
	configs := cli.StringSlice([]string{".renderizer.yaml"})

	// Remove args that are not processed by urfave/cli
	var args []string
	for _, arg := range os.Args {
		larg := strings.ToLower(arg)
		switch {
		case larg == "-c":
			fallthrough
		case strings.HasPrefix(arg, "--") && strings.Contains(arg, "="):
			settings.Arguments = append(settings.Arguments, arg)

		case strings.HasPrefix(arg, "--"):
			fallthrough
		default:
			args = append(args, arg)
		}
	}

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
			Name:   "settings, S",
			Usage:  "load the settings from the provided YAMLs",
			Value:  &configs,
			EnvVar: "RENDERIZER",
		},
		cli.StringFlag{
			Name:  "environment, env, E",
			Usage: "load the environment into the variable name instead of as 'env'",
		},
		cli.StringFlag{
			Name:        "missing, M",
			Usage:       "the 'missingkey' template option (default|zero|error)",
			Value:       "error",
			EnvVar:      "RENDERIZER_MISSINGKEY",
			Destination: &settings.MissingKey,
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

	app.Before = func(ctx *cli.Context) error {
		settings.Arguments = append(settings.Arguments, ctx.Args()...)

		switch settings.MissingKey {
		case "zero", "error", "default", "invalid":
		default:
			fmt.Fprintf(os.Stderr, "ERROR: Resetting invalid missingkey: %+v", settings.MissingKey)
			settings.MissingKey = "error"
		}

		if len(settings.ConfigFiles) == 0 {
			settings.Defaulted = true
			settings.ConfigFiles = []string{".renderizer.yaml"}
		}

		for _, config := range settings.ConfigFiles {
			in, err := ioutil.ReadFile(config)
			if err != nil {
				if !settings.Defaulted {
					return err
				}
			} else {
				yaml.Unmarshal(in, &settings.Config)
				if settings.Verbose && settings.Defaulted {
					log.Printf("used config: %+v", settings.ConfigFiles)
				}
			}
			if settings.Debugging {
				log.Printf("-settings:%#v", settings)
				log.Printf("loaded: %#v", settings.Config)
			} else if settings.Verbose {
				log.Printf("-settings:%+v", settings)
				log.Printf("loaded: %+v", settings.Config)
			}
		}

		return nil
	}

	app.Action = renderizer
	app.Run(args)
}
