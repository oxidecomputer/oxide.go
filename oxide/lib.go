package oxide

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

//go:generate go run ../internal/generate/generate.go

// TokenEnvVar is the environment variable that contains the token.
const TokenEnvVar = "OXIDE_TOKEN"

// HostEnvVar is the environment variable that contains the host.
const HostEnvVar = "OXIDE_HOST"

// NewClient creates a new client for the Oxide API.
// You need to pass in your API token to create the client.
func NewClient(token, userAgent, host string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("you need to pass in an API token to create the client")
	}

	baseURL, err := parseBaseURL(host)
	if err != nil {
		return nil, err
	}

	client := &Client{
		server: baseURL,
		token:  token,
	}

	// Ensure the server URL always has a trailing slash.
	if !strings.HasSuffix(client.server, "/") {
		client.server += "/"
	}

	uat := userAgentTransport{
		base:      http.DefaultTransport,
		userAgent: userAgent,
		client:    client,
	}

	client.client = &http.Client{
		Transport: uat,
		// We want a longer timeout since some of the files might take a bit.
		Timeout: 600 * time.Second,
	}

	return client, nil
}

// NewClientFromEnv creates a new client for the Oxide API, using the token
// stored in the environment variable `OXIDE_TOKEN` and the host stored in the
// environment variable `OXIDE_HOST`.
func NewClientFromEnv(userAgent string) (*Client, error) {
	token := os.Getenv(TokenEnvVar)
	if token == "" {
		return nil, fmt.Errorf("the environment variable %s must be set with your API token", TokenEnvVar)
	}

	host := os.Getenv(HostEnvVar)
	if host == "" {
		return nil, fmt.Errorf("the environment variable %s must be set with the host of the Oxide API", HostEnvVar)
	}

	return NewClient(token, userAgent, host)
}

// parseBaseURL parses the base URL from the server URL.
func parseBaseURL(baseURL string) (string, error) {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		// Assume https.
		baseURL = "https://" + baseURL
	}

	newBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	b := newBaseURL.String()

	// Ensure the server URL always has a trailing slash.
	if !strings.HasSuffix(b, "/") {
		b += "/"
	}

	return b, nil
}

// WithToken overrides the token used for authentication.
func (c *Client) WithToken(token string) {
	c.token = token
}

type userAgentTransport struct {
	userAgent string
	base      http.RoundTripper
	client    *Client
}

func (t userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.base == nil {
		return nil, errors.New("RoundTrip: no Transport specified")
	}

	newReq := *req
	newReq.Header = make(http.Header)
	for k, vv := range req.Header {
		newReq.Header[k] = vv
	}

	// Add the user agent header.
	newReq.Header["User-Agent"] = []string{t.userAgent}

	// Add the content-type header.
	newReq.Header["Content-Type"] = []string{"application/json"}

	// Add the authorization header.
	newReq.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", t.client.token)}

	return t.base.RoundTrip(&newReq)
}
