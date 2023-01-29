package format

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.nhat.io/httpmock/matcher"
)

func TestFormatValueInline(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		value    any
		expected string
	}{
		{
			scenario: "nil",
			expected: "<nil>",
		},
		{
			scenario: "ExactMatch",
			value:    matcher.Exact("expected"),
			expected: "expected",
		},
		{
			scenario: "[]byte",
			value:    []byte("expected"),
			expected: "expected",
		},
		{
			scenario: "string",
			value:    "expected",
			expected: "expected",
		},
		{
			scenario: "Callback",
			value: matcher.Match(func() matcher.Matcher {
				return matcher.Exact("expected")
			}),
			expected: "expected",
		},
		{
			scenario: "Body Matcher",
			value:    matcher.Body(`expected`),
			expected: "expected",
		},
		{
			scenario: "Matcher",
			value:    matcher.JSON("{}"),
			expected: "matcher.JSONMatcher(\"{}\")",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, formatValueInline(tc.value))
		})
	}
}

func TestFormatValueInline_Panic(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() {
		formatValueInline(42)
	})
}

func TestFormatType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		value    any
		expected string
	}{
		{
			scenario: "nil",
			expected: "",
		},
		{
			scenario: "ExactMatch",
			value:    matcher.Exact("expected"),
			expected: "",
		},
		{
			scenario: "[]byte",
			value:    []byte("expected"),
			expected: "",
		},
		{
			scenario: "string",
			value:    "expected",
			expected: "",
		},
		{
			scenario: "Callback",
			value: matcher.Match(func() matcher.Matcher {
				return matcher.Exact("expected")
			}),
			expected: "",
		},
		{
			scenario: "Body Matcher is nil",
			value:    (*matcher.BodyMatcher)(nil),
			expected: "",
		},
		{
			scenario: "Body Matcher is not nil",
			value:    matcher.Body(matcher.JSON(`{}`)),
			expected: " using matcher.JSONMatcher",
		},
		{
			scenario: "Matcher",
			value:    matcher.JSON("{}"),
			expected: " using matcher.JSONMatcher",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, formatType(tc.value))
		})
	}
}

func TestFormatValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		value    any
		expected string
	}{
		{
			scenario: "nil",
			expected: "<nil>",
		},
		{
			scenario: "[]byte",
			value:    []byte("expected"),
			expected: "expected",
		},
		{
			scenario: "string",
			value:    "expected",
			expected: "expected",
		},
		{
			scenario: "Callback",
			value: matcher.Match(func() matcher.Matcher {
				return matcher.Exact("expected")
			}),
			expected: "expected",
		},
		{
			scenario: "ExactMatch",
			value:    matcher.Exact("expected"),
			expected: "expected",
		},
		{
			scenario: "Customer Matcher without expectation",
			value:    matcher.Fn("", nil),
			expected: "matches custom expectation",
		},
		{
			scenario: "Customer Matcher with expectation",
			value:    matcher.Fn("expected", nil),
			expected: "expected",
		},
		{
			scenario: "Body Matcher is nil",
			value:    (*matcher.BodyMatcher)(nil),
			expected: "",
		},
		{
			scenario: "Body Matcher is not nil",
			value:    matcher.Body("expected"),
			expected: "expected",
		},
		{
			scenario: "Matcher",
			value:    matcher.JSON("{}"),
			expected: "{}",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, formatValue(tc.value))
		})
	}
}

func TestFormatValue_Panic(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() {
		formatValue(42)
	})
}
