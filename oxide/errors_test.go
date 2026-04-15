package oxide

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPError_Error(t *testing.T) {
	url, _ := url.Parse("http://127.0.0.1:12220/v1/disks/my-disk")
	// Go map iteration is random so we only add a single header
	header := make(http.Header)
	header.Add("Content-Type", "application/json")

	res := http.Response{
		StatusCode: 400,
		Header:     header,
		Request: &http.Request{
			Method: http.MethodPost,
			URL:    url,
		},
	}

	apiErr := ErrorResponse{
		ErrorCode: "ObjectAlreadyExists",
		Message:   "already exists: project \"my-disk\"",
		RequestId: "c42e6ade-69d5-4018-91f8-88bf53b5d026",
	}

	tests := []struct {
		name   string
		fields HTTPError
		want   string
	}{
		{
			name: "returns an error with populated oxide.Error type",
			fields: HTTPError{
				ErrorResponse: &apiErr,
				RawBody:       "{error: Some error}",
				HTTPResponse:  &res,
			},
			want: `POST http://127.0.0.1:12220/v1/disks/my-disk
----------- RESPONSE -----------
Status: 400 ObjectAlreadyExists
Message: already exists: project "my-disk"
RequestID: c42e6ade-69d5-4018-91f8-88bf53b5d026
------- RESPONSE HEADERS -------
Content-Type: [application/json]
`,
		},
		{
			name: "returns an error with raw body",
			fields: HTTPError{
				ErrorResponse: nil,
				RawBody:       "{error: Some error}",
				HTTPResponse:  &res,
			},
			want: `POST http://127.0.0.1:12220/v1/disks/my-disk
----------- RESPONSE -----------
Status: 400
Response Body: {error: Some error}
------- RESPONSE HEADERS -------
Content-Type: [application/json]
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := HTTPError{
				ErrorResponse: tt.fields.ErrorResponse,
				RawBody:       tt.fields.RawBody,
				HTTPResponse:  tt.fields.HTTPResponse,
			}
			got := err.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_NewHTTPError(t *testing.T) {
	url, _ := url.Parse("http://127.0.0.1:12220/v1/disks/my-disk")
	// Go map iteration is random so we only add a single header
	header := make(http.Header)
	header.Add("Content-Type", "application/json")

	res := http.Response{
		StatusCode: 300,
		Header:     header,
		Body:       io.NopCloser(strings.NewReader("some error")),
		Request: &http.Request{
			Method: http.MethodPost,
			URL:    url,
		},
	}

	res2 := http.Response{
		StatusCode: 400,
		Header:     header,
		Body: io.NopCloser(strings.NewReader(`{
  "request_id": "37a8ed33-b7ad-43b0-b2ce-1171d03f5324",
  "error_code": "ObjectAlreadyExists",
  "message": "already exists: project \"my-project\""
}
`)),
		Request: &http.Request{
			Method: http.MethodPost,
			URL:    url,
		},
	}

	tests := []struct {
		name string
		args *http.Response
		want error
	}{
		{
			name: "returns an error without populated oxide.Error type",
			args: &res,
			want: &HTTPError{
				HTTPResponse:  &res,
				RawBody:       "some error",
				ErrorResponse: nil,
			},
		},
		{
			name: "returns an error with populated oxide.Error type",
			args: &res2,
			want: &HTTPError{
				HTTPResponse: &res2,
				RawBody: `{
  "request_id": "37a8ed33-b7ad-43b0-b2ce-1171d03f5324",
  "error_code": "ObjectAlreadyExists",
  "message": "already exists: project \"my-project\""
}
`,
				ErrorResponse: &ErrorResponse{
					Message:   "already exists: project \"my-project\"",
					ErrorCode: "ObjectAlreadyExists",
					RequestId: "37a8ed33-b7ad-43b0-b2ce-1171d03f5324",
				},
			},
		},
		{
			name: "returns nil when is success response",
			args: &http.Response{
				StatusCode: 200,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader("success")),
				Request: &http.Request{
					Method: http.MethodPost,
					URL:    url,
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewHTTPError(tt.args)
			assert.Equal(t, tt.want, err)
		})
	}
}

func Test_NewHTTPError_correct_type(t *testing.T) {
	url, _ := url.Parse("http://127.0.0.1:12220/v1/disks/my-disk")
	// Go map iteration is random so we only add a single header
	header := make(http.Header)
	header.Add("Content-Type", "application/json")

	res := http.Response{
		StatusCode: 300,
		Header:     header,
		Body:       io.NopCloser(strings.NewReader("some error")),
		Request: &http.Request{
			Method: http.MethodPost,
			URL:    url,
		},
	}

	res2 := http.Response{
		StatusCode: 400,
		Header:     header,
		Body: io.NopCloser(strings.NewReader(`{
  "request_id": "37a8ed33-b7ad-43b0-b2ce-1171d03f5324",
  "error_code": "ObjectAlreadyExists",
  "message": "already exists: project \"my-project\""
}
`)),
		Request: &http.Request{
			Method: http.MethodPost,
			URL:    url,
		},
	}

	tests := []struct {
		name string
		args *http.Response
		want error
	}{
		{
			name: "returns an error without populated oxide.Error type",
			args: &res,
			want: &HTTPError{
				HTTPResponse:  &res,
				RawBody:       "some error",
				ErrorResponse: nil,
			},
		},
		{
			name: "returns an error with populated oxide.Error type",
			args: &res2,
			want: &HTTPError{
				HTTPResponse: &res2,
				RawBody: `{
  "request_id": "37a8ed33-b7ad-43b0-b2ce-1171d03f5324",
  "error_code": "ObjectAlreadyExists",
  "message": "already exists: project \"my-project\""
}
`,
				ErrorResponse: &ErrorResponse{
					Message:   "already exists: project \"my-project\"",
					ErrorCode: "ObjectAlreadyExists",
					RequestId: "37a8ed33-b7ad-43b0-b2ce-1171d03f5324",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewHTTPError(tt.args)

			var apiError *HTTPError
			assert.Equal(t, errors.As(err, &apiError), true)
		})
	}
}

func TestHTTPError_Is(t *testing.T) {
	url, _ := url.Parse("http://127.0.0.1:12220/v1/disks/my-disk")

	makeHTTPError := func(errorCode string) error {
		return &HTTPError{
			ErrorResponse: &ErrorResponse{
				ErrorCode: errorCode,
				Message:   "test message",
				RequestId: "test-request-id",
			},
			HTTPResponse: &http.Response{
				StatusCode: 400,
				Header:     make(http.Header),
				Request:    &http.Request{Method: http.MethodGet, URL: url},
			},
		}
	}

	makeStatusError := func(statusCode int) error {
		return &HTTPError{
			HTTPResponse: &http.Response{
				StatusCode: statusCode,
				Header:     make(http.Header),
				Request:    &http.Request{Method: http.MethodGet, URL: url},
			},
		}
	}

	tests := []struct {
		name      string
		err       error
		target    error
		wantMatch bool
	}{
		{
			name:      "matches ObjectNotFound",
			err:       makeHTTPError("ObjectNotFound"),
			target:    ErrObjectNotFound,
			wantMatch: true,
		},
		{
			name:      "matches ObjectAlreadyExists",
			err:       makeHTTPError("ObjectAlreadyExists"),
			target:    ErrObjectAlreadyExists,
			wantMatch: true,
		},
		{
			name:      "does not match different oxide error",
			err:       makeHTTPError("ObjectNotFound"),
			target:    ErrObjectAlreadyExists,
			wantMatch: false,
		},
		{
			name:      "does not match non-oxide error",
			err:       makeHTTPError("ObjectNotFound"),
			target:    errors.New("ObjectNotFound"),
			wantMatch: false,
		},
		{
			name:      "handles fmt.Errorf wrapping",
			err:       fmt.Errorf("creating disk: %w", makeHTTPError("ObjectNotFound")),
			target:    ErrObjectNotFound,
			wantMatch: true,
		},
		{
			name:      "matches status 404",
			err:       makeStatusError(404),
			target:    ErrHTTP404,
			wantMatch: true,
		},
		{
			name:      "matches status 500",
			err:       makeStatusError(500),
			target:    ErrHTTP500,
			wantMatch: true,
		},
		{
			name:      "does not match different status",
			err:       makeStatusError(404),
			target:    ErrHTTP500,
			wantMatch: false,
		},
		{
			name:      "status sentinel matches regardless of error_code",
			err:       makeHTTPError("ObjectNotFound"),
			target:    ErrHTTP400,
			wantMatch: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantMatch {
				assert.ErrorIs(t, tc.err, tc.target)
			} else {
				assert.NotErrorIs(t, tc.err, tc.target)
			}
		})
	}
}
