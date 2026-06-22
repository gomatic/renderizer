# renderizer

Render Go `text/template` files from the command line, with variables supplied as command-line flags, YAML settings files, and the process environment.

```sh
echo 'Hello, {{.Name}}!' | renderizer --stdin --name=World
# Hello, World!
```

## Install

```sh
go install github.com/gomatic/renderizer/cmd/renderizer@latest
```

## Usage

```
renderizer [options] [--name=value...] [template-file...]
```

Variables are supplied three ways and merged, with command-line variables winning over settings:

- **Command-line variables** — any `--name=value` becomes a template variable. Names nest with dots (`--a.b.c=1`), values are typed automatically (int, float, bool, time, string), and a repeated name builds a list. `-C` toggles title-casing of the variable names that follow it.
- **Settings files** — `--settings=file.yaml` (repeatable) loads variable defaults from YAML. With none given, an optional `.<name>.yaml` in the working directory is used if present.
- **Environment** — bound under `env` by default (`{{.env.HOME}}`); rename with `--environment=name`.

When no template file is given, renderizer reads stdin (a pipe, or `--stdin`), or discovers a default `renderizer.{yaml,json,html,txt,xml}[.tmpl]` in the working directory.

| Option | Aliases | Description |
|--------|---------|-------------|
| `--settings` | `-S` `-s` | Load settings from a YAML file (repeatable; env `RENDERIZER`) |
| `--missing` | `-M` `-m` | `missingkey` option: `default`, `zero`, `error`, `invalid` |
| `--environment` | `-E` `-e` `--env` | Variable name to bind the environment under (default `env`) |
| `--stdin` | `-c` | Read the template from stdin |
| `--testing` | `-T` | Make nondeterministic template functions reproducible |
| `--debugging` | `-D` `--debug` | Enable debug logging |
| `--verbose` | `-V` | Enable verbose logging |

Template functions come from [`gomatic/funcmap`](https://github.com/gomatic/funcmap) (`upper`, `lower`, `trim`, `replace`, `add`, `inc`, `environment`, `command_line`, the `ip*` helpers, and more) plus the full [Sprig v3](https://masterminds.github.io/sprig/) library (`b64enc`, `date`, `default`, `list`, `dict`, `fromJson`, `toRawJson`, …). On a name clash funcmap's function wins, so funcmap's signatures (e.g. the two-argument `trim`, the reversed-argument `sub`/`div`/`mod`) are preserved and Sprig supplies everything else.

## Analyze

`renderizer analyze <template-file>` (or piped stdin) infers the input data model a template requires and prints it as a YAML skeleton — scalars as `""`, ranged values as single-element lists, nested fields as maps:

```sh
$ renderizer analyze examples/basic/basic.txt.tmpl
Items:
    - ""
Name: ""
```

It walks the template's parse tree, recognizing field references, `range` (a list with inferred element fields), `with`/`if` scopes, and `$`/named range variables. Fields reached only through a function result are skipped, so the model is a sound lower bound on the required input, not necessarily exhaustive.

`renderizer version` prints the version (alongside the built-in `--version`/`-v`).

## Develop

```sh
make check   # vet, lint, staticcheck, govulncheck, and 100% coverage
make build   # build the binary into bin/
```

See [docs/architecture.md](docs/architecture.md) for the package layout.
