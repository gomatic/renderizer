// Package template parses and executes a single Go text/template against a
// data context, using the gomatic/funcmap function set. It is an
// implementation package: pure, with no IO and no CLI knowledge, so a renderer
// is reusable anywhere a template must be evaluated.
package template

import (
	"bytes"
	"fmt"
	"maps"
	"math/rand"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/gomatic/clock"
	"github.com/gomatic/funcmap"

	"github.com/gomatic/renderizer/internal/constants"
)

// testSeed seeds the deterministic random source used in testing mode.
const testSeed = 0

type (
	// MissingKey is the text/template "missingkey" option.
	MissingKey string
	// Name labels a template for error reporting.
	Name string
	// TestingEnabled, when set, swaps nondeterministic template functions for
	// fixed ones so rendered output is stable across runs.
	TestingEnabled bool
)

// validMissingKeys are the text/template "missingkey" option values; anything
// else normalizes to "error".
const defaultMissingKey MissingKey = "error"

// NormalizeMissingKey returns key when it is a valid text/template missingkey
// option and "error" otherwise.
func NormalizeMissingKey(key MissingKey) MissingKey {
	switch key {
	case "zero", "error", "default", "invalid":
		return key
	}
	return defaultMissingKey
}

// Funcs returns a fresh function set combining the Sprig v3 library with
// funcmap's own functions. funcmap is overlaid last so it wins on a name clash:
// existing templates keep funcmap's signatures (e.g. the two-argument trim and
// the reversed-argument sub/div/mod) while Sprig v3 supplies everything funcmap
// does not define. In testing mode the nondeterministic functions
// (command_line, now, started, rand) are overridden so output is reproducible.
func Funcs(testing TestingEnabled) template.FuncMap {
	funcs := template.FuncMap{}
	maps.Copy(funcs, sprig.TxtFuncMap())
	maps.Copy(funcs, funcmap.New(funcmap.WithV1Map()))
	if testing {
		fixed := clock.Now(clock.Format)
		generator := rand.New(rand.NewSource(testSeed))
		funcs["command_line"] = func() string { return "testing" }
		funcs["now"] = fixed
		funcs["started"] = fixed
		funcs["rand"] = func() int64 { return generator.Int63() }
	}
	return funcs
}

// Render parses source as a template named name with the given functions and
// missingkey option, then executes it against data, returning the rendered
// bytes. Parse and execute failures surface as distinct sentinels; a panic in a
// template function is recovered as ErrRenderPanic rather than crashing.
func Render(funcs template.FuncMap, missing MissingKey, name Name, source []byte, data any) (out []byte, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			out, err = nil, constants.ErrRenderPanic.With(nil, fmt.Sprint(recovered))
		}
	}()
	parsed, err := template.New(string(name)).
		Option("missingkey=" + string(missing)).
		Funcs(funcs).
		Parse(string(source))
	if err != nil {
		return nil, constants.ErrParseTemplate.With(err)
	}
	var rendered bytes.Buffer
	if err := parsed.Execute(&rendered, data); err != nil {
		return nil, constants.ErrExecuteTemplate.With(err)
	}
	return rendered.Bytes(), nil
}
