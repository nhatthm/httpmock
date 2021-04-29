package httpmock

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DoRequest calls DoRequestWithTimeout with 1 second timeout.
// nolint:thelper // It is called in DoRequestWithTimeout.
func DoRequest(
	t *testing.T,
	method, requestURI string,
	headers Header,
	body []byte,
) (int, map[string]string, []byte, time.Duration) {
	return DoRequestWithTimeout(t, method, requestURI, headers, body, time.Second)
}

// DoRequestWithTimeout sends a simple HTTP request for testing and returns the status code, response headers and
// response body along with the total execution time.
//
//   code, headers, body, _ = DoRequestWithTimeout(t, http.MethodGet, "/", nil, nil, 0)
func DoRequestWithTimeout(
	t *testing.T,
	method, requestURI string,
	headers Header,
	body []byte,
	timeout time.Duration,
) (int, map[string]string, []byte, time.Duration) {
	t.Helper()

	var reqBody io.Reader

	if body != nil {
		reqBody = strings.NewReader(string(body))
	}

	req, err := http.NewRequestWithContext(context.Background(), method, requestURI, reqBody)
	require.NoError(t, err, "could not create a new request")

	for header, value := range headers {
		req.Header.Set(header, value)
	}

	client := http.Client{Timeout: timeout}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	require.NoError(t, err, "could not make a request to mocked server")

	respCode := resp.StatusCode
	respHeaders := map[string]string(nil)

	if len(resp.Header) > 0 {
		respHeaders = map[string]string{}

		for header := range resp.Header {
			respHeaders[header] = resp.Header.Get(header)
		}
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err, "could not read response body")

	err = resp.Body.Close()
	require.NoError(t, err, "could not close response body")

	return respCode, respHeaders, respBody, elapsed
}

// GetBody returns request body and lets it re-readable.
func GetBody(r *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	err = r.Body.Close()
	if err != nil {
		return nil, err
	}

	r.Body = ioutil.NopCloser(bytes.NewReader(body))

	return body, err
}

// AssertHeaderContains asserts that the HTTP headers contains some specifics headers.
func AssertHeaderContains(t assert.TestingT, headers, contains Header) bool {
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

func requireNoErr(err error) {
	if err != nil {
		panic(err)
	}
}

// mergeHeaders merges a list of headers with some defaults. If a default header appears in the given headers, it
// will not be merged, no matter what the value is.
func mergeHeaders(headers, defaultHeaders map[string]string) map[string]string {
	result := make(map[string]string, len(headers)+len(defaultHeaders))

	for header, value := range defaultHeaders {
		result[textproto.CanonicalMIMEHeaderKey(header)] = value
	}

	for header, value := range headers {
		result[textproto.CanonicalMIMEHeaderKey(header)] = value
	}

	return result
}

func formatExpectedRequest(w io.Writer, method string, uri Matcher, header HeaderMatcher, body Matcher) {
	formatExpectedRequestTimes(w, method, uri, header, body, 0, 0)
}

func formatExpectedRequestTimes(w io.Writer, method string, uri Matcher, header HeaderMatcher, body Matcher, totalCalls, remainingCalls int) {
	expectedHeader := map[string]interface{}(nil)
	if header != nil {
		expectedHeader = make(map[string]interface{}, len(header))

		for header, matcher := range header {
			expectedHeader[header] = matcher
		}
	}

	expectedBody := []byte(nil)
	if body != nil {
		expectedBody = []byte(body.Expected())
	}

	formatRequestTimes(w, method, uri.Expected(), expectedHeader, expectedBody, totalCalls, remainingCalls)
}

func formatHTTPRequest(w io.Writer, method, uri string, header http.Header, body []byte) {
	expectedHeader := map[string]interface{}(nil)
	if header != nil {
		expectedHeader = make(map[string]interface{}, len(header))

		for key := range header {
			expectedHeader[key] = header.Get(key)
		}
	}

	formatRequestTimes(w, method, uri, expectedHeader, body, 0, 0)
}

func formatRequestTimes(w io.Writer, method string, uri interface{}, header map[string]interface{}, body interface{}, totalCalls, remainingCalls int) {
	_, _ = fmt.Fprintf(w, "%s %s", method, formatValueInline(uri))

	if remainingCalls > 0 && (totalCalls != 0 || remainingCalls != 1) {
		_, _ = fmt.Fprintf(w, " (called: %d time(s), remaining: %d time(s))", totalCalls, remainingCalls)
	}

	_, _ = fmt.Fprintln(w)

	if len(header) > 0 {
		_, _ = fmt.Fprintf(w, "%swith header:\n", indent)

		keys := make([]string, len(header))
		i := 0

		for k := range header {
			keys[i] = k
			i++
		}

		sort.Strings(keys)

		for _, key := range keys {
			_, _ = fmt.Fprintf(w, "%s%s%s: %s\n", indent, indent, key, formatValueInline(header[key]))
		}
	}

	if body != nil {
		bodyStr := formatValue(body)

		if bodyStr != "" {
			_, _ = fmt.Fprintf(w, "%swith body%s\n", indent, formatType(body))
			_, _ = fmt.Fprintf(w, "%s%s%s\n", indent, indent, bodyStr)
		}
	}
}

func formatValueInline(v interface{}) string {
	if v == nil {
		return "<nil>"
	}

	switch m := v.(type) {
	case *ExactMatch,
		[]byte,
		string:
		return formatValue(v)

	case Matcher:
		return fmt.Sprintf("%T(%q)", v, m.Expected())

	default:
		panic("unknown value type")
	}
}

func formatType(v interface{}) string {
	if v == nil {
		return ""
	}

	switch v.(type) {
	case *ExactMatch,
		[]byte,
		string:
		return ""

	default:
		return fmt.Sprintf(" using %T", v)
	}
}

func formatValue(v interface{}) string {
	if v == nil {
		return "<nil>"
	}

	switch m := v.(type) {
	case Matcher:
		return m.Expected()

	case []byte:
		return string(m)

	case string:
		return m

	default:
		panic("unknown value type")
	}
}
