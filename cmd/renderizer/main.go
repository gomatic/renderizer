package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/imdario/mergo"
	"github.com/kardianos/osext"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"

	"github.com/gomatic/renderizer/v2/pkg/renderizer"
)

var (
	version  = "2.0.9"
	commit   = "default"
	date     = "20060102T150405"
	selfn, _ = osext.Executable()
	selfz    = filepath.Base(selfn)
	semver   = version + "-" + date + "." + commit[:7]
	appver   = selfz + "/" + semver
)

type Settings struct {
	Defaulted bool
	// Configuration yaml
	ConfigFiles *cli.StringSlice
	//
	Options renderizer.Options
}

// RunResult contains the result of running the CLI
type RunResult struct {
	ExitCode int
	Error    error
}

// run executes the renderizer CLI with the given arguments
// It returns the exit code and any error that occurred
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) RunResult {
	// Reset settings for each run
	settings := Settings{
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

	app := cli.NewApp()
	app.Name = "renderizer"
	app.Usage = "renderizer [options] [--name=value...] template..."
	app.UsageText = "Template renderer"
	app.Version = appver
	app.EnableBashCompletion = true
	app.Writer = stdout
	app.ErrWriter = stderr
	// Override ExitErrHandler to return errors instead of calling os.Exit()
	// This allows tests to capture exit codes
	app.ExitErrHandler = func(ctx *cli.Context, err error) {
		// Don't call os.Exit() - just let the error be returned from app.Run()
		// The error will be handled by the caller
	}

	// Set environment variable
	_ = os.Setenv("RENDERIZER_VERSION", appver)

	configs := cli.StringSlice{}

	app.Commands = []*cli.Command{
		{
			Name:  "version",
			Usage: "Shows the app version",
			Action: func(ctx *cli.Context) error {
				_, _ = fmt.Fprintln(stdout, ctx.App.Version)
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
		// Check if stdin is available (not a terminal)
		fi, _ := os.Stdin.Stat()
		settings.Options.Stdin = settings.Options.Stdin || (fi.Mode()&os.ModeCharDevice) == 0

		// If stdin is provided in test, check if it has content
		// Only set Stdin=true if there's actual content, otherwise allow template discovery
		if stdin != nil && stdin != os.Stdin {
			// For seekable readers (like bytes.Reader), check if there's content
			if seeker, ok := stdin.(io.Seeker); ok {
				pos, _ := seeker.Seek(0, io.SeekCurrent)
				size, _ := seeker.Seek(0, io.SeekEnd)
				if _, err := seeker.Seek(pos, io.SeekStart); err != nil {
					log.Printf("warning: failed to restore seek position: %v", err)
				}
				if size > 0 {
					settings.Options.Stdin = true
				}
			}
			// For non-seekable readers, we can't check without consuming
			// so we don't set Stdin=true here
		}

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
				return cli.Exit("missing template name", 1)
			}

			mainName = strings.Split(strings.TrimLeft(filepath.Base(settings.Options.Templates[0]), "."), ".")[0]
		}

		switch settings.Options.MissingKey {
		case "zero", "error", "default", "invalid":
		default:
			_, _ = fmt.Fprintf(stderr, "ERROR: Resetting invalid missingkey: %+v", settings.Options.MissingKey)
			settings.Options.MissingKey = "error"
		}

		if len(configs.Value()) == 0 {
			settings.Defaulted = true
			settings.ConfigFiles = cli.NewStringSlice("." + mainName + ".yaml")
		} else {
			settings.ConfigFiles = &configs
		}

		for _, config := range settings.ConfigFiles.Value() {
			in, err := os.ReadFile(config)
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

	// Process args to separate flags from templates/arguments
	processedArgs := []string{args[0]} // program name
	if len(args) > 1 {
		next := false
		for _, arg := range args[1:] {
			larg := strings.ToLower(arg)
			if next {
				processedArgs = append(processedArgs, arg)
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
					processedArgs = append(processedArgs, arg)
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
					processedArgs = append(processedArgs, arg)
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
		// Save original stdin/stdout for restoration
		oldStdin := os.Stdin
		oldStdout := os.Stdout
		oldStderr := os.Stderr

		// Redirect stdout/stderr if custom writers provided
		// If the writer is already a *os.File (like from a pipe), use it directly
		// Otherwise, create a pipe and copy to the writer
		var stdoutPipeR, stdoutPipeW *os.File
		var stderrPipeR, stderrPipeW *os.File
		var stdoutDone, stderrDone chan struct{}

		if stdout != os.Stdout {
			if f, ok := stdout.(*os.File); ok {
				// Already a file (like from a pipe), use it directly
				os.Stdout = f
				defer func() {
					os.Stdout = oldStdout
				}()
			} else {
				// Need to use a pipe to redirect to the writer
				var err error
				stdoutPipeR, stdoutPipeW, err = os.Pipe()
				if err != nil {
					return err
				}
				stdoutDone = make(chan struct{})
				os.Stdout = stdoutPipeW
				defer func() {
					stdoutPipeW.Close()
					<-stdoutDone // Wait for copy to complete
					stdoutPipeR.Close()
					os.Stdout = oldStdout
				}()
				// Copy from pipe to the actual writer in a goroutine
				go func() {
					if _, err := io.Copy(stdout, stdoutPipeR); err != nil {
						log.Printf("warning: failed to copy stdout: %v", err)
					}
					close(stdoutDone)
				}()
			}
		}
		if stderr != os.Stderr {
			if f, ok := stderr.(*os.File); ok {
				// Already a file (like from a pipe), use it directly
				os.Stderr = f
				defer func() { os.Stderr = oldStderr }()
			} else {
				// Need to use a pipe to redirect to the writer
				var err error
				stderrPipeR, stderrPipeW, err = os.Pipe()
				if err != nil {
					return err
				}
				stderrDone = make(chan struct{})
				os.Stderr = stderrPipeW
				defer func() {
					stderrPipeW.Close()
					<-stderrDone // Wait for copy to complete
					stderrPipeR.Close()
					os.Stderr = oldStderr
				}()
				// Copy from pipe to the actual writer in a goroutine
				go func() {
					if _, err := io.Copy(stderr, stderrPipeR); err != nil {
						log.Printf("warning: failed to copy stderr: %v", err)
					}
					close(stderrDone)
				}()
			}
		}

		// Handle custom stdin for testing
		// renderizer reads from os.Stdin, so we need to replace it temporarily
		if stdin != nil && stdin != os.Stdin {
			// Create a temporary file and copy stdin content to it
			tmpFile, err := os.CreateTemp("", "renderizer-stdin-*")
			if err != nil {
				return err
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			// Copy stdin to temp file
			if _, err := io.Copy(tmpFile, stdin); err != nil {
				return err
			}
			if _, err := tmpFile.Seek(0, 0); err != nil {
				return fmt.Errorf("failed to seek temp file: %w", err)
			}

			// Replace os.Stdin with our temp file
			os.Stdin = tmpFile
			defer func() { os.Stdin = oldStdin }()
		}

		// Call renderizer - it now returns exit code instead of calling os.Exit()
		// Note: renderizer uses fmt.Println which writes to os.Stdout
		// So we need os.Stdout redirected above
		exitCode, err := renderizer.Render(settings.Options)
		if err != nil {
			return err
		}
		if exitCode != 0 {
			// Return a cli.Exit error with the exit code
			// This will be caught by app.Run() and returned as an error
			// cli.Exit implements cli.ExitCoder interface
			return cli.Exit("", exitCode)
		}
		return nil
	}

	err := app.Run(processedArgs)
	if err != nil {
		if exitErr, ok := err.(cli.ExitCoder); ok {
			return RunResult{ExitCode: exitErr.ExitCode(), Error: err}
		}
		return RunResult{ExitCode: 1, Error: err}
	}

	// Note: renderizer.Render() calls os.Exit(), so if we get here in production,
	// it means os.Exit() was called. In tests, we need to handle this differently.
	return RunResult{ExitCode: 0, Error: nil}
}

func main() {
	result := run(os.Args, os.Stdin, os.Stdout, os.Stderr)
	if result.Error != nil {
		if exitErr, ok := result.Error.(cli.ExitCoder); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
	os.Exit(result.ExitCode)
}
