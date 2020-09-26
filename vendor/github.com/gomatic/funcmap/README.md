# funcmap

[![Build Status](https://travis-ci.org/gomatic/funcmap.svg?branch=master)](https://travis-ci.org/gomatic/funcmap)

Go template functions.

    import "github.com/gomatic/funcmap"

...

    template.New(name).
        Funcs(funcmap.Map).
        Parse(templateSource).
        Execute(&result, templateVariables)
