package httpmock_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nhatthm/httpmock"
)

type (
	Server = httpmock.Server
	Header = httpmock.Header
)

type TestingT struct {
	strings.Builder
}

func (t *TestingT) Errorf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(t, format, args...)
}

func (t *TestingT) FailNow() {
	panic("failed")
}

func (t *TestingT) Cleanup(_ func()) {
	// Do nothing.
}

func T() *TestingT {
	return &TestingT{}
}

func TestServer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario        string
		mockServer      func(s *Server)
		method          string
		uri             string
		headers         Header
		body            []byte
		waitTime        time.Duration
		expectedCode    int
		expectedHeaders Header
		expectedBody    string
		expectedError   bool
	}{
		{
			scenario:        "no expectation",
			mockServer:      func(s *Server) {},
			expectedCode:    http.StatusInternalServerError,
			expectedHeaders: Header{},
			expectedBody:    `unexpected request received: GET /`,
			expectedError:   true,
		},
		{
			scenario: "expected different method",
			mockServer: func(s *Server) {
				s.Expect(http.MethodPost, "/")
			},
			expectedCode:    http.StatusInternalServerError,
			expectedHeaders: Header{},
			expectedBody: `Expected: POST /
Actual: GET /
    with header:
        Accept-Encoding: gzip
        User-Agent: Go-http-client/1.1
Error: method "POST" expected, "GET" received
`,
			expectedError: true,
		},
		{
			scenario: "expected different uri",
			mockServer: func(s *Server) {
				s.Expect(http.MethodGet, "/path")
			},
			expectedCode:    http.StatusInternalServerError,
			expectedHeaders: Header{},
			expectedBody: `Expected: GET /path
Actual: GET /
    with header:
        Accept-Encoding: gzip
        User-Agent: Go-http-client/1.1
Error: request uri "/path" expected, "/" received
`,
			expectedError: true,
		},
		{
			scenario: "expected header",
			mockServer: func(s *Server) {
				s.Expect(http.MethodGet, "/").
					WithHeader("Content-Type", "application/json")
			},
			expectedCode:    http.StatusInternalServerError,
			expectedHeaders: Header{},
			expectedBody: `Expected: GET /
    with header:
        Content-Type: application/json
Actual: GET /
    with header:
        Accept-Encoding: gzip
        User-Agent: Go-http-client/1.1
Error: header "Content-Type" with value "application/json" expected, "" received
`,
			expectedError: true,
		},
		{
			scenario: "expected body",
			mockServer: func(s *Server) {
				s.Expect(http.MethodGet, "/").
					WithBody(`{"foo":"bar"}`).
					WithHeader("Content-Type", "application/json")
			},
			body:            []byte(`{"foo":"baz"}`),
			headers:         Header{"Content-Type": "application/json"},
			expectedCode:    http.StatusInternalServerError,
			expectedHeaders: Header{},
			expectedBody: `Expected: GET /
    with header:
        Content-Type: application/json
    with body:
        {"foo":"bar"}
Actual: GET /
    with header:
        Accept-Encoding: gzip
        Content-Length: 13
        Content-Type: application/json
        User-Agent: Go-http-client/1.1
    with body:
        {"foo":"baz"}
Error: expected request body: "{\"foo\":\"bar\"}", received: "{\"foo\":\"baz\"}"
`,
			expectedError: true,
		},
		{
			scenario: "success",
			mockServer: func(s *Server) {
				s.WithDefaultResponseHeaders(Header{"Content-Type": "application/json"})
				s.Expect(http.MethodPost, "/create").
					WithBody(`{"foo":"bar"}`).
					WithHeader("Content-Type", "application/json").
					ReturnCode(http.StatusCreated).
					ReturnHeader("X-ID", "1").
					Return(`{"id":1,"foo":"bar"}`)
			},
			method:       http.MethodPost,
			uri:          "/create",
			body:         []byte(`{"foo":"bar"}`),
			headers:      Header{"Content-Type": "application/json"},
			expectedCode: http.StatusCreated,
			expectedHeaders: Header{
				"X-ID":         "1",
				"Content-Type": "application/json",
			},
			expectedBody: `{"id":1,"foo":"bar"}`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			serverT := T()
			s := httpmock.MockServer(serverT, tc.mockServer)

			defer s.Close()

			if tc.method == "" {
				tc.method = http.MethodGet
			}

			if tc.uri == "" {
				tc.uri = "/"
			}

			code, headers, body, _ := request(t, s.URL(),
				tc.method, tc.uri,
				tc.headers,
				tc.body,
				tc.waitTime,
			)

			assert.Equal(t, tc.expectedCode, code)
			assertHeaders(t, tc.expectedHeaders, headers)
			assert.Equal(t, tc.expectedBody, string(body))

			if tc.expectedError {
				assert.Contains(t, serverT.String(), tc.expectedBody)
			} else {
				assert.Empty(t, serverT.String())
			}
		})
	}
}

func TestServer_Repeatability(t *testing.T) {
	t.Parallel()

	testingT := T()

	s := httpmock.MockServer(testingT, func(s *httpmock.Server) {
		s.Expect(http.MethodGet, "/").
			Return(`hello world!`).
			Twice()
	})

	request := func() (int, []byte) {
		code, _, body, _ := request(t, s.URL(), http.MethodGet, "/", nil, nil, 0)

		return code, body
	}
	expectedCode := http.StatusOK
	expectedBody := []byte(`hello world!`)

	// 1st request is ok.
	code, body := request()

	assert.Equal(t, expectedCode, code)
	assert.Equal(t, expectedBody, body)
	assert.Empty(t, testingT.String())

	// 2nd request is ok.
	code, body = request()

	assert.Equal(t, expectedCode, code)
	assert.Equal(t, expectedBody, body)
	assert.Empty(t, testingT.String())

	// 3rd request is unexpected.
	code, body = request()
	errorMsg := `unexpected request received: GET /`

	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, []byte(errorMsg), body)
	assert.Contains(t, testingT.String(), errorMsg)
}

func TestServer_Wait(t *testing.T) {
	t.Parallel()

	waitTime := 100 * time.Millisecond
	expectedDelay := waitTime - 3*time.Millisecond

	testCases := []struct {
		scenario      string
		mockServer    func(s *Server)
		expectedDelay time.Duration
	}{
		{
			scenario: "no delay",
			mockServer: func(s *Server) {
				s.Expect(http.MethodGet, "/")
			},
		},
		{
			scenario: "wait until",
			mockServer: func(s *Server) {
				s.Expect(http.MethodGet, "/").
					WaitUntil(time.After(waitTime))
			},
			expectedDelay: expectedDelay,
		},
		{
			scenario: "after",
			mockServer: func(s *Server) {
				s.Expect(http.MethodGet, "/").
					After(waitTime)
			},
			expectedDelay: expectedDelay,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			testingT := T()
			s := httpmock.MockServer(testingT, tc.mockServer)

			code, _, _, elapsed := request(t, s.URL(), http.MethodGet, "/", nil, nil, tc.expectedDelay)
			expectedCode := http.StatusOK

			assert.Equal(t, expectedCode, code)
			assert.LessOrEqual(t, tc.expectedDelay, elapsed,
				"unexpected delay, expected: %s, got %s", tc.expectedDelay, elapsed,
			)
		})
	}
}

func TestServer_ExpectationsWereMet(t *testing.T) {
	t.Parallel()

	testingT := T()

	s := httpmock.MockServer(testingT, func(s *httpmock.Server) {
		s.Expect(http.MethodGet, "/").Times(3)
		s.Expect(http.MethodGet, "/path")
	})

	request := func() int {
		code, _, _, _ := request(t, s.URL(), http.MethodGet, "/", nil, nil, 0)

		return code
	}

	expectedCode := http.StatusOK

	// 1st request is ok.
	assert.Equal(t, expectedCode, request())

	expectedErr := `there are remaining expectations that were not met:
- GET / (called: 1 time(s), remaining: 2 time(s))
- GET /path
`
	assert.EqualError(t, s.ExpectationsWereMet(), expectedErr)

	// 2nd request is ok.
	assert.Equal(t, expectedCode, request())

	expectedErr = `there are remaining expectations that were not met:
- GET / (called: 2 time(s), remaining: 1 time(s))
- GET /path
`
	assert.EqualError(t, s.ExpectationsWereMet(), expectedErr)
}

func request(
	t *testing.T,
	baseURL string,
	method, uri string,
	headers Header,
	body []byte,
	waitTime time.Duration,
) (int, map[string]string, []byte, time.Duration) {
	t.Helper()

	var reqBody io.Reader

	if body != nil {
		reqBody = strings.NewReader(string(body))
	}

	req, err := http.NewRequest(method, baseURL+uri, reqBody)
	require.NoError(t, err, "could not create a new request")

	for header, value := range headers {
		req.Header.Set(header, value)
	}

	timeout := waitTime + time.Second
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

func assertHeaders(t *testing.T, expected, headers Header) bool {
	t.Helper()

	if len(expected) == 0 {
		return true
	}

	expectedHeaders := make(Header, len(expected))
	actualHeaders := make(Header, len(expected))

	for header := range expected {
		headerKey := http.CanonicalHeaderKey(header)
		expectedHeaders[headerKey] = expected[header]

		if value, ok := headers[headerKey]; ok {
			actualHeaders[headerKey] = value
		}
	}

	return assert.Equal(t, expectedHeaders, actualHeaders)
}
