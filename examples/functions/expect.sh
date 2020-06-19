#!/usr/bin/env bash
name=$(basename ${PWD})
renderizer
renderizer ${name}.txt.tmpl
