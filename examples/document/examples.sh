#!/usr/bin/env bash
name=$(basename ${PWD})
renderizer
renderizer ${name}.html.tmpl --settings=.${name}.yaml
renderizer ${name}.html.tmpl --items=apple --items=banana --items=cherry --foo=true
renderizer ${name}.html.tmpl --items=apple --items=banana --items=cherry --foo=true --settings=.${name}.yaml
renderizer ${name}.html.tmpl --settings .${name}.yaml
renderizer ${name}.html.tmpl --items=apple --items=banana --items=cherry --foo=true --settings .${name}.yaml
