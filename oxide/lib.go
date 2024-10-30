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
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
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

	// The directory to look for Oxide CLI configuration files. Defaults
	// to $HOME/.config/oxide if unset.
	ConfigDir string

	// The name of the Oxide CLI profile to use for authentication.
	// The Host and Token fields will override their respective values
	// provided by the profile.
	Profile string

	// Whether to use the default profile listed in the Oxide CLI
	// config.toml file for authentication. Will be overridden by
	// the Profile field.
	UseDefaultProfile bool
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// Base URL of the Oxide API including the scheme. For example,
	// https://api.oxide.computer.
	host string

	// Oxide API authentication token.
	token string

	// HTTP client to make API requests.
	client *http.Client

	// The user agent string to add to every API request.
	userAgent string
}

type hostCreds struct {
	host  string
	token string
}

// NewClient creates a new client for the Oxide API. To authenticate with
// environment variables, set OXIDE_HOST and OXIDE_TOKEN accordingly. Pass in a
// non-nil *Config to set the various configuration options on a Client. When
// setting the host and token through the *Config, these will override any set
// environment variables.
func NewClient(cfg *Config) (*Client, error) {
	token := os.Getenv(TokenEnvVar)
	host := os.Getenv(HostEnvVar)
	userAgent := defaultUserAgent()
	httpClient := &http.Client{
		Timeout: 600 * time.Second,
	}

	errs := make([]error, 0)

	// Layer in the user-provided configuration if provided.
	if cfg != nil {
		var fileCreds *hostCreds
		if cfg.Profile != "" || cfg.UseDefaultProfile {
			var err error
			fileCreds, err = profileCredentials(*cfg)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to read config: %w", err))
			}
		}

		// Use explicit host over profile.
		if cfg.Host != "" {
			host = cfg.Host
		} else if fileCreds != nil {
			host = fileCreds.host
		}

		// Use explicit token over profile.
		if cfg.Token != "" {
			token = cfg.Token
		} else if fileCreds != nil {
			token = fileCreds.token
		}

		if cfg.UserAgent != "" {
			userAgent = cfg.UserAgent
		}

		if cfg.HTTPClient != nil {
			httpClient = cfg.HTTPClient
		}
	}

	host, err := parseBaseURL(host)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed parsing host address: %w", err))
	}

	if token == "" {
		errs = append(errs, errors.New("token is required"))
	}

	// To aggregate the validation errors above.
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("invalid client configuration:\n%w", err)
	}

	client := &Client{
		token:     token,
		host:      host,
		userAgent: userAgent,
		client:    httpClient,
	}

	return client, nil
}

// defaultUserAgent builds and returns the default user agent string.
func defaultUserAgent() string {
	return fmt.Sprintf("oxide.go/%s", version)
}

func profileCredentials(cfg Config) (*hostCreds, error) {
	configDir := cfg.ConfigDir
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("unable to find user's home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config", "oxide")
	}

	profile := cfg.Profile

	// Use explicitly configured profile over default when both are set.
	if cfg.UseDefaultProfile && profile == "" {
		var err error
		profile, err = defaultProfile(filepath.Join(configDir, "config.toml"))
		if err != nil {
			return nil, fmt.Errorf("failed to get default profile: %w", err)
		}
	}

	fileCreds, err := readCredentials(filepath.Join(configDir, "credentials.toml"), profile)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials for profile %q: %w", profile, err)
	}

	return fileCreds, nil
}

// defaultProfile returns the default profile from config.toml, if present.
func defaultProfile(configPath string) (string, error) {
	configFile, err := toml.LoadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to open config: %w", err)
	}

	if profileName := configFile.Get("default-profile"); profileName != nil {
		return profileName.(string), nil
	}

	return "", errors.New("no default profile set")
}

// readCredentials returns the token and host associated with a profile.
func readCredentials(credentialsPath, profileName string) (*hostCreds, error) {
	if profileName == "" {
		return nil, errors.New("no profile name provided")
	}

	credentialsFile, err := toml.LoadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %v", credentialsPath, err)
	}

	profile, ok := credentialsFile.Get("profile." + profileName).(*toml.Tree)
	if !ok {
		return nil, errors.New("profile not found")
	}

	token, ok := profile.Get("token").(string)
	if !ok {
		return nil, errors.New("token not found")
	}

	host, ok := profile.Get("host").(string)
	if !ok {
		return nil, errors.New("host not found")
	}

	return &hostCreds{host: host, token: token}, nil
}

// parseBaseURL parses the base URL from the host URL.
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

	// Ensure the host URL always has a trailing slash.
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
