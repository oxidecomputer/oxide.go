// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"crypto/tls"
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

const (
	// TokenEnvVar is the environment variable that contains the token.
	TokenEnvVar = "OXIDE_TOKEN"

	// HostEnvVar is the environment variable that contains the host.
	HostEnvVar = "OXIDE_HOST"

	// ProfileEnvVar is the environment variable that contains the credentials
	// profile to use.
	ProfileEnvVar = "OXIDE_PROFILE"

	// credentialsFile is the name of the file the Oxide CLI stores credentials in.
	credentialsFile = "credentials.toml"

	// configFile is the name of the file the Oxide CLI stores its config in.
	configFile = "config.toml"

	// defaultConfigDir is the default path used by the Oxide CLI for configuration files.
	defaultConfigDir = ".config" + string(filepath.Separator) + "oxide"
)

// ClientOption configures [Client] during calls to [NewClient].
type ClientOption interface {
	apply(cfg *clientConfig) error
}

// clientOptionFunc provides a simpler type to implement the [ClientOption]
// interface using closure functions. It allows the [clientConfig] to remain
// unexported and out of the documentation for external users.
type clientOptionFunc func(cfg *clientConfig) error

// apply modifies cfg using the given clientOptionFunc, implementing
// [ClientOption].
func (c clientOptionFunc) apply(cfg *clientConfig) error {
	return c(cfg)
}

// clientConfig holds the configuration for a [Client] while it's being
// constructed via [NewClient].
type clientConfig struct {
	host              string
	token             string
	profile           string
	useDefaultProfile bool
	configDir         string
	httpClient        *http.Client
	userAgent         string

	// These fields track whether the options were set from [ClientOption]. This
	// is used to determine whether values set via environment variables should
	// be overriden.
	hostSetFromOption           bool
	tokenSetFromOption          bool
	profileSetFromOption        bool
	defaultProfileSetFromOption bool
}

// WithHost sets the Oxide host for [Client] (e.g.,
// https://oxide.sys.example.com).
func WithHost(host string) ClientOption {
	return clientOptionFunc(func(cfg *clientConfig) error {
		cfg.host = host
		cfg.hostSetFromOption = true
		return nil
	})
}

// WithToken sets the API token for [Client].
func WithToken(token string) ClientOption {
	return clientOptionFunc(func(cfg *clientConfig) error {
		cfg.token = token
		cfg.tokenSetFromOption = true
		return nil
	})
}

// WithProfile sets the profile name within the credentials file to use for
// authentication. This is mutually exclusive with [WithHost], [WithToken], and
// [WithDefaultProfile].
func WithProfile(profile string) ClientOption {
	return clientOptionFunc(func(cfg *clientConfig) error {
		cfg.profile = profile
		cfg.profileSetFromOption = true
		return nil
	})
}

// WithDefaultProfile uses the default profile within the credentials file
// to use for authentication. This is mutually exclusive with [WithHost],
// [WithToken], and [WithProfile].
func WithDefaultProfile() ClientOption {
	return clientOptionFunc(func(cfg *clientConfig) error {
		cfg.useDefaultProfile = true
		cfg.defaultProfileSetFromOption = true
		return nil
	})
}

// WithConfigDir sets the directory to search for the Oxide credentials file.
func WithConfigDir(dir string) ClientOption {
	return clientOptionFunc(func(cfg *clientConfig) error {
		cfg.configDir = dir
		return nil
	})
}

// WithTimeout sets the timeout for the HTTP client. This option is overriden if
// [WithHTTPClient] is set.
func WithTimeout(timeout time.Duration) ClientOption {
	return clientOptionFunc(func(cfg *clientConfig) error {
		if cfg.httpClient == nil {
			cfg.httpClient = defaultHTTPClient()
		}
		cfg.httpClient.Timeout = timeout
		return nil
	})
}

// WithHTTPClient sets a custom HTTP client, replacing the default HTTP client
// entirely. This overrides [WithTimeout], [WithInsecureSkipVerify], and
// [WithUserAgent] and should only be used in advanced use cases such as
// configuring a proxy or changing TLS configuration.
func WithHTTPClient(client *http.Client) ClientOption {
	return clientOptionFunc(func(cfg *clientConfig) error {
		cfg.httpClient = client
		return nil
	})
}

// WithInsecureSkipVerify disables TLS certificate verification. This is
// insecure and should only be used for testing or in controlled environments.
// This option is overridden if [WithHTTPClient] is set.
func WithInsecureSkipVerify() ClientOption {
	return clientOptionFunc(func(cfg *clientConfig) error {
		if cfg.httpClient == nil {
			cfg.httpClient = defaultHTTPClient()
		}
		transport, ok := cfg.httpClient.Transport.(*http.Transport)
		if !ok || transport == nil {
			transport = http.DefaultTransport.(*http.Transport).Clone()
		}
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
		cfg.httpClient.Transport = transport
		return nil
	})
}

// WithUserAgent sets the user agent string for the client. This option is
// overriden if [WithHTTPClient] is set.
func WithUserAgent(userAgent string) ClientOption {
	return clientOptionFunc(func(cfg *clientConfig) error {
		cfg.userAgent = userAgent
		return nil
	})
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

type authCredentials struct {
	host  string
	token string
}

// NewClient creates an Oxide API client. When called with no options, it reads
// configuration from environment variables `OXIDE_HOST` and `OXIDE_TOKEN`, or
// `OXIDE_PROFILE`. When called with one or more [ClientOption], it configure
// the client accordingly, overriding values from environment variables. When
// the same [ClientOption] is passed multiple times, the last argument wins.
//
// The [WithHost] and [WithToken] options are mutually exclusive with
// [WithProfile] and [WithDefaultProfile].
func NewClient(opts ...ClientOption) (*Client, error) {
	cfg := &clientConfig{
		host:       os.Getenv(HostEnvVar),
		token:      os.Getenv(TokenEnvVar),
		profile:    os.Getenv(ProfileEnvVar),
		userAgent:  defaultUserAgent(),
		httpClient: defaultHTTPClient(),
	}

	var optErrs []error
	for _, opt := range opts {
		if err := opt.apply(cfg); err != nil {
			optErrs = append(optErrs, err)
		}
	}
	if err := errors.Join(optErrs...); err != nil {
		return nil, fmt.Errorf("failed to apply options:\n%w", err)
	}

	// Validate conflicting options.
	if (cfg.profileSetFromOption || cfg.defaultProfileSetFromOption) && (cfg.hostSetFromOption || cfg.tokenSetFromOption) {
		return nil, errors.New("cannot authenticate with both a profile and host/token")
	}
	if cfg.profileSetFromOption && cfg.defaultProfileSetFromOption {
		return nil, errors.New("cannot authenticate with both default profile and a defined profile")
	}

	// Options override environment variables.
	if cfg.hostSetFromOption || cfg.tokenSetFromOption {
		if !cfg.profileSetFromOption {
			cfg.profile = ""
		}
		if !cfg.defaultProfileSetFromOption {
			cfg.useDefaultProfile = false
		}
	}
	if cfg.profileSetFromOption || cfg.defaultProfileSetFromOption {
		if !cfg.hostSetFromOption {
			cfg.host = ""
		}
		if !cfg.tokenSetFromOption {
			cfg.token = ""
		}
	}

	if cfg.profile != "" || cfg.useDefaultProfile {
		configDir := cfg.configDir
		if configDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("unable to find user's home directory: %w", err)
			}
			configDir = filepath.Join(homeDir, defaultConfigDir)
		}

		authCredentials, err := getProfile(configDir, cfg.profile, cfg.useDefaultProfile)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve profile: %w", err)
		}

		cfg.host = authCredentials.host
		cfg.token = authCredentials.token
	}

	errs := make([]error, 0)
	host, err := parseBaseURL(cfg.host)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed parsing host address: %w", err))
	}

	if cfg.token == "" {
		errs = append(errs, errors.New("token is required"))
	}

	// To aggregate the validation errors above.
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("invalid client configuration:\n%w", err)
	}

	client := &Client{
		token:     cfg.token,
		host:      host,
		userAgent: cfg.userAgent,
		client:    cfg.httpClient,
	}

	return client, nil
}

// defaultHTTPClient builds and returns the default HTTP client.
func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 600 * time.Second,
	}
}

// defaultUserAgent builds and returns the default user agent string.
func defaultUserAgent() string {
	return fmt.Sprintf("oxide.go/%s", sdkVersion)
}

// getProfile determines the path of the user's credentials file
// and returns the host and token for the requested profile.
func getProfile(configDir string, profile string, useDefault bool) (*authCredentials, error) {
	// Use explicitly configured profile over default when both are set.
	if useDefault && profile == "" {
		configPath := filepath.Join(configDir, configFile)

		var err error
		profile, err = defaultProfile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get default profile from %q: %w", configPath, err)
		}
	}

	credentialsPath := filepath.Join(configDir, credentialsFile)
	fileCreds, err := parseCredentialsFile(credentialsPath, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials for profile %q from %q: %w", profile, credentialsPath, err)
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

// parseCredentialsFile parses a credentials.toml and returns the token and host
// associated with the requested profile.
func parseCredentialsFile(credentialsPath, profileName string) (*authCredentials, error) {
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

	var hostTokenErr error
	token, ok := profile.Get("token").(string)
	if !ok {
		hostTokenErr = errors.New("token not found")
	}

	host, ok := profile.Get("host").(string)
	if !ok {
		hostTokenErr = errors.Join(errors.New("host not found"))
	}

	return &authCredentials{host: host, token: token}, hostTokenErr
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
	req.Header.Set("API-Version", openAPIVersion)

	// Add the parameters to the url.
	if err := expandURL(req.URL, params); err != nil {
		return nil, fmt.Errorf("expanding URL with parameters failed: %v", err)
	}

	// Add queries if any
	addQueries(req.URL, queries)

	return req, nil
}

type Request struct {
	Method string
	Path   string
	Body   io.Reader
	Params map[string]string
	Query  map[string]string
}

// MakeRequest takes a `Request` that defines the desired API request, builds
// the URI using the configured API host, and sends the request to the API.
func (c *Client) MakeRequest(ctx context.Context, req Request) (*http.Response, error) {
	uri := resolveRelative(c.host, req.Path)
	httpReq, err := c.buildRequest(ctx, req.Body, req.Method, uri, req.Params, req.Query)
	if err != nil {
		return nil, fmt.Errorf("building request failed: %v", err)
	}
	return c.client.Do(httpReq)
}
