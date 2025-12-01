# oxide.go

[![Go Reference](https://pkg.go.dev/badge/github.com/oxidecomputer/oxide.go.svg)](https://pkg.go.dev/github.com/oxidecomputer/oxide.go)

_**IMPORTANT:** This SDK is under heavy development and will have constant breaking changes._

The Go [API](https://docs.oxide.computer) client for administrating an Oxide rack.

To contribute to this repository make sure you read the contributing [documentation](./CONTRIBUTING.md).

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
	cfg := oxide.Config{
		Host:  "https://api.oxide.computer",
		Token: "oxide-abc123",
	}
	client, err := oxide.NewClient(&cfg)
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

1. Explicit configuration: Set `Host` and `Token` in the `Config`:

   ```go
   cfg := oxide.Config{
       Host:  "https://api.oxide.computer",
       Token: "oxide-abc123",
   }
   ```

1. Environment variables: Set `OXIDE_HOST` and `OXIDE_TOKEN`:

   ```bash
   export OXIDE_HOST="https://api.oxide.computer"
   export OXIDE_TOKEN="oxide-abc123"
   ```

1. Oxide profile: Use a profile from the Oxide config file:
   - Set `Profile` in the `Config`:
     ```go
     cfg := oxide.Config{
         Profile: "my-profile",
     }
     ```
   - Or set the `OXIDE_PROFILE` environment variable:
     ```bash
     export OXIDE_PROFILE="my-profile"
     ```

1. Default profile: Use the default profile from the Oxide config file:
   ```go
   cfg := oxide.Config{
       UseDefaultProfile: true,
   }
   ```

When using profiles, the client reads from the Oxide CLI configuration files located at `$HOME/.config/oxide/credentials.toml` (or a custom directory via `Config.ConfigDir`).
Values defined in `Config` have higher precedence and override environment variables. Configuring both profile and host/token authentication is disallowed and will return an error from oxide.NewClient, as well configuring both a profile and the default profile.
