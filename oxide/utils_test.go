package oxide

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type expandTest struct {
	in         string
	expansions map[string]string
	want       string
}

const testServerURL = "https://example.com"

var expandTests = []expandTest{
	// no expansions
	{
		"",
		map[string]string{},
		testServerURL,
	},
	// multiple expansions, no escaping
	{
		"file/convert/{{.srcFormat}}/{{.outputFormat}}",
		map[string]string{
			"srcFormat":    "step",
			"outputFormat": "obj",
		},
		testServerURL + "/file/convert/step/obj",
	},
}

func TestExpandURL(t *testing.T) {
	for i, test := range expandTests {
		uri := resolveRelative(testServerURL, test.in)
		u, err := url.Parse(uri)
		if err != nil {
			t.Fatalf("parsing url %q failed: %v", test.in, err)
		}
		expandURL(u, test.expansions)
		got := u.String()
		if got != test.want {
			t.Errorf("got %q expected %q in test %d", got, test.want, i+1)
		}
	}
}

func Test_addQueries(t *testing.T) {
	type args struct {
		query map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "keeps URL the same if no query params supplied",
			args: args{query: map[string]string{}},
			want: "https://example.com",
		},
		{
			name: "keeps URL the same if no query values supplied",
			args: args{query: map[string]string{
				"organization": "",
				"project":      "",
			}},
			want: "https://example.com",
		},
		{
			name: "adds query parameters successfully",
			args: args{query: map[string]string{
				"organization": "myorg",
				"project":      "prod",
			}},
			want: "https://example.com?organization=myorg&project=prod",
		},
	}
	for _, tt := range tests {
		u, err := url.Parse(testServerURL)
		if err != nil {
			t.Fatalf("parsing url %q failed: %v", testServerURL, err)
		}

		t.Run(tt.name, func(t *testing.T) {
			addQueries(u, tt.args.query)
			got := u.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_buildRequest(t *testing.T) {
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
				}},
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildRequest(tt.args.body, tt.args.method, tt.args.uri, tt.args.params, tt.args.queries)
			if err != nil {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want.Method, got.Method)
			assert.Equal(t, tt.want.Body, got.Body)
			assert.Equal(t, tt.want.URL.Host, got.URL.Host)
			assert.Equal(t, tt.want.URL.Path, got.URL.Path)
			assert.Equal(t, tt.want.URL.RawPath, got.URL.RawPath)
			assert.Equal(t, tt.want.URL.RawQuery, got.URL.RawQuery)
		})
	}
}
