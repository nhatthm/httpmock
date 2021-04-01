package httpmock

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/path", nil)
}

func newTestRequestWithBody(body string) *http.Request {
	return httptest.NewRequest(http.MethodGet, "/path", strings.NewReader(body))
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
				{Method: http.MethodGet},
			},
			expectedError: `Expected: GET 
Actual: POST /
Error: method "GET" expected, "POST" received
`,
		},
		{
			scenario: "uri is not matched",
			request:  newTestRequest(),
			expectations: []*Request{
				{Method: http.MethodGet, RequestURI: "/path2"},
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
					RequestURI:    "/path",
					RequestHeader: map[string]string{"Authorization": "Bearer token"},
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
			scenario: "body is not matched",
			request:  newTestRequestWithBody("body"),
			expectations: []*Request{
				{
					Method:      http.MethodGet,
					RequestURI:  "/path",
					RequestBody: []byte("expected body"),
				},
			},
			expectedError: `Expected: GET /path
    with body:
        expected body
Actual: GET /path
    with body:
        body
Error: expected request body: "expected body", received: "body"
`,
		},
		{
			scenario: "no repeatability",
			request:  newTestRequest(),
			expectations: []*Request{
				{Method: http.MethodGet, RequestURI: "/path"},
			},
			expectedRequest:  &Request{Method: http.MethodGet, RequestURI: "/path"},
			expectedRequests: []*Request{},
		},
		{
			scenario: "repeatability is 1",
			request:  newTestRequest(),
			expectations: []*Request{
				{Method: http.MethodGet, RequestURI: "/path", Repeatability: 1},
			},
			expectedRequest:  &Request{Method: http.MethodGet, RequestURI: "/path"},
			expectedRequests: []*Request{},
		},
		{
			scenario: "repeatability is 2",
			request:  newTestRequest(),
			expectations: []*Request{
				{Method: http.MethodGet, RequestURI: "/path", Repeatability: 2},
			},
			expectedRequest: &Request{Method: http.MethodGet, RequestURI: "/path", Repeatability: 1},
			expectedRequests: []*Request{
				{Method: http.MethodGet, RequestURI: "/path", Repeatability: 1},
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

func TestExactURIMatcher(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario    string
		expectation string
		uri         string
		expected    bool
	}{
		{
			scenario:    "equal",
			expectation: "example.org/path/to/file.txt",
			uri:         "example.org/path/to/file.txt",
			expected:    true,
		},
		{
			scenario:    "not equal",
			expectation: "example.org/path/to/file.txt",
			uri:         "example.org/path/to/file2.txt",
			expected:    false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			result := ExactURIMatcher()(tc.expectation, tc.uri)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExactHeaderMatcher(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario    string
		expectation string
		header      string
		expected    bool
	}{
		{
			scenario:    "equal",
			expectation: "Bearer token",
			header:      "Bearer token",
			expected:    true,
		},
		{
			scenario:    "not equal",
			expectation: "Bearer token",
			header:      "Bearer token2",
			expected:    false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			result := ExactHeaderMatcher()(tc.expectation, tc.header)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExactBodyMatcher(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario    string
		expectation string
		body        string
		expected    bool
	}{
		{
			scenario:    "equal",
			expectation: "body",
			body:        "body",
			expected:    true,
		},
		{
			scenario:    "not equal",
			expectation: "body",
			body:        "body 2",
			expected:    false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			result := ExactBodyMatcher()([]byte(tc.expectation), []byte(tc.body))

			assert.Equal(t, tc.expected, result)
		})
	}
}
