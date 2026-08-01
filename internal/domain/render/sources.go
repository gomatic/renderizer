package render

import (
	"fmt"
	"path/filepath"

	"github.com/gomatic/renderizer/internal/constants"
)

// Resolving which templates to render: explicit files, piped stdin, or the
// conventional file discovered from the working directory. Discovery is the
// part that can surprise, so it is isolated from the rendering it feeds.

// resolveSources decides what to render: explicit templates, stdin, a
// discovered default file, or — failing all — a missing-template error.
func resolveSources(cfg Config) ([]templateSource, error) {
	if len(cfg.Templates) > 0 {
		return fileSources(cfg.Templates), nil
	}
	if bool(cfg.StdinEnabled) {
		return []templateSource{{name: "stdin", isStdin: true}}, nil
	}
	if name, ok := discover(cfg); ok {
		return []templateSource{{name: name}}, nil
	}
	return nil, constants.ErrMissingTemplate
}

// fileSources maps template paths to file sources.
func fileSources(templates TemplateFiles) []templateSource {
	sources := make([]templateSource, len(templates))
	for i, name := range templates {
		sources[i] = templateSource{name: name}
	}
	return sources
}

// discover looks for a default template file across the candidate base names.
func discover(cfg Config) (string, bool) {
	for _, base := range bases(cfg.Getwd) {
		for _, candidate := range candidates(base) {
			if cfg.Exists(candidate) {
				return candidate, true
			}
		}
	}
	return "", false
}

// baseName is a default-discovery base file name (the working directory name
// or the tool's fallback), tried against each candidate extension.
type baseName string

// bases returns the base names tried during discovery: the working directory
// name (when available) and the fallback.
func bases(getwd GetwdFunc) []baseName {
	if dir, err := getwd(); err == nil {
		return []baseName{baseName(filepath.Base(dir)), defaultBase}
	}
	return []baseName{defaultBase}
}

// candidates enumerates the file names tried for a base name, matching the
// historical discovery order (type extension × optional .tmpl suffix).
func candidates(base baseName) []string {
	suffixes := []string{".tmpl", ""}
	types := []string{"yaml", "json", "html", "txt", "xml", ""}
	names := make([]string, 0, len(suffixes)*len(types))
	for _, suffix := range suffixes {
		for _, typ := range types {
			names = append(names, fmt.Sprintf("%s.%s%s", base, typ, suffix))
		}
	}
	return names
}
