#!/usr/bin/env bash
name=$(basename ${PWD})
renderizer
renderizer ${name}.txt.tmpl --settings=.${name}.yaml
renderizer ${name}.txt.tmpl --name=Renderizer --items=one --items=two --items=three
renderizer ${name}.txt.tmpl --name=Renderizer --items=one --items=two --items=three --settings=.${name}.yaml
renderizer ${name}.txt.tmpl --settings .${name}.yaml
renderizer ${name}.txt.tmpl --name=Renderizer --items=one --items=two --items=three --settings .${name}.yaml
