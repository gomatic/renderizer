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
type Renderizer interface {
	Render() error
}

//
func New(options ...Option) Renderizer {
	settings := Options{
		Config:      map[string]interface{}{},
		Capitalize:  true,
		MissingKey:  "error",
		TimeFormat:  "20060102T150405",
		Environment: "env",
		Arguments:   []string{},
		Templates:   []string{},
	}

	for _, opt := range options {
		if opt != nil {
			opt(&settings)
		}
	}

	return &settings
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
		if settings.Verbose {
			log.Printf("file: %+v", file)
		}
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
			vs := Values{}
			for k, v := range globalContext {
				vs[k] = Value{Value: v}
			}
			data, err = Renderer(Context{
				Name:       file,
				Functions:  funcmap.Map,
				MissingKey: settings.MissingKey,
				Values:     vs,
			}, r)
			if err != nil {
				log.Printf("%s: %s", file, err)
				return 1
			}
			fmt.Println(string(data))

			return 0
		}()
	}

	os.Exit(status)
	return nil
}

//
type Value struct {
	TitleCase bool
	Name      string
	Value     interface{}
}

//
type Values map[string]Value

//
func (v Values) Map() (m map[string]interface{}) {
	m = map[string]interface{}{}
	for k, v := range v {
		m[k] = v.Value
	}
	return m
}

type Context struct {
	Name       string
	MissingKey string
	Functions  template.FuncMap
	Values     Values
}

//
func Renderer(c Context, r io.Reader) ([]byte, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if c.Functions == nil {
		c.Functions = funcmap.Map
	}
	if c.MissingKey == "" {
		c.MissingKey = "error"
	}
	tmpl, err := template.New(c.Name).
		Option(fmt.Sprintf("missingkey=%s", c.MissingKey)).
		Funcs(c.Functions).
		Parse(string(data))
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, c.Values.Map())
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
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
