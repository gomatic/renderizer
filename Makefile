export APP_NAME := $(notdir $(shell pwd))
DESC :=
PROJECT_URL := "https://github.com/gomatic/$(APP_NAME)"

ROOT := $(shell git rev-parse --show-toplevel)
include $(ROOT)/.env.mk
include $(ROOT)/.version.mk

#

PREFIX ?= usr/local

#

.DEFAULT_GOAL := snapshot

#

ALL_SOURCES = $(filter-out vendor/%, $(wildcard *.go */*.go */*/*.go))
MAIN_SOURCES = $(filter-out %_test.go, $(ALL_SOURCES))
TEST_SOURCES = $(filter %_test.go, $(ALL_SOURCES))

#

.PHONY: build snapshot
build snapshot: lint test $(BINARY) # Build snapshot

.PHONY: all
all: release ## Make everything

$(BINARY): $(MAIN_SOURCES)
	go tool github.com/goreleaser/goreleaser --snapshot --rm-dist --skip-publish --skip-validate

.PHONY: release
release: lint test ## Build releases
	$(MAKE) full-release BUILD=

.PHONY: full-release
full-release: tag
	go tool github.com/goreleaser/goreleaser --rm-dist --skip-publish

.PHONY: tag
tag:
	git tag -f -m $(TAG) $(TAG)

.PHONY: vet test
vet test: ## Run tests or vet
	go $@ ./...

.PHONY: lint
lint: vet ## Run tests or vet
	go tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run

.PHONY: fmt
fmt: ## Run tests or vet
	go tool golang.org/x/tools/cmd/goimports -w cmd pkg

.PHONY: tests
tests: snapshot examples
	go test -v -race -count=5 -coverprofile=coverage.out ./...

.PHONY: examples
examples:
	@scripts/test-iterate test examples

.PHONY: clean
clean:
	rm -rf build

.PHONY: help
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
