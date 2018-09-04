package main

import (
	"strconv"
	"time"
)

// Convert to []interface{}
func toInterfaceSlice(x interface{}) (i []interface{}) {
	switch y := x.(type) {
	case []string:
		i = make([]interface{}, len(y))
		for j, z := range y {
			i[j] = z
		}
	case []int64:
		i = make([]interface{}, len(y))
		for j, z := range y {
			i[j] = z
		}
	case []float64:
		i = make([]interface{}, len(y))
		for j, z := range y {
			i[j] = z
		}
	case []bool:
		i = make([]interface{}, len(y))
		for j, z := range y {
			i[j] = z
		}
	case []*time.Time:
		i = make([]interface{}, len(y))
		for j, z := range y {
			i[j] = z
		}
	}
	return i
}

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
