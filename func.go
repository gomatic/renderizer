package main

import (
	"strings"
	"text/template"
	"time"
)

//
var funcs = template.FuncMap{
	"add":       func(a, b int) int { return a + b },
	"inc":       func(a int) int { return a + 1 },
	"now":       func() time.Time { return time.Now().UTC() },
	"lower":     func(s string) string { return strings.ToLower(s) },
	"upper":     func(s string) string { return strings.ToUpper(s) },
	"trim":      func(s string, cut string) string { return strings.Trim(s, cut) },
	"trimLeft":  func(s string, cut string) string { return strings.TrimLeft(s, cut) },
	"trimRight": func(s string, cut string) string { return strings.TrimRight(s, cut) },
}
