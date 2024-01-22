// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// TokenEnvVar is the environment variable that contains the token.
const TokenEnvVar = "OXIDE_TOKEN"

// HostEnvVar is the environment variable that contains the host.
const HostEnvVar = "OXIDE_HOST"

// Config is the configuration that can be set on a Client.
type Config struct {
	// Base URL of the Oxide API including the scheme. For example,
	// https://api.oxide.computer.
	Host string

	// Oxide API authentication token.
	Token string

	// A custom HTTP client to use for the Client instead of the default.
	HTTPClient *http.Client

	// A custom user agent string to add to every request instead of the
	// default.
	UserAgent string
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// Base URL of the Oxide API including the scheme. For example,
	// https://api.oxide.computer.
	server string

	// Oxide API authentication token.
	token string

	// HTTP client to make API requests.
	client *http.Client

	// The user agent string to add to every API request.
	userAgent string
}

// NewClient creates a new client for the Oxide API. To authenticate with
// environment variables, set OXIDE_HOST and OXIDE_TOKEN accordingly. Pass in a
// non-nil *Config to set the various configuration options on a Client. When
// setting the host and token through the *Config, these will override any set
// environment variables.
func NewClient(cfg *Config) (*Client, error) {
	token := os.Getenv(TokenEnvVar)
	server := os.Getenv(HostEnvVar)
	userAgent := defaultUserAgent()
	httpClient := &http.Client{
		Timeout: 600 * time.Second,
	}

	// Layer in the user-provided configuration if provided.
	if cfg != nil {
		if cfg.Host != "" {
			server = cfg.Host
		}

		if cfg.Token != "" {
			token = cfg.Token
		}

		if cfg.UserAgent != "" {
			userAgent = cfg.UserAgent
		}

		if cfg.HTTPClient != nil {
			httpClient = cfg.HTTPClient
		}
	}

	var errServer error
	server, err := parseBaseURL(server)
	if err != nil {
		errServer = fmt.Errorf("failed parsing host address: %w", err)
	}

	var errToken error
	if token == "" {
		errToken = errors.New("token is required")
	}

	// To aggregate the validation errors above.
	if err := errors.Join(errServer, errToken); err != nil {
		return nil, fmt.Errorf("invalid client configuration:\n%w", err)
	}

	client := &Client{
		token:     token,
		server:    server,
		userAgent: userAgent,
		client:    httpClient,
	}

	return client, nil
}

// defaultUserAgent builds and returns the default user agent string.
func defaultUserAgent() string {
	return fmt.Sprintf("oxide.go/%s", version)
}

// parseBaseURL parses the base URL from the server URL.
func parseBaseURL(baseURL string) (string, error) {
	if baseURL == "" {
		return "", errors.New("host address is empty")
	}

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

// buildRequest creates an HTTP request to interact with the Oxide API.
func (c *Client) buildRequest(ctx context.Context, body io.Reader, method, uri string, params, queries map[string]string) (*http.Request, error) {
	// Create the request.
	req, err := http.NewRequestWithContext(ctx, method, uri, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	// Add the parameters to the url.
	if err := expandURL(req.URL, params); err != nil {
		return nil, fmt.Errorf("expanding URL with parameters failed: %v", err)
	}

	// Add queries if any
	addQueries(req.URL, queries)

	return req, nil
}
