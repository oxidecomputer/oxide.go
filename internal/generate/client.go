package main

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// TODO: Is is necessary to generate this?
// Generate the client.go file.
func generateClient(doc *openapi3.T) error {
	f, err := openGeneratedFile("../../oxide/client.go")
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, `// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.oxide.computer for example.
	server string

	// Client is the *http.Client for performing requests.
	client *http.Client

	// token is the API token used for authentication.
	token string
}`)

	return nil
}
