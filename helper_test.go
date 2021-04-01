package httpmock_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/httpmock"
)

func TestGetBody(t *testing.T) {
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
