package render

import (
	"path/filepath"
	"strings"

	"github.com/gomatic/renderizer/internal/environment"
	"github.com/gomatic/renderizer/internal/settings"
	"github.com/gomatic/renderizer/internal/variables"
)

// Building the data context: the merged view of settings files, command-line
// assignments and the environment that every template is rendered against.
// Precedence lives here and nowhere else — a value's source decides whether it
// wins, and getting that wrong silently renders the wrong output.

// buildContext assembles the template data: command-line variables, then
// settings (which only fill names the variables did not set), then the
// environment map.
func buildContext(cfg Config) (variables.Context, error) {
	format := variables.TimeFormat(cfg.TimeFormat)
	data, err := variables.Assignments(cfg.Assignments, variables.Capitalization(cfg.CapitalizeEnabled), format)
	if err != nil {
		return nil, err
	}
	loaded, err := settings.Load(settings.ReadFile(cfg.ReadFile), settingsFiles(cfg), format)
	if err != nil {
		return nil, err
	}
	mergeDefaults(data, loaded)
	addEnvironment(cfg, data)
	return data, nil
}

// mergeDefaults fills data with settings values for names the command line did
// not set, recursing into maps so nested settings fill nested gaps. Command-line
// variables always win, so a name present in both keeps the command-line value.
func mergeDefaults(data, loaded variables.Context) {
	for key, value := range loaded {
		existing, present := data[key]
		if !present {
			data[key] = value
			continue
		}
		mergeNested(existing, value)
	}
}

// mergeNested deep-merges incoming into existing only when both are maps;
// otherwise the existing (command-line) value is kept.
func mergeNested(existing, incoming any) {
	existingMap, isMap := existing.(map[string]any)
	incomingMap, incomingIsMap := incoming.(map[string]any)
	if isMap && incomingIsMap {
		mergeDefaults(existingMap, incomingMap)
	}
}

// addEnvironment binds the environment map under the configured key.
func addEnvironment(cfg Config, data variables.Context) {
	if cfg.Environment == "" {
		return
	}
	data[string(cfg.Environment)] = map[string]string(environment.Load(environment.Environ(cfg.Environ)))
}

// settingsFiles returns the explicit --settings files, or the optional implicit
// default when none were given.
func settingsFiles(cfg Config) []settings.File {
	if len(cfg.Settings) == 0 {
		return []settings.File{{Path: "." + mainName(cfg) + ".yaml", IsOptional: true}}
	}
	files := make([]settings.File, len(cfg.Settings))
	for i, path := range cfg.Settings {
		files[i] = settings.File{Path: path}
	}
	return files
}

// mainName derives the base name for the default settings file from the first
// template, the working directory, or the fallback.
func mainName(cfg Config) string {
	if len(cfg.Templates) > 0 {
		base := filepath.Base(cfg.Templates[0])
		return strings.Split(strings.TrimLeft(base, "."), ".")[0]
	}
	if dir, err := cfg.Getwd(); err == nil {
		return filepath.Base(dir)
	}
	return defaultBase
}

// templateSource names a single render input: an explicit/discovered file, or
// stdin.
type templateSource struct {
	name    string
	isStdin bool
}
