# AGENTS.md

This file provides context for AI coding assistants working on this codebase.

## Project Overview

**oxide.go** is the official Go SDK for administrating an Oxide rack. It's an OpenAPI-generated client library providing programmatic access to the Oxide infrastructure management API.

## Build and Test Commands

```bash
make all          # Run fmt, lint, test, staticcheck, vet
make test         # Run tests
make lint         # Run golangci-lint
make generate     # Regenerate SDK from OpenAPI spec
```

Note: prefer to run make targets like `make generate` over one-off `go generate` or `go run` commands. The make targets include additional steps and context that are needed for correct behavior.

Note: unless otherwise specified, we should update generated code and run unit tests before considering a change complete.

## Key Patterns

### Code Generation

The SDK is generated from the Oxide API's OpenAPI specification:

- Source spec version tracked in `VERSION_OMICRON`
- Generator in `internal/generate/`
- Templates in `internal/generate/templates/`
- **Do not manually edit** `types.go`, `paths.go`, `responses.go`, or `version.go`

### Finding the OpenAPI Specification

To fetch the upstream OpenAPI spec manually:

1. Read `VERSION_OMICRON` to get the Omicron commit hash (e.g., `06c0808`)
2. Fetch `https://raw.githubusercontent.com/oxidecomputer/omicron/{commit}/openapi/nexus/nexus-latest.json`
3. This file is a symlink - the response body contains the versioned filename (e.g., `nexus-2025120300.0.0.json`)
4. Fetch the actual spec at `https://raw.githubusercontent.com/oxidecomputer/omicron/{commit}/openapi/nexus/{versioned-filename}`

See `getOpenAPISpecURL` in `internal/generate/main.go` for the full implementation.

For local testing, set `OPENAPI_SPEC_PATH` environment variable to use a local spec file instead.

**Claude skill:** Use `/fetch-spec` to automatically fetch the spec and save it to `/tmp/oxide-openapi-spec.json`. Supports arguments like `/fetch-spec endpoint disk` or `/fetch-spec schema Instance`.

## Code Style

- Mozilla Public License headers on all files
- Run `make fmt` before committing
- Run `make lint` to check for issues
- Unexported functions/constants by default
- Clear godoc comments on public APIs
