# oxide.go

[![Go Reference](https://pkg.go.dev/badge/github.com/oxidecomputer/oxide.go.svg)](https://pkg.go.dev/github.com/oxidecomputer/oxide.go)

_**IMPORTANT:** This SDK is under heavy development and will have constant breaking changes._

The Go [API](https://docs.oxide.computer) client for administrating an Oxide rack.

To contribute to this repository make sure you read the contributing [documentation](./CONTRIBUTING.md).

## Getting started

Make sure you have installed [Go](https://go.dev/dl/) 1.17 or above.

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
   client, err := oxide.NewClient("<auth token>", "<user-agent>", "<host>")
	if err != nil {
		panic(err)
	}

	resp, err := client.OrganizationCreate(
		&oxide.OrganizationCreate{
			Description: "sample org",
			Name:        oxide.Name("sre"),
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", resp)
}
```
