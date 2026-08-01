package render

import (
	"context"
	"io"
	"log/slog"

	"gopkg.in/yaml.v3"

	"github.com/gomatic/renderizer/internal/constants"
	"github.com/gomatic/renderizer/internal/domain"
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
func Run(_ context.Context, logger *slog.Logger, cfg Config, _ ...domain.Argument) (Result, error) {
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
	if bool(cfg.VerboseEnabled) {
		logger.Info("Resolved templates.", "templates", sourceNames(sources))
	}
	if bool(cfg.DebuggingEnabled) {
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

// renderAll renders every source against data and concatenates the output,
// terminating each rendered block with a newline as the historical tool did.
func renderAll(cfg Config, data variables.Context, sources []templateSource) (Result, error) {
	funcs := template.Funcs(template.TestingEnabled(cfg.TestingEnabled))
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
func renderOne(
	cfg Config,
	funcs map[string]any,
	missing template.MissingKey,
	data variables.Context,
	source templateSource,
) ([]byte, error) {
	bytes, err := read(cfg, source)
	if err != nil {
		return nil, err
	}
	return template.Render(funcs, missing, template.Name(source.name), bytes, map[string]any(data))
}

// read returns the bytes of a source: stdin or a file via the injected reader.
func read(cfg Config, source templateSource) ([]byte, error) {
	if source.isStdin {
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
