package planner

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.nhat.io/httpmock/format"
	"go.nhat.io/httpmock/value"
)

// Error represents an error that occurs while matching a request.
type Error struct {
	expected Expectation
	actual   *http.Request

	messageFormat string
	messageArgs   []any
}

func (e Error) formatExpected(w io.Writer) {
	format.ExpectedRequest(w,
		e.expected.Method(),
		e.expected.URIMatcher(),
		e.expected.HeaderMatcher(),
		e.expected.BodyMatcher(),
	)
}

func (e Error) formatActual(w io.Writer) {
	body, err := value.GetBody(e.actual)
	if err != nil {
		body = []byte(fmt.Sprintf("could not read request body: %s", err.Error()))
	}

	format.HTTPRequest(w, e.actual.Method, e.actual.RequestURI, e.actual.Header, body)
}

// Error satisfies the error interface.
func (e Error) Error() string {
	var sb strings.Builder

	_, _ = fmt.Fprint(&sb, "Expected: ")
	e.formatExpected(&sb)
	_, _ = fmt.Fprint(&sb, "Actual: ")
	e.formatActual(&sb)
	_, _ = fmt.Fprint(&sb, "Error: ")
	_, _ = fmt.Fprintf(&sb, e.messageFormat, e.messageArgs...)
	_, _ = fmt.Fprint(&sb, "\n")

	return sb.String()
}

// NewError creates a new Error.
func NewError(expected Expectation, request *http.Request, messageFormat string, messageArgs ...any) *Error {
	return &Error{
		expected:      expected,
		actual:        request,
		messageFormat: messageFormat,
		messageArgs:   messageArgs,
	}
}
