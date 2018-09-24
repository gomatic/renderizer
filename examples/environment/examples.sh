#!/usr/bin/env bash
name=$(basename ${PWD})
renderizer
renderizer ${name}.txt.tmpl
# Note the need for -C otherwise PWD becomes Pwd
renderizer ${name}.txt.tmpl --environment=elsewhere -C --env.PWD=${PWD}
