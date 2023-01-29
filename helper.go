package httpmock

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.nhat.io/httpmock/test"
)

// DoRequest calls DoRequestWithTimeout with 1 second timeout.
// nolint:thelper // It is called in DoRequestWithTimeout.
func DoRequest(
	tb testing.TB,
	method, requestURI string,
	headers Header,
	body []byte,
) (int, map[string]string, []byte, time.Duration) {
	return DoRequestWithTimeout(tb, method, requestURI, headers, body, time.Second)
}

// DoRequestWithTimeout sends a simple HTTP requestExpectation for testing and returns the status code, response headers and
// response body along with the total execution time.
//
//	code, headers, body, _ = DoRequestWithTimeout(t, http.MethodGet, "/", map[string]string{}, nil, 0)
func DoRequestWithTimeout(
	tb testing.TB,
	method, requestURI string,
	headers Header,
	body []byte,
	timeout time.Duration,
) (int, map[string]string, []byte, time.Duration) {
	tb.Helper()

	var reqBody io.Reader

	if body != nil {
		reqBody = strings.NewReader(string(body))
	}

	req, err := http.NewRequestWithContext(context.Background(), method, requestURI, reqBody)
	require.NoError(tb, err, "could not create a new request")

	for header, value := range headers {
		req.Header.Set(header, value)
	}

	client := http.Client{Timeout: timeout}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	require.NoError(tb, err, "could not make a request to mocked server")

	respCode := resp.StatusCode
	respHeaders := map[string]string(nil)

	if len(resp.Header) > 0 {
		respHeaders = map[string]string{}

		for header := range resp.Header {
			respHeaders[header] = resp.Header.Get(header)
		}
	}

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(tb, err, "could not read response body")

	err = resp.Body.Close()
	require.NoError(tb, err, "could not close response body")

	return respCode, respHeaders, respBody, elapsed
}

// FailResponse responds a failure to client.
func FailResponse(w http.ResponseWriter, format string, args ...any) error {
	w.WriteHeader(http.StatusInternalServerError)

	_, err := fmt.Fprintf(w, format, args...)

	return err
}

// AssertHeaderContains asserts that the HTTP headers contains some specifics headers.
func AssertHeaderContains(t test.T, headers, contains Header) bool {
	if len(contains) == 0 {
		return true
	}

	expectedHeaders := make(Header, len(contains))
	actualHeaders := make(Header, len(contains))

	for header := range contains {
		headerKey := http.CanonicalHeaderKey(header)
		expectedHeaders[headerKey] = contains[header]

		if value, ok := headers[headerKey]; ok {
			actualHeaders[headerKey] = value
		}
	}

	return assert.Equal(t, expectedHeaders, actualHeaders)
}
