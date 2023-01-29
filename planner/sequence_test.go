package planner_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.nhat.io/httpmock/matcher"
	"go.nhat.io/httpmock/mock/http"
	plannermock "go.nhat.io/httpmock/mock/planner"
	"go.nhat.io/httpmock/planner"
)

func TestSequence(t *testing.T) {
	t.Parallel()

	const payload = `{"id": 42}`

	testCases := []struct {
		scenario        string
		request         *http.Request
		mockExpectation plannermock.ExpectationMocker
		expectedRequest bool
		expectedRemain  int
		expectedError   string
	}{
		{
			scenario: "method mismatched",
			request:  http.BuildRequest().WithMethod(http.MethodPost).Build(),
			mockExpectation: plannermock.MockExpectation(func(e *plannermock.Expectation) {
				e.On("Method").Maybe().Return(http.MethodGet)
				e.On("URIMatcher").Maybe().Return(matcher.Match("/"))
				e.On("HeaderMatcher").Maybe().Return(nil)
				e.On("BodyMatcher").Maybe().Return(nil)
			}),
			expectedRemain: 1,
			expectedError: `Expected: GET /
Actual: POST /
Error: method "GET" expected, "POST" received
`,
		},
		{
			scenario: "uri mismatched",
			request:  http.BuildRequest().WithURI("/users").Build(),
			mockExpectation: plannermock.MockExpectation(func(e *plannermock.Expectation) {
				e.On("Method").Maybe().Return(http.MethodGet)
				e.On("URIMatcher").Maybe().Return(matcher.Match("/"))
				e.On("HeaderMatcher").Maybe().Return(nil)
				e.On("BodyMatcher").Maybe().Return(nil)
			}),
			expectedRemain: 1,
			expectedError: `Expected: GET /
Actual: GET /users
Error: request uri "/" expected, "/users" received
`,
		},
		{
			scenario: "header mismatched",
			request: http.BuildRequest().
				WithHeader(`Authorization`, `Bearer foobar`).
				Build(),
			mockExpectation: plannermock.MockExpectation(func(e *plannermock.Expectation) {
				e.On("Method").Maybe().Return(http.MethodGet)
				e.On("URIMatcher").Maybe().Return(matcher.Match("/"))
				e.On("HeaderMatcher").Maybe().Return(matcher.HeaderMatcher{
					"Authorization": matcher.Match(`Bearer token`),
				})
				e.On("BodyMatcher").Maybe().Return(nil)
			}),
			expectedRemain: 1,
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
			scenario: "payload mismatched",
			request:  http.BuildRequest().Build(),
			mockExpectation: plannermock.MockExpectation(func(e *plannermock.Expectation) {
				e.On("Method").Maybe().Return(http.MethodGet)
				e.On("URIMatcher").Maybe().Return(matcher.Match("/"))
				e.On("HeaderMatcher").Maybe().Return(nil)
				e.On("BodyMatcher").Maybe().Return(matcher.Body(payload))
			}),
			expectedRemain: 1,
			expectedError: `Expected: GET /
    with body
        {"id": 42}
Actual: GET /
Error: expected request body: {"id": 42}, received: 
`,
		},
		{
			scenario: "success - unlimited times",
			request: http.BuildRequest().
				WithBody(payload).
				WithHeader("Authorization", "Bearer foobar").
				Build(),
			mockExpectation: plannermock.MockExpectation(func(e *plannermock.Expectation) {
				e.On("Method").Maybe().Return(http.MethodGet)
				e.On("URIMatcher").Maybe().Return(matcher.Match("/"))
				e.On("HeaderMatcher").Maybe().Return(matcher.HeaderMatcher{
					"Authorization": matcher.Match(regexp.MustCompile(`^Bearer `)),
				})
				e.On("BodyMatcher").Maybe().Return(matcher.Body(payload))
				e.On("RemainTimes").Return(uint(0))
			}),
			expectedRemain:  1,
			expectedRequest: true,
		},
		{
			scenario: "success - once",
			request: http.BuildRequest().
				WithBody(payload).
				WithHeader("Authorization", "Bearer foobar").
				Build(),
			mockExpectation: plannermock.MockExpectation(func(e *plannermock.Expectation) {
				e.On("Method").Maybe().Return(http.MethodGet)
				e.On("URIMatcher").Maybe().Return(matcher.Match("/"))
				e.On("HeaderMatcher").Maybe().Return(matcher.HeaderMatcher{
					"Authorization": matcher.Match(regexp.MustCompile(`^Bearer `)),
				})
				e.On("BodyMatcher").Maybe().Return(matcher.Body(payload))
				e.On("RemainTimes").Return(uint(1))
			}),
			expectedRequest: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			p := planner.Sequence()

			p.Expect(tc.mockExpectation(t))

			result, err := p.Plan(tc.request)
			remain := p.Remain()

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}

			assert.Equal(t, tc.expectedRequest, result != nil)
			assert.Len(t, remain, tc.expectedRemain)
		})
	}
}

func TestSequence_Empty(t *testing.T) {
	t.Parallel()

	p := planner.Sequence()

	assert.True(t, p.IsEmpty())

	p.Expect(plannermock.NoMockExpectation(t))

	assert.False(t, p.IsEmpty())

	p.Reset()

	assert.True(t, p.IsEmpty())
}

func TestSequence_Reset(t *testing.T) {
	t.Parallel()

	e := plannermock.NoMockExpectation(t)
	p := planner.Sequence()

	p.Expect(e)

	assert.Equal(t, []planner.Expectation{e}, p.Remain())

	p.Reset()

	assert.Empty(t, p.Remain())
}
