package oxide

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// HTTPError is an error returned by a failed API call.
type HTTPError struct {
	// URL is the URL that was being accessed when the error occurred.
	// It will always be populated.
	URL *url.URL

	// StatusCode is the HTTP response status code and will always be populated.
	StatusCode int

	// RequestMethod is the HTTP request method and will always be populated.
	RequestMethod string

	// ErrorResponse is the API's Error response type.
	ErrorResponse *Error

	// RawBody is the raw response returned by the server.
	RawBody string

	// Header contains the response header fields from the server.
	Header http.Header
}

// Error converts the HTTPError type to a readable string.
func (err HTTPError) Error() string {
	output := new(bytes.Buffer)
	fmt.Fprintln(output, "\n------- REQUEST -------")
	fmt.Fprintf(output, "%s %s\n", err.RequestMethod, err.URL)

	for k, v := range err.Header {
		fmt.Fprintf(output, "%s: %s\n", k, v)
	}

	fmt.Fprintln(output, "------- RESPONSE -------")
	if err.ErrorResponse != nil {
		fmt.Fprintf(output, "Status: %d %s\n", err.StatusCode, err.ErrorResponse.ErrorCode)
		fmt.Fprintf(output, "Message: %s\n", err.ErrorResponse.Message)
		fmt.Fprintf(output, "RequestID: %s\n", err.ErrorResponse.RequestId)
	} else {
		// In the very unlikely case that the error response was not able to be parsed
		// into oxide.Error, we will return the raw body of the response.
		fmt.Fprintf(output, "Status: %d\n", err.StatusCode)
		fmt.Fprintf(output, "Response Body: %v\n", err.RawBody)
	}

	return fmt.Sprintf("%v", output.String())
}

// checkResponse returns an error (of type *HTTPError) if the response
// status code is 3xx or greater.
func checkResponse(res *http.Response) error {
	if res.StatusCode <= 300 {
		return nil
	}

	// We want the API error to be returned, so in the unlikely case
	// that io.ReadAll returns with an error we'll leave it as an empty string
	slurp, _ := io.ReadAll(res.Body)

	e := HTTPError{
		RequestMethod: res.Request.Method,
		StatusCode:    res.StatusCode,
		URL:           res.Request.URL,
		Header:        res.Header,
		RawBody:       string(slurp),
	}

	var apiError Error
	if err := json.Unmarshal([]byte(slurp), &apiError); err != nil {
		// We return the error as is even with an unmarshal error,
		// as it already contains the RawBody.
		return e
	}

	e.ErrorResponse = &apiError

	return e
}
