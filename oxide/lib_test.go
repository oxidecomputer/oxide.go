// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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

	// Test body.
	createBody := &struct {
		Name string `json:"name,omitempty"`
		Size int    `json:"size,omitempty"`
	}{
		Name: "hi",
		Size: 1073741824,
	}
	reqBody := new(bytes.Buffer)
	err := json.NewEncoder(reqBody).Encode(createBody)
	require.NoError(t, err)

	rCloser := io.NopCloser(reqBody)

	// Test client.
	host := "127.0.0.1:12220"
	c, err := NewClient(
		WithHost(fmt.Sprintf("http://%s", host)),
		WithToken("foo"),
	)
	require.NoError(t, err)

	type args struct {
		body    io.Reader
		method  string
		path    string
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
				path:   "/v1/disks",
				params: map[string]string{},
				queries: map[string]string{
					"project": "prod",
				},
			},
			want: &http.Request{
				Method: "POST",
				URL: &url.URL{
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
				path:    "/v1/disks",
				params:  map[string]string{},
				queries: map[string]string{},
			},
			want: &http.Request{
				Method: "GET",
				URL: &url.URL{
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
				path:   "/v1/disks/{{.disk}}",
				params: map[string]string{
					"disk": "hi",
				},
				queries: map[string]string{},
			},
			want: &http.Request{
				Method: "DELETE",
				URL: &url.URL{
					Path:     "/v1/disks/hi",
					RawPath:  "/v1/disks/hi",
					RawQuery: "",
				},
				Body: nil,
			},
		},
		{
			name: "builds request successfully with empty query values",
			args: args{
				body:   nil,
				method: http.MethodDelete,
				path:   "/v1/disks/{{.disk}}",
				params: map[string]string{
					"disk": "hi",
				},
				queries: map[string]string{
					"organization": "",
					"project":      "",
				},
			},
			want: &http.Request{
				Method: "DELETE",
				URL: &url.URL{
					Path:     "/v1/disks/hi",
					RawPath:  "/v1/disks/hi",
					RawQuery: "",
				},
				Body: nil,
			},
		},
		{
			name: "sanitize inputs",
			args: args{
				body:   nil,
				method: http.MethodGet,
				path:   "/v1/disks/{{.disk}}",
				params: map[string]string{
					"disk": "../../projects/secret",
				},
				queries: map[string]string{
					"test": "test?other=test",
				},
			},
			want: &http.Request{
				Method: "GET",
				URL: &url.URL{
					Path:     "/v1/disks/../../projects/secret",
					RawPath:  "/v1/disks/..%2F..%2Fprojects%2Fsecret",
					RawQuery: "test=test%3Fother%3Dtest",
				},
				Body: nil,
			},
		},
		{
			name: "fails on a malformed path",
			args: args{
				body:   nil,
				method: http.MethodDelete,
				path:   "/v1/disks/{{.disk}}",
				params: map[string]string{
					"risk": "hi",
				},
				queries: map[string]string{},
			},
			wantErr: "map has no entry for key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, err := c.buildRequest(
				t.Context(),
				tt.args.body,
				tt.args.method,
				fmt.Sprintf("http://%s%s", host, tt.args.path),
				tt.args.params,
				tt.args.queries,
			)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				// Only asserting values that we care about
				assert.Equal(t, host, req.URL.Host)
				assert.Equal(t, tt.want.Method, req.Method)
				assert.Equal(t, tt.want.Body, req.Body)
				assert.Equal(t, tt.want.URL.Path, req.URL.Path)
				assert.Equal(t, tt.want.URL.RawPath, req.URL.RawPath)
				assert.Equal(t, tt.want.URL.RawQuery, req.URL.RawQuery)
			}
		})
	}
}

func Test_NewClient(t *testing.T) {
	tt := map[string]struct {
		options        func(string) []ClientOption
		setHome        bool
		env            map[string]string
		expectedClient *Client
		expectedError  string
	}{
		"succeeds with valid client from options": {
			options: func(string) []ClientOption {
				return []ClientOption{
					WithHost("http://localhost"),
					WithToken("foo"),
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
		"succeeds with valid client from env and options": {
			env: map[string]string{
				"OXIDE_HOST":  "http://localhost",
				"OXIDE_TOKEN": "foo",
			},
			options: func(string) []ClientOption {
				return []ClientOption{
					WithUserAgent("bob"),
					WithHTTPClient(&http.Client{
						Timeout: 500 * time.Second,
					}),
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
		"succeeds with config, overrides env": {
			env: map[string]string{
				"OXIDE_PROFILE": "file",
			},
			options: func(string) []ClientOption {
				return []ClientOption{
					WithHost("http://localhost"),
					WithToken("foo"),
				}
			},
			setHome: true,
			expectedClient: &Client{
				host:  "http://localhost/",
				token: "foo",
				client: &http.Client{
					Timeout: 600 * time.Second,
				},
				userAgent: defaultUserAgent(),
			},
		},
		"succeeds with profile": {
			env: map[string]string{
				"OXIDE_HOST":  "",
				"OXIDE_TOKEN": "",
			},
			options: func(string) []ClientOption {
				return []ClientOption{
					WithProfile("file"),
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
		"succeeds with profile from env": {
			env: map[string]string{
				"OXIDE_PROFILE": "file",
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
			options: func(string) []ClientOption {
				return []ClientOption{
					WithDefaultProfile(),
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
			options: func(oxideDir string) []ClientOption {
				return []ClientOption{
					WithDefaultProfile(),
					WithConfigDir(oxideDir),
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
			options: func(oxideDir string) []ClientOption {
				return []ClientOption{
					WithProfile("other"),
					WithConfigDir(oxideDir),
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
			options: func(string) []ClientOption {
				return []ClientOption{
					WithProfile("other"),
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
		"succeeds with host and token from different sources ": {
			env: map[string]string{
				"OXIDE_TOKEN": "foo",
			},
			options: func(string) []ClientOption {
				return []ClientOption{
					WithHost("http://localhost"),
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
		"fails with missing address using options": {
			env: map[string]string{
				"OXIDE_HOST":  "",
				"OXIDE_TOKEN": "",
			},
			options: func(string) []ClientOption {
				return []ClientOption{
					WithToken("foo"),
				}
			},
			expectedError: "invalid client configuration:\nfailed parsing host address: host address is empty",
		},
		"fails with missing token using options": {
			env: map[string]string{
				"OXIDE_HOST":  "",
				"OXIDE_TOKEN": "",
			},
			options: func(string) []ClientOption {
				return []ClientOption{
					WithHost("http://localhost"),
				}
			},
			expectedError: "invalid client configuration:\ntoken is required",
		},
		"fails with missing address using env variables": {
			env: map[string]string{
				"OXIDE_HOST":  "",
				"OXIDE_TOKEN": "foo",
			},
			expectedError: "invalid client configuration:\nfailed parsing host address: host address is empty",
		},
		"fails with missing token using env variables": {
			env: map[string]string{
				"OXIDE_HOST":  "http://localhost",
				"OXIDE_TOKEN": "",
			},
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
			options: func(string) []ClientOption {
				return []ClientOption{
					WithConfigDir("/not/a/valid/directory"),
					WithDefaultProfile(),
				}
			},
			expectedError: "unable to retrieve profile: failed to get default profile from \"/not/a/valid/directory/config.toml\": failed to open config: open /not/a/valid/directory/config.toml: no such file or directory",
		},
		"fails with invalid profile": {
			options: func(string) []ClientOption {
				return []ClientOption{
					WithProfile("not-a-profile"),
				}
			},
			setHome:       true,
			expectedError: "unable to retrieve profile: failed to get credentials for profile \"not-a-profile\" from \"<OXIDE_DIR>/credentials.toml\": profile not found",
		},
		"fails with profile and host": {
			options: func(string) []ClientOption {
				return []ClientOption{
					WithHost("http://localhost"),
					WithProfile("file"),
				}
			},
			setHome:       true,
			expectedError: "cannot authenticate with both a profile and host/token",
		},
		"fails with profile and token": {
			options: func(string) []ClientOption {
				return []ClientOption{
					WithToken("foo"),
					WithProfile("file"),
				}
			},
			setHome:       true,
			expectedError: "cannot authenticate with both a profile and host/token",
		},
		"fails with profile and default profile": {
			options: func(string) []ClientOption {
				return []ClientOption{
					WithProfile("file"),
					WithDefaultProfile(),
				}
			},
			setHome:       true,
			expectedError: "cannot authenticate with both default profile and a defined profile",
		},
	}

	for testName, testCase := range tt {
		t.Run(testName, func(t *testing.T) {
			for key, val := range testCase.env {
				t.Setenv(key, val)
			}

			tmpDir := setupConfig(t)
			oxideDir := filepath.Join(tmpDir, ".config", "oxide")

			originalHome := os.Getenv("HOME")
			if testCase.setHome {
				t.Setenv("HOME", tmpDir)
			} else {
				// Ensure we don't read from your actual credentials.toml.
				os.Unsetenv("HOME")
			}

			t.Cleanup(func() {
				os.Setenv("HOME", originalHome)
				require.NoError(t, os.RemoveAll(tmpDir))
			})

			var opts []ClientOption
			if testCase.options != nil {
				opts = testCase.options(oxideDir)
			}
			c, err := NewClient(opts...)

			if testCase.expectedError != "" {
				assert.EqualError(
					t,
					err,
					strings.ReplaceAll(testCase.expectedError, "<OXIDE_DIR>", oxideDir),
				)
			} else {
				assert.NoError(t, err)
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
	require.NoError(
		t,
		os.WriteFile(filepath.Join(oxideDir, "credentials.toml"), credentials, 0o600),
	)
	require.NoError(
		t,
		os.WriteFile(
			filepath.Join(oxideDir, "config.toml"),
			[]byte(`default-profile = "file"`),
			0o644,
		),
	)

	return tmpDir
}

func Test_MakeRequest(t *testing.T) {
	tests := []struct {
		name          string
		request       Request
		expectedQuery string
	}{
		{
			name: "request without optional fields",
			request: Request{
				Method: http.MethodGet,
				Path:   "/v1/projects",
			},
			expectedQuery: "",
		},
		{
			name: "request with all fields",
			request: Request{
				Method: http.MethodPost,
				Path:   "/v1/projects",
				Body:   strings.NewReader(`{"name":"my-project"}`),
				Params: map[string]string{
					"project": "my-project",
				},
				Query: map[string]string{
					"project": "my-project",
				},
			},
			expectedQuery: "project=my-project",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedRequest *http.Request
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					capturedRequest = r
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(`{"status":"ok"}`))
					require.NoError(t, err)
				}),
			)
			defer server.Close()

			client, err := NewClient(
				WithHost(server.URL),
				WithToken("test-token"),
			)
			require.NoError(t, err)

			resp, err := client.MakeRequest(context.Background(), tc.request)
			require.NoError(t, err)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			require.NotNil(t, capturedRequest)
			assert.Equal(t, tc.request.Method, capturedRequest.Method)
			assert.Equal(t, tc.request.Path, capturedRequest.URL.Path)
			assert.Equal(t, tc.expectedQuery, capturedRequest.URL.RawQuery)

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func Test_NewClient_HTTPOptions(t *testing.T) {
	customClient := &http.Client{
		Timeout: 99 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}

	tests := []struct {
		name                   string
		options                []ClientOption
		expectedTimeout        time.Duration
		expectedInsecureVerify bool
	}{
		{
			name: "timeout then insecure skip verify",
			options: []ClientOption{
				WithHost("https://localhost"),
				WithToken("test-token"),
				WithTimeout(30 * time.Second),
				WithInsecureSkipVerify(),
			},
			expectedTimeout:        30 * time.Second,
			expectedInsecureVerify: true,
		},
		{
			name: "insecure skip verify then timeout",
			options: []ClientOption{
				WithHost("https://localhost"),
				WithToken("test-token"),
				WithInsecureSkipVerify(),
				WithTimeout(30 * time.Second),
			},
			expectedTimeout:        30 * time.Second,
			expectedInsecureVerify: true,
		},
		{
			name: "WithHTTPClient overrides other options",
			options: []ClientOption{
				WithHost("https://localhost"),
				WithToken("test-token"),
				WithTimeout(30 * time.Second),
				WithInsecureSkipVerify(),
				WithHTTPClient(customClient),
			},
			expectedTimeout:        99 * time.Second,
			expectedInsecureVerify: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(tc.options...)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedTimeout, client.client.Timeout)

			transport, ok := client.client.Transport.(*http.Transport)
			require.True(t, ok, "expected *http.Transport")
			require.NotNil(t, transport.TLSClientConfig)
			assert.Equal(t, tc.expectedInsecureVerify, transport.TLSClientConfig.InsecureSkipVerify)
		})
	}
}
