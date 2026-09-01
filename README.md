# Oxide Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/oxidecomputer/oxide.go.svg)](https://pkg.go.dev/github.com/oxidecomputer/oxide.go)

The Oxide Go SDK enables Go programs to interact with the [Oxide
API](https://docs.oxide.computer).

## Version Policy

This project adheres to [Semantic Versioning](https://semver.org/). It is
currently at major version zero (e.g., v0.Y.Z). Anything may change at any time
and the public API should not be considered stable.

## Minimum Supported Go Version

The Go version specified in [go.mod](./go.mod) is the minimum supported Go
version for this project.

## Usage

Use `go get` to fetch this module as a dependency.

```console
go get github.com/oxidecomputer/oxide.go@latest
```

### Example

```go
package main

import (
	"context"
	"fmt"

	"github.com/oxidecomputer/oxide.go/oxide"
)

func main() {
	client, err := oxide.NewClient(
		oxide.WithHost("https://oxide.sys.example.com"),
		oxide.WithToken("oxide-token-abc123"),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	params := oxide.ProjectCreateParams{
		Body: &oxide.ProjectCreate{
			Name:        oxide.Name("my-project"),
			Description: "A project created by the Go SDK.",
		},
	}

	project, err := client.ProjectCreate(ctx, params)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created project:\n%#v\n", project)
}
```

### Authentication

The Go SDK supports several authentication methods.

1. Explicit options: Use `WithHost` and `WithToken`:

   ```go
   client, err := oxide.NewClient(
       oxide.WithHost("https://oxide.sys.example.com"),
       oxide.WithToken("oxide-token-abc123"),
   )
   ```

1. Environment variables: Set `OXIDE_HOST` and `OXIDE_TOKEN`:

   ```bash
   export OXIDE_HOST="https://oxide.sys.r3.oxide-preview.com"
   export OXIDE_TOKEN="oxide-token-abc123"
   ```

   Then create the client with no options:

   ```go
   client, err := oxide.NewClient()
   ```

1. Oxide profile: Use a profile from the Oxide credentials file:

   ```go
   client, err := oxide.NewClient(oxide.WithProfile("my-profile"))
   ```

   Or set the `OXIDE_PROFILE` environment variable:

   ```bash
   export OXIDE_PROFILE="my-profile"
   ```

   Then create the client with no options:

   ```go
   client, err := oxide.NewClient()
   ```

1. Default profile: Use the default profile from the Oxide credentials file:

   ```go
   client, err := oxide.NewClient(oxide.WithDefaultProfile())
   ```

When using profiles, the client reads from the Oxide credentials file
located at `$HOME/.config/oxide/credentials.toml`, or a custom directory via
`WithConfigDir`:

```go
client, err := oxide.NewClient(
	oxide.WithProfile("my-profile"),
	oxide.WithConfigDir("/path/to/oxide/config"),
)
```

Options override environment variables. Configuring `WithProfile` or
`WithDefaultProfile` together with `WithHost` or `WithToken` returns an error.
Configuring `WithProfile` together with `WithDefaultProfile` also returns an
error.

## Contributing

Read [CONTRIBUTING.md](./CONTRIBUTING.md) before contributing to this project.
