package oxide

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ErrorCode represents an Oxide API error matched by error_code. Use `errors.Is` to test SDK
// errors against known error codes.
//
// These error codes are derived from the Error enum in omicron:
// https://github.com/oxidecomputer/omicron/blob/0ef43032/common/src/api/external/error.rs
//
// Some upstream variants (e.g. TypeVersionMismatch) are omitted because they serialize
// to the same error_code as another variant and can't be disambiguated. We also omit the "Not
// Found" code as described below.
//
// TODO: Encode error types in the OpenAPI spec upstream so that we don't have to maintain this
// mapping in the SDK.
type ErrorCode struct {
	code string
}

func (e *ErrorCode) Error() string {
	return e.code
}

var (
	// ErrObjectNotFound indicates that the requested resource doesn't exist. Note that this doesn't
	// cover all possible 404 errors from the API. Omicron also uses a generic "Not Found" error
	// code to represent non-resource not found errors, such as SCIM authz errors and errors from
	// internal endpoints. The web server also returns a 404 without an error code when the request
	// path doesn't match any known route. However, these errors are semantically distinct from
	// ErrObjectNotFound, and represent internal errors that the SDK shouldn't be concerned with.
	ErrObjectNotFound       error = &ErrorCode{"ObjectNotFound"}
	ErrObjectAlreadyExists  error = &ErrorCode{"ObjectAlreadyExists"}
	ErrInvalidRequest       error = &ErrorCode{"InvalidRequest"}
	ErrInvalidValue         error = &ErrorCode{"InvalidValue"}
	ErrUnauthenticated      error = &ErrorCode{"Unauthorized"}
	ErrForbidden            error = &ErrorCode{"Forbidden"}
	ErrInternalError        error = &ErrorCode{"Internal"}
	ErrServiceUnavailable   error = &ErrorCode{"ServiceNotAvailable"}
	ErrInsufficientCapacity error = &ErrorCode{"InsufficientCapacity"}
	ErrConflict             error = &ErrorCode{"Conflict"}
	ErrGone                 error = &ErrorCode{"Gone"}
)

// StatusError represents an Oxide API error matched by HTTP status code. Use `errors.Is` to
// test SDK errors against known statuses.
type StatusError struct {
	status int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("HTTP %d", e.status)
}

var (
	ErrHTTP400 error = &StatusError{400}
	ErrHTTP401 error = &StatusError{401}
	ErrHTTP403 error = &StatusError{403}
	ErrHTTP404 error = &StatusError{404}
	ErrHTTP409 error = &StatusError{409}
	ErrHTTP410 error = &StatusError{410}
	ErrHTTP500 error = &StatusError{500}
	ErrHTTP503 error = &StatusError{503}
	ErrHTTP507 error = &StatusError{507}
)

// Is implements errors.Is. We allow testing against both error code and http status errors.
func (e *HTTPError) Is(target error) bool {
	switch t := target.(type) {
	case *ErrorCode:
		return e.ErrorResponse != nil && e.ErrorResponse.ErrorCode == t.code
	case *StatusError:
		return e.HTTPResponse != nil && e.HTTPResponse.StatusCode == t.status
	}
	return false
}

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
	if err.HTTPResponse.Request.URL != nil {
		fmt.Fprintf(
			output,
			"%s %s\n",
			err.HTTPResponse.Request.Method,
			err.HTTPResponse.Request.URL,
		)
	} else {
		// This case is extremely unlikely, just adding to avoid a panic due to a nil pointer
		fmt.Fprintf(output, "%s <URL unavailable>\n", err.HTTPResponse.Request.Method)
	}
	fmt.Fprintln(output, "----------- RESPONSE -----------")
	if err.ErrorResponse != nil {
		fmt.Fprintf(
			output,
			"Status: %d %s\n",
			err.HTTPResponse.StatusCode,
			err.ErrorResponse.ErrorCode,
		)
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

// NewHTTPError returns an error of type *HTTPError if the response
// status code is 3xx or greater.
func NewHTTPError(res *http.Response) error {
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
		return &e
	}

	e.ErrorResponse = &apiError

	return &e
}
