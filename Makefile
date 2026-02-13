SHELL = /usr/bin/env bash -eo pipefail
NAME := oxide

# If this session isn't interactive, then we don't want to allocate a
# TTY, which would fail, but if it is interactive, we do want to attach
# so that the user can send e.g. ^C through.
INTERACTIVE := $(shell [ -t 0 ] && echo 1 || echo 0)
ifeq ($(INTERACTIVE), 1)
	DOCKER_FLAGS += -t
endif

# Set our default go compiler
GO := go
export GOBIN = $(shell pwd)/bin
VERSION := $(shell cat $(CURDIR)/VERSION)

.PHONY: generate
generate:
	@ echo "+ Generating SDK..."
	@ go generate ./...
	@ echo "+ Formatting generated SDK..."
	@ $(MAKE) fmt
	@ echo "+ Tidying up modules..."
	@ go mod tidy

.PHONY: build
build: $(NAME) ## Builds a dynamic package. This is to be used for CI purposes only.

$(NAME): $(wildcard *.go) $(wildcard */*.go)
	@echo "+ $@"
	$(GO) build -tags "$(BUILDTAGS)" ${GO_LDFLAGS} -o $(NAME) ./internal/generate/

all: generate test fmt lint ## Runs a fmt, lint, and test.

.PHONY: fmt
fmt: ## Formats Go code including long line wrapping.
	@ echo "+ Formatting Go code..."
	@ go tool -modfile tools/go.mod golangci-lint fmt

.PHONY: fmt-md
fmt-md: ## Formats markdown files with prettier.
	@ echo "+ Formatting markdown files..."
	@ npx prettier --write "**/*.md"

.PHONY: lint
lint: ## Verifies `golangci-lint` passes.
	@ echo "+ Running Go linters..."
	@ go tool -modfile tools/go.mod golangci-lint run

.PHONY: test
test: ## Runs the go tests.
	@ echo "+ Running Go tests..."
	@ $(GO) test -v -tags "$(BUILDTAGS)" ./...

.PHONY: golden-fixtures
golden-fixtures: ## Refreshes golden test fixtures. Requires OXIDE_HOST, OXIDE_TOKEN, and OXIDE_PROJECT.
	@ echo "+ Refreshing golden test fixtures..."
	@ $(GO) run ./oxide/testdata/main.go


.PHONY: tag
tag: ## Create a new git tag to prepare to build a release.
	git tag -sa $(VERSION) -m "$(VERSION)"
	@echo "Run git push origin $(VERSION) to push your new tag to GitHub and trigger a release."

.PHONY: changelog
## Creates a changelog prior to a release
changelog: tools-private
	@ echo "+ Creating changelog..."
	@ $(GOBIN)/whatsit changelog create --repository oxidecomputer/oxide.go --new-version $(VERSION) --config ./.changelog/$(VERSION).toml

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# whatsit is a Rust tool used for changelog generation, installed via cargo.
VERSION_WHATSIT:=053446d

tools-private: $(GOBIN)/whatsit

$(GOBIN):
	@ mkdir -p $(GOBIN)

# TODO: actually release a version of whatsit to use the tag flag
$(GOBIN)/whatsit: | $(GOBIN)
	@ echo "-> Installing whatsit $(VERSION_WHATSIT)..."
	@ cargo install --git ssh://git@github.com/oxidecomputer/whatsit.git --rev $(VERSION_WHATSIT) --branch main --root ./ 
