package httpmock

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"github.com/stretchr/testify/require"
)

// Server is a Mock server.
type Server struct {
	// Represents the requests that are expected of a server.
	ExpectedRequests []*Request

	// Holds the requested that were made to this server.
	Requests []Request

	// Test server.
	server *httptest.Server

	// test is An optional variable that holds the test struct, to be used when an
	// invalid mock call was made.
	test TestingT

	mu sync.Mutex

	// defaultResponseHeader contains a list of default headers that will be sent to client.
	defaultResponseHeader Header

	// matchRequest matches a request with one of the expectations.
	matchRequest RequestMatcher
}

// NewServer creates mocked server.
func NewServer(t TestingT) *Server {
	s := Server{
		test:         t,
		matchRequest: DefaultRequestMatcher(),
	}
	s.server = httptest.NewServer(&s)

	return &s
}

// WithDefaultResponseHeaders sets the default response headers of the server.
func (s *Server) WithDefaultResponseHeaders(headers Header) *Server {
	s.defaultResponseHeader = headers

	return s
}

// WithRequestMatcher sets the RequestMatcher of the server.
func (s *Server) WithRequestMatcher(matcher RequestMatcher) *Server {
	s.matchRequest = matcher

	return s
}

// URL returns the current URL of the httptest.Server.
func (s *Server) URL() string {
	return s.server.URL
}

// Close closes mocked server.
func (s *Server) Close() {
	s.server.Close()
}

// Expect adds a new expected request.
//
//    Server.Expect(http.MethodGet, "/path").
func (s *Server) Expect(method, requestURI string) *Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := newRequest(s, method, requestURI)
	s.ExpectedRequests = append(s.ExpectedRequests, c)

	return c
}

// ExpectGet adds a new expected http.MethodGet request.
//
//   Server.ExpectGet("/path")
func (s *Server) ExpectGet(requestURI string) *Request {
	return s.Expect(http.MethodGet, requestURI)
}

// ExpectHead adds a new expected http.MethodHead request.
//
//   Server.ExpectHead("/path")
func (s *Server) ExpectHead(requestURI string) *Request {
	return s.Expect(http.MethodHead, requestURI)
}

// ExpectPost adds a new expected http.MethodPost request.
//
//   Server.ExpectPost("/path")
func (s *Server) ExpectPost(requestURI string) *Request {
	return s.Expect(http.MethodPost, requestURI)
}

// ExpectPut adds a new expected http.MethodPut request.
//
//   Server.ExpectPut("/path")
func (s *Server) ExpectPut(requestURI string) *Request {
	return s.Expect(http.MethodPut, requestURI)
}

// ExpectPatch adds a new expected http.MethodPatch request.
//
//   Server.ExpectPatch("/path")
func (s *Server) ExpectPatch(requestURI string) *Request {
	return s.Expect(http.MethodPatch, requestURI)
}

// ExpectDelete adds a new expected http.MethodDelete request.
//
//   Server.ExpectDelete("/path")
func (s *Server) ExpectDelete(requestURI string) *Request {
	return s.Expect(http.MethodDelete, requestURI)
}

// ExpectationsWereMet checks whether all queued expectations were met in order.
// If any of them was not met - an error is returned.
func (s *Server) ExpectationsWereMet() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.ExpectedRequests) == 0 {
		return nil
	}

	var sb strings.Builder

	sb.WriteString("there are remaining expectations that were not met:\n")

	for _, expected := range s.ExpectedRequests {
		sb.WriteString("- ")
		formatRequestTimes(&sb,
			expected.Method,
			expected.RequestURI,
			expected.RequestHeader,
			expected.RequestBody,
			expected.totalCalls,
			expected.Repeatability,
		)
	}

	// nolint:goerr113
	return errors.New(sb.String())
}

// ServeHTTP serves the request.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.ExpectedRequests) == 0 {
		body, err := GetBody(r)
		if err == nil && len(body) > 0 {
			s.failResponsef(w, "unexpected request received: %s %s, body:\n%s", r.Method, r.RequestURI, string(body))
		} else {
			s.failResponsef(w, "unexpected request received: %s %s", r.Method, r.RequestURI)
		}

		return
	}

	expected, expectedRequests, err := s.matchRequest(r, s.ExpectedRequests)
	if err != nil {
		s.failResponsef(w, err.Error())

		return
	}

	// Update expectations.
	s.ExpectedRequests = expectedRequests

	// Log the request.
	expected.totalCalls++
	s.Requests = append(s.Requests, *expected)

	// Block if specified.
	if expected.waitFor != nil {
		<-expected.waitFor
	} else {
		time.Sleep(expected.waitTime)
	}

	for header, value := range mergeHeaders(expected.ResponseHeader, s.defaultResponseHeader) {
		w.Header().Set(header, value)
	}

	w.WriteHeader(expected.StatusCode)

	body, err := expected.Do(r)
	require.NoError(s.test, err)

	_, err = w.Write(body)
	require.NoError(s.test, err)
}

func (s *Server) failResponsef(w http.ResponseWriter, format string, args ...interface{}) {
	body := fmt.Sprintf(format, args...)
	s.test.Errorf(body)

	w.WriteHeader(http.StatusInternalServerError)
	_, err := w.Write([]byte(body))
	require.NoError(s.test, err, "could not write response: %q", body)
}
