package main

import (
	"text/template"
	"time"
)

//
var funcs = template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"inc": func(a int) int { return a + 1 },
	"now": func() time.Time { return time.Now().UTC() },
}
