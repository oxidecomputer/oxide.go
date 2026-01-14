# oxide.go

[![Go Reference](https://pkg.go.dev/badge/github.com/oxidecomputer/oxide.go.svg)](https://pkg.go.dev/github.com/oxidecomputer/oxide.go)

_**IMPORTANT:** This SDK is under heavy development and will have constant breaking changes._

The Go [API](https://docs.oxide.computer) client for administrating an Oxide rack.

To contribute to this repository make sure you read the contributing
[documentation](./CONTRIBUTING.md).

## Getting started

Make sure you have installed [Go](https://go.dev/dl/) 1.21.x or above.

### Installation

Use `go get` inside your module dependencies directory

```console
go get github.com/oxidecomputer/oxide.go@latest
```

### Usage example

```Go
package main

import (
	"fmt"

	"github.com/oxidecomputer/oxide.go/oxide"
)

func main() {
	client, err := oxide.NewClient(
		oxide.WithHost("https://api.oxide.computer"),
		oxide.WithToken("oxide-abc123"),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	params := oxide.ProjectCreateParams{
		Body: &oxide.ProjectCreate{
			Description: "A sample project",
			Name:        oxide.Name("my-project"),
		},
	}

	resp, err := client.ProjectCreate(ctx, params)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", resp)
}
```

### Authentication

The client supports several authentication methods.

1. Explicit options: Use `WithHost` and `WithToken`:

   ```go
   client, err := oxide.NewClient(
       oxide.WithHost("https://api.oxide.computer"),
       oxide.WithToken("oxide-abc123"),
   )
   ```

1. Environment variables: Set `OXIDE_HOST` and `OXIDE_TOKEN`:

   ```bash
   export OXIDE_HOST="https://api.oxide.computer"
   export OXIDE_TOKEN="oxide-abc123"
   ```

   Then create the client with no options:

   ```go
   client, err := oxide.NewClient()
   ```

1. Oxide profile: Use a profile from the Oxide config file:

   ```go
   client, err := oxide.NewClient(oxide.WithProfile("my-profile"))
   ```

   Or set the `OXIDE_PROFILE` environment variable:

   ```bash
   export OXIDE_PROFILE="my-profile"
   ```

1. Default profile: Use the default profile from the Oxide config file:

   ```go
   client, err := oxide.NewClient(oxide.WithDefaultProfile())
   ```

When using profiles, the client reads from the Oxide credentials file located at
`$HOME/.config/oxide/credentials.toml`, or a custom directory via `WithConfigDir`.

Options override environment variables. Configuring both profile and host/token options is
disallowed and will return an error, as will configuring both `WithProfile` and
`WithDefaultProfile`.
