package planner_test

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.nhat.io/httpmock/matcher"
	"go.nhat.io/httpmock/mock/http"
	plannermock "go.nhat.io/httpmock/mock/planner"
	"go.nhat.io/httpmock/planner"
)

func TestMatchURI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		uri           any
		expectedError string
	}{
		{
			scenario: "match panic",
			uri: matcher.Fn("<panic>", func(any) (bool, error) {
				panic("match panic")
			}),
			expectedError: `Expected: GET <panic>
Actual: GET /
Error: could not match request uri: match panic
`,
		},
		{
			scenario: "match error",
			uri: matcher.Fn("<error>", func(any) (bool, error) {
				return false, errors.New("match error")
			}),
			expectedError: `Expected: GET <error>
Actual: GET /
Error: could not match request uri: match error
`,
		},
		{
			scenario: "mismatched",
			uri:      matcher.Match("/users"),
			expectedError: `Expected: GET /users
Actual: GET /
Error: request uri "/users" expected, "/" received
`,
		},
		{
			scenario: "matched",
			uri:      matcher.Match("/"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			expected := plannermock.MockExpectation(func(e *plannermock.Expectation) {
				e.On("URIMatcher").Return(tc.uri)
				e.On("Method").Maybe().Return(http.MethodGet)
				e.On("HeaderMatcher").Maybe().Return(nil)
				e.On("BodyMatcher").Maybe().Return(nil)
			})(t)

			err := planner.MatchURI(expected, http.BuildRequest().Build())

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
		headerMatcher matcher.HeaderMatcher
		request       *http.Request
		expectedError string
	}{
		{
			scenario: "no header",
		},
		{
			scenario: "match panic",
			headerMatcher: map[string]matcher.Matcher{
				"Authorization": matcher.Fn("<panic>", func(any) (bool, error) {
					panic("match panic")
				}),
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
			headerMatcher: map[string]matcher.Matcher{
				"Authorization": matcher.Fn("<error>", func(any) (bool, error) {
					return false, errors.New("match error")
				}),
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
			headerMatcher: map[string]matcher.Matcher{
				"Authorization": matcher.Match("Bearer token"),
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
			headerMatcher: map[string]matcher.Matcher{
				"Authorization": matcher.Match(regexp.MustCompile(`^Bearer `)),
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

			expected := plannermock.MockExpectation(func(e *plannermock.Expectation) {
				e.On("HeaderMatcher").Return(tc.headerMatcher)
				e.On("URIMatcher").Maybe().Return(matcher.Match("/"))
				e.On("Method").Maybe().Return(http.MethodGet)
				e.On("BodyMatcher").Maybe().Return(nil)
			})(t)

			err := planner.MatchHeader(expected, tc.request)

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
		bodyMatcher   *matcher.BodyMatcher
		request       *http.Request
		expectedError string
	}{
		{
			scenario: "no expect",
		},
		{
			scenario: "match panic",
			bodyMatcher: matcher.Body(matcher.Fn("<panic>", func(any) (bool, error) {
				panic("match panic")
			})),
			request: http.BuildRequest().Build(),
			expectedError: `Expected: GET /
    with body
        <panic>
Actual: GET /
Error: could not match body: match panic
`,
		},
		{
			scenario:    "match error",
			bodyMatcher: matcher.Body(payload),
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
			scenario:    "mismatched",
			bodyMatcher: matcher.Body(`{"id":1}`),
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
			scenario:    "mismatched with empty expectation",
			bodyMatcher: matcher.Body(``),
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
			scenario:    "matched",
			bodyMatcher: matcher.Body(matcher.JSON(`{"id": 42}`)),
			request: http.BuildRequest().
				WithBody(payload).
				Build(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			expected := plannermock.MockExpectation(func(e *plannermock.Expectation) {
				e.On("BodyMatcher").Return(tc.bodyMatcher)
				e.On("URIMatcher").Maybe().Return(matcher.Match("/"))
				e.On("Method").Maybe().Return(http.MethodGet)
				e.On("HeaderMatcher").Maybe().Return(nil)
			})(t)

			err := planner.MatchBody(expected, tc.request)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
