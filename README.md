# renderizer

[![Build Status](https://travis-ci.org/gomatic/renderizer.svg?branch=master)](https://travis-ci.org/gomatic/renderizer)

Render Go text templates from the command line.

    go get github.com/gomatic/renderizer

Supports providing top-level name/value pairs on the command line:

    echo 'Hello, {{.User}}' | renderizer --user=${USER}

And read from the environment: 

    echo 'Hello, {{.env.User}}' | renderizer

## Usage:

    renderizer [OPTIONS] [--name=value]... [template-file]...

## Examples

Render the `pod.yaml` using values from `test/.renderizer.yaml`:

    renderizer --settings=example/pod/.pod.yaml example/pod/pod.yaml.tmpl

Or set `RENDERIZER` in the environment:

    RENDERIZER=example/.pod.yaml renderizer example/pod/pod.yaml.tmpl

Alternatively, it'll try `.renderizer.yaml` in the current directory.

    (cd example/pod; renderizer pod.yaml)

Next, override the `deployment` value to render the "dev" `pod.yaml` (after `cd example/pod`):

    renderizer --deployment=dev --name='spaced out' pod.yaml

For more examples, see the [`examples`](examples) folder.

# Configuration

### Settings

Settings can be loaded from a yaml:

    renderizer --settings=.renderizer.yaml --name=value --top=first template-file

### Capitalization `-C`

This is a positional toggle flag.

Variable names are converted to title case by default. It can be disabled for any subsequent variables:

    renderizer --name=value -C --top=first template-file

Sets:

    Name: value
    top: first

### Missing Keys

Control the missingkeys template-engine option:

    renderizer --missing=zero --top=first template-file

### Environment

Provide a name for the environment variables:

    renderizer --environment=env template-file

It defaults to `env` which is effectively the same as the above `--environment=env`.

## Template Functions

For the full list, see [functions.txt.tmpl](examples/functions/functions.txt.tmpl)

- `add` - `func(a, b int) int`
- `cleanse` - `func(s string) string` - remove `[^[:alpha:]]`
- `commandLine` - `func() string` the command line
- `environment` - `map[string]string` - the runtime environment
- `inc` - `func(a int) int`
- `join` - `func(a []interface, sep string) string`
- `lower` - `strings.ToLower`
- `now` - `time.Now`
- `replace` - `strings.Replace`
- `trim` - `strings.Trim`
- `trimLeft` - `strings.TrimLeft`
- `trimRight` - `strings.TrimRight`
- `upper` - `strings.ToUpper`
