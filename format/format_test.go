package format_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.nhat.io/httpmock/format"
	"go.nhat.io/httpmock/matcher"
)

func TestExpectedRequest(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	uri := matcher.Exact("/users")

	header := matcher.HeaderMatcher{
		"Authorization": matcher.Match(`Bearer token`),
	}

	body := matcher.Body(matcher.JSON(`{"id": 42}`))

	format.ExpectedRequest(buf, http.MethodGet, uri, header, body)

	assert.Equal(t, expectedStringWithoutTimes(), buf.String())
}

func TestExpectedRequestTimes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario       string
		totalCalls     int
		remainingCalls int
		expected       string
	}{
		{
			scenario:       "0 call, remain 0",
			totalCalls:     0,
			remainingCalls: 0,
			expected:       expectedStringWithoutTimes(),
		},
		{
			scenario:       "0 call, remain 1",
			totalCalls:     0,
			remainingCalls: 1,
			expected:       expectedStringWithoutTimes(),
		},
		{
			scenario:       "0 call, remain 2",
			totalCalls:     0,
			remainingCalls: 2,
			expected:       expectedStringWithTimes(0, 2),
		},
		{
			scenario:       "1 call, remain 0",
			totalCalls:     1,
			remainingCalls: 0,
			expected:       expectedStringWithoutTimes(),
		},
		{
			scenario:       "1 call, remain 1",
			totalCalls:     1,
			remainingCalls: 1,
			expected:       expectedStringWithTimes(1, 1),
		},
		{
			scenario:       "1 call, remain 2",
			totalCalls:     1,
			remainingCalls: 2,
			expected:       expectedStringWithTimes(1, 2),
		},
		{
			scenario:       "2 call, remain 0",
			totalCalls:     2,
			remainingCalls: 0,
			expected:       expectedStringWithoutTimes(),
		},
		{
			scenario:       "2 call, remain 1",
			totalCalls:     2,
			remainingCalls: 1,
			expected:       expectedStringWithTimes(2, 1),
		},
		{
			scenario:       "2 call, remain 2",
			totalCalls:     2,
			remainingCalls: 2,
			expected:       expectedStringWithTimes(2, 2),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			buf := new(bytes.Buffer)
			uri := matcher.Exact("/users")

			header := matcher.HeaderMatcher{
				"Authorization": matcher.Match(`Bearer token`),
			}

			body := matcher.Body(matcher.JSON(`{"id": 42}`))

			format.ExpectedRequestTimes(buf, http.MethodGet, uri, header, body, tc.totalCalls, tc.remainingCalls)

			assert.Equal(t, tc.expected, buf.String())
		})
	}
}

func TestHTTPRequest(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	uri := "/users"

	header := http.Header{
		"Authorization": {`Bearer token`},
	}

	body := []byte(`{"id": 42}`)

	format.HTTPRequest(buf, http.MethodGet, uri, header, body)

	assert.Equal(t, expectedStringWithoutMatcher(), buf.String())
}

func expectedStringWithoutMatcher() string {
	return `GET /users
    with header:
        Authorization: Bearer token
    with body
        {"id": 42}
`
}

func expectedStringWithoutTimes() string {
	return `GET /users
    with header:
        Authorization: Bearer token
    with body using matcher.JSONMatcher
        {"id": 42}
`
}

func expectedStringWithTimes(called, remaining int) string {
	return fmt.Sprintf(`GET /users (called: %d time(s), remaining: %d time(s))
    with header:
        Authorization: Bearer token
    with body using matcher.JSONMatcher
        {"id": 42}
`, called, remaining)
}
