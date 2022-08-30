package matcher_test

import (
	"errors"
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.nhat.io/httpmock/matcher"
)

func TestHeaderMatcher_Match(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		matcher       matcher.HeaderMatcher
		header        http.Header
		expectedError string
	}{
		{
			scenario: "nil",
		},
		{
			scenario: "empty",
			matcher:  matcher.HeaderMatcher{},
		},
		{
			scenario: "match error",
			matcher: matcher.HeaderMatcher{
				"Authorization": matcher.Fn("", func(interface{}) (bool, error) {
					return false, errors.New("match error")
				}),
			},
			expectedError: `could not match header: match error`,
		},
		{
			scenario: "mismatched",
			matcher: matcher.HeaderMatcher{
				"Authorization": matcher.Match("Bearer token"),
			},
			header: map[string][]string{
				"Authorization": {"Bearer foobar"},
			},
			expectedError: `header "Authorization" with value "Bearer token" expected, "Bearer foobar" received`,
		},
		{
			scenario: "matched",
			matcher: matcher.HeaderMatcher{
				"Authorization": matcher.Match(regexp.MustCompile(`Bearer .*`)),
			},
			header: map[string][]string{
				"Authorization": {"Bearer foobar"},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := tc.matcher.Match(tc.header)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
