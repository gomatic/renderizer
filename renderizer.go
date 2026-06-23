// Package renderizer is the public library API for rendering and analyzing Go
// text/templates with the renderizer function set (Sprig v3 overlaid by
// gomatic/funcmap). It is a thin facade over the same internal engine the CLI
// uses; the implementation packages stay internal so the public surface is just
// what an embedding program needs: obtain the functions, render, and analyze.
package renderizer

import (
	"text/template"

	"github.com/gomatic/renderizer/internal/inspect"
	tmpl "github.com/gomatic/renderizer/internal/template"
)

type (
	// MissingKey is the text/template "missingkey" option
	// (default|zero|error|invalid); any other value normalizes to "error".
	MissingKey string
	// Name labels a template for error reporting.
	Name string
	// Template is raw Go text/template source.
	Template []byte
)

// Funcs returns a fresh standard function set: Sprig v3 overlaid by
// gomatic/funcmap. Callers may add their own functions to the returned map
// before rendering or analyzing.
func Funcs() template.FuncMap {
	return tmpl.Funcs(false)
}

// Render parses source (labeled name) with funcs and the missingkey option,
// executes it against data, and returns the rendered bytes.
func Render(funcs template.FuncMap, missing MissingKey, name Name, source Template, data any) ([]byte, error) {
	return tmpl.Render(funcs, tmpl.NormalizeMissingKey(tmpl.MissingKey(missing)), tmpl.Name(name), source, data)
}

// Analyze infers the input data a template requires and returns it as a YAML
// skeleton (scalars "", ranged values single-element lists, nested fields maps).
// funcs must contain every function the template calls so it parses.
func Analyze(funcs template.FuncMap, name Name, source Template) ([]byte, error) {
	model, err := inspect.Analyze(funcs, inspect.Name(name), source)
	if err != nil {
		return nil, err
	}
	return inspect.Skeleton(model), nil
}
