package httpmock_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/httpmock"
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
	body, err := httpmock.GetBody(req)

	assert.Equal(t, expectedBody, body)
	assert.NoError(t, err)

	// 2nd read.
	body, err = httpmock.GetBody(req)

	assert.Equal(t, expectedBody, body)
	assert.NoError(t, err)
}

func TestGetBody_ReadError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("read error")
	req := httptest.NewRequest(http.MethodGet, "/", newReader("body", expectedErr, nil))

	body, err := httpmock.GetBody(req)

	assert.Nil(t, body)
	assert.Equal(t, expectedErr, err)
}

func TestGetBody_CloseError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("close error")
	req := httptest.NewRequest(http.MethodGet, "/", newReader("body", nil, expectedErr))

	body, err := httpmock.GetBody(req)

	assert.Nil(t, body)
	assert.Equal(t, expectedErr, err)
}
