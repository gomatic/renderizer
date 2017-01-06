package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v2"
)

//
func usage(out io.Writer) {
	fmt.Fprintf(out, `usage: renderizer [options] [--name=value...] template...
options:

  -S=settings.yaml            Load the settings from the provided yaml.
                              Initializes from RENDERIZER environment.
                              If not set, tries .renderizer.yaml in the current directory.
  -E[=name]                   Load the environment into the variables map as _env
                              or name if provided
  -M[=(default|zero|error)]   The missingkey template option. Default: error
                              Initializes from RENDERIZER_MISSINGKEY environment.
  -V                          Enable verbose output.
  -D                          Enable debug output.

`)
}

//
type Settings struct {
	// Capitalization is a positional toggles. The following variable names are capitalized (title-case).
	Capitalize bool
	// Set the Missing Key template option. Defaults to "error".
	MissingKey string
	// Configuration yaml
	Config string
	// Add the environment map to the variables.
	Environment string
	//
	TimeFormat string
	//
	Debugging bool
	//
	Verbose bool
	//
	CommandLine string
}

//
var settings = Settings{
	Capitalize: true,
	MissingKey: "error",
	TimeFormat: "20060102T150405",
}

//
func main() {

	if len(os.Args) == 1 {
		usage(os.Stdout)
		return
	}

	vars := map[string]interface{}{}
	load := map[string]interface{}{}
	args := []string{}

	settings.CommandLine = commandLine()

	// Initialize some settings from the environment.

	if m, exists := os.LookupEnv("RENDERIZER_MISSINGKEY"); exists {
		settings.MissingKey = m
	}
	if m, exists := os.LookupEnv("RENDERIZER"); exists {
		settings.Config = m
	}

	// Process settings flags.

	params := []string{}
	for _, arg := range os.Args[1:] {
		la := len(arg)
		if la == 0 {
			continue
		}
		if arg[0] != '-' || (la > 2 && arg[:2] == "--") {
			if strings.ToLower(arg) == "--help" {
				usage(os.Stdout)
				return
			}
			params = append(params, arg)
			continue
		}
		switch arg[1:][0] {
		case 'h', 'H':
			usage(os.Stdout)
			return
		case 's', 'S':
			nv := strings.SplitN(arg, "=", 2)
			if len(nv) != 1 {
				settings.Config = nv[1]
			}
		case 'e', 'E':
			nv := strings.SplitN(arg, "=", 2)
			if len(nv) != 1 {
				settings.Environment = nv[1]
			}
		case 'v', 'V':
			settings.Verbose = true
		case 'd', 'D':
			settings.Debugging = true
		case 'm', 'M':
			nv := strings.SplitN(arg, "=", 2)
			if len(nv) == 1 {
				settings.MissingKey = "error"
			} else {
				settings.MissingKey = nv[1]
			}
		}
	}

	switch settings.MissingKey {
	case "zero", "error", "default", "invalid":
	default:
		fmt.Fprintf(os.Stderr, "ERROR: Resetting invalid missingkey: %+v", settings.MissingKey)
		settings.MissingKey = "error"
	}

	// Load the settings.

	forced := false
	if settings.Config == "" {
		forced = true
		settings.Config = ".renderizer.yaml"
	}

	if settings.Config != "" {
		in, err := ioutil.ReadFile(settings.Config)
		if err != nil {
			if !forced {
				log.Println(err)
			}
		} else {
			yaml.Unmarshal(in, &load)
			if settings.Verbose && forced {
				log.Printf("used config: %+v", settings.Config)
			}
		}
		if settings.Debugging {
			log.Printf("-settings:%#v", settings)
			log.Printf("loaded: %#v", load)
		} else if settings.Verbose {
			log.Printf("-settings:%+v", settings)
			log.Printf("loaded: %+v", load)
		}
	}

	// Iterate the remaining arguments for variable overrides and file names.

	i := 0
	for a, arg := range params {
		if len(arg) > 1 && arg[0] == '-' {
			switch arg[1:][0] {
			case 'c', 'C':
				settings.Capitalize = !settings.Capitalize
				if settings.Verbose {
					log.Printf("-capitalize:%v", settings.Capitalize)
				}
				continue
			}

			arg = strings.TrimLeft(arg, "-")
			nv := strings.SplitN(arg, "=", 2)
			n := nv[0]
			i += 1

			if settings.Capitalize {
				n = fmt.Sprintf("%s%s", strings.ToTitle(n[:1]), strings.ToLower(n[1:]))
			}

			var v interface{}
			if len(nv) == 1 {
				v = true
			} else {
				v = typer(nv[1])
			}

			if settings.Debugging {
				log.Printf("arg:%d var:%d name:%s value:%+v", a, i, n, v)
			}

			// This is probably a overkill but I thought it a good idea to keep the types strong as much as possible.
			if vs, exists := vars[n]; exists {
				switch x := vs.(type) {
				case string:
					switch y := v.(type) {
					case string:
						vars[n] = []string{x, y}
					default:
						vars[n] = []interface{}{x, y}
					}
				case int64:
					switch y := v.(type) {
					case int64:
						vars[n] = []int64{x, y}
					default:
						vars[n] = []interface{}{x, y}
					}
				case float64:
					switch y := v.(type) {
					case float64:
						vars[n] = []float64{x, y}
					default:
						vars[n] = []interface{}{x, y}
					}
				case bool:
					switch y := v.(type) {
					case bool:
						vars[n] = []bool{x, y}
					default:
						vars[n] = []interface{}{x, y}
					}
				case *time.Time:
					switch y := v.(type) {
					case *time.Time:
						vars[n] = []*time.Time{x, y}
					default:
						vars[n] = []interface{}{x, y}
					}
				case []string:
					switch y := v.(type) {
					case string:
						vars[n] = append(x, y)
					default:
						vars[n] = append(toi(x), y)
					}
				case []int64:
					switch y := v.(type) {
					case int64:
						vars[n] = append(x, y)
					default:
						vars[n] = append(toi(x), y)
					}
				case []float64:
					switch y := v.(type) {
					case float64:
						vars[n] = append(x, y)
					default:
						vars[n] = append(toi(x), y)
					}
				case []bool:
					switch y := v.(type) {
					case bool:
						vars[n] = append(x, y)
					default:
						vars[n] = append(toi(x), y)
					}
				case []*time.Time:
					switch y := v.(type) {
					case *time.Time:
						vars[n] = append(x, y)
					default:
						vars[n] = append(toi(x), y)
					}
				case []interface{}:
					vars[n] = append(x, v)
				}
			} else {
				vars[n] = v
			}
		} else {
			args = append(args, arg)
		}
	}

	if len(args) == 0 {
		usage(os.Stdout)
		os.Exit(1)
	}

	if len(load) == 0 && len(vars) == 0 {
		usage(os.Stderr)
		fmt.Fprintln(os.Stderr, "WARNING: No variables provided.")
	}

	// Copy the loaded keys in the vars unless provided on the command line.

	for n, v := range load {
		if _, exists := vars[n]; !exists {
			vars[n] = v
		}
	}

	// Dump the settings

	if settings.Debugging {
		log.Printf("vars: %#v", vars)
	} else if settings.Verbose {
		log.Printf("vars: %+v", vars)
	}

	// Execute each template

	for _, arg := range args {
		file, err := ioutil.ReadFile(arg)
		if err != nil {
			log.Println(err)
			continue
		}

		tmpl, err := template.New(arg).
			Option(fmt.Sprintf("missingkey=%s", settings.MissingKey)).
			Funcs(funcs).
			Parse(string(file))

		if err != nil {
			log.Print(err)
		}
		var sqlBuffer bytes.Buffer
		err = tmpl.Execute(&sqlBuffer, vars)
		if err != nil {
			log.Print(err)
		}
		fmt.Println(string(sqlBuffer.Bytes()))
	}
}

// Reproduce a command line string that reflects a usable command line.
func commandLine() string {

	quoter := func(e string) string {
		if !strings.Contains(e, " ") {
			return e
		}
		p := strings.SplitN(e, "=", 2)
		if strings.Contains(p[0], " ") {
			p[0] = `"` + strings.Replace(p[0], `"`, `\"`, -1) + `"`
		}
		if len(p) == 1 {
			return p[0]
		}
		return p[0] + `="` + strings.Replace(p[1], `"`, `\"`, -1) + `"`
	}
	each := func(s []string) (o []string) {
		o = make([]string, len(s))
		for i, t := range s {
			o[i] = quoter(t)
		}
		return
	}
	return filepath.Base(os.Args[0]) + " " + strings.Join(each(os.Args[1:]), " ")
}