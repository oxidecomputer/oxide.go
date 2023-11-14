package oxide

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPError_Error(t *testing.T) {
	url, _ := url.Parse("http://127.0.0.1:12220/v1/disks/my-disk")
	header := http.Header{}
	header.Add("X-Request-Id", "99b40b0a-234f-4e10-87b4-081b0432ad19")
	header.Add("Date", "Tue, 14 Nov 2023 06:57:09 GMT")
	header.Add("Content-Length", "152")
	header.Add("Content-Type", "application/json")
	tests := []struct {
		name   string
		fields HTTPError
		want   string
	}{
		{
			name: "returns an error with raw body",
			fields: HTTPError{
				URL:           url,
				StatusCode:    400,
				RequestMethod: http.MethodDelete,
				ErrorResponse: nil,
				RawBody:       "{error: Some error}",
				Header:        header,
			},
			want: `
------- REQUEST -------
DELETE http://127.0.0.1:12220/v1/disks/my-disk
X-Request-Id: [99b40b0a-234f-4e10-87b4-081b0432ad19]
Date: [Tue, 14 Nov 2023 06:57:09 GMT]
Content-Length: [152]
Content-Type: [application/json]
------- RESPONSE -------
Status: 400
Response Body: {error: Some error}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := HTTPError{
				URL:           tt.fields.URL,
				StatusCode:    tt.fields.StatusCode,
				RequestMethod: tt.fields.RequestMethod,
				ErrorResponse: tt.fields.ErrorResponse,
				RawBody:       tt.fields.RawBody,
				Header:        tt.fields.Header,
			}
			got := err.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}
