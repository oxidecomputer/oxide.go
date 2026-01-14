# Fetch OpenAPI Specification

## Description

Use this skill when you need to look up details about the Oxide API, including:
- Finding endpoint definitions, paths, or HTTP methods
- Looking up schema/type definitions from the API spec
- Understanding request/response formats for API operations
- Investigating API changes or additions
- Comparing generated code against the source OpenAPI spec

## Instructions

1. Read the `VERSION_OMICRON` file to get the Omicron commit hash

2. Check if a cached spec exists for this version:

   ```bash
   test -f /tmp/oxide-openapi-spec-{commit}.json && echo "cached"
   ```

   If the file exists, skip to step 5 - the spec is already cached. Report that the cached spec is
   being used.

3. Use curl to get the symlink target (the versioned filename):

   ```bash
   curl -sL "https://raw.githubusercontent.com/oxidecomputer/omicron/{commit}/openapi/nexus/nexus-latest.json"
   ```

   The response body contains the versioned filename (e.g., `nexus-2026010800.0.0-1844ae.json`)

4. Download the actual OpenAPI spec:

   ```bash
   curl -sL "https://raw.githubusercontent.com/oxidecomputer/omicron/{commit}/openapi/nexus/{versioned-filename}" -o /tmp/oxide-openapi-spec-{commit}.json
   ```

5. Report what you found:
   - The Omicron commit version
   - Whether the spec was fetched fresh or loaded from cache
   - Confirm the spec is available at `/tmp/oxide-openapi-spec-{commit}.json`

6. The spec is now available at `/tmp/oxide-openapi-spec-{commit}.json` for further analysis

## Optional Arguments

If the user provides arguments like:

- `--endpoint <name>` or `endpoint <name>`: Search for and display details about a specific endpoint
- `--schema <name>` or `schema <name>` or `type <name>`: Search for and display a specific
  schema/type definition
- `--search <term>`: Search endpoints and schemas for the given term

Use `jq` or read the temp file to fulfill these queries.

## Output Format

- **Default to YAML** when displaying schemas or endpoint details (use `yq` or `jq ... | yq -P`)
- Only use JSON or table format if the user explicitly requests it (e.g., `--json` or `--table`)

## Notes

- The spec is large (~2MB JSON), so summarize rather than output the entire thing
- If fetching fails, suggest checking if VERSION_OMICRON contains a valid commit hash
- Reference `internal/generate/main.go:getOpenAPISpecURL` for the canonical implementation
