package httpmock

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const indent = "    "

// ErrUnsupportedDataType represents that the data type is not supported.
var ErrUnsupportedDataType = errors.New("unsupported data type")

// RequestMatcherError represents an error that occurs while matching a request.
type RequestMatcherError struct {
	expected *Request
	actual   *http.Request
	message  string
	args     []interface{}
}

func (e RequestMatcherError) formatExpected(w io.Writer) {
	formatExpectedRequest(w, e.expected.Method, e.expected.RequestURI, e.expected.RequestHeader, e.expected.RequestBody)
}

func (e RequestMatcherError) formatActual(w io.Writer) {
	body, err := GetBody(e.actual)
	requireNoErr(err)

	formatHTTPRequest(w, e.actual.Method, e.actual.RequestURI, e.actual.Header, body)
}

// Error satisfies the error interface.
func (e RequestMatcherError) Error() string {
	var sb strings.Builder

	_, _ = fmt.Fprint(&sb, "Expected: ")
	e.formatExpected(&sb)
	_, _ = fmt.Fprint(&sb, "Actual: ")
	e.formatActual(&sb)
	_, _ = fmt.Fprint(&sb, "Error: ")
	_, _ = fmt.Fprintf(&sb, e.message, e.args...)
	_, _ = fmt.Fprint(&sb, "\n")

	return sb.String()
}

// MatcherError instantiates a new RequestMatcherError.
func MatcherError(expected *Request, request *http.Request, message string, args ...interface{}) *RequestMatcherError {
	return &RequestMatcherError{
		expected: expected,
		actual:   request,
		message:  message,
		args:     args,
	}
}
