// Package inspect infers the input data model a Go text/template requires by
// walking its parse tree, and renders that model as a YAML skeleton. It is an
// implementation package: pure, with no IO and no CLI knowledge.
//
// It recognizes field references (`.A.B`), `range` (marking the ranged value a
// list and inferring its element fields), `with` and `if` scopes, and `$`/named
// range variables. Constructs it cannot attribute to the input — fields reached
// through a function result or a computed chain — are skipped, so the model is a
// sound lower bound on the required input, not necessarily exhaustive.
package inspect

import (
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/gomatic/renderizer/internal/constants"
)

type (
	// Name labels the template for parse-error reporting.
	Name string

	// Fields maps a field name to its inferred sub-model.
	Fields map[string]*Field

	// Field is one node of the inferred model. List marks a value reached via
	// range (a collection); Fields holds its nested or element fields.
	Field struct {
		Fields Fields
		List   bool
	}

	// Model is the inferred input data model: the top-level fields a template
	// reads from its data.
	Model struct {
		Fields Fields
	}
)

// Analyze parses source and infers the data model it reads. funcs must contain
// every function the template calls so parsing succeeds; missingkey=zero keeps
// analysis independent of the data.
func Analyze(funcs template.FuncMap, name Name, source []byte) (Model, error) {
	parsed, err := template.New(string(name)).Funcs(funcs).Option("missingkey=zero").Parse(string(source))
	if err != nil {
		return Model{}, constants.ErrParseTemplate.With(err)
	}
	model := Model{Fields: Fields{}}
	root := scope{root: model.Fields, dot: model.Fields, vars: map[string]Fields{}}
	if parsed.Tree != nil {
		walk(parsed.Root, root)
	}
	return model, nil
}

// Skeleton renders the model as a YAML document of placeholder values: scalars
// as empty strings, lists as a single example element, and nested fields as
// maps. An empty model renders as an explanatory comment.
func Skeleton(model Model) []byte {
	if len(model.Fields) == 0 {
		return []byte("# template requires no input data\n")
	}
	// yaml.Marshal of plain map/slice/string values is infallible.
	out, _ := yaml.Marshal(build(model.Fields))
	return out
}

// build turns a field set into a placeholder value tree.
func build(fields Fields) any {
	if len(fields) == 0 {
		return ""
	}
	placeholder := make(map[string]any, len(fields))
	for name, field := range fields {
		placeholder[name] = value(field)
	}
	return placeholder
}

// value turns a single field into its placeholder, wrapping list fields in a
// one-element slice.
func value(field *Field) any {
	inner := build(field.Fields)
	if field.List {
		return []any{inner}
	}
	return inner
}
