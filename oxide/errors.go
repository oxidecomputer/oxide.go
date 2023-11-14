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
	// Message is the server response message and is only populated when
	// explicitly referenced by the JSON server response.
	//Message string
	// TODO: Add request method
	//RequestMethod string

	// Body is the raw response returned by the server.
	// It is often but not always JSON, depending on how the request fails.
	Body string
	// Header contains the response header fields from the server.
	Header http.Header
}

// APIError is an error returned by a failed API call.
type APIError struct {
	// URL is the URL that was being accessed when the error occurred.
	// It will always be populated.
	URL *url.URL

	// StatusCode is the HTTP response status code and will always be populated.
	StatusCode int

	RequestMethod string

	ErrorResponse Error

	// Header contains the response header fields from the server.
	Header http.Header
}

// Error converts the Error type to a readable string.
func (err HTTPError) Error() string {
	//	if err.Message != "" {
	//		return fmt.Sprintf("HTTP %d: %s (%s)", err.StatusCode, err.Message, err.URL)
	//	}

	return fmt.Sprintf("HTTP %d (%s) BODY -> %v", err.StatusCode, err.URL, err.Body)
}

// Error converts the Error type to a readable string.
func (err APIError) Error() string {
	output := new(bytes.Buffer)
	fmt.Fprintln(output, "\n------- REQUEST -------")
	fmt.Fprintf(output, "%s %s\n", err.RequestMethod, err.URL)

	//	fmt.Fprintln(output, "------- HEADERS -------")
	for k, v := range err.Header {
		//fmt.Println(k, v)
		fmt.Fprintf(output, "%s: %s\n", k, v)
	}

	fmt.Fprintln(output, "------- RESPONSE -------")
	fmt.Fprintf(output, "Status: %d %s\n", err.StatusCode, err.ErrorResponse.ErrorCode)
	fmt.Fprintf(output, "Message: %s\n", err.ErrorResponse.Message)
	fmt.Fprintf(output, "RequestID: %s\n", err.ErrorResponse.RequestId)

	return fmt.Sprintf("%v", output.String())
}

// checkResponse returns an error (of type *HTTPError) if the response
// status code is not 2xx.
func checkResponse(res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}

	slurp, _ := io.ReadAll(res.Body)

	var apiError Error
	if err := json.Unmarshal([]byte(slurp), &apiError); err != nil {
		return &HTTPError{
			URL:        res.Request.URL,
			StatusCode: res.StatusCode,
			Body:       string(slurp),
			Header:     res.Header,
			//	Message:    "",
		}
	}

	return &APIError{
		RequestMethod: res.Request.Method,
		StatusCode:    res.StatusCode,
		URL:           res.Request.URL,
		Header:        res.Header,
		ErrorResponse: apiError,
	}
}
