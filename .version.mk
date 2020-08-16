VERSION_MAJOR := 2
VERSION_MINOR := 0
VERSION_PATCH := 11

ifdef BUILD
BUILD := -$(BUILD)
else
BUILD := -$(shell date +%s)
endif

TAG := v$(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_PATCH)
FULL_VERSION := $(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_PATCH)$(BUILD)
VERSION ?= $(FULL_VERSION)

export COMMIT_TIME := $(shell git log -n1 --format=%ad --date=format:'%Y%m%dT%H' $(COMMIT_HASH))
