package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/gomatic/funcmap"
	"gopkg.in/yaml.v2"
)

//
func usage(out io.Writer) {
	fmt.Fprintln(out, `usage: renderizer [options] [--name=value...] template...
options:

  -S=settings.yaml            Load the settings from the provided yaml.
                              Initializes from RENDERIZER environment.
                              If not set, tries .renderizer.yaml in the current directory.
  -E=name                     Load the environment into the variable name instead of as env.
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
	Config []string
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
}

//
func main() {

	vars := map[string]interface{}{}
	load := map[string]interface{}{}
	args := []string{}
	settings.Config = []string{}

	// Initialize some settings from the environment.

	if m, exists := os.LookupEnv("RENDERIZER_MISSINGKEY"); exists {
		settings.MissingKey = m
	}
	if m, exists := os.LookupEnv("RENDERIZER"); exists {
		settings.Config = append(settings.Config, m)
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
				settings.Config = append(settings.Config, nv[1])
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
	if len(settings.Config) == 0 {
		forced = true
		settings.Config = []string{".renderizer.yaml"}
	}

	for _, config := range settings.Config {
		in, err := ioutil.ReadFile(config)
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
				case int:
					switch y := v.(type) {
					case int:
						vars[n] = []int64{int64(x), int64(y)}
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

	files := args
	if len(args) == 0 {
		if len(args) < 2 {
			stat, _ := os.Stdin.Stat()
			isTTY := (stat.Mode() & os.ModeCharDevice) != 0
			if isTTY {
				log.Println("source: stdin")
			}
			files = []string{""}
		}
	}

	// Copy the loaded keys in the vars unless provided on the command line.

	for n, v := range load {
		if _, exists := vars[n]; !exists {
			switch x := v.(type) {
			case int8:
				vars[n] = int64(x)
			case int16:
				vars[n] = int64(x)
			case int32:
				vars[n] = int64(x)
			case int:
				vars[n] = int64(x)
			default:
				vars[n] = v
			}
		}
	}

	if settings.Environment != "" {
		v := make(map[string]string)
		for _, item := range os.Environ() {
			splits := strings.Split(item, "=")
			v[splits[0]] = strings.Join(splits[1:], "=")
		}
		vars[settings.Environment] = v
	}

	// Dump the settings

	if settings.Debugging {
		log.Printf("vars: %#v", vars)
	} else if settings.Verbose {
		log.Printf("vars: %+v", vars)
	}

	// Execute each template

	for _, arg := range files {
		var file []byte
		if arg == "" {
			f, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				log.Println(err)
				continue
			}
			file = f
		} else {
			f, err := ioutil.ReadFile(arg)
			if err != nil {
				log.Println(err)
				continue
			}
			file = f
		}

		tmpl, err := template.New(arg).
			Option(fmt.Sprintf("missingkey=%s", settings.MissingKey)).
			Funcs(funcmap.Map).
			Parse(string(file))

		if err != nil {
			log.Print(err)
		}
		var sqlBuffer bytes.Buffer
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("PANIC: %+v", r)
				}
			}()
			err = tmpl.Execute(&sqlBuffer, vars)
			if err != nil {
				log.Print(err)
			}
		}()
		fmt.Println(string(sqlBuffer.Bytes()))
	}
}
