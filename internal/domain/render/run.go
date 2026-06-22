package render

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gomatic/renderizer/internal/constants"
	"github.com/gomatic/renderizer/internal/environment"
	"github.com/gomatic/renderizer/internal/settings"
	"github.com/gomatic/renderizer/internal/template"
	"github.com/gomatic/renderizer/internal/variables"
)

// defaultBase is the fallback base name for default settings/template discovery.
const defaultBase = "renderizer"

// Result is the outcome of a render: the concatenated rendered output, ready to
// be written verbatim to the command's writer.
type Result struct {
	Output []byte
}

// Run builds the template data context, resolves the templates to render, and
// renders each, returning the concatenated output. It holds no presentation
// logic; the caller writes Result.Output.
func Run(_ context.Context, logger *slog.Logger, cfg Config) (Result, error) {
	data, err := buildContext(cfg)
	if err != nil {
		return Result{}, err
	}
	sources, err := resolveSources(cfg)
	if err != nil {
		return Result{}, err
	}
	logResolution(logger, cfg, data, sources)
	return renderAll(cfg, data, sources)
}

// logResolution emits the verbose template/source summary and the debug context
// dump, gated on the corresponding flags so the dump cost is only paid when
// requested.
func logResolution(logger *slog.Logger, cfg Config, data variables.Context, sources []templateSource) {
	logger.Debug("Rendering templates.", "count", len(sources))
	if bool(cfg.Verbose) {
		logger.Info("Resolved templates.", "templates", sourceNames(sources))
	}
	if bool(cfg.Debugging) {
		logger.Debug("Template context.", "data", dump(data))
	}
}

// sourceNames lists the resolved source names for logging.
func sourceNames(sources []templateSource) []string {
	names := make([]string, len(sources))
	for i, source := range sources {
		names[i] = source.name
	}
	return names
}

// dump renders the context as YAML for debug logging. Marshaling a plain
// map/slice/scalar tree is infallible.
func dump(data variables.Context) string {
	out, _ := yaml.Marshal(map[string]any(data))
	return string(out)
}

// buildContext assembles the template data: command-line variables, then
// settings (which only fill names the variables did not set), then the
// environment map.
func buildContext(cfg Config) (variables.Context, error) {
	format := variables.TimeFormat(cfg.TimeFormat)
	data, err := variables.Assignments(cfg.Assignments, variables.Capitalization(cfg.Capitalize), format)
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
		return []settings.File{{Path: "." + mainName(cfg) + ".yaml", Optional: true}}
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
	name  string
	stdin bool
}

// resolveSources decides what to render: explicit templates, stdin, a
// discovered default file, or — failing all — a missing-template error.
func resolveSources(cfg Config) ([]templateSource, error) {
	if len(cfg.Templates) > 0 {
		return fileSources(cfg.Templates), nil
	}
	if bool(cfg.Stdin) {
		return []templateSource{{name: "stdin", stdin: true}}, nil
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

// bases returns the base names tried during discovery: the working directory
// name (when available) and the fallback.
func bases(getwd GetwdFunc) []string {
	if dir, err := getwd(); err == nil {
		return []string{filepath.Base(dir), defaultBase}
	}
	return []string{defaultBase}
}

// candidates enumerates the file names tried for a base name, matching the
// historical discovery order (type extension × optional .tmpl suffix).
func candidates(base string) []string {
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

// renderAll renders every source against data and concatenates the output,
// terminating each rendered block with a newline as the historical tool did.
func renderAll(cfg Config, data variables.Context, sources []templateSource) (Result, error) {
	funcs := template.Funcs(template.TestingEnabled(cfg.Testing))
	missing := template.NormalizeMissingKey(template.MissingKey(cfg.MissingKey))
	var output []byte
	for _, source := range sources {
		rendered, err := renderOne(cfg, funcs, missing, data, source)
		if err != nil {
			return Result{Output: output}, err
		}
		output = append(append(output, rendered...), '\n')
	}
	return Result{Output: output}, nil
}

// renderOne reads and renders a single source.
func renderOne(cfg Config, funcs map[string]any, missing template.MissingKey, data variables.Context, source templateSource) ([]byte, error) {
	bytes, err := read(cfg, source)
	if err != nil {
		return nil, err
	}
	return template.Render(funcs, missing, template.Name(source.name), bytes, map[string]any(data))
}

// read returns the bytes of a source: stdin or a file via the injected reader.
func read(cfg Config, source templateSource) ([]byte, error) {
	if source.stdin {
		data, err := io.ReadAll(cfg.Source)
		if err != nil {
			return nil, constants.ErrReadTemplate.With(err)
		}
		return data, nil
	}
	data, err := cfg.ReadFile(source.name)
	if err != nil {
		return nil, constants.ErrOpenTemplate.With(err, source.name)
	}
	return data, nil
}
