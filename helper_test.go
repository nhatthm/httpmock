package httpmock

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type reader struct {
	upstream *strings.Reader
	readErr  error
	closeErr error
}

func (r *reader) Read(p []byte) (n int, err error) {
	if r.readErr != nil {
		return 0, r.readErr
	}

	return r.upstream.Read(p)
}

func (r *reader) Close() error {
	return r.closeErr
}

func newReader(s string, readErr, closeErr error) *reader {
	return &reader{
		upstream: strings.NewReader(s),
		readErr:  readErr,
		closeErr: closeErr,
	}
}

func TestGetBody(t *testing.T) {
	t.Parallel()

	expectedBody := []byte("body")
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("body"))

	// 1st read.
	body, err := GetBody(req)

	assert.Equal(t, expectedBody, body)
	assert.NoError(t, err)

	// 2nd read.
	body, err = GetBody(req)

	assert.Equal(t, expectedBody, body)
	assert.NoError(t, err)
}

func TestGetBody_ReadError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("read error")
	req := httptest.NewRequest(http.MethodGet, "/", newReader("body", expectedErr, nil))

	body, err := GetBody(req)

	assert.Nil(t, body)
	assert.Equal(t, expectedErr, err)
}

func TestGetBody_CloseError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("close error")
	req := httptest.NewRequest(http.MethodGet, "/", newReader("body", nil, expectedErr))

	body, err := GetBody(req)

	assert.Nil(t, body)
	assert.Equal(t, expectedErr, err)
}

func TestRequireNoErr_NoPanic(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		requireNoErr(nil)
	})
}

func TestRequireNoErr_Panic(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() {
		requireNoErr(errors.New("error"))
	})
}

func TestFormatValueInline(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		value    interface{}
		expected string
	}{
		{
			scenario: "nil",
			expected: "<nil>",
		},
		{
			scenario: "ExactMatch",
			value:    Exact("expected"),
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
			scenario: "Matcher",
			value:    JSON("{}"),
			expected: "*httpmock.JSONMatch(\"{}\")",
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
		value    interface{}
		expected string
	}{
		{
			scenario: "nil",
			expected: "",
		},
		{
			scenario: "ExactMatch",
			value:    Exact("expected"),
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
			scenario: "Matcher",
			value:    JSON("{}"),
			expected: " using *httpmock.JSONMatch",
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
		value    interface{}
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
			scenario: "ExactMatch",
			value:    Exact("expected"),
			expected: "expected",
		},
		{
			scenario: "Matcher",
			value:    JSON("{}"),
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
