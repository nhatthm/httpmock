package httpmock

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type errReader struct{}

func (r errReader) Read([]byte) (int, error) {
	return 0, errors.New("read error")
}

func newTestRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/path", nil)
}

func newTestRequestWithBody(body string) *http.Request {
	return httptest.NewRequest(http.MethodGet, "/path", strings.NewReader(body))
}

func newTestRequestWithBodyError() *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/path", &errReader{})

	return r
}

func TestExactMatch_Expected(t *testing.T) {
	t.Parallel()

	m := Exact("value")
	expected := "value"

	assert.Equal(t, expected, m.Expected())
}

func TestExactMatch_Match(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		actual   string
		expected bool
	}{
		{
			scenario: "match",
			actual:   "value",
			expected: true,
		},
		{
			scenario: "no match",
			actual:   "mismatch",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			m := Exact("value")

			assert.Equal(t, tc.expected, m.Match(tc.actual))
		})
	}
}

func TestExactfMatch_Match(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		format   string
		args     []interface{}
		actual   string
		expected bool
	}{
		{
			scenario: "match",
			format:   "Bearer %s",
			args:     []interface{}{"token"},
			actual:   "Bearer token",
			expected: true,
		},
		{
			scenario: "no match",
			format:   "Bearer %s",
			args:     []interface{}{"token"},
			actual:   "Bearer unknown",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			m := Exactf(tc.format, tc.args...)

			assert.Equal(t, tc.expected, m.Match(tc.actual))
		})
	}
}

func TestJSONMatch_Expected(t *testing.T) {
	t.Parallel()

	m := JSON("{}")
	expected := "{}"

	assert.Equal(t, expected, m.Expected())
}

func TestJSONMatch_Match(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		json     string
		actual   string
		expected bool
	}{
		{
			scenario: "match",
			json: `{
	"username": "user"
}`,
			actual:   `{"username": "user"}`,
			expected: true,
		},
		{
			scenario: "match with <ignore-diff>",
			json:     `{"username": "<ignore-diff>"}`,
			actual:   `{"username": "user"}`,
			expected: true,
		},
		{
			scenario: "no match",
			json:     "{}",
			actual:   "[]",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			m := JSON(tc.json)

			assert.Equal(t, tc.expected, m.Match(tc.actual))
		})
	}
}

func TestRegexMatch_Expected(t *testing.T) {
	t.Parallel()

	m := Regex(regexp.MustCompile(".*"))
	expected := ".*"

	assert.Equal(t, expected, m.Expected())
}

func TestRegexMatch_Match(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		matcher  *RegexMatch
		actual   string
		expected bool
	}{
		{
			scenario: "match with regexp",
			matcher:  Regex(regexp.MustCompile(".*")),
			actual:   `hello`,
			expected: true,
		},
		{
			scenario: "match with regexp pattern",
			matcher:  RegexPattern(".*"),
			actual:   `hello`,
			expected: true,
		},
		{
			scenario: "no match",
			matcher:  RegexPattern("^[0-9]+$"),
			actual:   "mismatch",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.matcher.Match(tc.actual))
		})
	}
}

func TestValueMatcher(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		value    interface{}
		expected Matcher
	}{
		{
			scenario: "matcher",
			value:    Exact("expected"),
			expected: Exact("expected"),
		},
		{
			scenario: "[]byte",
			value:    []byte("expected"),
			expected: Exact("expected"),
		},
		{
			scenario: "string",
			value:    "expected",
			expected: Exact("expected"),
		},
		{
			scenario: "regexp",
			value:    regexp.MustCompile(".*"),
			expected: RegexPattern(".*"),
		},
		{
			scenario: "fmt.Stringer",
			value:    time.UTC,
			expected: Exact("UTC"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, ValueMatcher(tc.value))
		})
	}
}

func TestValueMatcher_Match(t *testing.T) {
	t.Parallel()

	m := ValueMatcher(func() Matcher {
		return Exact("expected")
	})

	assert.Equal(t, "expected", m.Expected())
	assert.True(t, m.Match("expected"))
	assert.False(t, m.Match("Mismatch"))
}

func TestValueMatcher_Panic(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() {
		ValueMatcher(42)
	})
}

func TestSequentialRequestMatcher(t *testing.T) {
	t.Parallel()

	requestWithWrongHeader := newTestRequest()
	requestWithWrongHeader.Header.Set("Authorization", "token")

	testCases := []struct {
		scenario         string
		request          *http.Request
		expectations     []*Request
		expectedRequest  *Request
		expectedRequests []*Request
		expectedError    string
	}{
		{
			scenario: "method is not matched",
			request:  httptest.NewRequest(http.MethodPost, "/", nil),
			expectations: []*Request{
				{Method: http.MethodGet, RequestURI: Exact("/")},
			},
			expectedError: `Expected: GET /
Actual: POST /
Error: method "GET" expected, "POST" received
`,
		},
		{
			scenario: "uri is not matched",
			request:  newTestRequest(),
			expectations: []*Request{
				{Method: http.MethodGet, RequestURI: Exact("/path2")},
			},
			expectedError: `Expected: GET /path2
Actual: GET /path
Error: request uri "/path2" expected, "/path" received
`,
		},
		{
			scenario: "header is not matched",
			request:  requestWithWrongHeader,
			expectations: []*Request{
				{
					Method:        http.MethodGet,
					RequestURI:    Exact("/path"),
					RequestHeader: HeaderMatcher{"Authorization": Exact("Bearer token")},
				},
			},
			expectedError: `Expected: GET /path
    with header:
        Authorization: Bearer token
Actual: GET /path
    with header:
        Authorization: token
Error: header "Authorization" with value "Bearer token" expected, "token" received
`,
		},
		{
			scenario: "could not read body",
			request:  newTestRequestWithBodyError(),
			expectations: []*Request{
				{
					Method:      http.MethodGet,
					RequestURI:  Exact("/path"),
					RequestBody: Exact("expected body"),
				},
			},
			expectedError: `Expected: GET /path
    with body
        expected body
Actual: GET /path
    with body
        could not read request body: read error
Error: could not read request body: read error
`,
		},
		{
			scenario: "body is not matched",
			request:  newTestRequestWithBody("body"),
			expectations: []*Request{
				{
					Method:      http.MethodGet,
					RequestURI:  Exact("/path"),
					RequestBody: Exact("expected body"),
				},
			},
			expectedError: `Expected: GET /path
    with body
        expected body
Actual: GET /path
    with body
        body
Error: expected request body: "expected body", received: "body"
`,
		},
		{
			scenario: "unlimited repeatability",
			request:  newTestRequest(),
			expectations: []*Request{
				{Method: http.MethodGet, RequestURI: Exact("/path")},
			},
			expectedRequest: &Request{Method: http.MethodGet, RequestURI: Exact("/path")},
			expectedRequests: []*Request{
				{Method: http.MethodGet, RequestURI: Exact("/path")},
			},
		},
		{
			scenario: "repeatability is 1",
			request:  newTestRequest(),
			expectations: []*Request{
				{Method: http.MethodGet, RequestURI: Exact("/path"), Repeatability: 1},
			},
			expectedRequest:  &Request{Method: http.MethodGet, RequestURI: Exact("/path")},
			expectedRequests: []*Request{},
		},
		{
			scenario: "repeatability is 2",
			request:  newTestRequest(),
			expectations: []*Request{
				{Method: http.MethodGet, RequestURI: Exact("/path"), Repeatability: 2},
			},
			expectedRequest: &Request{Method: http.MethodGet, RequestURI: Exact("/path"), Repeatability: 1},
			expectedRequests: []*Request{
				{Method: http.MethodGet, RequestURI: Exact("/path"), Repeatability: 1},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			expected, expectedRequests, err := SequentialRequestMatcher()(tc.request, tc.expectations)

			assert.Equal(t, tc.expectedRequest, expected)
			assert.Equal(t, tc.expectedRequests, expectedRequests)

			if tc.expectedError == "" {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
