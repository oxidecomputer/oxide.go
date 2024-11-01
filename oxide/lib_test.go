// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_buildRequest(t *testing.T) {
	t.Parallel()

	type dummyCreate struct {
		Name string `json:"name,omitempty"`
		Size int    `json:"size,omitempty"`
	}
	createBody := &dummyCreate{
		Name: "hi",
		Size: 1073741824,
	}
	reqBody := new(bytes.Buffer)
	if err := json.NewEncoder(reqBody).Encode(createBody); err != nil {
		t.Errorf("encoding json body request failed: %v", err)
		return
	}

	rCloser := io.NopCloser(reqBody)

	type args struct {
		body    io.Reader
		method  string
		uri     string
		params  map[string]string
		queries map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    *http.Request
		wantErr string
	}{
		{
			name: "builds request successfully with body",
			args: args{
				body:   reqBody,
				method: http.MethodPost,
				uri:    "http://127.0.0.1:12220/v1/disks",
				params: map[string]string{},
				queries: map[string]string{
					"project": "prod",
				},
			},
			want: &http.Request{
				Method: "POST",
				URL: &url.URL{
					Scheme:   "http",
					Host:     "127.0.0.1:12220",
					Path:     "/v1/disks",
					RawPath:  "/v1/disks",
					RawQuery: "project=prod",
				},
				Body: rCloser,
			},
		},
		{
			name: "builds request successfully without body or params",
			args: args{
				body:    nil,
				method:  http.MethodGet,
				uri:     "http://127.0.0.1:12220/v1/disks",
				params:  map[string]string{},
				queries: map[string]string{},
			},
			want: &http.Request{
				Method: "GET",
				URL: &url.URL{
					Scheme:   "http",
					Host:     "127.0.0.1:12220",
					Path:     "/v1/disks",
					RawPath:  "/v1/disks",
					RawQuery: "",
				},
				Body: nil,
			},
		},
		{
			name: "builds request successfully with params",
			args: args{
				body:   nil,
				method: http.MethodDelete,
				uri:    "http://127.0.0.1:12220/v1/disks/{{.disk}}",
				params: map[string]string{
					"disk": "hi",
				},
				queries: map[string]string{},
			},
			want: &http.Request{
				Method: "DELETE",
				URL: &url.URL{
					Scheme:   "http",
					Host:     "127.0.0.1:12220",
					Path:     "/v1/disks/hi",
					RawPath:  "/v1/disks/hi",
					RawQuery: "",
				},
				Body: nil,
			},
		},
		// TODO: Create a check that verifies that path is not malformed
		//		{
		//			name: "fails on a malformed path",
		//			args: args{
		//				body:   nil,
		//				method: http.MethodDelete,
		//				uri:    "http://127.0.0.1:12220/v1/disks/{{.disk}}",
		//				params: map[string]string{
		//					"risk": "hi",
		//				},
		//				queries: map[string]string{},
		//			},
		//			wantErr: "Some error that doesn't exist yet",
		//		},
	}

	// Just to get a client to call buildRequest on.
	c, err := NewClient(&Config{
		Host:  "http://localhost:3000",
		Token: "foo",
	})
	if err != nil {
		t.Fatalf("failed creating api client: %v", err)
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := c.buildRequest(context.TODO(), tt.args.body, tt.args.method, tt.args.uri, tt.args.params, tt.args.queries)
			if err != nil {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			// Only asserting values that we care about
			assert.Equal(t, tt.want.Method, got.Method)
			assert.Equal(t, tt.want.Body, got.Body)
			assert.Equal(t, tt.want.URL.Host, got.URL.Host)
			assert.Equal(t, tt.want.URL.Path, got.URL.Path)
			assert.Equal(t, tt.want.URL.RawPath, got.URL.RawPath)
			assert.Equal(t, tt.want.URL.RawQuery, got.URL.RawQuery)
		})
	}
}

func Test_NewClient(t *testing.T) {
	tt := map[string]struct {
		config         func(string) *Config
		setHome        bool
		env            map[string]string
		expectedClient *Client
		expectedError  string
	}{
		"succeeds with valid client from config": {
			config: func(string) *Config {
				return &Config{
					Host:  "http://localhost",
					Token: "foo",
				}
			},
			expectedClient: &Client{
				host:  "http://localhost/",
				token: "foo",
				client: &http.Client{
					Timeout: 600 * time.Second,
				},
				userAgent: defaultUserAgent(),
			},
		},
		"succeeds with valid client from env": {
			env: map[string]string{
				"OXIDE_HOST":  "http://localhost",
				"OXIDE_TOKEN": "foo",
			},
			expectedClient: &Client{
				host:  "http://localhost/",
				token: "foo",
				client: &http.Client{
					Timeout: 600 * time.Second,
				},
				userAgent: defaultUserAgent(),
			},
		},
		"succeeds with valid client from env and config": {
			env: map[string]string{
				"OXIDE_HOST":  "http://localhost",
				"OXIDE_TOKEN": "foo",
			},
			config: func(string) *Config {
				return &Config{
					UserAgent: "bob",
					HTTPClient: &http.Client{
						Timeout: 500 * time.Second,
					},
				}
			},
			expectedClient: &Client{
				host:  "http://localhost/",
				token: "foo",
				client: &http.Client{
					Timeout: 500 * time.Second,
				},
				userAgent: "bob",
			},
		},
		"succeeds with profile": {
			config: func(string) *Config {
				return &Config{
					Profile: "file",
				}
			},
			setHome: true,
			expectedClient: &Client{
				host:  "http://file-host/",
				token: "file-token",
				client: &http.Client{
					Timeout: 600 * time.Second,
				},
				userAgent: defaultUserAgent(),
			},
		},
		"succeeds with default profile": {
			config: func(string) *Config {
				return &Config{
					UseDefaultProfile: true,
				}
			},
			setHome: true,
			expectedClient: &Client{
				host:  "http://file-host/",
				token: "file-token",
				client: &http.Client{
					Timeout: 600 * time.Second,
				},
				userAgent: defaultUserAgent(),
			},
		},
		"succeeds with config dir and default profile": {
			config: func(oxideDir string) *Config {
				return &Config{
					UseDefaultProfile: true,
					ConfigDir:         oxideDir,
				}
			},
			expectedClient: &Client{
				host:  "http://file-host/",
				token: "file-token",
				client: &http.Client{
					Timeout: 600 * time.Second,
				},
				userAgent: defaultUserAgent(),
			},
		},
		"succeeds with config dir and profile": {
			config: func(oxideDir string) *Config {
				return &Config{
					Profile:   "other",
					ConfigDir: oxideDir,
				}
			},
			expectedClient: &Client{
				host:  "http://other-host/",
				token: "other-token",
				client: &http.Client{
					Timeout: 600 * time.Second,
				},
				userAgent: defaultUserAgent(),
			},
		},
		"succeeds with profile, overrides env": {
			env: map[string]string{
				"OXIDE_HOST":  "http://localhost",
				"OXIDE_TOKEN": "foo",
			},
			config: func(string) *Config {
				return &Config{
					Profile: "other",
				}
			},
			setHome: true,
			expectedClient: &Client{
				host:  "http://other-host/",
				token: "other-token",
				client: &http.Client{
					Timeout: 600 * time.Second,
				},
				userAgent: defaultUserAgent(),
			},
		},
		"fails with missing address using config": {
			env: map[string]string{
				"OXIDE_HOST":  "",
				"OXIDE_TOKEN": "",
			},
			config: func(string) *Config {
				return &Config{
					Token: "foo",
				}
			},
			expectedError: "invalid client configuration:\nfailed parsing host address: host address is empty",
		},
		"fails with missing token using config": {
			env: map[string]string{
				"OXIDE_HOST":  "",
				"OXIDE_TOKEN": "",
			},
			config: func(string) *Config {
				return &Config{
					Host: "http://localhost",
				}
			},
			expectedError: "invalid client configuration:\ntoken is required",
		},
		"fails with missing address using env variables": {
			env: map[string]string{
				"OXIDE_HOST":  "",
				"OXIDE_TOKEN": "foo",
			},
			config:        nil,
			expectedError: "invalid client configuration:\nfailed parsing host address: host address is empty",
		},
		"fails with missing token using env variables": {
			env: map[string]string{
				"OXIDE_HOST":  "http://localhost",
				"OXIDE_TOKEN": "",
			},
			config:        nil,
			expectedError: "invalid client configuration:\ntoken is required",
		},
		"fails with missing address and token": {
			env: map[string]string{
				"OXIDE_HOST":  "",
				"OXIDE_TOKEN": "",
			},
			expectedError: "invalid client configuration:\nfailed parsing host address: host address is empty\ntoken is required",
		},
		"fails with invalid config dir": {
			config: func(string) *Config {
				return &Config{
					ConfigDir:         "/not/a/valid/directory",
					UseDefaultProfile: true,
				}
			},
			expectedError: "invalid client configuration:\nunable to retrieve profile: failed to get default profile from \"/not/a/valid/directory/config.toml\": failed to open config: open /not/a/valid/directory/config.toml: no such file or directory\nfailed parsing host address: host address is empty\ntoken is required",
		},
		"fails with invalid profile": {
			config: func(string) *Config {
				return &Config{
					Profile: "not-a-profile",
				}
			},
			setHome:       true,
			expectedError: "invalid client configuration:\nunable to retrieve profile: failed to get credentials for profile \"not-a-profile\" from \"<OXIDE_DIR>/credentials.toml\": profile not found\nfailed parsing host address: host address is empty\ntoken is required",
		},
		"fails with invalid profile and default profile": {
			config: func(oxideDir string) *Config {
				return &Config{
					ConfigDir:         oxideDir,
					Profile:           "not-a-profile",
					UseDefaultProfile: true,
				}
			},
			setHome:       true,
			expectedError: "invalid client configuration:\nunable to retrieve profile: failed to get credentials for profile \"not-a-profile\" from \"<OXIDE_DIR>/credentials.toml\": profile not found\nfailed parsing host address: host address is empty\ntoken is required",
		},
		"fails with profile and host": {
			config: func(oxideDir string) *Config {
				return &Config{
					Host:    "http://localhost",
					Profile: "file",
				}
			},
			setHome:       true,
			expectedError: "invalid client configuration:\nunable to retrieve profile: cannot authenticate with both a profile and host/token\ntoken is required",
		},
		"fails with profile and token": {
			config: func(oxideDir string) *Config {
				return &Config{
					Token:   "foo",
					Profile: "file",
				}
			},
			setHome:       true,
			expectedError: "invalid client configuration:\nunable to retrieve profile: cannot authenticate with both a profile and host/token\nfailed parsing host address: host address is empty",
		},
	}

	for testName, testCase := range tt {
		t.Run(testName, func(t *testing.T) {
			for key, val := range testCase.env {
				os.Setenv(key, val)
			}

			tmpDir := setupConfig(t)
			oxideDir := filepath.Join(tmpDir, ".config", "oxide")

			originalHome := os.Getenv("HOME")
			if testCase.setHome {
				os.Setenv("HOME", tmpDir)
			} else {
				// Ensure we don't read from your actual credentials.toml.
				os.Unsetenv("HOME")
			}

			t.Cleanup(func() {
				for key := range testCase.env {
					os.Unsetenv(key)
				}
				os.Setenv("HOME", originalHome)
				require.NoError(t, os.RemoveAll(tmpDir))
			})

			var config *Config
			if testCase.config != nil {
				config = testCase.config(oxideDir)
			}
			c, err := NewClient(config)

			if testCase.expectedError != "" {
				assert.EqualError(t, err, strings.ReplaceAll(testCase.expectedError, "<OXIDE_DIR>", oxideDir))
			}

			assert.Equal(t, testCase.expectedClient, c)
		})
	}
}

func setupConfig(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	oxideDir := filepath.Join(tmpDir, ".config", "oxide")
	require.NoError(t, os.MkdirAll(oxideDir, 0o755))

	credentials := []byte(`
[profile.file]
host = "http://file-host"
token = "file-token"
user = "file-user"

[profile.other]
host = "http://other-host"
token = "other-token"
user = "other-user"
`)
	require.NoError(t, os.WriteFile(filepath.Join(oxideDir, "credentials.toml"), credentials, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(oxideDir, "config.toml"), []byte(`default-profile = "file"`), 0o644))

	return tmpDir
}
