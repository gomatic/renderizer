export APP_NAME := $(notdir $(shell pwd))
DESC := 
PROJECT_URL := "https://github.com/gomatic/$(APP_NAME)"

SOURCES = $(wildcard *.go)

.PHONY : $(SOURCES)
.PHONY : all release build vet test
.PHONY : help
.DEFAULT_GOAL := ci

PREFIX ?= usr/local

ci: env vet test

all: release vet test ## Make everything

build: # Build darwin
	goreleaser --config .goreleaser-darwin.yml --debug --rm-dist --skip-publish --skip-validate

release: ## Build releases
	goreleaser --rm-dist --skip-publish --skip-validate
	scripts/rename-release

env:
	go env

vet test: ## Run tests or vet
	go $@ -mod=vendor ./...


help: ## This help.
	@echo $(APP_NAME)
	@echo SOURCES=$(SOURCES)
	@echo Targets:
	@awk 'BEGIN {FS = ":.*?## "} / [#][#] / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
