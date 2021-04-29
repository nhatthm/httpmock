package httpmock

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/stretchr/testify/assert"
	"github.com/swaggest/assertjson"
)

// Matcher determines if the actual matches the expectation.
type Matcher interface {
	Match(actual string) bool
	Expected() string
}

// ExactMatch matches by exact value.
type ExactMatch struct {
	expected string
}

// Expected returns the expectation.
func (m *ExactMatch) Expected() string {
	return m.expected
}

// Match determines if the actual is expected.
func (m *ExactMatch) Match(actual string) bool {
	return assert.ObjectsAreEqual(m.expected, actual)
}

// JSONMatch matches by json with <ignore-diff> support.
type JSONMatch struct {
	expected string
}

// Expected returns the expectation.
func (m *JSONMatch) Expected() string {
	return m.expected
}

// Match determines if the actual is expected.
func (m *JSONMatch) Match(actual string) bool {
	return assertjson.FailNotEqual([]byte(m.expected), []byte(actual)) == nil
}

// RegexMatch matches by regex.
type RegexMatch struct {
	regexp *regexp.Regexp
}

// Expected returns the expectation.
func (m *RegexMatch) Expected() string {
	return m.regexp.String()
}

// Match determines if the actual is expected.
func (m *RegexMatch) Match(actual string) bool {
	return m.regexp.MatchString(actual)
}

// CallbackMatch matches by calling a function.
type CallbackMatch struct {
	callback func() Matcher
	upstream Matcher
}

func (m *CallbackMatch) matcher() Matcher {
	if m.upstream == nil {
		m.upstream = m.callback()
	}

	return m.upstream
}

// Expected returns the expectation.
func (m *CallbackMatch) Expected() string {
	return m.matcher().Expected()
}

// Match determines if the actual is expected.
func (m *CallbackMatch) Match(actual string) bool {
	return m.matcher().Match(actual)
}

// Match creates a callback matcher.
func Match(callback func() Matcher) Matcher {
	return &CallbackMatch{
		callback: callback,
	}
}

// Exact matches two objects by their exact values.
func Exact(expected string) *ExactMatch {
	return &ExactMatch{expected: expected}
}

// JSON matches two json objects with <ignore-diff> support.
func JSON(expected string) *JSONMatch {
	return &JSONMatch{expected: expected}
}

// RegexPattern matches two objects by using regex.
func RegexPattern(pattern string) *RegexMatch {
	return &RegexMatch{regexp: regexp.MustCompile(pattern)}
}

// Regex matches two objects by using regex.
func Regex(regexp *regexp.Regexp) *RegexMatch {
	return &RegexMatch{regexp: regexp}
}

// ValueMatcher returns a matcher according to its type.
func ValueMatcher(v interface{}) Matcher {
	switch val := v.(type) {
	case Matcher:
		return val

	case func() Matcher:
		return Match(val)

	case []byte:
		return Exact(string(val))

	case string:
		return Exact(val)

	case *regexp.Regexp:
		return Regex(val)

	case fmt.Stringer:
		return Exact(val.String())

	default:
		panic(fmt.Errorf("%w: unexpected request body data type: %T", ErrUnsupportedDataType, v))
	}
}

// RequestMatcher matches a request with one of the expectations.
type RequestMatcher func(r *http.Request, expectations []*Request) (*Request, []*Request, error)

// SequentialRequestMatcher matches a request in sequence.
func SequentialRequestMatcher() RequestMatcher {
	return func(r *http.Request, expectedRequests []*Request) (*Request, []*Request, error) {
		expected := expectedRequests[0]

		if expected.Method != r.Method {
			return nil, nil, MatcherError(expected, r,
				"method %q expected, %q received", expected.Method, r.Method,
			)
		}

		if !expected.RequestURI.Match(r.RequestURI) {
			return nil, nil, MatcherError(expected, r,
				"request uri %q expected, %q received", expected.RequestURI.Expected(), r.RequestURI,
			)
		}

		for header, matcher := range expected.RequestHeader {
			value := r.Header.Get(header)
			if !matcher.Match(value) {
				return nil, nil, MatcherError(expected, r,
					"header %q with value %q expected, %q received", header, matcher.Expected(), value,
				)
			}
		}

		if expected.RequestBody != nil {
			body, err := GetBody(r)
			if err != nil {
				return nil, nil, MatcherError(expected, r, "could not read request body: %s", err.Error())
			}

			if !expected.RequestBody.Match(string(body)) {
				return nil, nil, MatcherError(expected, r, "expected request body: %q, received: %q", expected.RequestBody.Expected(), string(body))
			}
		}

		return expected, nextExpectations(expected, expectedRequests), nil
	}
}

func nextExpectations(expected *Request, expectedRequests []*Request) []*Request {
	if expected.Repeatability == 0 {
		return expectedRequests
	}

	if expected.Repeatability > 0 {
		expected.Repeatability--

		if expected.Repeatability > 0 {
			return expectedRequests
		}
	}

	return expectedRequests[1:]
}
