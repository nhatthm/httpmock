package planner

import (
	"errors"
	"regexp"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.nhat.io/httpmock/matcher"
	"go.nhat.io/httpmock/mock/http"
	"go.nhat.io/httpmock/request"
)

func TestMatchURI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		uri           interface{}
		expectedError string
	}{
		{
			scenario: "match panic",
			uri: matcher.Fn("<panic>", func(interface{}) (bool, error) {
				panic("match panic")
			}),
			expectedError: `Expected: GET <panic>
Actual: GET /
Error: could not match request uri: match panic
`,
		},
		{
			scenario: "match error",
			uri: matcher.Fn("<error>", func(interface{}) (bool, error) {
				return false, errors.New("match error")
			}),
			expectedError: `Expected: GET <error>
Actual: GET /
Error: could not match request uri: match error
`,
		},
		{
			scenario: "mismatched",
			uri:      "/users",
			expectedError: `Expected: GET /users
Actual: GET /
Error: request uri "/users" expected, "/" received
`,
		},
		{
			scenario: "matched",
			uri:      "/",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			expected := request.NewRequest(&sync.Mutex{}, http.MethodGet, tc.uri)

			err := MatchURI(expected, http.BuildRequest().Build())

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestMatchHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		expect        func(r *request.Request)
		request       *http.Request
		expectedError string
	}{
		{
			scenario: "no header",
			expect:   func(r *request.Request) {},
		},
		{
			scenario: "match panic",
			expect: func(r *request.Request) {
				r.WithHeader("Authorization", matcher.Fn("<panic>", func(interface{}) (bool, error) {
					panic("match panic")
				}))
			},
			request: http.BuildRequest().Build(),
			expectedError: `Expected: GET /
    with header:
        Authorization: <panic>
Actual: GET /
Error: could not match header: match panic
`,
		},
		{
			scenario: "match error",
			expect: func(r *request.Request) {
				r.WithHeader("Authorization", matcher.Fn("<error>", func(interface{}) (bool, error) {
					return false, errors.New("match error")
				}))
			},
			request: http.BuildRequest().Build(),
			expectedError: `Expected: GET /
    with header:
        Authorization: <error>
Actual: GET /
Error: could not match header: match error
`,
		},
		{
			scenario: "mismatched",
			expect: func(r *request.Request) {
				r.WithHeader("Authorization", "Bearer token")
			},
			request: http.BuildRequest().
				WithHeader("Authorization", "Bearer foobar").
				Build(),
			expectedError: `Expected: GET /
    with header:
        Authorization: Bearer token
Actual: GET /
    with header:
        Authorization: Bearer foobar
Error: header "Authorization" with value "Bearer token" expected, "Bearer foobar" received
`,
		},
		{
			scenario: "matched",
			expect: func(r *request.Request) {
				r.WithHeader("Authorization", regexp.MustCompile(`^Bearer `))
			},
			request: http.BuildRequest().
				WithHeader("Authorization", "Bearer foobar").
				Build(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			expected := request.NewRequest(&sync.Mutex{}, http.MethodGet, "/")

			tc.expect(expected)

			err := MatchHeader(expected, tc.request)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestMatchBody(t *testing.T) {
	t.Parallel()

	const payload = `{"id":42}`

	testCases := []struct {
		scenario      string
		expect        func(r *request.Request)
		request       *http.Request
		expectedError string
	}{
		{
			scenario: "no expect",
			expect:   func(r *request.Request) {},
		},
		{
			scenario: "match panic",
			expect: func(r *request.Request) {
				r.WithBody(matcher.Fn("<panic>", func(interface{}) (bool, error) {
					panic("match panic")
				}))
			},
			request: http.BuildRequest().Build(),
			expectedError: `Expected: GET /
    with body
        <panic>
Actual: GET /
Error: could not match body: match panic
`,
		},
		{
			scenario: "match error",
			expect: func(r *request.Request) {
				r.WithBody(payload)
			},
			request: http.BuildRequest().
				WithBody(payload).
				WithBodyReadError(errors.New(`read error`)).
				Build(),
			expectedError: `Expected: GET /
    with body
        {"id":42}
Actual: GET /
    with body
        could not read request body: read error
Error: could not match body: read error
`,
		},
		{
			scenario: "mismatched",
			expect: func(r *request.Request) {
				r.WithBody(`{"id":1}`)
			},
			request: http.BuildRequest().
				WithBody(payload).
				Build(),
			expectedError: `Expected: GET /
    with body
        {"id":1}
Actual: GET /
    with body
        {"id":42}
Error: expected request body: {"id":1}, received: {"id":42}
`,
		},
		{
			scenario: "mismatched with empty expectation",
			expect: func(r *request.Request) {
				r.WithBody(``)
			},
			request: http.BuildRequest().
				WithBody(payload).
				Build(),
			expectedError: `Expected: GET /
Actual: GET /
    with body
        {"id":42}
Error: body does not match expectation, received: {"id":42}
`,
		},
		{
			scenario: "matched",
			request: http.BuildRequest().
				WithBody(payload).
				Build(),
			expect: func(r *request.Request) {
				r.WithBody(matcher.JSON(`{"id": 42}`))
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			expected := request.NewRequest(&sync.Mutex{}, http.MethodGet, "/")

			tc.expect(expected)

			err := MatchBody(expected, tc.request)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
