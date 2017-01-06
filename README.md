# renderizer

[![Build Status](https://travis-ci.org/gomatic/renderizer.svg?branch=master)](https://travis-ci.org/gomatic/renderizer)

Render Go text templates from the command line.

    go get github.com/gomatic/renderizer

Supports providing top-level name/value pairs on the command line:

    renderizer --name=value --top=first template-file

Sets:

    Name: value
    Top: first

**NOTE:**

- Template values are provided `--name=value`.
- Renderizer controls are provided using `-NAME[=value]`.

## Example

First:

    git clone https://github.com/gomatic/renderizer.git
    cd renderizer
    go build

Render the `pod.yaml` using values from `test/.renderizer.yaml`:

    renderizer -S=test/.renderizer.yaml test/pod.yaml

Or set `RENDERIZER` in the environment:

    RENDERIZER=test/.renderizer.yaml renderizer test/pod.yaml

Alternatively, it'll try `.renderizer.yaml` in the current directory.

    (cd test; renderizer pod.yaml)

Next, override the `deployment` value to render the "dev" `pod.yaml` (after `cd test/`):

    renderizer --deployment=dev --name='spaced out' pod.yaml

## Configuration

### Settings `-S=`

Settings can be loaded from a yaml:

    renderizer -S=renderizer.yaml --name=value --top=first template-file

### Capitalization `-C`

This is a positional toggle flag.

Variable names are converted to title case by default. It can be disabled for any subsequent variables:

    renderizer --name=value -C --top=first template-file

Sets:

    Name: value
    top: first

### Missing Keys `-M=`

Control the missingkeys template-engine option:

    renderizer -M=zero --top=first template-file

### Environment `-E=`

Provide a default value for missing environment variables:

    renderizer -E=--missing-- template-file

## Template Functions

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
