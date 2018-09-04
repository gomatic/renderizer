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
	"github.com/urfave/cli"
)

//
func renderizer(ctx *cli.Context) error {

	globalContext := map[string]interface{}{}
	args := []string{}

	// Iterate the remaining arguments for variable overrides and file names.

	i := 0
	for a, arg := range settings.Arguments {
		if len(arg) == 0 {
			continue
		} else if arg[0] != '-' {
			args = append(args, arg)
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

		nameValuePair := strings.SplitN(strings.TrimLeft(arg, "-"), "=", 2)
		currentName := nameValuePair[0]
		i += 1

		if settings.Capitalize {
			currentName = fmt.Sprintf("%s%s", strings.ToTitle(currentName[:1]), strings.ToLower(currentName[1:]))
		}

		var currentValue interface{}
		if len(nameValuePair) == 1 { // i.e. a boolean
			currentValue = true
		} else {
			currentValue = typer(nameValuePair[1])
		}

		if settings.Debugging {
			log.Printf("file:%d var:%d name:%s value:%+v", a, i, currentName, currentValue)
		}

		// Build lists and try to keep the type of the slice consistent with the values.
		if values, exists := globalContext[currentName]; exists {
			switch valuesType := values.(type) {
			case string:
				switch y := currentValue.(type) {
				case string:
					globalContext[currentName] = []string{valuesType, y}
				default:
					globalContext[currentName] = []interface{}{valuesType, y}
				}
			case int:
				switch y := currentValue.(type) {
				case int:
					globalContext[currentName] = []int64{int64(valuesType), int64(y)}
				default:
					globalContext[currentName] = []interface{}{valuesType, y}
				}
			case int64:
				switch y := currentValue.(type) {
				case int64:
					globalContext[currentName] = []int64{valuesType, y}
				default:
					globalContext[currentName] = []interface{}{valuesType, y}
				}
			case float64:
				switch y := currentValue.(type) {
				case float64:
					globalContext[currentName] = []float64{valuesType, y}
				default:
					globalContext[currentName] = []interface{}{valuesType, y}
				}
			case bool:
				switch y := currentValue.(type) {
				case bool:
					globalContext[currentName] = []bool{valuesType, y}
				default:
					globalContext[currentName] = []interface{}{valuesType, y}
				}
			case *time.Time:
				switch y := currentValue.(type) {
				case *time.Time:
					globalContext[currentName] = []*time.Time{valuesType, y}
				default:
					globalContext[currentName] = []interface{}{valuesType, y}
				}
			case []string:
				switch y := currentValue.(type) {
				case string:
					globalContext[currentName] = append(valuesType, y)
				default:
					globalContext[currentName] = append(toInterfaceSlice(valuesType), y)
				}
			case []int64:
				switch y := currentValue.(type) {
				case int64:
					globalContext[currentName] = append(valuesType, y)
				default:
					globalContext[currentName] = append(toInterfaceSlice(valuesType), y)
				}
			case []float64:
				switch y := currentValue.(type) {
				case float64:
					globalContext[currentName] = append(valuesType, y)
				default:
					globalContext[currentName] = append(toInterfaceSlice(valuesType), y)
				}
			case []bool:
				switch y := currentValue.(type) {
				case bool:
					globalContext[currentName] = append(valuesType, y)
				default:
					globalContext[currentName] = append(toInterfaceSlice(valuesType), y)
				}
			case []*time.Time:
				switch y := currentValue.(type) {
				case *time.Time:
					globalContext[currentName] = append(valuesType, y)
				default:
					globalContext[currentName] = append(toInterfaceSlice(valuesType), y)
				}
			case []interface{}:
				globalContext[currentName] = append(valuesType, currentValue)
			}
		} else {
			globalContext[currentName] = currentValue
		}
	}

	// If there's no files, read from stdin.
	files := args
	if len(args) == 0 {
		stat, _ := os.Stdin.Stat()
		isTTY := (stat.Mode() & os.ModeCharDevice) != 0
		if isTTY {
			log.Println("source: stdin")
		}
		files = []string{""}
	}

	// Copy any loaded keys into the globalContext unless they already exist, i.e. they were provided on the command line.
	for n, v := range settings.Config {
		if _, exists := globalContext[n]; !exists {
			switch x := v.(type) {
			case int8:
				globalContext[n] = int64(x)
			case int16:
				globalContext[n] = int64(x)
			case int32:
				globalContext[n] = int64(x)
			case int:
				globalContext[n] = int64(x)
			default:
				globalContext[n] = v
			}
		}
	}

	if settings.Environment != "" || len(args) == 0 {
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
		log.Printf("globalContext: %+v", globalContext)
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
			fmt.Println(string(data))

			return 0
		}()
	}

	os.Exit(status)
	return nil
}
