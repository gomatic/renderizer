// Package environment turns a process environment listing into a name/value
// map for injection into a template context. The environment source is an
// injected seam so the conversion is testable without reading the real
// process environment.
package environment

import "strings"

// Environ returns the process environment as "KEY=VALUE" strings. os.Environ
// satisfies it in production; tests inject a fixed slice.
type Environ func() []string

// Variables is a decoded environment map.
type Variables map[string]string

// Load splits each "KEY=VALUE" entry on the first '=' into the resulting map.
// An entry with no '=' maps the whole entry to an empty value.
func Load(environ Environ) Variables {
	variables := Variables{}
	for _, entry := range environ() {
		key, value, _ := strings.Cut(entry, "=")
		variables[key] = value
	}
	return variables
}
