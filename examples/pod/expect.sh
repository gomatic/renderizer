#!/usr/bin/env bash
name=$(basename ${PWD})
renderizer --testing
renderizer ${name}.yaml.tmpl --testing
