export APP_NAME := $(notdir $(shell pwd))
DESC :=
PROJECT_URL := "https://github.com/gomatic/$(APP_NAME)"

ROOT := $(shell git rev-parse --show-toplevel)
include $(ROOT)/.env.mk
include $(ROOT)/.version.mk

#

PREFIX ?= usr/local

#

.PHONY : all release snapshot vet test examples tag
.PHONY : help
.DEFAULT_GOAL := snapshot

#

ALL_SOURCES = $(filter-out vendor/%, $(wildcard *.go */*.go */*/*.go))
MAIN_SOURCES = $(filter-out %_test.go, $(ALL_SOURCES))
TEST_SOURCES = $(filter %_test.go, $(ALL_SOURCES))

#

build snapshot: vet test $(BINARY) # Build snapshot

all: release ## Make everything

$(BINARY): $(MAIN_SOURCES)
	goreleaser --snapshot --rm-dist --skip-publish --skip-validate

release: vet test ## Build releases
	$(MAKE) full-release BUILD=

full-release: tag
	goreleaser --rm-dist --skip-publish

tag:
	git tag -f -m $(TAG) $(TAG)

vet test: ## Run tests or vet
	go $@ ./...

tests examples: snapshot
	@scripts/test-iterate test examples

clean:
	rm -rf build

help: ## This help.
	@echo $(APP_NAME)
	@echo MAIN_SOURCES=$(MAIN_SOURCES)
	@echo TEST_SOURCES=$(TEST_SOURCES)
	@echo BUILD=$(BUILD)
	@echo FULL_VERSION=$(FULL_VERSION)
	@echo ROOT=$(ROOT)
	@echo GOOS=$(GOOS)
	@echo GOARCH=$(GOARCH)
	@echo DIST=$(DIST)
	@echo BINARY=$(BINARY)
	@echo VERSION=$(VERSION)
	@echo Targets:
	@awk 'BEGIN {FS = ":.*?## "} / [#][#] / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
