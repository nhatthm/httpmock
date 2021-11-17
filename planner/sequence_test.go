package planner

import (
	"regexp"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/httpmock/matcher"
	"github.com/nhatthm/httpmock/mock/http"
	"github.com/nhatthm/httpmock/request"
)

func TestSequence(t *testing.T) {
	t.Parallel()

	const payload = `{"id": 42}`

	testCases := []struct {
		scenario        string
		request         *http.Request
		expectations    []*request.Request
		expectedRequest bool
		expectedRemain  int
		expectedError   string
	}{
		{
			scenario: "urmethodi mismatched",
			request:  http.BuildRequest().WithMethod(http.MethodPost).Build(),
			expectations: []*request.Request{
				request.NewRequest(&sync.Mutex{}, http.MethodGet, "/").Once(),
			},
			expectedRemain: 1,
			expectedError: `Expected: GET /
Actual: POST /
Error: method "GET" expected, "POST" received
`,
		},
		{
			scenario: "uri mismatched",
			request:  http.BuildRequest().WithURI("/users").Build(),
			expectations: []*request.Request{
				request.NewRequest(&sync.Mutex{}, http.MethodGet, "/").Once(),
			},
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
			expectations: []*request.Request{
				request.NewRequest(&sync.Mutex{}, http.MethodGet, "/").Once().
					WithHeader("Authorization", `Bearer token`),
			},
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
			expectations: []*request.Request{
				request.NewRequest(&sync.Mutex{}, http.MethodGet, "/").Once().
					WithBody(payload),
			},
			expectedRemain: 1,
			expectedError: `Expected: GET /
    with body
        {"id": 42}
Actual: GET /
Error: expected request body: {"id": 42}, received: 
`,
		},
		{
			scenario: "success",
			request: http.BuildRequest().
				WithBody(payload).
				WithHeader("Authorization", "Bearer foobar").
				Build(),
			expectations: []*request.Request{
				request.NewRequest(&sync.Mutex{}, http.MethodGet, "/").Once().
					WithHeader("Authorization", regexp.MustCompile(`^Bearer `)).
					WithBody(matcher.JSON(`{"id":42}`)),
			},
			expectedRequest: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			p := Sequence()

			for _, r := range tc.expectations {
				p.Expect(r)
			}

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

	p := Sequence()

	assert.True(t, p.IsEmpty())

	p.Expect(request.NewRequest(nil, http.MethodGet, "/"))

	assert.False(t, p.IsEmpty())

	p.Reset()

	assert.True(t, p.IsEmpty())
}
