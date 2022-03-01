// Code generated by `generate`. DO NOT EDIT.

package oxide

import "net/http"

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.oxide.computer for example.
	server string

	// Client is the *http.Client for performing requests.
	client *http.Client

	// token is the API token used for authentication.
	token string
}
