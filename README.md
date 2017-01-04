# renderizer

[![Build Status](https://travis-ci.org/gomatic/renderizer.svg?branch=master)](https://travis-ci.org/gomatic/renderizer)

Render Go text templates from the command line.

    go get github.com/gomatic/renderizer

Supports providing top-level name/value pairs on the command line:

    renderizer --name=value --top=first template-file

**NOTE:** that _all_ parameter values are provided `name=value`.

## Example

First:

    git clone https://github.com/gomatic/renderizer.git
    cd renderizer
    go build

Generate the master pod.yaml:

    renderizer -S=test/.renderizer.yaml test/pod.yaml

 or, set `RENDERIZER` in the environment:

    RENDERIZER=test/.renderizer.yaml renderizer test/pod.yaml

alternatively, it'll try `.renderizer.yaml` in the current directory.

    (cd test; renderizer pod.yaml)

Generate the dev pod.yaml (after `cd test/`):

    renderizer pod.yaml --deployment=dev

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

Add the environment to the variables map as `_env`:

    renderizer -E template-file

or name the map key:

    renderizer -E=environ template-file

## Functions

### `add`
### `inc`
### `now`
### `lower`
### `upper`
### `trim`
### `trimLeft`
### `trimRight`
