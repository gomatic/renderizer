.PHONY : test example examples

ROOT := $(shell git rev-parse --show-toplevel)
include $(ROOT)/.env.mk
export PATH := $(DIST):$(PATH)

test example examples:
	@./expect.sh
