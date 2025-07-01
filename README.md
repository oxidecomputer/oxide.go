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
