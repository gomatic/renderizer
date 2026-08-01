package version

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gomatic/renderizer/internal/domain"
)

// Result is the outcome of a version query: the formatted version line, ready
// to be written verbatim to the command's writer.
type Result struct {
	Output []byte
}

// Run formats the "<app> version <build>" line. It holds no presentation logic
// beyond the line itself; the caller writes Result.Output. The signature
// mirrors the other domain Runs (context, logger, config) so every command
// wires through the same seam, even though formatting needs neither the
// context nor the logger.
func Run(_ context.Context, _ *slog.Logger, cfg Config, _ ...domain.Argument) (Result, error) {
	return Result{Output: fmt.Appendf(nil, "%s version %s\n", cfg.App, cfg.Build)}, nil
}
