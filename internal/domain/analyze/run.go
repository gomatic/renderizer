package analyze

import (
	"context"
	"io"
	"log/slog"

	"github.com/gomatic/renderizer/internal/constants"
	"github.com/gomatic/renderizer/internal/inspect"
	"github.com/gomatic/renderizer/internal/template"
)

// Result is the outcome of an analysis: the YAML data-model skeleton.
type Result struct {
	Output []byte
}

// Run reads the template and infers its input data model, returning the YAML
// skeleton. Template functions are only needed so parsing succeeds — the
// analysis never executes the template — so the default function set is used.
func Run(_ context.Context, logger *slog.Logger, cfg Config) (Result, error) {
	source, name, err := read(cfg)
	if err != nil {
		return Result{}, err
	}
	model, err := inspect.Analyze(template.Funcs(false), inspect.Name(name), source)
	if err != nil {
		return Result{}, err
	}
	logger.Debug("Analyzed template.", "template", name)
	return Result{Output: inspect.Skeleton(model)}, nil
}

// read returns the template bytes and a display name from a file or stdin.
func read(cfg Config) ([]byte, string, error) {
	if cfg.Template == "" {
		data, err := io.ReadAll(cfg.Source)
		if err != nil {
			return nil, "", constants.ErrReadTemplate.With(err)
		}
		return data, "stdin", nil
	}
	data, err := cfg.ReadFile(string(cfg.Template))
	if err != nil {
		return nil, "", constants.ErrOpenTemplate.With(err, string(cfg.Template))
	}
	return data, string(cfg.Template), nil
}
