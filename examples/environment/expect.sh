#!/usr/bin/env bash
name=$(basename "${PWD}")
export ONLY_ENVIRONMENT=environment
export CAN_OVERRIDE=environment
renderizer
renderizer --environment=elsewhere -C --env.CAN_OVERRIDE=command-line --env.ENVIRONMENT=command-line
renderizer ${name}.txt.tmpl
renderizer ${name}.txt.tmpl --environment=elsewhere -C --env.CAN_OVERRIDE=command-line --env.ENVIRONMENT=command-line
