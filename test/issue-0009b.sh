#!/bin/sh

parameters=(
	# issue 9
	"--a.b=1 --a.b=2 --a.b=3"
	"--a.b=A --a.b=B --a.b=C"
	"--a.b=1 --a.b=A --a.b=3"
	"--a.b=A --a.b=2 --a.b=3"
	"--a.b.c=kills:a.b --a.b=A --a.b=B --a.b=C"
)
