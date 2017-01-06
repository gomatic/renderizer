package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"
)

//
var funcs = template.FuncMap{
	"add":         func(a, b int) int { return a + b },
	"cleanse":     cleanse(`[^[:alpha:]]`),
	"commandLine": func() string { return settings.CommandLine },
	"environment": environment(),
	"inc":         func(a int) int { return a + 1 },
	"join":        joiner,
	"lower":       strings.ToLower,
	"now":         time.Now,
	"replace":     strings.Replace,
	"trim":        strings.Trim,
	"trimLeft":    strings.TrimLeft,
	"trimRight":   strings.TrimRight,
	"upper":       strings.ToUpper,
}

//
func joiner(y []interface{}, sep string) (s string) {
	if y == nil {
		return
	}
	i := make([]string, len(y))
	for j, z := range y {
		i[j] = fmt.Sprintf("%v", z)
	}
	return strings.Join(i, sep)
}

//
func cleanse(r string) func(string) string {
	re := regexp.MustCompile(r)
	return func(s string) string {
		return re.ReplaceAllString(s, "")
	}
}

//
func environment() func(string) string {
	env := make(map[string]string)
	for _, item := range os.Environ() {
		splits := strings.Split(item, "=")
		env[splits[0]] = splits[1]
	}
	return func(n string) string {
		v, exists := env[n]
		if !exists {
			v = settings.Environment
		}
		return v
	}
}
