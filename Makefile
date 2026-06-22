# gomatic/build — shared Go toolchain (canonical source)
#
# Single source of truth for how every gomatic Go library and service runs
# vet / lint / staticcheck / tests / formatting and builds binaries.
#
# This Makefile and its scripts/ are COPIED VERBATIM into every gomatic Go repo
# and live in-tree there — each repo owns its own copy. The gomatic repos are
# spread across many orgs, so the shared-checkout indirection used elsewhere (a
# sibling gomatic/build clone with BUILD_HOME pointed at it, or a `make -f
# ${BUILD_HOME}/Makefile` alias) is deliberately NOT used here. A repo just runs
# its own in-tree copy directly:
#
#   make lint test build           # from inside any repo
#
# gomatic/build is the UPSTREAM that every copy tracks; refresh a repo's copy
# from canonical with `make build-self-update` (scripts/self-update.sh). The only
# things a copy edits inline are the per-repo knobs below (COVER_PKGS, BINARIES).
#
# CI (.github/workflows of each repo) checks out only that repo and runs the
# in-tree Makefile directly — no gomatic/build checkout, no BUILD_HOME override:
#   - uses: actions/checkout@v5
#   - uses: actions/setup-go@v5
#     with: { go-version: '1.26', check-latest: true, cache: true }
#   - run: make ci
#
# Everything else is derived from the repo's own source of truth:
#   BINARIES   <- the `id:` values under `builds:` in ./.goreleaser.yml
#   SUBMODULES <- nested go.mod dirs (excluding vendor/testdata/fixtures)
# Override either on the command line for the rare repo that needs to.
#
# The tools are NOT baked here: every repo pins its own toolchain in its go.mod
# `tool (...)` stanza, and this Makefile runs each one with `go tool <name>` from
# $(CURDIR). The version is whatever that repo pinned — built and cached by the
# go command on first use. In the Docker image the build cache is pre-warmed by
# `make tools`, so CI never compiles a tool; locally `go tool` builds on first
# use and reuses the cache thereafter.

# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.DEFAULT_GOAL := test

.PHONY: help
help: ## This help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m[ target... ]\033[0m\n"} /^[a-zA-Z_0-9@\/ -]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(sort $(MAKEFILE_LIST))

# --------------------------------------------------------------------------- #
# gomatic/build location + tools
# --------------------------------------------------------------------------- #
# This Makefile's own directory, regardless of the consumer's $(CURDIR). Used to
# locate the shared configs (golangci.complexity.yaml) and scripts/.
BUILD_HOME ?= $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

# Tools run via `go tool` against the CONSUMER's own go.mod `tool (...)` stanza:
# every consumer repo pins its toolchain there, so the version is whatever that
# repo pinned — built and cached by the go command on first use. Nothing is
# baked into a bin/ and nothing is put on PATH; the consumer's tool stanza is
# the single source of tool versions. (In the Docker image the build cache is
# pre-warmed by `make tools`, so CI never compiles a tool from scratch.)
GO ?= go

GOLANGCI_LINT := $(GO) tool golangci-lint
STATICCHECK   := $(GO) tool staticcheck
GOFUMPT       := $(GO) tool gofumpt
GOTESTSUM     := $(GO) tool gotestsum
GOVULNCHECK   := $(GO) tool govulncheck
GORELEASER    := $(GO) tool goreleaser

# The CANONICAL tool set — identical in every gomatic repo, no exceptions. This
# is the single source of truth for which tools a repo installs; `make tools`
# brings any repo's go.mod `tool (...)` stanza into compliance with it. Each
# tool is named twice: TOOL_PKGS (the `go get -tool` import path) and TOOL_NAMES
# (the `go tool <name>` command). Repo-specific code GENERATORS (buf,
# protoc-gen-*) are layered on top by the repos that generate code — the quality
# toolchain below is universal.
TOOL_PKGS := \
	github.com/golangci/golangci-lint/v2/cmd/golangci-lint \
	github.com/goreleaser/goreleaser/v2 \
	github.com/uudashr/gocognit/cmd/gocognit \
	golang.org/x/tools/cmd/goimports \
	golang.org/x/vuln/cmd/govulncheck \
	gotest.tools/gotestsum \
	honnef.co/go/tools/cmd/staticcheck \
	mvdan.cc/gofumpt
TOOL_NAMES := golangci-lint goreleaser gocognit goimports govulncheck gotestsum staticcheck gofumpt

# Install the canonical set into THIS module's go.mod and build each so the go
# cache is warm (the Docker image bakes that layer). Idempotent: re-running it on
# a compliant repo is a no-op. `go get -tool` adds any missing tool; `go mod
# tidy` (and `go mod vendor` where the repo vendors) reconcile the graph.
.PHONY: tools
tools: ## Install the canonical tool set (uniform across every repo)
	$(GO) get -tool $(TOOL_PKGS)
	$(GO) mod tidy
	@test -d vendor && $(GO) mod vendor || true
	@$(foreach n,$(TOOL_NAMES),$(GO) tool $(n) --help >/dev/null 2>&1 || true;)

.PHONY: tools-version
tools-version: ## Print the version of every pinned tool
	@$(GOLANGCI_LINT) version
	@$(STATICCHECK) -version
	@$(GOFUMPT) --version
	@$(GOTESTSUM) --version
	@$(GORELEASER) --version | grep -i gitversion

# --------------------------------------------------------------------------- #
# Consumer configuration (derived; override on the command line if ever needed)
# --------------------------------------------------------------------------- #
# BINARIES: the `id:` values under `builds:` in the consumer's goreleaser config.
# Derived with yq (not awk over raw YAML, which mis-parsed a templated id like
# `id: "{{ .ProjectName }}"` into `--id "{{` and broke the build). The config is
# either .goreleaser.yaml or .goreleaser.yml; an id that is itself a
# `{{ .ProjectName }}` template — or absent — resolves to the project_name value,
# while literal ids pass through unchanged.
# The yq pipeline: bind project_name as $pn, take each build id (falling back to
# $pn when a build sets none), and replace a `{{ ... }}` template id with $pn —
# so a literal id passes through and a `{{ .ProjectName }}` id resolves to the
# project name. One yq call, no shell loop.
GORELEASER_CONFIG ?= $(firstword $(wildcard .goreleaser.yaml .goreleaser.yml))
BINARIES ?= $(shell test -n '$(GORELEASER_CONFIG)' && yq '.project_name as $$pn | .builds[] | select(.skip != true) | .id // $$pn | sub("\{\{.*\}\}", $$pn)' '$(GORELEASER_CONFIG)' 2>/dev/null)

# SUBMODULES: nested modules (own go.mod), excluding vendored/test-fixture mods
# and Terraform-downloaded module sources under .terraform/.
SUBMODULES ?= $(patsubst ./%/,%,$(dir $(shell find . -mindepth 2 -name go.mod -not -path '*/vendor/*' -not -path '*/testdata/*' -not -path '*/fixtures/*' -not -path '*/.terraform/*' -not -path '*/tools/*')))

BUILD_DIR ?= bin
GOOS      ?= $(shell go env GOOS)
GOARCH    ?= $(shell go env GOARCH)

COVERAGE_FOLDER ?= var
GO_TEST_FORMAT  ?= standard-verbose

# Coverage gate: COVER_PKGS is the set of packages whose AGGREGATE statement
# coverage must reach COVER_THRESHOLD (default: every package at 100%). A
# consumer narrows COVER_PKGS — BEFORE the include — to drop composition roots
# (cmd/*, whose os.Exit/signal wiring is exercised by integration tests) and
# committed generated trees (src/proto, src/grammar, internal/gen) whose lines
# are machine-authored, e.g.:
#   COVER_PKGS = $(shell go list ./... | grep -v /cmd/ | grep -v /src/proto)
# COVERPKG is the comma-joined form `go test -coverpkg` needs (so coverage is
# attributed across the same set, not just the package under test).
comma := ,
empty :=
space := $(empty) $(empty)
# cmd/* is the composition root (os.Exit/signal/flag wiring), exercised by the
# end-to-end tests but excluded from the 100% statement-coverage gate.
COVER_PKGS      ?= $(shell go list ./... | grep -v /cmd/)
COVER_MODE      ?= atomic
COVER_THRESHOLD ?= 100.0%
COVERPKG        ?= $(subst $(space),$(comma),$(strip $(COVER_PKGS)))

export BUILD_NUMBER           ?= 9999999
export GORELEASER_BUILD_FLAGS ?=

$(BUILD_DIR) $(COVERAGE_FOLDER):
	mkdir -p $@

##@ CI

# `ci` is the HARD CI gate — the single source of truth for what every push must
# pass. It is a SUPERSET of `check` (the developer gate): the same static,
# vulnerability and 100%-`cover` gates, plus the race detector (`test-all`) and
# cross-platform compilation (`build-all`) that `check` skips for local speed.
# `vulncheck` and `cover` were once held out of CI during a soft rollout;
# consumers are green, so they are enforced on every push now — coverage and
# vulnerabilities can no longer silently regress in CI.
.PHONY: ci
ci: lint staticcheck vulncheck cover test-all build-all ## Aggregate target for CI builds

# True CI parity: run the real `ci` recipe INSIDE the baked toolchain image,
# so it uses the pinned tools and the exact base environment CI runs in — not the
# host's Go. The consumer checkout is bind-mounted and its own Makefile drives
# `ci` (BUILD_HOME is baked into the image, so the include resolves, and any
# consumer ci extensions run too). Everything is vendored, so no module cache
# mount is needed. -it only when attached to a TTY, so it works in scripts.
BUILD_IMAGE ?= $(DOCKER_REGISTRY)/build:$(DOCKER_IMAGE_TAG)
DOCKER_TTY  := $(shell test -t 0 && echo -it)
.PHONY: ci-local
ci-local: ## Run the CI aggregate inside the baked image, exactly as CI does
	docker run --rm $(DOCKER_TTY) \
		--volume $(CURDIR):$(CURDIR) \
		--workdir $(CURDIR) \
		$(BUILD_IMAGE) \
		make ci

##@ Code Quality

# `check` is the comprehensive DEVELOPER gate: run it locally before pushing. It
# is the static + `vulncheck` + 100%-`cover` core that `ci` ALSO enforces, so a
# local `check` pass predicts a green CI; `ci` is a superset that additionally
# runs `test-all` (race) and `build-all` (cross-compile). The complexity linters
# are part of `lint` now (folded into .golangci.yaml).
.PHONY: check
check: lint staticcheck vulncheck cover ## Full developer gate (CI runs this + race & cross-compile)

# Per-submodule vet targets via a static pattern rule (NOT `vet-%:` — GNU make
# skips implicit/pattern rules for phony targets; a static pattern with an
# explicit target list fires correctly).
VET_SUBMODULES := $(addprefix vet@,$(SUBMODULES))
.PHONY: vet $(VET_SUBMODULES)
vet: $(VET_SUBMODULES) ## Run go vet (root module + submodules)
	go vet ./...
$(VET_SUBMODULES): vet@%:
	go vet -C $* ./...

.PHONY: lint
lint: vet ## Run golangci-lint (incl. the central complexity linters)
	$(GOLANGCI_LINT) run

.PHONY: staticcheck
staticcheck: ## Run staticcheck
	$(STATICCHECK) ./...

# VULNCHECK_SCAN is the govulncheck precision knob. The default `symbol` builds
# the call graph and reports only vulnerabilities actually reachable from a
# called symbol. A repo whose generics trip the known x/vuln source-mode panic
# (`ForEachElement called on type containing *types.TypeParam`, govulncheck
# <=v1.4.0) sets `VULNCHECK_SCAN = package`: that skips call-graph construction
# (so no panic) and flags any vulnerable *package* that is imported at all —
# strictly more conservative than symbol scanning, never less, so it can't hide
# a finding. Drop back to `symbol` once upstream fixes the panic.
VULNCHECK_SCAN ?= symbol

.PHONY: vulncheck
vulncheck: ## Run govulncheck
	$(GOVULNCHECK) -mode=source -scan=$(VULNCHECK_SCAN) ./...

##@ Test

TEST_SUBMODULES := $(addprefix test@,$(SUBMODULES))
.PHONY: test $(TEST_SUBMODULES)
test: $(TEST_SUBMODULES) ## Run tests (root module + submodules)
	$(GOTESTSUM) --format $(GO_TEST_FORMAT) -- ./...
$(TEST_SUBMODULES): test@%:
	go test -C $* ./...

TESTALL_SUBMODULES := $(addprefix test-all@,$(SUBMODULES))
.PHONY: test-all $(TESTALL_SUBMODULES)
test-all: $(COVERAGE_FOLDER) $(TESTALL_SUBMODULES) ## Run all tests with race detection + coverage
	CGO_ENABLED=1 $(GOTESTSUM) --format $(GO_TEST_FORMAT) -- -race -short -coverprofile=$(COVERAGE_FOLDER)/coverage.out ./...
$(TESTALL_SUBMODULES): test-all@%:
	cd $* && CGO_ENABLED=1 go test -race -short ./...

.PHONY: coverage
coverage: $(COVERAGE_FOLDER) ## Run tests with coverage
	$(GOTESTSUM) --format $(GO_TEST_FORMAT) -- -coverprofile=$(COVERAGE_FOLDER)/coverage.out ./...

# The coverage GATE: run COVER_PKGS and fail unless aggregate statement coverage
# is exactly COVER_THRESHOLD. This is the 100%-coverage enforcement every gomatic
# consumer shares (folded into `check`); a consumer scopes COVER_PKGS to its own
# tested set. Lists the sub-100% functions on failure so the miss is actionable.
.PHONY: cover
cover: $(COVERAGE_FOLDER) ## Run tests and assert COVER_THRESHOLD coverage of COVER_PKGS
	$(GOTESTSUM) --format $(GO_TEST_FORMAT) -- -covermode=$(COVER_MODE) -coverpkg=$(COVERPKG) -coverprofile=$(COVERAGE_FOLDER)/coverage.out $(COVER_PKGS)
	@total=$$(go tool cover -func=$(COVERAGE_FOLDER)/coverage.out | awk '/^total:/{print $$3}'); \
	echo "total coverage: $$total"; \
	[ "$$total" = "$(COVER_THRESHOLD)" ] || { echo "coverage $$total below $(COVER_THRESHOLD):"; go tool cover -func=$(COVERAGE_FOLDER)/coverage.out | awk '$$3 != "100.0%"'; exit 1; }

# Build-tag-gated tests (integration, e2e, ...). The shared `test`/`test-all`
# targets run only UNTAGGED unit tests; anything behind a `//go:build <tag>`
# constraint is invisible to them and runs here instead. gomatic/build owns the
# invocation (gotestsum + the pinned toolchain); the CONSUMER owns whatever
# environment those tests need — a live Postgres, env vars, fixtures — by
# exporting it before calling this (or wrapping it in a consumer target).
#
#   make test-tag TAG=integration
#   make test-tag TAG=integration TEST_TAG_ARGS='-run TestFoo -timeout 10m'
#   make test-tag TAG=e2e         TEST_TAG_PKG=./e2e/...
#
# TEST_TAG_PKG defaults to ./... ; TEST_TAG_ARGS appends any extra `go test`
# flags (-run, -timeout, -parallel, -bench, ...).
TEST_TAG_PKG  ?= ./...
TEST_TAG_ARGS ?=
.PHONY: test-tag
test-tag: ## Run build-tag-gated tests: make test-tag TAG=integration [TEST_TAG_ARGS='-run X']
	@test -n "$(TAG)" || { echo "ERROR: TAG is required, e.g. 'make test-tag TAG=integration'"; exit 1; }
	$(GOTESTSUM) --format $(GO_TEST_FORMAT) -- -tags $(TAG) $(TEST_TAG_ARGS) $(TEST_TAG_PKG)

# Tag-as-target sugar: `make test-tag-e2e` == `make test-tag TAG=e2e`. The stem
# $* is the tag. It delegates to test-tag so the actual invocation stays single-
# sourced; TEST_TAG_ARGS / TEST_TAG_PKG given on the command line propagate to
# the sub-make. Deliberately NOT phony: .PHONY takes no patterns, and GNU make
# skips IMPLICIT pattern rules for phony targets (see the vet note above). The
# recipe creates no file named after the target, so it re-runs like a phony
# target would. (Pattern targets don't appear in `make help`.)
test-tag-%:
	@$(MAKE) test-tag TAG=$*

# `integration` is the near-universal tag, so it gets a named, help-listed
# shortcut rather than every consumer wrapping `test-tag TAG=integration`. The
# consumer still owns the ENVIRONMENT these tests need (live Postgres, env
# vars) — export it before invoking, e.g. PGHOST/PGPORT/... — gomatic/build owns
# only the run.
#
# Scope: run from ./integration/... when that folder exists (the common layout —
# tagged tests collected in their own dir), else the whole module (./...) for
# repos that scatter `//go:build integration` files alongside their packages.
INTEGRATION_PKG := $(if $(wildcard integration),./integration/...,./...)
.PHONY: test-integration
test-integration: ## Run `integration`-tagged tests (needs the consumer's env, e.g. Postgres)
	@$(MAKE) test-tag TAG=integration TEST_TAG_PKG=$(INTEGRATION_PKG)

##@ Build

.PHONY: build
build: $(BINARIES) ## Build all binaries for the current platform

# Build hook. Deliberately a no-op: generated code is committed and formatting
# is enforced by the gate, so `make build` must NOT regenerate or rewrite the
# tree (regeneration is fragile across tool versions and would dirty the working
# tree mid-build). Run `make fmt` / `make generate` explicitly when you want them.
.PHONY: pre-build
pre-build:

# Build one binary via goreleaser single-target snapshot, then drop a stable
# unversioned symlink beside the arch-suffixed artifact.
.PHONY: $(BINARIES)
$(BINARIES): pre-build | $(BUILD_DIR)
	$(GORELEASER) build --single-target --snapshot --clean --id $@
	cp dist/$@-$(GOOS)-$(GOARCH) $(BUILD_DIR)/$@-$(GOOS)-$(GOARCH)
	@rm -f $(BUILD_DIR)/$@
	@ln -sf $@-$(GOOS)-$(GOARCH) $(BUILD_DIR)/$@

.PHONY: build-all
build-all: pre-build ## Build binaries for all platforms
	$(if $(BINARIES),$(GORELEASER) build --snapshot --clean,@echo "no BINARIES (no builds: in .goreleaser.yml) — nothing to build")

.PHONY: release
release: pre-build ## Create a release with goreleaser
	$(GORELEASER) release --clean

.PHONY: release-snapshot
release-snapshot: pre-build ## Create a snapshot release (no git tag required)
	$(GORELEASER) release --snapshot --clean

##@ Docker

# Centralized image build. Everything is derived from sources gomatic/build already
# knows, so a consumer sets nothing in the common case:
#   DOCKER_IMAGE      <- ghcr.io/gomatic/<repo-dir-name>
#   DOCKER_ENTRYPOINT <- the first id under builds: in .goreleaser.yml
# The consumer's Dockerfile is expected to `FROM gomatic/runtime` (built from this
# repo's runtime.dockerfile), which bakes the unprivileged user, certs and the
# distroless base — so the consumer Dockerfile is just COPY + ENTRYPOINT.
DOCKER_REGISTRY   ?= ghcr.io/gomatic
DOCKER_IMAGE      ?= $(DOCKER_REGISTRY)/$(notdir $(CURDIR))
DOCKER_IMAGE_TAG  ?= latest
DOCKER_ENTRYPOINT ?= $(firstword $(BINARIES))
DOCKER_PLATFORM   ?= linux/$(GOARCH)
DOCKERFILE        ?= Dockerfile

# OCI provenance labels stamped onto every image identically. Revision is the CI
# build number; source is the consumer's origin remote. --build-arg
# ENTRYPOINT_BIN lets the shared runtime.dockerfile / consumer Dockerfile pick
# the binary without hardcoding it.
DOCKER_LABELS := \
	--label org.opencontainers.image.revision=$(BUILD_NUMBER) \
	--label org.opencontainers.image.source=$(shell git config --get remote.origin.url 2>/dev/null) \
	--label org.opencontainers.image.title=$(notdir $(CURDIR))

.PHONY: docker
docker: build-all ## Build the docker image (single-arch, for the current platform)
	docker build \
		--platform $(DOCKER_PLATFORM) \
		--build-arg ENTRYPOINT_BIN=$(DOCKER_ENTRYPOINT) \
		$(DOCKER_LABELS) \
		--file $(DOCKERFILE) \
		--tag $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG) .

# Multi-arch manifest list from both arches goreleaser already produced under
# dist/. build-all emits both amd64 and arm64, but a per-consumer single-arch
# `docker build` threw one away; buildx assembles the manifest so the published
# tag serves both.
DOCKER_BUILDX_PLATFORMS ?= linux/amd64,linux/arm64
.PHONY: docker-buildx
docker-buildx: build-all ## Build + push a multi-arch image manifest (needs buildx)
	docker buildx build \
		--platform $(DOCKER_BUILDX_PLATFORMS) \
		--build-arg ENTRYPOINT_BIN=$(DOCKER_ENTRYPOINT) \
		$(DOCKER_LABELS) \
		--file $(DOCKERFILE) \
		--tag $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG) \
		--push .

##@ Utilities

.PHONY: fmt
fmt: ## Format code
	$(GOFUMPT) -l -w .

.PHONY: generate
generate: ## Generate code
	go generate ./...

# Tidy/vendor/verify the root module AND every nested module. Mirrors the
# test@% / vet@% fan-out: a static pattern rule over an explicit target list
# (phony pattern rules don't fire as implicit rules — see the vet note above).
DEPS_SUBMODULES := $(addprefix deps@,$(SUBMODULES))
.PHONY: deps tidy $(DEPS_SUBMODULES)
deps tidy: $(DEPS_SUBMODULES) ## Tidy and verify dependencies (root module + submodules)
	go mod tidy -go=$(shell go mod edit -json | jq -r .Go)
	go mod vendor
	go mod verify
$(DEPS_SUBMODULES): deps@%:
	cd $* && go mod tidy -go=$$(go mod edit -json | jq -r .Go) && go mod vendor && go mod verify

# Update THIS repo (the shared toolchain) to canonical gomatic/build. Operates on
# $(BUILD_HOME), never $(CURDIR): a consumer runs this from inside its own
# checkout to refresh gomatic/build itself. The update logic (find the canonical
# gomatic/build remote, fast-forward, fail loudly if stale) lives in
# scripts/self-update.sh; see there for layout/fork handling. There are no tool
# binaries to rebuild — each consumer's `go tool` resolves its own pinned
# versions on demand.
.PHONY: build-self-update
build-self-update: ## Update gomatic/build from canonical remote
	@$(BUILD_HOME)/scripts/self-update.sh $(BUILD_HOME)

# Fail if the working tree is dirty. Run after fmt/generate/tidy to prove a
# consumer committed everything those targets produce — the typical CI failure
# is a checked-in repo that drifted from `go generate` / `gofumpt` / `go mod
# vendor` output. Any repo can hang this off its `ci` target.
.PHONY: verify
verify: ## Fail if the working tree has uncommitted changes
	@git diff --exit-code || { echo "ERROR: working tree is dirty (run 'make fmt generate tidy' and commit)"; exit 1; }

# Cheap environment sanity check: is the local gomatic/build clone behind canonical
# gomatic/build, and does the consumer's Go version match the toolchain gomatic/build
# pins? Catches the "green in CI, stale locally" class. Read-only and best-effort
# — never fails the build (warnings only), so it's safe to run anywhere. Logic
# lives in scripts/doctor.sh.
.PHONY: doctor
doctor: tools-version ## Diagnose gomatic/build freshness + Go version drift
	@$(BUILD_HOME)/scripts/doctor.sh $(BUILD_HOME)

.PHONY: clean
clean: ## Clean build + test artifacts
	rm -rf dist/
	rm -rf $(foreach b,$(BINARIES),$(BUILD_DIR)/$(b)*)
	rm -rf *.test *.out coverage* $(COVERAGE_FOLDER)/coverage*
