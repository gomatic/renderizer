package main

import (
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

// Coerce m slices to type-specific slices where applicable
func retyper(m map[string]interface{}) {
	for k, v := range m {
		switch vt := v.(type) {
		case map[string]interface{}:
			retyper(vt)
		case []interface{}:
			if len(vt) == 1 {
				m[k] = vt[0]
				return
			}
			kind := reflect.Invalid
			for i, value := range vt {
				t := reflect.TypeOf(value)
				if i == 0 {
					kind = t.Kind()
					continue
				}
				// If the types are different or invalid, do not create a type-specific slice.
				if kind == reflect.Invalid || t.Kind() != kind {
					return
				}
			}
			// Create the type-specific list. This set of types must correspond to the types returned from `typer()`.
			switch kind {
			case reflect.Bool:
				rt := make([]bool, len(vt))
				for i, value := range vt {
					rt[i] = value.(bool)
				}
				m[k] = rt
			case reflect.Int64:
				rt := make([]int64, len(vt))
				for i, value := range vt {
					rt[i] = value.(int64)
				}
				m[k] = rt
			case reflect.Float64:
				rt := make([]float64, len(vt))
				for i, value := range vt {
					rt[i] = value.(float64)
				}
				m[k] = rt
			case reflect.String:
				rt := make([]string, len(vt))
				for i, value := range vt {
					rt[i] = value.(string)
				}
				m[k] = rt
			}
		default:
			log.Printf("WARNING: unexpected %[1]T %#[1]v", vt)
		}
	}
}
