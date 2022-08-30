package format

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.nhat.io/httpmock/matcher"
)

func TestIsNil(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		value    interface{}
		expected bool
	}{
		{
			scenario: "nil",
			expected: true,
		},
		{
			scenario: "nil slice",
			value:    ([]int)(nil),
			expected: true,
		},
		{
			scenario: "nil interface",
			value:    (error)(nil),
			expected: true,
		},
		{
			scenario: "nil pointer",
			value:    (*matcher.BodyMatcher)(nil),
			expected: true,
		},
		{
			scenario: "empty slice",
			value:    []int{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, isNil(tc.value))
		})
	}
}
