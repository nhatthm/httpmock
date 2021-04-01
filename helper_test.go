package httpmock

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBody(t *testing.T) {
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
