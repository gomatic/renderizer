# Architecture

renderizer renders Go `text/template` files from three merged variable sources — command-line flags, YAML settings, and the process environment — and analyzes templates to infer their input data model. It follows the [`gomatic/template.cli`](https://github.com/gomatic/template.cli) layered structure.

## Public library

The module root (`package renderizer`, [`renderizer.go`](../renderizer.go)) is a thin, minimal facade for embedding the engine in other Go programs (e.g. `lono`): `Funcs` (the Sprig + funcmap set, extensible by the caller), `Render`, and `Analyze`. It wraps `internal/template` and `internal/inspect` only — the implementation packages stay `internal/`, so the public surface is exactly those three functions plus their value types (`MissingKey`, `Name`, `Template`). The CLI and the library share one engine; neither tier depends on the other.

## Tiers

Dependencies flow one direction: composition root → app → domain → implementation. No tier imports one to its left.

### Composition root — `cmd/renderizer`

`main.go` is the only place that touches the outside world. It:

1. tokenizes `os.Args` via `internal/variables.Tokenize`;
2. assembles an `app.Runtime` carrying the real IO seams (`os.ReadFile`, `os.Stat`, `os.Getwd`, `os.Environ`, stdin) and the parsed assignments;
3. builds the root command via `app/commands/render.Command` and attaches the `analyze` and `version` subcommands;
4. runs it and maps the resulting error to an exit code via `app.ExitCode`.

It holds no behavior and is excluded from the 100% coverage gate (`COVER_PKGS` in the `Makefile`); its behavior is verified by the end-to-end tests in `main_test.go`.

### app — `internal/app`

The CLI command definitions and the shared seams between the framework and the domain.

- **`commands/<name>/command.go`** — one package per command (`render`, `analyze`, `version`). Each `Command(rt)` returns a `cli.Command` that binds flags to a domain `Config` and wires its action; the action assembles the config from flags + runtime seams, runs the domain, and writes the output. `render` is renderizer's default action, so its `Command` is used as the root and carries the global flags; `analyze` and `version` are subcommands.
- **`action.go`** — `Write`, the shared seam that writes a command's raw output and propagates its error (mirroring template.cli's `app.Default`, but for bytes rather than JSON).
- **`logger.go`** — `NewLogger` (slog level from the verbosity flags).
- **`exit.go`** — `ExitCode` (sentinel error → historical exit code: read=2, parse=4, execute=8, panic=15, otherwise 1).
- **`runtime.go`** — `Runtime`, the injected IO seams and parsed arguments passed from the composition root to each command.

### domain — `internal/domain/{render,analyze}`

Orchestration only. `Run(ctx, logger, cfg) (Result, error)`:

1. **builds the data context** — command-line variables (`variables.Assignments`), then settings defaults (`settings.Load`, filling only names the command line did not set), then the environment map (`environment.Load`);
2. **resolves the templates** — explicit files, else stdin, else a discovered default, else `ErrMissingTemplate`;
3. **renders each** — `template.Render` against the context — and concatenates the output.

`internal/domain/analyze` mirrors this for the `analyze` command: it reads a template (file or stdin) and delegates to `internal/inspect` to infer the data model.

Neither domain imports urfave/cli or contains IO of its own; every filesystem and environment access goes through an injected seam on its `Config`, so every branch is reachable from a test.

### implementation

- **`internal/variables`** — the cli/v3-compatible argument handling. `Tokenize` splits argv into known flags (for cli/v3), arbitrary `--name=value` assignments, and template files. `Assignments` builds a typed, possibly nested context from the assignments, honoring the `-C` capitalization toggle, dotted nesting, repeated-name lists, and automatic value typing.
- **`internal/settings`** — loads and merges YAML settings files, retyping their values to match command-line typing.
- **`internal/template`** — builds the function set from the Sprig v3 library overlaid with `gomatic/funcmap` (funcmap wins on name clashes; nondeterministic functions are overridden in testing mode) and parses/executes one template, recovering panics as `ErrRenderPanic`.
- **`internal/environment`** — turns a `KEY=VALUE` listing into a map.
- **`internal/inspect`** — walks a template's parse tree to infer its required input data model and renders it as a YAML skeleton (the `analyze` command).
- **`internal/constants`** — the sentinel `Error` type and the errors the program can emit.

## The cli/v3 constraint

urfave/cli v3 rejects undefined flags, but renderizer's defining feature is passing arbitrary `--name=value` variables. The resolution is the `Tokenize` pre-pass: arbitrary assignments and the positional `-C` toggle are extracted from argv *before* cli/v3 parses, so cli/v3 only ever sees its declared flags. This preserves the full historical CLI surface while staying within a conventional CLI framework.

## Quality gate

`make check` runs `go vet`, `golangci-lint` (cognitive complexity ≤ 7), `staticcheck`, `govulncheck`, and a 100%-statement-coverage assertion over every package except `cmd/`. The tools are pinned in the `go.mod` `tool` stanza and run via `go tool`.
