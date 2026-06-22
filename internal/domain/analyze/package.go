// Package analyze orchestrates the analyze command: it reads a template from a
// file or stdin and infers the input data model the template requires,
// rendering it as a YAML skeleton. It delegates the parse-tree analysis to
// internal/inspect and holds no CLI or output-formatting logic. This is the
// domain tier between the app/cmd tier and the implementation packages.
package analyze
