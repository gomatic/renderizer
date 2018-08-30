#!/bin/sh
# Example of passing classic shell list to renderizer
# Appropriate for classic shell scripts that use space-delimited list variables

list2yml() {
  local name
  name=$1
  shift
  echo "$name:"
  for item
  do
     echo "- $item"
  done
}

ITEMS="apple banana cherry"

val2yml() {
  echo "$1: $2"
}

list2yml items $ITEMS > sample.yml
val2yml foo true >> sample.yml

# Use -m=default so we can check whether a variable is set
renderizer -m=default -S=sample.yml sample.html.tpl
rm sample.yml
