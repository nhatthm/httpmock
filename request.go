package httpmock

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Header is a list of HTTP headers.
type Header map[string]string

// RequestHandler handles the request and returns a result or an error.
type RequestHandler func(r *http.Request) ([]byte, error)

// Request is an expectation.
type Request struct {
	parent *Server

	// Method is the expected HTTP Method of the given request.
	Method string
	// RequestURI is the expected HTTP Request URI of the given request.
	// The uri does not need to be exactly same but satisfies the URIMatcher.
	RequestURI string
	// RequestHeader is a list of expected headers of the given request.
	RequestHeader Header
	// RequestBody is the expected body of the given request.
	RequestBody []byte

	// StatusCode is the response code when the request is handled.
	StatusCode int
	// ResponseHeader is a list of response headers to be sent to client when the request is handled.
	ResponseHeader Header
	// Do handles the request and returns a result or an error.
	Do RequestHandler

	// The number of times to return the return arguments when setting
	// expectations. 0 means to always return the value.
	Repeatability int

	// Amount of times this call has been called
	totalCalls int

	// Holds a channel that will be used to block the Do until it either
	// receives a message or is closed. nil means it returns immediately.
	waitFor <-chan time.Time

	waitTime time.Duration
}

func newRequest(parent *Server, method string, requestURI string) *Request {
	return &Request{
		parent:        parent,
		Method:        method,
		StatusCode:    http.StatusOK,
		RequestURI:    requestURI,
		Repeatability: 0,
		waitFor:       nil,

		Do: func(r *http.Request) ([]byte, error) {
			return nil, nil
		},
	}
}

func (r *Request) lock() {
	r.parent.mu.Lock()
}

func (r *Request) unlock() {
	r.parent.mu.Unlock()
}

// WithHeader sets an expected header of the given request.
//
//    Server.Expect(http.MethodGet, "/path").
//    	WithHeader("foo": "bar")
//nolint:unparam
func (r *Request) WithHeader(header, value string) *Request {
	r.lock()
	defer r.unlock()

	if r.RequestHeader == nil {
		r.RequestHeader = Header{}
	}

	r.RequestHeader[header] = value

	return r
}

// WithHeaders sets a list of expected headers of the given request.
//
//    Server.Expect(http.MethodGet, "/path").
//    	WithHeaders(httpmock.Header{"foo": "bar"})
func (r *Request) WithHeaders(headers Header) *Request {
	r.lock()
	defer r.unlock()
	r.RequestHeader = headers

	return r
}

// WithBody sets the expected body of the given request.
//
//    Server.Expect(http.MethodGet, "/path").
//    	WithBody("hello world!")
func (r *Request) WithBody(body interface{}) *Request {
	r.lock()
	defer r.unlock()

	switch body := body.(type) {
	case []byte:
		r.RequestBody = body

	case string:
		r.RequestBody = []byte(body)

	case fmt.Stringer:
		r.RequestBody = []byte(body.String())

	default:
		panic(fmt.Errorf("%w: unexpected request body data type: %T", ErrUnsupportedDataType, body))
	}

	return r
}

// WithBodyJSON marshals the object and use it as the expected body of the given request.
//
//    Server.Expect(http.MethodGet, "/path").
//    	WithBodyJSON(map[string]string{"foo": "bar"})
// nolint:unparam
func (r *Request) WithBodyJSON(v interface{}) *Request {
	body, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return r.WithBody(body)
}

// ReturnCode sets the response code.
//
//    Server.Expect(http.MethodGet, "/path").
//    	ReturnCode(http.StatusBadRequest)
func (r *Request) ReturnCode(code int) *Request {
	r.lock()
	defer r.unlock()
	r.StatusCode = code

	return r
}

// ReturnHeader sets a response header.
//
//    Server.Expect(http.MethodGet, "/path").
//    	ReturnHeader("foo", "bar")
//nolint:unparam
func (r *Request) ReturnHeader(header, value string) *Request {
	r.lock()
	defer r.unlock()

	if r.ResponseHeader == nil {
		r.ResponseHeader = Header{}
	}

	r.ResponseHeader[header] = value

	return r
}

// ReturnHeaders sets a list of response headers.
//
//    Server.Expect(http.MethodGet, "/path").
//    	ReturnHeaders(httpmock.Header{"foo": "bar"})
func (r *Request) ReturnHeaders(headers Header) *Request {
	r.lock()
	defer r.unlock()
	r.ResponseHeader = headers

	return r
}

// Return sets the result to return to client.
//
//    Server.Expect(http.MethodGet, "/path").
//    	Return("hello world!")
//nolint:unparam
func (r *Request) Return(v interface{}) *Request {
	var body []byte

	switch value := v.(type) {
	case []byte:
		body = value

	case string:
		body = []byte(value)

	case fmt.Stringer:
		body = []byte(value.String())

	default:
		panic(fmt.Errorf("%w: unexpected response data type: %T", ErrUnsupportedDataType, body))
	}

	return r.RequestHandler(func(_ *http.Request) ([]byte, error) {
		return body, nil
	})
}

// ReturnJSON marshals the object using json.Marshal and uses it as the result to return to client.
//
//    Server.Expect(http.MethodGet, "/path").
//    	ReturnJSON(map[string]string{"foo": "bar"})
func (r *Request) ReturnJSON(body interface{}) *Request {
	return r.RequestHandler(func(_ *http.Request) ([]byte, error) {
		return json.Marshal(body)
	})
}

// ReturnFile reads the file using ioutil.ReadFile and uses it as the result to return to client.
//
//    Server.Expect(http.MethodGet, "/path").
//    	ReturnFile("resources/fixtures/response.txt")
// nolint:unparam
func (r *Request) ReturnFile(filePath string) *Request {
	filePath = filepath.Join(".", filepath.Clean(filePath))

	if _, err := os.Stat(filePath); err != nil {
		panic(err)
	}

	return r.RequestHandler(func(_ *http.Request) ([]byte, error) {
		// nolint:gosec // filePath is cleaned above.
		return ioutil.ReadFile(filePath)
	})
}

// RequestHandler sets the handler to handle a given request.
//
//    Server.Expect(http.MethodGet, "/path").
//		RequestHandler(func(_ *http.Request) ([]byte, error) {
//			return []byte("hello world!"), nil
//		})
func (r *Request) RequestHandler(handler RequestHandler) *Request {
	r.lock()
	defer r.unlock()
	r.Do = handler

	return r
}

// Once indicates that that the mock should only return the value once.
//
//    Server.Expect(http.MethodGet, "/path").
//    	Return("hello world!").
//    	Once()
func (r *Request) Once() *Request {
	return r.Times(1)
}

// Twice indicates that that the mock should only return the value twice.
//
//    Server.Expect(http.MethodGet, "/path").
//    	Return("hello world!").
//    	Twice()
func (r *Request) Twice() *Request {
	return r.Times(2)
}

// Times indicates that that the mock should only return the indicated number
// of times.
//
//    Server.Expect(http.MethodGet, "/path").
//    	Return("hello world!").
//    	Times(5)
func (r *Request) Times(i int) *Request {
	r.lock()
	defer r.unlock()
	r.Repeatability = i

	return r
}

// WaitUntil sets the channel that will block the mock's return until its closed
// or a message is received.
//
//    Server.Expect(http.MethodGet, "/path").
//    	WaitUntil(time.After(time.Second)).
//    	Return("hello world!")
func (r *Request) WaitUntil(w <-chan time.Time) *Request {
	r.lock()
	defer r.unlock()
	r.waitFor = w

	return r
}

// After sets how long to block until the call returns
//
//    Server.Expect(http.MethodGet, "/path").
//    	After(time.Second).
//    	Return("hello world!")
func (r *Request) After(d time.Duration) *Request {
	r.lock()
	defer r.unlock()
	r.waitTime = d

	return r
}
