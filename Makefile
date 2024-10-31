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
generate: tools
	@ echo "+ Generating SDK..."
	@ go generate ./...
	@ echo "+ Updating imports..."
	@ $(GOBIN)/goimports -w oxide/*.go
	@ echo "+ Formatting generated SDK..."
	@ gofmt -s -w oxide/*.go
	@ echo "+ Tidying up modules..."
	@ go mod tidy

.PHONY: build
build: $(NAME) ## Builds a dynamic package.

$(NAME): $(wildcard *.go) $(wildcard */*.go)
	@echo "+ $@"
	$(GO) build -tags "$(BUILDTAGS)" ${GO_LDFLAGS} -o $(NAME) .

all: generate test fmt lint staticcheck vet ## Runs a fmt, lint, test, staticcheck, and vet.

.PHONY: fmt
fmt: ## Verifies all files have been `gofmt`ed.
	@ echo "+ Verifying all files have been gofmt-ed..."
	@if [[ ! -z "$(shell gofmt -s -d . | grep -v -e internal/generate/test_generated -e internal/generate/test_utils | tee /dev/stderr)" ]]; then \
		exit 1; \
	fi

.PHONY: lint
lint: tools ## Verifies `golangci-lint` passes.
	@ echo "+ Running Go linters..."
	@ $(GOBIN)/golangci-lint run -E gofmt

.PHONY: test
test: ## Runs the go tests.
	@ echo "+ Running Go tests..."
	@ $(GO) test -v -tags "$(BUILDTAGS)" ./...

.PHONY: vet
vet: ## Verifies `go vet` passes.
	@ echo "+ Verifying go vet passes..."
	@if [[ ! -z "$(shell $(GO) vet ./... | tee /dev/stderr)" ]]; then \
		exit 1; \
	fi

.PHONY: staticcheck
staticcheck: tools ## Verifies `staticcheck` passes.
	@ echo "+ Verifying staticcheck passes..."
	@if [[ ! -z "$(shell $(GOBIN)/staticcheck ./... | tee /dev/stderr)" ]]; then \
		exit 1; \
	fi

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

# The following installs the necessary tools within the local /bin directory.
# This way linting tools don't need to be downloaded/installed every time you
# want to run the linters or generate the SDK.
VERSION_DIR:=$(GOBIN)/versions
VERSION_GOIMPORTS:=v0.24.0
VERSION_GOLANGCILINT:=v1.61.0
VERSION_STATICCHECK:=2024.1.1
VERSION_WHATSIT:=1f5eb3ea

tools: $(GOBIN)/golangci-lint $(GOBIN)/goimports $(GOBIN)/staticcheck

tools-private: $(GOBIN)/whatsit

$(GOBIN):
	@ mkdir -p $(GOBIN)

$(VERSION_DIR): | $(GOBIN)
	@ mkdir -p $(GOBIN)/versions

$(VERSION_DIR)/.version-golangci-lint-$(VERSION_GOLANGCILINT): | $(VERSION_DIR)
	@ rm -f $(VERSION_DIR)/.version-golangci-lint-*
	@ echo $(VERSION_GOLANGCILINT) > $(VERSION_DIR)/.version-golangci-lint-$(VERSION_GOLANGCILINT)

$(GOBIN)/golangci-lint: $(VERSION_DIR)/.version-golangci-lint-$(VERSION_GOLANGCILINT) | $(GOBIN)
	@ echo "-> Installing golangci-lint..."
	@ curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(GOBIN) $(VERSION_GOLANGCILINT)

$(VERSION_DIR)/.version-goimports-$(VERSION_GOIMPORTS): | $(VERSION_DIR)
	@ rm -f $(VERSION_DIR)/.version-goimports-*
	@ echo $(VERSION_GOIMPORTS) > $(VERSION_DIR)/.version-goimports-$(VERSION_GOIMPORTS)

$(GOBIN)/goimports: $(VERSION_DIR)/.version-goimports-$(VERSION_GOIMPORTS) | $(GOBIN)
	@ echo "-> Installing goimports..."
	@ go install golang.org/x/tools/cmd/goimports@$(VERSION_GOIMPORTS)

$(VERSION_DIR)/.version-staticcheck-$(VERSION_STATICCHECK): | $(VERSION_DIR)
	@ rm -f $(VERSION_DIR)/.version-staticcheck-*
	@ echo $(VERSION_STATICCHECK) > $(VERSION_DIR)/.version-staticcheck-$(VERSION_STATICCHECK)

$(GOBIN)/staticcheck: $(VERSION_DIR)/.version-staticcheck-$(VERSION_STATICCHECK) | $(GOBIN)
	@ echo "-> Installing staticcheck..."
	@ go install honnef.co/go/tools/cmd/staticcheck@$(VERSION_STATICCHECK)

$(VERSION_DIR)/.version-whatsit-$(VERSION_WHATSIT): | $(VERSION_DIR)
	@ rm -f $(VERSION_DIR)/.version-whatsit-*
	@ echo $(VERSION_WHATSIT) > $(VERSION_DIR)/.version-whatsit-$(VERSION_WHATSIT)

# TODO: actually release a version of whatsit to use the tag flag
$(GOBIN)/whatsit: $(VERSION_DIR)/.version-whatsit-$(VERSION_WHATSIT) | $(GOBIN)
	@ echo "-> Installing whatsit..."
	@ cargo install --git ssh://git@github.com/oxidecomputer/whatsit.git --rev $(VERSION_WHATSIT) --branch main --root ./ 
