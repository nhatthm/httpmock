package httpmock_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/nhatthm/httpmock"
	"github.com/nhatthm/httpmock/mock/planner"
)

type (
	Server = httpmock.Server
	Header = httpmock.Header
)

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
			scenario:        "no expectation with body",
			mockServer:      func(s *Server) {},
			body:            []byte(`foobar`),
			expectedCode:    http.StatusInternalServerError,
			expectedHeaders: Header{},
			expectedBody:    "unexpected request received: GET /, body:\nfoobar",
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
				s.ExpectGet("/path")
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
				s.ExpectGet("/").
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
				s.ExpectGet("/").
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
    with body
        {"foo":"bar"}
Actual: GET /
    with header:
        Accept-Encoding: gzip
        Content-Length: 13
        Content-Type: application/json
        User-Agent: Go-http-client/1.1
    with body
        {"foo":"baz"}
Error: expected request body: {"foo":"bar"}, received: {"foo":"baz"}
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
			s := httpmock.MockServer(tc.mockServer).WithTest(serverT)

			defer s.Close()

			if tc.method == "" {
				tc.method = http.MethodGet
			}

			if tc.uri == "" {
				tc.uri = "/"
			}

			code, headers, body, _ := doRequest(t, s.URL(),
				tc.method, tc.uri,
				tc.headers,
				tc.body,
				tc.waitTime,
			)

			assert.Equal(t, tc.expectedCode, code)
			httpmock.AssertHeaderContains(t, headers, tc.expectedHeaders)
			assert.Equal(t, tc.expectedBody, string(body))

			if tc.expectedError {
				assert.Contains(t, serverT.String(), tc.expectedBody)
			} else {
				assert.Empty(t, serverT.String())
			}
		})
	}
}

func TestServer_WithDefaultHeaders(t *testing.T) {
	t.Parallel()

	testingT := T()

	s := httpmock.MockServer(func(s *httpmock.Server) {
		s.WithTest(testingT)

		s.WithDefaultResponseHeaders(httpmock.Header{
			"Content-Type": "application/json",
		})

		s.ExpectGet("/text").
			ReturnHeader("Content-Type", "text/plain").
			ReturnHeader("X-ID", "1").
			Return(`hello world!`)

		s.ExpectGet("/json").
			Return(`{"foo":"bar"}`)
	})

	defer s.Close()

	request := func(uri string) (int, map[string]string, []byte) {
		code, headers, body, _ := doRequest(t, s.URL(), http.MethodGet, uri, nil, nil, 0)

		return code, headers, body
	}

	expectedCode := http.StatusOK

	// 1st request is ok.
	expectedHeaders := map[string]string{"Content-Type": "text/plain", "X-ID": "1"}
	expectedBody := []byte(`hello world!`)

	code, headers, body := request("/text")

	assert.Equal(t, expectedCode, code)
	httpmock.AssertHeaderContains(t, headers, expectedHeaders)
	assert.Equal(t, expectedBody, body)
	assert.Empty(t, testingT.String())

	// 2nd request is ok.
	expectedHeaders = map[string]string{"Content-Type": "application/json"}
	expectedBody = []byte(`{"foo":"bar"}`)

	code, headers, body = request("/json")

	assert.Equal(t, expectedCode, code)
	httpmock.AssertHeaderContains(t, headers, expectedHeaders)
	assert.Equal(t, expectedBody, body)
	assert.Empty(t, testingT.String())

	assert.NoError(t, s.ExpectationsWereMet())
}

func TestServer_WithPlanner(t *testing.T) {
	t.Parallel()

	testingT := T()

	p := planner.Mock(func(p *planner.Planner) {
		p.On("Expect", mock.Anything)

		p.On("IsEmpty").Return(false)

		p.On("Plan", mock.Anything).
			Return(nil, errors.New("you shall not pass"))
	})(t)

	s := httpmock.New(func(s *httpmock.Server) {
		s.WithPlanner(p)

		s.ExpectGet("/").
			Return(`hello world!`)
	})(testingT)

	code, _, body, _ := doRequest(t, s.URL(), http.MethodGet, "/", nil, nil, 0)

	expectedCode := http.StatusInternalServerError
	expectedBody := []byte(`you shall not pass`)

	assert.Equal(t, expectedCode, code)
	assert.Equal(t, expectedBody, body)
	assert.Equal(t, string(expectedBody), testingT.String())
}

func TestServer_WithPlanner_Panic(t *testing.T) {
	t.Parallel()

	expected := `could not change planner: planner is not empty`

	assert.PanicsWithError(t, expected, func() {
		s := httpmock.NewServer()

		s.ExpectGet("/").
			Return(`hello world!`)

		s.WithPlanner(planner.NoMockPlanner(t))
	})
}

func TestServer_Repeatability(t *testing.T) {
	t.Parallel()

	testingT := T()

	s := httpmock.MockServer(func(s *httpmock.Server) {
		s.WithTest(testingT)

		s.ExpectGet("/").
			Return(`hello world!`).
			Twice()
	})

	defer s.Close()

	request := func() (int, []byte) {
		code, _, body, _ := doRequest(t, s.URL(), http.MethodGet, "/", nil, nil, 0)

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
				s.ExpectGet("/")
			},
		},
		{
			scenario: "wait until",
			mockServer: func(s *Server) {
				s.ExpectGet("/").
					WaitUntil(time.After(waitTime))
			},
			expectedDelay: expectedDelay,
		},
		{
			scenario: "after",
			mockServer: func(s *Server) {
				s.ExpectGet("/").
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
			s := httpmock.MockServer(tc.mockServer).WithTest(testingT)

			defer s.Close()

			code, _, _, elapsed := doRequest(t, s.URL(), http.MethodGet, "/", nil, nil, tc.expectedDelay)
			expectedCode := http.StatusOK

			assert.Equal(t, expectedCode, code)
			assert.LessOrEqual(t, tc.expectedDelay, elapsed,
				"unexpected delay, expected: %s, got %s", tc.expectedDelay, elapsed,
			)

			assert.NoError(t, s.ExpectationsWereMet())
		})
	}
}

func TestServer_ExpectationsWereNotMet(t *testing.T) {
	t.Parallel()

	testingT := T()

	s := httpmock.MockServer(func(s *httpmock.Server) {
		s.WithTest(testingT)

		s.ExpectGet("/").Times(3)
		s.ExpectGet("/unlimited").UnlimitedTimes()
		s.ExpectGet("/path")
	})

	defer s.Close()

	request := func(uri string) int {
		code, _, _, _ := doRequest(t, s.URL(), http.MethodGet, uri, nil, nil, 0)

		return code
	}

	expectedCode := http.StatusOK

	// 1st request is ok.
	assert.Equal(t, expectedCode, request("/"))

	expectedErr := `there are remaining expectations that were not met:
- GET / (called: 1 time(s), remaining: 2 time(s))
- GET /unlimited
- GET /path
`
	assert.EqualError(t, s.ExpectationsWereMet(), expectedErr)

	// 2nd request is ok.
	assert.Equal(t, expectedCode, request("/"))

	expectedErr = `there are remaining expectations that were not met:
- GET / (called: 2 time(s), remaining: 1 time(s))
- GET /unlimited
- GET /path
`
	assert.EqualError(t, s.ExpectationsWereMet(), expectedErr)
}

func TestServer_ExpectationsWereNotMet_UnlimitedRequest(t *testing.T) {
	t.Parallel()

	testingT := T()

	s := httpmock.MockServer(func(s *httpmock.Server) {
		s.WithTest(testingT)

		s.ExpectGet("/").UnlimitedTimes()
		s.ExpectGet("/path")
	})

	defer s.Close()

	request := func(uri string) int {
		code, _, _, _ := doRequest(t, s.URL(), http.MethodGet, uri, nil, nil, 0)

		return code
	}

	expectedCode := http.StatusOK

	// 1st request is ok.
	assert.Equal(t, expectedCode, request("/"))

	expectedErr := `there are remaining expectations that were not met:
- GET /path
`
	assert.EqualError(t, s.ExpectationsWereMet(), expectedErr)

	// 2nd request is ok.
	assert.Equal(t, expectedCode, request("/"))

	expectedErr = `there are remaining expectations that were not met:
- GET /path
`
	assert.EqualError(t, s.ExpectationsWereMet(), expectedErr)
}

func TestServer_ExpectationsWereMet_UnlimitedRequest(t *testing.T) {
	t.Parallel()

	testingT := T()

	s := httpmock.MockServer(func(s *httpmock.Server) {
		s.WithTest(testingT)

		s.ExpectGet("/").UnlimitedTimes()
	})

	defer s.Close()

	request := func(uri string) int {
		code, _, _, _ := doRequest(t, s.URL(), http.MethodGet, uri, nil, nil, 0)

		return code
	}

	expectedCode := http.StatusOK

	assert.Equal(t, expectedCode, request("/"))
	assert.NoError(t, s.ExpectationsWereMet())
}

func TestServer_ResetExpectations(t *testing.T) {
	t.Parallel()

	s := httpmock.MockServer(func(s *httpmock.Server) {
		s.ExpectGet("/").Times(3)
	})

	defer s.Close()

	s.ResetExpectations()

	assert.NoError(t, s.ExpectationsWereMet())
}

// nolint:thelper // It is called in DoRequestWithTimeout.
func doRequest(
	t *testing.T,
	baseURL string,
	method, uri string,
	headers Header,
	body []byte,
	waitTime time.Duration,
) (int, map[string]string, []byte, time.Duration) {
	return httpmock.DoRequestWithTimeout(t,
		method, baseURL+uri,
		headers, body,
		waitTime+time.Second,
	)
}
