package value_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.nhat.io/httpmock/mock/http"
	"go.nhat.io/httpmock/value"
)

func TestString_Success(t *testing.T) {
	t.Parallel()

	const expected = `id:42`

	testCases := []struct {
		scenario string
		input    any
		expected string
	}{
		{
			scenario: "string",
			input:    expected,
		},
		{
			scenario: "[]byte",
			input:    []byte(expected),
		},
		{
			scenario: "fmt.Stringer",
			input:    bytes.NewBufferString(expected),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			actual := value.String(tc.input)

			assert.Equal(t, expected, actual)
		})
	}
}

func TestString_Panic(t *testing.T) {
	t.Parallel()

	assert.PanicsWithError(t, `unsupported data type`, func() {
		value.String(42)
	})
}

func TestGetBody(t *testing.T) {
	t.Parallel()

	expectedBody := []byte("body")
	req := http.BuildRequest().WithBody("body").Build()

	// 1st read.
	body, err := value.GetBody(req)

	assert.Equal(t, expectedBody, body)
	assert.NoError(t, err)

	// 2nd read.
	body, err = value.GetBody(req)

	assert.Equal(t, expectedBody, body)
	assert.NoError(t, err)
}

func TestGetBody_ReadError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("read error")
	req := http.BuildRequest().WithBodyReadError(expectedErr).Build()

	body, err := value.GetBody(req)

	assert.Nil(t, body)
	assert.Equal(t, expectedErr, err)
}

func TestGetBody_CloseError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("close error")
	req := http.BuildRequest().WithBodyCloseError(expectedErr).Build()

	body, err := value.GetBody(req)

	assert.Nil(t, body)
	assert.Equal(t, expectedErr, err)
}
