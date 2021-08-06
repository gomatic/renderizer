package renderizer

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gomatic/clock"
	"github.com/gomatic/funcmap"
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
)

//
type Options struct {
	// Capitalization is a positional toggles. The following variable names are capitalized (title-case).
	Capitalize bool
	// Set the Missing Key template option. Defaults to "error".
	MissingKey string
	//
	Config map[string]interface{}
	//
	Arguments []string
	//
	Templates []string
	// Add the environment map to the variables.
	Environment string
	//
	TimeFormat string
	//
	Stdin bool
	//
	Debugging bool
	//
	Verbose bool
	//
	Testing bool
	//
	Output string
}

//
type Renderizer interface {
	Render() error
}

//
func New(settings Options) Renderizer {
	return &settings
}

//
func Render(settings Options) error {
	return settings.Render()
}

//
func (settings *Options) Render() error {

	if settings.Testing {
		rand.Seed(0)
		funcmap.UseClock(clock.Format)
		funcmap.Map["command_line"] = func() string { return "testing" }
	}

	globalContext := map[string]interface{}{}

	// Iterate the remaining arguments for variable overrides and file names.

	for a, arg := range settings.Arguments {
		if len(arg) == 0 {
			continue
		} else if arg[0] != '-' {
			continue
		}

		switch arg[1:][0] {
		case 'c', 'C':
			settings.Capitalize = !settings.Capitalize
			if settings.Verbose {
				log.Printf("-capitalize:%v", settings.Capitalize)
			}
			continue
		}

		currentContext := map[string]interface{}{}

		nameValuePair := strings.SplitN(strings.TrimLeft(arg, "-"), "=", 2)
		currentName := nameValuePair[0]

		// Iterate dotted notation and construct a map from it.
		var (
			local = currentContext       // Iterative map-reference into globalContext
			leaf  map[string]interface{} // Keep a reference to the leaf node
			last  string                 // Keep the leaf node's map-key
		)
		for _, name := range strings.Split(currentName, ".") {
			if settings.Capitalize {
				name = fmt.Sprintf("%s%s", strings.ToTitle(name[:1]), strings.ToLower(name[1:]))
			}

			local[name] = map[string]interface{}{}
			leaf, last, local = local, name, local[name].(map[string]interface{})
		}

		var currentValue interface{}
		if len(nameValuePair) == 1 { // i.e. a boolean
			currentValue = []interface{}{true}
		} else {
			currentValue = []interface{}{settings.typer(nameValuePair[1])}
		}
		leaf[last] = currentValue

		if settings.Debugging {
			log.Printf("index:%d name:%s value:%+v", a, currentName, currentValue)
		}

		if settings.Debugging {
			log.Printf("currentContext: %[1]T %#[1]v", currentContext)
		}
		mergo.Merge(&globalContext, currentContext, mergo.WithAppendSlice)
		if settings.Debugging {
			log.Printf("globalContext: %[1]T %#[1]v", globalContext)
		}
	}

	globalContext = settings.Retyper(globalContext, retypeSingleElementSlice)

	// If there's no files, read from stdin.
	files := settings.Templates
	if len(files) == 0 {
		if settings.Stdin && settings.Verbose {
			log.Println("source: stdin")
		}
		files = []string{""}
	}

	// Copy any loaded keys into the globalContext unless they already exist, i.e. they were provided on the command line.
	mergo.Merge(&globalContext, settings.Config)

	if settings.Environment != "" || len(files) == 0 {
		v := make(map[string]string)
		for _, item := range os.Environ() {
			splits := strings.Split(item, "=")
			v[splits[0]] = strings.Join(splits[1:], "=")
		}
		globalContext[settings.Environment] = v
	}

	// Dump the settings

	if settings.Debugging {
		log.Printf("globalContext: %#v", globalContext)
	} else if settings.Verbose {
		if o, err := yaml.Marshal(globalContext); err != nil {
			log.Printf("globalContext: %+v", globalContext)
		} else {
			log.Printf("globalContext:\n%s", o)
		}
	}

	// Execute each template

	status := 0
	for _, file := range files {
		status |= func() (status int) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("PANIC: %+v", r)
					status = 15
				}
			}()

			var err error
			var data []byte
			var r io.ReadCloser
			if file == "" {
				r = os.Stdin
			} else {
				r, err = os.Open(file)
				if err != nil {
					log.Println(err)
					return 1
				}
				defer r.Close()
			}
			f, err := ioutil.ReadAll(r)
			if err != nil {
				log.Println(err)
				return 2
			}
			data = f

			tmpl, err := template.New(file).
				Option(fmt.Sprintf("missingkey=%s", settings.MissingKey)).
				Funcs(funcmap.Map).
				Parse(string(data))
			if err != nil {
				log.Print(err)
				return 4
			}

			var b bytes.Buffer
			err = tmpl.Execute(&b, globalContext)
			if err != nil {
				log.Print(err)
				return 8
			}

			data = b.Bytes()
			fh := os.Stdout

			if settings.Output != "" {
				if fh, err = os.Create(settings.Output); err != nil {
					log.Print(err)
					return 8
				}
			}

			if _, err = fh.Write(data); err != nil {
				log.Print(err)
				return 8
			}

			return 0
		}()
	}

	os.Exit(status)
	return nil
}

// Transform a string into the best type.
func (settings Options) typer(d string) (result interface{}) {
	v := string(d)
	if parsedValue, err := strconv.ParseInt(v, 10, 64); err == nil {
		result = parsedValue
	} else if parsedValue, err := strconv.ParseFloat(v, 64); err == nil {
		result = parsedValue
	} else if parsedValue, err := strconv.ParseBool(v); err == nil {
		result = parsedValue
	} else if parsedValue, err := time.Parse(settings.TimeFormat, v); err == nil { // TODO parameterize format
		result = parsedValue
	} else {
		result = v
	}
	return
}

type retyperOptions func(*retyperConfig)
type retyperConfig struct {
	retypeSingleElementSlice bool
}

func retypeSingleElementSlice(config *retyperConfig) { config.retypeSingleElementSlice = true }

// Coerce m slices to type-specific slices where applicable
func (settings Options) retyping(source map[string]interface{}, config retyperConfig) map[string]interface{} {
	for k, v := range source {
		switch vt := v.(type) {
		case map[interface{}]interface{}:
			f := make(map[string]interface{}, len(vt))
			for e, s := range vt {
				f[fmt.Sprintf("%v", e)] = s
			}
			settings.retyping(f, config)
			source[k] = f
		case map[string]string:
			f := make(map[string]interface{}, len(vt))
			for e, s := range vt {
				f[e] = s
			}
			settings.retyping(f, config)
			source[k] = f
		case int:
			source[k] = int64(vt)
		case map[string]interface{}:
			settings.retyping(vt, config)
		case []interface{}:
			if config.retypeSingleElementSlice && len(vt) == 1 {
				source[k] = vt[0]
				continue
			}
			kind := reflect.Invalid
			valid := func() bool {
				for i, value := range vt {
					t := reflect.TypeOf(value)
					if i == 0 {
						kind = t.Kind()
						continue
					}
					// If the types are different or invalid, do not create a type-specific slice.
					if kind == reflect.Invalid || t.Kind() != kind {
						return false
					}
				}
				return true
			}()
			if !valid {
				source[k] = vt
				continue
			}
			// Create the type-specific list. This set of types must correspond to the types returned from `typer()`.
			switch kind {
			case reflect.Bool:
				rt := make([]bool, len(vt))
				for i, value := range vt {
					rt[i] = value.(bool)
				}
				source[k] = rt
			case reflect.Int64:
				rt := make([]int64, len(vt))
				for i, value := range vt {
					rt[i] = value.(int64)
				}
				source[k] = rt
			case reflect.Float64:
				rt := make([]float64, len(vt))
				for i, value := range vt {
					rt[i] = value.(float64)
				}
				source[k] = rt
			case reflect.String:
				rt := make([]string, len(vt))
				for i, value := range vt {
					rt[i] = value.(string)
				}
				source[k] = rt
			default:
				source[k] = vt
			}
		case string:
			source[k] = settings.typer(vt)
		case bool, int64, float64:
			source[k] = vt
		default:
			source[k] = vt
			log.Printf("WARNING: unexpected %[1]T %#[1]v", vt)
		}
	}
	return source
}

// Coerce m slices to type-specific slices where applicable
func (settings Options) Retyper(source map[string]interface{}, options ...retyperOptions) map[string]interface{} {
	config := retyperConfig{}
	for _, f := range options {
		f(&config)
	}
	d := settings.retyping(source, config)
	return d
}
