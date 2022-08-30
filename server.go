package httpmock

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/stretchr/testify/require"

	"go.nhat.io/httpmock/format"
	"go.nhat.io/httpmock/planner"
	"go.nhat.io/httpmock/request"
	"go.nhat.io/httpmock/test"
	"go.nhat.io/httpmock/value"
)

// Server is a Mock server.
type Server struct {
	// Holds the requested that were made to this server.
	Requests []request.Request

	// Test server.
	server  *httptest.Server
	planner planner.Planner

	// test is An optional variable that holds the test struct, to be used when an
	// invalid MockServer call was made.
	test test.T
	mu   sync.Mutex

	// defaultRequestOptions contains a list of default options what will be applied to every new requests.
	defaultRequestOptions []func(r *request.Request)
	// defaultResponseHeader contains a list of default headers that will be sent to client.
	defaultResponseHeader map[string]string
}

// NewServer creates a new server.
func NewServer() *Server {
	s := Server{
		test:    test.NoOpT(),
		planner: planner.Sequence(),
	}

	s.server = httptest.NewServer(&s)

	return &s
}

// WithPlanner sets the planner.
func (s *Server) WithPlanner(p planner.Planner) *Server {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.planner.IsEmpty() {
		panic(errors.New("could not change planner: planner is not empty")) // nolint: goerr113
	}

	s.planner = p

	return s
}

// WithTest sets the *testing.T of the server.
func (s *Server) WithTest(t test.T) *Server {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.test = t

	return s
}

// WithDefaultRequestOptions sets the default request options of the server.
func (s *Server) WithDefaultRequestOptions(opt func(r *request.Request)) *Server {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.defaultRequestOptions = append(s.defaultRequestOptions, opt)

	return s
}

// WithDefaultResponseHeaders sets the default response headers of the server.
func (s *Server) WithDefaultResponseHeaders(headers map[string]string) *Server {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.defaultResponseHeader = headers

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
//	Server.Expect(httpmock.MethodGet, "/path").
func (s *Server) Expect(method string, requestURI interface{}) *request.Request {
	expect := request.NewRequest(&s.mu, method, requestURI).Once()

	for _, o := range s.defaultRequestOptions {
		o(expect)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.planner.Expect(expect)

	return expect
}

// ExpectGet adds a new expected http.MethodGet request.
//
//	Server.ExpectGet("/path")
func (s *Server) ExpectGet(requestURI interface{}) *request.Request {
	return s.Expect(MethodGet, requestURI)
}

// ExpectHead adds a new expected http.MethodHead request.
//
//	Server.ExpectHead("/path")
func (s *Server) ExpectHead(requestURI interface{}) *request.Request {
	return s.Expect(MethodHead, requestURI)
}

// ExpectPost adds a new expected http.MethodPost request.
//
//	Server.ExpectPost("/path")
func (s *Server) ExpectPost(requestURI interface{}) *request.Request {
	return s.Expect(MethodPost, requestURI)
}

// ExpectPut adds a new expected http.MethodPut request.
//
//	Server.ExpectPut("/path")
func (s *Server) ExpectPut(requestURI interface{}) *request.Request {
	return s.Expect(MethodPut, requestURI)
}

// ExpectPatch adds a new expected http.MethodPatch request.
//
//	Server.ExpectPatch("/path")
func (s *Server) ExpectPatch(requestURI interface{}) *request.Request {
	return s.Expect(MethodPatch, requestURI)
}

// ExpectDelete adds a new expected http.MethodDelete request.
//
//	Server.ExpectDelete("/path")
func (s *Server) ExpectDelete(requestURI interface{}) *request.Request {
	return s.Expect(MethodDelete, requestURI)
}

// ExpectationsWereMet checks whether all queued expectations were met in order.
// If any of them was not met - an error is returned.
func (s *Server) ExpectationsWereMet() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.planner.IsEmpty() {
		return nil
	}

	var (
		sb    strings.Builder
		count int
	)

	sb.WriteString("there are remaining expectations that were not met:\n")

	for _, expected := range s.planner.Remain() {
		repeat := request.Repeatability(expected)
		calls := request.NumCalls(expected)

		if repeat < 1 && calls > 0 {
			continue
		}

		sb.WriteString("- ")
		format.ExpectedRequestTimes(&sb,
			request.Method(expected),
			request.URIMatcher(expected),
			request.HeaderMatcher(expected),
			request.BodyMatcher(expected),
			calls,
			repeat,
		)

		count++
	}

	if count == 0 {
		return nil
	}

	// nolint:goerr113
	return errors.New(sb.String())
}

// ServeHTTP serves the request.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.planner.IsEmpty() {
		body, err := value.GetBody(r)
		if err == nil && len(body) > 0 {
			s.failResponsef(w, "unexpected request received: %s %s, body:\n%s", r.Method, r.RequestURI, string(body))
		} else {
			s.failResponsef(w, "unexpected request received: %s %s", r.Method, r.RequestURI)
		}

		return
	}

	expected, err := s.planner.Plan(r)
	if err != nil {
		s.failResponsef(w, err.Error())

		return
	}

	// Log the request.
	request.CountCall(expected)

	s.Requests = append(s.Requests, *expected)

	err = request.Handle(expected, w, r, s.defaultResponseHeader)
	require.NoError(s.test, err)
}

func (s *Server) failResponsef(w http.ResponseWriter, format string, args ...interface{}) {
	body := fmt.Sprintf(format, args...)
	s.test.Errorf(body)

	err := request.FailResponse(w, body)

	require.NoError(s.test, err, "could not write response: %q", body)
}

// ResetExpectations resets all the expectations.
func (s *Server) ResetExpectations() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Requests = nil

	s.planner.Reset()
}
