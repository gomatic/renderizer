// Package settings loads renderizer variable defaults from YAML files and
// merges them into a single context. It is an implementation package: it knows
// nothing about the CLI, taking its file reader as an injected seam so every
// path is testable without touching the filesystem.
package settings

import (
	"dario.cat/mergo"
	"gopkg.in/yaml.v3"

	"github.com/gomatic/renderizer/internal/constants"
	"github.com/gomatic/renderizer/internal/variables"
)

// ReadFile reads a named file's bytes. os.ReadFile satisfies it in production;
// tests inject a fake.
type ReadFile func(name string) ([]byte, error)

// File is a settings file to load. An optional file that does not exist is
// skipped rather than failing — this is how the implicit default settings file
// stays optional.
type File struct {
	Path     string
	Optional bool
}

// Load reads each file in order, parses and retypes its YAML, and merges the
// results into one context. Later files append to slices from earlier ones.
func Load(read ReadFile, files []File, format variables.TimeFormat) (variables.Context, error) {
	merged := map[string]any{}
	for _, file := range files {
		loaded, err := loadFile(read, file, format)
		if err != nil {
			return nil, err
		}
		if err := mergeInto(merged, loaded); err != nil {
			return nil, err
		}
	}
	return merged, nil
}

// loadFile reads and parses one settings file, returning a nil map when an
// optional file is absent.
func loadFile(read ReadFile, file File, format variables.TimeFormat) (map[string]any, error) {
	data, err := read(file.Path)
	if err != nil {
		if file.Optional {
			return nil, nil
		}
		return nil, constants.ErrReadSettings.With(err, file.Path)
	}
	return parse(data, format)
}

// parse unmarshals YAML into a map and retypes its leaves.
func parse(data []byte, format variables.TimeFormat) (map[string]any, error) {
	loaded := map[string]any{}
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		return nil, constants.ErrParseSettings.With(err)
	}
	return variables.Retype(loaded, format, false), nil
}

// mergeInto merges src into dst, appending slices, leaving dst unchanged when
// src is nil.
func mergeInto(dst, src map[string]any) error {
	if src == nil {
		return nil
	}
	if err := mergo.Merge(&dst, src, mergo.WithAppendSlice); err != nil {
		return constants.ErrMergeContext.With(err)
	}
	return nil
}
