# renderizer — agent instructions

renderizer is a Go CLI that renders Go `text/template` files from command-line variables, YAML settings, and the environment. It follows the [`gomatic/template.cli`](https://github.com/gomatic/template.cli) layered structure and the gomatic Go quality gate. **Read [`docs/architecture.md`](../docs/architecture.md) before changing anything.**

It also exposes a minimal **public library** at the module root (`package renderizer` — `Funcs`, `Render`, `Analyze`) so other Go programs (e.g. `lono`) embed the same engine. Keep that surface minimal: the implementation packages stay `internal/`; add to the facade only what an embedder genuinely needs.

## The tiers (app → domain → implementation)

Dependencies flow one direction only.

1. **composition root** — [`cmd/renderizer/main.go`](../cmd/renderizer/main.go). Tokenizes argv, assembles the runtime IO seams (`os.ReadFile`, `os.Stat`, `os.Getwd`, `os.Environ`, stdin), wires the command packages, and maps the error to an exit code. No behavior — excluded from the coverage gate.
2. **app** — [`internal/app`](../internal/app). The command definitions and shared seams: [`commands/<name>/command.go`](../internal/app/commands) (one per command — `render` is the root default action, `analyze` and `version` are subcommands), plus `action.go` (the `Write` output seam), `logger.go`, `exit.go`, and `runtime.go`. Each `command.go` binds flags to a domain `Config` and wires its action; business logic stays in the domain. 100% covered.
3. **domain** — [`internal/domain/<name>`](../internal/domain): `package.go`, `types.go`, `config.go`, `run.go` (+ `run_test.go`). Orchestration only — `render` builds the data context, resolves the templates, and renders each; `analyze` infers a template's data model. Delegates all real work to implementation packages. Never imports urfave/cli.
4. **implementation** — [`internal/variables`](../internal/variables) (the cli/v3-compatible argument tokenizer plus typed nested-map builder), [`internal/settings`](../internal/settings), [`internal/template`](../internal/template), [`internal/environment`](../internal/environment), [`internal/inspect`](../internal/inspect) (the analyze parse-tree walker), and [`internal/constants`](../internal/constants) (sentinel errors). Pure, reusable, no CLI knowledge.

## The CLI technique

urfave/cli v3 rejects undefined flags, but renderizer's identity is arbitrary `--name=value` variables. [`internal/variables.Tokenize`](../internal/variables/tokenize.go) pre-splits argv into known flags (handed to cli/v3), arbitrary variable assignments (parsed by this package), and template files — so cli/v3 never sees an undefined flag. Preserve this seam when changing the CLI.

## Hard rules

- Named types for every `Config` field and function parameter — no bare `string`/`bool`/`int`.
- Private by default; constant sentinel errors in `internal/constants`, matched with `errors.Is`.
- 100% statement coverage (cmd excluded), cognitive complexity ≤ 7 per function, `gofumpt`-clean.
- `make check` must exit zero before a change is complete.
