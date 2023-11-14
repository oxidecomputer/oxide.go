package oxide

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTPError is an error returned by a failed API call.
type HTTPError struct {
	// ErrorResponse is the API's Error response type.
	ErrorResponse *ErrorResponse

	// HTTPResponse is the raw HTTP response returned by the server.
	HTTPResponse *http.Response

	// RawBody is the raw response body returned by the server.
	RawBody string
}

// Error converts the HTTPError type to a readable string.
func (err HTTPError) Error() string {
	output := new(bytes.Buffer)
	fmt.Fprintf(output, "%s %s\n", err.HTTPResponse.Request.Method, err.HTTPResponse.Request.URL)

	fmt.Fprintln(output, "----------- RESPONSE -----------")
	if err.ErrorResponse != nil {
		fmt.Fprintf(output, "Status: %d %s\n", err.HTTPResponse.StatusCode, err.ErrorResponse.ErrorCode)
		fmt.Fprintf(output, "Message: %s\n", err.ErrorResponse.Message)
		fmt.Fprintf(output, "RequestID: %s\n", err.ErrorResponse.RequestId)
	} else {
		// In the very unlikely case that the error response was not able to be parsed
		// into oxide.Error, we will return the raw body of the response.
		fmt.Fprintf(output, "Status: %d\n", err.HTTPResponse.StatusCode)
		fmt.Fprintf(output, "Response Body: %v\n", err.RawBody)
	}

	fmt.Fprintln(output, "------- RESPONSE HEADERS -------")
	for k, v := range err.HTTPResponse.Header {
		fmt.Fprintf(output, "%s: %s\n", k, v)
	}

	return fmt.Sprintf("%v", output.String())
}

// TODO: export this function and change name
// checkResponse returns an error (of type *HTTPError) if the response
// status code is 3xx or greater.
func checkResponse(res *http.Response) error {
	if res.StatusCode <= 299 {
		return nil
	}

	// We want the API error to be returned, so in the unlikely case
	// that io.ReadAll returns with an error we'll leave it as an empty string
	slurp, _ := io.ReadAll(res.Body)

	e := HTTPError{
		HTTPResponse: res,
		RawBody:      string(slurp),
	}

	var apiError ErrorResponse
	if err := json.Unmarshal(slurp, &apiError); err != nil {
		// We return the error as is even with an unmarshal error,
		// as it already contains the RawBody.
		return e
	}

	e.ErrorResponse = &apiError

	return e
}
