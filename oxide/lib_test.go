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
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildRequest(context.TODO(), tt.args.body, tt.args.method, tt.args.uri, tt.args.params, tt.args.queries)
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
