package matcher_test

import (
	"errors"
	"io"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/httpmock/matcher"
	"github.com/nhatthm/httpmock/mock/http"
)

func TestBodyMatcher_Match(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario       string
		body           interface{}
		request        *http.Request
		expectedResult bool
		expectedActual string
		expectedError  error
	}{
		{
			scenario: "decode error",
			request: http.BuildRequest().
				WithBodyReadError(errors.New("read error")).
				Build(),
			expectedActual: `<could not decode>`,
			expectedError:  errors.New(`read error`),
		},
		{
			scenario:       "mismatched",
			body:           "foobar",
			expectedActual: `foo`,
			request: http.BuildRequest().
				WithBody(`foo`).
				Build(),
		},
		{
			scenario: "mismatched",
			body:     regexp.MustCompile(`^hello`),
			request: http.BuildRequest().
				WithBody("hello world").
				Build(),
			expectedResult: true,
			expectedActual: `hello world`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			m := matcher.Body(tc.body)
			matched, err := m.Match(tc.request)

			assert.Equal(t, tc.expectedResult, matched)
			assert.Equal(t, tc.expectedActual, m.Actual())
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestBodyMatcher_Match_ReuseBody(t *testing.T) {
	t.Parallel()

	expected := `foobar`

	r := http.BuildRequest().WithBody(expected).Build()
	m := matcher.Body(expected)
	_, err := m.Match(r)

	assert.NoError(t, err)

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close() // nolint: errcheck

	assert.Equal(t, expected, string(body))
	assert.NoError(t, err)
}

func TestBodyMatcher_Matcher(t *testing.T) {
	t.Parallel()

	m := matcher.Body(`payload`)

	assert.Equal(t, matcher.Match(`payload`), m.Matcher())
}

func TestBodyMatcher_Expected(t *testing.T) {
	t.Parallel()

	expected := `foobar`
	m := matcher.Body(expected)

	assert.Equal(t, expected, m.Expected())
}
