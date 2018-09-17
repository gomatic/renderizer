package main

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"
)

// Transform a string into the best type.
func typer(d string) (result interface{}) {
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
func retyping(source map[string]interface{}, config retyperConfig) map[string]interface{} {
	for k, v := range source {
		switch vt := v.(type) {
		case map[interface{}]interface{}:
			f := make(map[string]interface{}, len(vt))
			for e, s := range vt {
				f[fmt.Sprintf("%v", e)] = s
			}
			retyping(f, config)
			source[k] = f
		case map[string]string:
			f := make(map[string]interface{}, len(vt))
			for e, s := range vt {
				f[e] = s
			}
			retyping(f, config)
			source[k] = f
		case int:
			source[k] = int64(vt)
		case map[string]interface{}:
			retyping(vt, config)
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
			source[k] = typer(vt)
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
func retyper(source map[string]interface{}, options ...retyperOptions) map[string]interface{} {
	config := retyperConfig{}
	for _, f := range options {
		f(&config)
	}
	d := retyping(source, config)
	return d
}
