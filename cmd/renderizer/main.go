package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomatic/renderizer/pkg/renderizer"
	"github.com/imdario/mergo"
	"github.com/kardianos/osext"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

var (
	version  = "2.0.5"
	commit   = "unknown"
	date     = "20060102T150405"
	selfn, _ = osext.Executable()
	selfz    = filepath.Base(selfn)
	semver   = version + "-" + date + "." + commit[:7]
	appver   = selfz + "/" + semver
)

//
type Settings struct {
	Defaulted bool
	// Configuration yaml
	ConfigFiles *cli.StringSlice
	//
	Options renderizer.Options
}

//
var settings = Settings{
	ConfigFiles: &cli.StringSlice{},
	Options: renderizer.Options{
		Config:      map[string]interface{}{},
		Capitalize:  true,
		MissingKey:  "error",
		TimeFormat:  "20060102T150405",
		Environment: "env",
		Arguments:   []string{},
		Templates:   []string{},
	},
}

//
func main() {
	app := cli.NewApp()
	app.Name = "renderizer"
	app.Usage = "renderizer [options] [--name=value...] template..."
	app.UsageText = "Template renderer"
	app.Version = appver
	app.EnableBashCompletion = true

	_ = os.Setenv("RENDERIZER_VERSION", appver)

	configs := cli.StringSlice{}

	app.Commands = []*cli.Command{
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
		&cli.StringSliceFlag{
			Name:    "settings",
			Aliases: []string{"S", "s"},
			Usage:   `load the settings from the provided YAMLs (default: ".renderizer.yaml")`,
			Value:   &configs,
			EnvVars: []string{"RENDERIZER"},
		},
		&cli.StringFlag{
			Name:        "missing",
			Aliases:     []string{"M", "m"},
			Usage:       "the 'missingkey' template option (default|zero|error)",
			Value:       "error",
			EnvVars:     []string{"RENDERIZER_MISSINGKEY"},
			Destination: &settings.Options.MissingKey,
		},
		&cli.StringFlag{
			Name:        "environment",
			Aliases:     []string{"env", "E", "e"},
			Usage:       fmt.Sprintf("load the environment into the variable name instead of as '%s'", settings.Options.Environment),
			Value:       settings.Options.Environment,
			EnvVars:     []string{"RENDERIZER_ENVIRONMENT"},
			Destination: &settings.Options.Environment,
		},
		&cli.BoolFlag{
			Name:        "stdin",
			Aliases:     []string{"c"},
			Usage:       "read from stdin",
			Destination: &settings.Options.Stdin,
		},
		&cli.BoolFlag{
			Name:        "testing",
			Aliases:     []string{"T"},
			Usage:       "configure runtime to provide consistent output",
			EnvVars:     []string{"RENDERIZER_TESTING"},
			Destination: &settings.Options.Testing,
		},
		&cli.BoolFlag{
			Name:        "debugging",
			Aliases:     []string{"debug", "D"},
			Usage:       "enable debugging server",
			EnvVars:     []string{"RENDERIZER_DEBUG"},
			Destination: &settings.Options.Debugging,
		},
		&cli.BoolFlag{
			Name:        "verbose",
			Aliases:     []string{"V"},
			Usage:       "enable verbose output",
			EnvVars:     []string{"RENDERIZER_VEBOSE"},
			Destination: &settings.Options.Verbose,
		},
	}

	app.Before = func(ctx *cli.Context) error {

		fi, _ := os.Stdin.Stat()

		settings.Options.Stdin = settings.Options.Stdin || (fi.Mode()&os.ModeCharDevice) == 0

		settings.Options.Arguments = append(settings.Options.Arguments, ctx.Args().Slice()...)

		mainName := ""
		folderName, err := os.Getwd()
		bases := []string{"renderizer"}
		if err != nil {
			log.Println(err)
			mainName = "renderizer"
		} else {
			mainName = filepath.Base(folderName)
			bases = []string{mainName, "renderizer"}
		}

		if len(settings.Options.Templates) == 0 && !settings.Options.Stdin {
			// Try default the template name

			name := func() string {
				for _, base := range bases {
					for _, ext := range []string{".tmpl", ""} {
						for _, try := range []string{"yaml", "json", "html", "txt", "xml", ""} {
							name := fmt.Sprintf("%s.%s%s", base, try, ext)
							if _, err := os.Stat(name); err == nil {
								if settings.Options.Verbose {
									log.Printf("using template: %+v", name)
								}
								return name
							}
						}
					}
				}
				return ""
			}()
			if name != "" {
				settings.Options.Templates = append(settings.Options.Templates, name)
			}

			if len(settings.Options.Templates) == 0 {
				return cli.NewExitError("missing template name", 1)
			}

			mainName = strings.Split(strings.TrimLeft(filepath.Base(settings.Options.Templates[0]), "."), ".")[0]
		}

		switch settings.Options.MissingKey {
		case "zero", "error", "default", "invalid":
		default:
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: Resetting invalid missingkey: %+v", settings.Options.MissingKey)
			settings.Options.MissingKey = "error"
		}

		if len(configs.Value()) == 0 {
			settings.Defaulted = true
			settings.ConfigFiles = cli.NewStringSlice("." + mainName + ".yaml")
		} else {
			settings.ConfigFiles = &configs
		}

		for _, config := range settings.ConfigFiles.Value() {
			in, err := ioutil.ReadFile(config)
			if err != nil {
				if !settings.Defaulted {
					return err
				}
			} else {
				loaded := map[string]interface{}{}
				err := yaml.Unmarshal(in, &loaded)
				if err != nil {
					return err
				}
				if settings.Options.Debugging || settings.Options.Verbose {
					log.Printf("using settings: %+v", settings.ConfigFiles)
				}
				loaded = settings.Options.Retyper(loaded)
				if settings.Options.Debugging {
					log.Printf("loaded: %s = %#v", config, loaded)
				} else if settings.Options.Verbose {
					log.Printf("loaded: %s = %+v", config, loaded)
				}
				_ = mergo.Merge(&settings.Options.Config, loaded)
			}
		}

		if settings.Options.Debugging {
			log.Printf("--settings:%#v", settings)
		} else if settings.Options.Verbose {
			log.Printf("--settings:%+v", settings)
		}

		return nil
	}

	// Remove args that are not processed by urfave/cli
	args := []string{os.Args[0]}
	if len(os.Args) > 1 {
		next := false
		for _, arg := range os.Args[1:] {
			larg := strings.ToLower(arg)
			if next {
				args = append(args, arg)
				next = false
				continue
			}
			// TODO convert all '--name value' parameters to --name=value
			if strings.HasPrefix(larg, "--") {
				flag := larg
				parts := strings.SplitN(larg, "=", 2)
				if len(parts) == 2 {
					flag = parts[0]
				}
				switch flag[2:] {
				case "settings", "missing", "environment", "env":
					// If the flag requires a parameter but it is not specified with an =, grab the next argument too.
					if !strings.Contains(larg, "=") {
						next = true
					}
					fallthrough
				case "debug", "verbose", "testing", "version", "stdin", "help":
					args = append(args, arg)
					continue
				}
			} else if strings.HasPrefix(larg, "-") {
				switch arg[1:] {
				case "C":
				case "S":
					if !strings.Contains(arg, "=") {
						next = true
					}
					fallthrough
				default:
					args = append(args, arg)
					continue
				}
			} else {
				settings.Options.Templates = append(settings.Options.Templates, arg)
				continue
			}

			settings.Options.Arguments = append(settings.Options.Arguments, arg)
		}
	}

	app.Action = func(_ *cli.Context) error {
		return renderizer.Render(settings.Options)
	}

	_ = app.Run(args)
}
