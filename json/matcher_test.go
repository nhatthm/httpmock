package json_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nhatthm/httpmock"
	"github.com/nhatthm/httpmock/json"
	"github.com/stretchr/testify/assert"
)

func newTestRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/path", nil)
}

func newTestRequestWithBody(body string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/path", strings.NewReader(body))
}

func TestSequentialRequestMatcher(t *testing.T) {
	t.Parallel()

	requestWithWrongHeader := newTestRequest()
	requestWithWrongHeader.Header.Set("Authorization", "token")

	testCases := []struct {
		scenario         string
		request          *http.Request
		expectations     []*httpmock.Request
		expectedRequest  *httpmock.Request
		expectedRequests []*httpmock.Request
		expectedError    string
	}{
		{
			scenario: "method is not matched",
			request:  httptest.NewRequest(http.MethodPost, "/", nil),
			expectations: []*httpmock.Request{
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
			expectations: []*httpmock.Request{
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
			expectations: []*httpmock.Request{
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
			expectations: []*httpmock.Request{
				{
					Method:      http.MethodPost,
					RequestURI:  "/path",
					RequestBody: []byte("expected body"),
				},
			},
			expectedError: `Expected: POST /path
    with body:
        expected body
Actual: POST /path
    with body:
        body
Error: expected request body: "expected body", received: "body"
`,
		},
		{
			scenario: "body has <ignored-diff> and not equal",
			request:  newTestRequestWithBody(`{"value": 2, "diff": "foobar"}`),
			expectations: []*httpmock.Request{
				{
					Method:      http.MethodPost,
					RequestURI:  "/path",
					RequestBody: []byte(`{"value": 1, "diff": "<ignore-diff>"}`),
				},
			},
			expectedError: `Expected: POST /path
    with body:
        {"value": 1, "diff": "<ignore-diff>"}
Actual: POST /path
    with body:
        {"value": 2, "diff": "foobar"}
Error: expected request body: "{\"value\": 1, \"diff\": \"<ignore-diff>\"}", received: "{\"value\": 2, \"diff\": \"foobar\"}"
`,
		},
		{
			scenario: "body has <ignored-diff> and equal",
			request:  newTestRequestWithBody(`{"value": 1, "diff": "foobar"}`),
			expectations: []*httpmock.Request{
				{
					Method:      http.MethodPost,
					RequestURI:  "/path",
					RequestBody: []byte(`{"value": 1, "diff": "<ignore-diff>"}`),
				},
			},
			expectedRequest: &httpmock.Request{
				Method:      http.MethodPost,
				RequestURI:  "/path",
				RequestBody: []byte(`{"value": 1, "diff": "<ignore-diff>"}`),
			},
			expectedRequests: []*httpmock.Request{},
		},
		{
			scenario: "no repeatability",
			request:  newTestRequest(),
			expectations: []*httpmock.Request{
				{Method: http.MethodGet, RequestURI: "/path"},
			},
			expectedRequest:  &httpmock.Request{Method: http.MethodGet, RequestURI: "/path"},
			expectedRequests: []*httpmock.Request{},
		},
		{
			scenario: "repeatability is 1",
			request:  newTestRequest(),
			expectations: []*httpmock.Request{
				{Method: http.MethodGet, RequestURI: "/path", Repeatability: 1},
			},
			expectedRequest:  &httpmock.Request{Method: http.MethodGet, RequestURI: "/path"},
			expectedRequests: []*httpmock.Request{},
		},
		{
			scenario: "repeatability is 2",
			request:  newTestRequest(),
			expectations: []*httpmock.Request{
				{Method: http.MethodGet, RequestURI: "/path", Repeatability: 2},
			},
			expectedRequest: &httpmock.Request{Method: http.MethodGet, RequestURI: "/path", Repeatability: 1},
			expectedRequests: []*httpmock.Request{
				{Method: http.MethodGet, RequestURI: "/path", Repeatability: 1},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			expected, expectedRequests, err := json.RequestMatcher()(t, tc.request, tc.expectations)

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
