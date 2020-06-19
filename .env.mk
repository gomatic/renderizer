ROOT ?= $(shell git rev-parse --show-toplevel)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
DIST := $(ROOT)/build/dist/renderizer_$(GOOS)_$(GOARCH)
BINARY := $(DIST)/renderizer
