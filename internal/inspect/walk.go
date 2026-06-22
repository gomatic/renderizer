package inspect

import (
	"maps"
	"text/template/parse"
)

// scope tracks where references resolve while walking: root is the data passed
// to the template ($), dot is the current `.`, and vars binds range/with
// variables to their field sets.
type scope struct {
	root Fields
	dot  Fields
	vars map[string]Fields
}

// withDot returns a scope whose `.` resolves into fields.
func (s scope) withDot(fields Fields) scope {
	return scope{root: s.root, dot: fields, vars: s.vars}
}

// bind returns a scope with name bound to fields, copying the variable map so
// the binding does not leak to sibling scopes.
func (s scope) bind(name string, fields Fields) scope {
	vars := make(map[string]Fields, len(s.vars)+1)
	maps.Copy(vars, s.vars)
	vars[name] = fields
	return scope{root: s.root, dot: s.dot, vars: vars}
}

// walk dispatches a node to its handler; nodes that read no data are ignored.
func walk(node parse.Node, s scope) {
	switch typed := node.(type) {
	case *parse.ListNode:
		walkList(typed, s)
	case *parse.ActionNode:
		walkPipe(typed.Pipe, s)
	case *parse.RangeNode:
		walkRange(typed, s)
	case *parse.WithNode:
		walkWith(typed, s)
	case *parse.IfNode:
		walkBranch(typed.Pipe, typed.List, typed.ElseList, s)
	}
}

// walkList walks each child of a list node.
func walkList(node *parse.ListNode, s scope) {
	if node == nil {
		return
	}
	for _, child := range node.Nodes {
		walk(child, s)
	}
}

// walkPipe records every field referenced in a pipeline's commands. The pipe is
// always present: actions, if conditions, and parenthesized arguments never
// carry a nil pipe.
func walkPipe(pipe *parse.PipeNode, s scope) {
	for _, command := range pipe.Cmds {
		for _, arg := range command.Args {
			walkArg(arg, s)
		}
	}
}

// walkArg records the fields a single pipeline argument reads.
func walkArg(arg parse.Node, s scope) {
	switch typed := arg.(type) {
	case *parse.FieldNode:
		record(s.dot, typed.Ident)
	case *parse.VariableNode:
		recordVariable(typed.Ident, s)
	case *parse.ChainNode:
		walkArg(typed.Node, s)
	case *parse.PipeNode:
		walkPipe(typed, s)
	}
}

// walkBranch walks an if/else: the condition pipe and both branches, all with
// the same dot (a plain if does not shift scope).
func walkBranch(pipe *parse.PipeNode, list, elseList *parse.ListNode, s scope) {
	walkPipe(pipe, s)
	walk(list, s)
	walk(elseList, s)
}

// walkRange marks the ranged value a list, then walks the body with `.` (and the
// range value variable) bound to the element, and the else branch with the
// original scope.
func walkRange(node *parse.RangeNode, s scope) {
	element := rangeElement(node.Pipe, s)
	body := bindRangeVars(node.Pipe, element, s.withDot(element))
	walk(node.List, body)
	walk(node.ElseList, s)
}

// walkWith shifts `.` (and any declared variable) to the with value for the
// body, and walks the else branch with the original scope.
func walkWith(node *parse.WithNode, s scope) {
	field := pipeField(node.Pipe, s)
	body := s
	if field != nil {
		body = bindRangeVars(node.Pipe, field.Fields, s.withDot(field.Fields))
	}
	walk(node.List, body)
	walk(node.ElseList, s)
}

// rangeElement marks the ranged field a list and returns its element fields. A
// range over a non-field expression yields an anonymous element.
func rangeElement(pipe *parse.PipeNode, s scope) Fields {
	field := pipeField(pipe, s)
	if field == nil {
		return Fields{}
	}
	field.List = true
	return field.Fields
}

// bindRangeVars binds the value variable of a range/with declaration (the last
// declared variable) to element, leaving any index/key variable unmodeled.
func bindRangeVars(pipe *parse.PipeNode, element Fields, s scope) scope {
	if pipe == nil || len(pipe.Decl) == 0 {
		return s
	}
	value := pipe.Decl[len(pipe.Decl)-1]
	return s.bind(value.Ident[0], element)
}

// pipeField returns the field referenced by the last argument of a pipeline's
// last command — the value a range/with operates on — or nil when it is not a
// plain field or variable reference. A range/with pipe always has at least one
// command, each with at least one argument.
func pipeField(pipe *parse.PipeNode, s scope) *Field {
	command := pipe.Cmds[len(pipe.Cmds)-1]
	return argField(command.Args[len(command.Args)-1], s)
}

// argField returns the model field a field/variable argument refers to, or nil.
func argField(arg parse.Node, s scope) *Field {
	switch typed := arg.(type) {
	case *parse.FieldNode:
		return record(s.dot, typed.Ident)
	case *parse.VariableNode:
		base, rest := resolveVariable(typed.Ident, s)
		return record(base, rest)
	}
	return nil
}

// recordVariable records the fields read through a variable reference.
func recordVariable(ident []string, s scope) {
	base, rest := resolveVariable(ident, s)
	record(base, rest)
}

// resolveVariable resolves a variable's leading identifier to a field set and
// returns the remaining path. `$` is the root data; a known range/with variable
// resolves to its bound fields; anything else resolves to nil.
func resolveVariable(ident []string, s scope) (Fields, []string) {
	switch {
	case ident[0] == "$":
		return s.root, ident[1:]
	case known(s.vars, ident[0]):
		return s.vars[ident[0]], ident[1:]
	default:
		return nil, nil
	}
}

// known reports whether name is a bound variable.
func known(vars map[string]Fields, name string) bool {
	_, ok := vars[name]
	return ok
}

// record descends path under fields, creating nodes as needed, and returns the
// leaf field. It is a no-op returning nil when fields is nil or path is empty.
func record(fields Fields, path []string) *Field {
	if fields == nil || len(path) == 0 {
		return nil
	}
	var leaf *Field
	current := fields
	for _, name := range path {
		field, ok := current[name]
		if !ok {
			field = &Field{Fields: Fields{}}
			current[name] = field
		}
		leaf = field
		current = field.Fields
	}
	return leaf
}
