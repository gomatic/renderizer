package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gomatic/renderizer/pkg/renderizer"
	"github.com/imdario/mergo"
	"github.com/kardianos/osext"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

var (
	version    = "0.0.0"
	commitHash = "default"
	commitTime = "20060102T15"
	tag        = "v" + version
	buildTime  = "20060102T150405"
	fullSemver = semver(version, commitHash, commitTime, buildTime, tag)
	selfn, _   = osext.Executable()
	selfz      = filepath.Base(selfn)
	appver     = selfz + "/" + fullSemver
)

//
func semver(version, commitHash, commitTime, buildTime, tag string) string {
	if len(commitHash) < 8 {
		commitHash = "default"
	}
	if commitTime == "" {
		commitTime = buildTime
	}
	return version + "-" + commitTime + "." + commitHash[:7]
}

//
type Settings struct {
	renderizer.Options

	Defaulted bool
	// Configuration yaml
	ConfigFiles *cli.StringSlice
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
	settings.ConfigFiles = &cli.StringSlice{}
	app.Commands = Commands()
	app.Flags = Flags()
	app.Before = before
	app.Action = action

	_ = os.Setenv("RENDERIZER_VERSION", appver)

	_ = app.Run(customCommandLine())
}

//
func Commands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "version",
			Usage: "Shows the app version",
			Action: func(ctx *cli.Context) error {
				fmt.Println(ctx.App.Version)
				return nil
			},
		},
	}
}

func Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "settings",
			Aliases: []string{"S", "s"},
			Usage:   `load the settings from the provided YAMLs (default: ".renderizer.yaml")`,
			Value:   settings.ConfigFiles,
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
			EnvVars:     []string{"RENDERIZER_VERBOSE"},
			Destination: &settings.Options.Verbose,
		},
	}
}

//
func before(ctx *cli.Context) error {
	if settings.Verbose {
		log.Printf("before/start settings: %#v", settings)
	}
	defer func() {
		if settings.Verbose {
			log.Printf("before/end settings: %#v", settings)
		}
	}()

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
		// Try to default the template name
		if name := findTemplate(bases); name != "" {
			if settings.Options.Verbose {
				log.Printf("using template: %+v", name)
			}
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

	if len(settings.ConfigFiles.Value()) == 0 {
		settings.Defaulted = true
		settings.ConfigFiles = cli.NewStringSlice("." + mainName + ".yaml")
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

//
func action(_ *cli.Context) error {
	return renderizer.New(
		renderizer.WithArguments(settings.Arguments),
		renderizer.WithCapitalization(settings.Capitalize),
		renderizer.WithConfig(settings.Config),
		renderizer.WithDebugging(settings.Debugging),
		renderizer.WithEnvironment(settings.Environment),
		renderizer.WithMissingKeyString(settings.MissingKey),
		renderizer.WithStdin(settings.Stdin),
		renderizer.WithTemplates(settings.Templates),
		renderizer.WithTesting(settings.Testing),
		renderizer.WithTimeFormat(settings.TimeFormat),
		renderizer.WithVerbose(settings.Verbose),
	).Render()
}

//
func customCommandLine() []string {
	// Remove args that are not processed by urfave/cli
	args := []string{os.Args[0]}
	if len(os.Args) <= 1 {
		return args
	}
	prior := ""
	for _, arg := range os.Args[1:] {
		if prior != "" {
			args = append(args, prior+"="+arg)
			prior = ""
			continue
		}

		lowerArg := strings.ToLower(arg)
		isFlag := strings.HasPrefix(lowerArg, "-")

		flag := arg

		if strings.HasPrefix(lowerArg, "--") {
			parts := strings.SplitN(lowerArg, "=", 2)
			flag = parts[0]
		}

		switch flag {
		case "-S", "--settings", "--missing", "--environment", "--env":
			// These flags require values.
			if !strings.Contains(lowerArg, "=") {
				prior = arg
			}
			continue
		case "--debug", "--verbose", "--testing", "--version", "--stdin", "--help":
			args = append(args, arg)

		case "-C":
			// Passes through to renderizer
		}

		if isFlag {
			settings.Options.Arguments = append(settings.Options.Arguments, arg)
			continue
		}

		settings.Options.Templates = append(settings.Options.Templates, arg)
	}

	return args
}

// Will search for a file names like:
//   ^[.]?{names}([.](yaml|json|html|txt|xml))?([.]((go[.]?)?tm?pl|rzt)
func findTemplate(names []string) string {
	bases := strings.Replace(strings.Join(names, "|"), ".", "[.]", -1)
	types := strings.Join([]string{"yaml", "json", "html", "txt", "xml", ""}, "|")
	reName := regexp.MustCompile(fmt.Sprintf("^([.]?%s)([.](%s))?([.](tm?pl|rzt))?$", bases, types))
	fns, err := ioutil.ReadDir(".")
	if err != nil {
		log.Print(err)
	} else {
		for _, fn := range fns {
			if fn.IsDir() {
				continue
			}
			if reName.MatchString(fn.Name()) {
				return fn.Name()
			}
		}
	}
	return ""
}
