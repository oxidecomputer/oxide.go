# Fetch OpenAPI Specification

Fetch the upstream Oxide API OpenAPI specification from the Omicron repository.

## Instructions

Always fetch fresh to ensure the spec matches the current `VERSION_OMICRON`.

1. Read the `VERSION_OMICRON` file to get the Omicron commit hash

2. Use WebFetch to get the symlink target:
   - URL: `https://raw.githubusercontent.com/oxidecomputer/omicron/{commit}/openapi/nexus/nexus-latest.json`
   - The response body contains the versioned filename (e.g., `nexus-2025120300.0.0.json`)

3. Download the actual OpenAPI spec to a temp file using curl:
   ```bash
   curl -sL "https://raw.githubusercontent.com/oxidecomputer/omicron/{commit}/openapi/nexus/{versioned-filename}" -o /tmp/oxide-openapi-spec.json
   ```

4. Report what you found:
   - The Omicron commit version
   - The versioned spec filename
   - Confirm the spec was saved to `/tmp/oxide-openapi-spec.json`
   - The API version from the spec's `info.version` field
   - A brief summary: count of paths (endpoints) and count of components/schemas

5. The spec is now available at `/tmp/oxide-openapi-spec.json` for further analysis

## Optional Arguments

If the user provides arguments like:
- `--endpoint <name>` or `endpoint <name>`: Search for and display details about a specific endpoint
- `--schema <name>` or `schema <name>` or `type <name>`: Search for and display a specific schema/type definition
- `--search <term>`: Search endpoints and schemas for the given term

Use `jq` or read the temp file to fulfill these queries.

## Notes

- The spec is large (~2MB JSON), so summarize rather than output the entire thing
- If fetching fails, suggest checking if VERSION_OMICRON contains a valid commit hash
- Reference `internal/generate/main.go:getOpenAPISpecURL` for the canonical implementation
