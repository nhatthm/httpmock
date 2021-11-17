package request

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nhatthm/httpmock/matcher"
	"github.com/nhatthm/httpmock/must"
	"github.com/nhatthm/httpmock/value"
)

// Request is an expectation.
type Request struct {
	locker sync.Locker

	// method is the expected HTTP method of the given request.
	method string
	// requestURI is the expected HTTP request URI of the given request.
	// The uri does not need to be exactly same but satisfies the matcher.
	requestURI matcher.Matcher
	// requestHeader is a list of expected headers of the given request.
	requestHeader matcher.HeaderMatcher
	// requestBody is the expected body of the given request.
	requestBody *matcher.BodyMatcher

	// responseCode is the response code when the request is handled.
	responseCode int
	// responseHeader is a list of response headers to be sent to client when the request is handled.
	responseHeader Header

	run func(r *http.Request) ([]byte, error)

	// The number of times to return the return arguments when setting
	// expectations. 0 means to always return the value.
	repeatability int

	// Amount of times this request has been executed.
	totalCalls int

	// Holds a channel that will be used to block the Do until it either
	// receives a message or is closed. nil means it returns immediately.
	waitFor <-chan time.Time

	waitTime time.Duration
}

// NewRequest creates a new request expectation.
func NewRequest(locker sync.Locker, method string, requestURI interface{}) *Request {
	return &Request{
		locker:        locker,
		method:        method,
		responseCode:  http.StatusOK,
		requestURI:    matcher.Match(requestURI),
		repeatability: 0,
		waitFor:       nil,

		run: func(r *http.Request) ([]byte, error) {
			return nil, nil
		},
	}
}

func (r *Request) lock() {
	r.locker.Lock()
}

func (r *Request) unlock() {
	r.locker.Unlock()
}

// WithHeader sets an expected header of the given request.
//
//    Server.Expect(httpmock.MethodGet, "/path").
//    	WithHeader("foo", "bar")
//nolint:unparam
func (r *Request) WithHeader(header string, value interface{}) *Request {
	r.lock()
	defer r.unlock()

	if r.requestHeader == nil {
		r.requestHeader = matcher.HeaderMatcher{}
	}

	r.requestHeader[header] = matcher.Match(value)

	return r
}

// WithHeaders sets a list of expected headers of the given request.
//
//    Server.Expect(httpmock.MethodGet, "/path").
//    	WithHeaders(map[string]interface{}{"foo": "bar"})
func (r *Request) WithHeaders(headers map[string]interface{}) *Request {
	for header, value := range headers {
		r.WithHeader(header, value)
	}

	return r
}

// WithBody sets the expected body of the given request. It could be []byte, string, fmt.Stringer, or a Matcher.
//
//    Server.Expect(httpmock.MethodGet, "/path").
//    	WithBody("hello world!")
func (r *Request) WithBody(body interface{}) *Request {
	r.lock()
	defer r.unlock()

	r.requestBody = matchBody(body)

	return r
}

// WithBodyf formats according to a format specifier and use it as the expected body of the given request.
//
//    Server.Expect(httpmock.MethodGet, "/path").
//    	WithBodyf("hello %s", "john)
func (r *Request) WithBodyf(format string, args ...interface{}) *Request {
	return r.WithBody(fmt.Sprintf(format, args...))
}

// WithBodyJSON marshals the object and use it as the expected body of the given request.
//
//    Server.Expect(httpmock.MethodGet, "/path").
//    	WithBodyJSON(map[string]string{"foo": "bar"})
// nolint:unparam
func (r *Request) WithBodyJSON(v interface{}) *Request {
	body, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return r.WithBody(matcher.JSON(string(body)))
}

// ReturnCode sets the response code.
//
//    Server.Expect(httpmock.MethodGet, "/path").
//    	ReturnCode(httpmock.StatusBadRequest)
func (r *Request) ReturnCode(code int) *Request {
	r.lock()
	defer r.unlock()

	r.responseCode = code

	return r
}

// ReturnHeader sets a response header.
//
//    Server.Expect(httpmock.MethodGet, "/path").
//    	ReturnHeader("foo", "bar")
func (r *Request) ReturnHeader(header, value string) *Request {
	r.lock()
	defer r.unlock()

	if r.responseHeader == nil {
		r.responseHeader = map[string]string{}
	}

	r.responseHeader[header] = value

	return r
}

// ReturnHeaders sets a list of response headers.
//
//    Server.Expect(httpmock.MethodGet, "/path").
//    	ReturnHeaders(httpmock.Header{"foo": "bar"})
func (r *Request) ReturnHeaders(headers Header) *Request {
	r.lock()
	defer r.unlock()

	r.responseHeader = headers

	return r
}

// Return sets the result to return to client.
//
//    Server.Expect(httpmock.MethodGet, "/path").
//    	Return("hello world!")
func (r *Request) Return(v interface{}) *Request {
	body := []byte(value.String(v))

	return r.Run(func(*http.Request) ([]byte, error) {
		return body, nil
	})
}

// Returnf formats according to a format specifier and use it as the result to return to client.
//
//    Server.Expect(httpmock.MethodGet, "/path").
//    	Returnf("hello %s", "john")
func (r *Request) Returnf(format string, args ...interface{}) *Request {
	return r.Return(fmt.Sprintf(format, args...))
}

// ReturnJSON marshals the object using json.Marshal and uses it as the result to return to client.
//
//    Server.Expect(httpmock.MethodGet, "/path").
//    	ReturnJSON(map[string]string{"foo": "bar"})
func (r *Request) ReturnJSON(body interface{}) *Request {
	return r.Run(func(*http.Request) ([]byte, error) {
		return json.Marshal(body)
	})
}

// ReturnFile reads the file using ioutil.ReadFile and uses it as the result to return to client.
//
//    Server.Expect(httpmock.MethodGet, "/path").
//    	ReturnFile("resources/fixtures/response.txt")
// nolint:unparam
func (r *Request) ReturnFile(filePath string) *Request {
	filePath = filepath.Join(".", filepath.Clean(filePath))

	_, err := os.Stat(filePath)
	must.NotFail(err)

	return r.Run(func(*http.Request) ([]byte, error) {
		// nolint:gosec // filePath is cleaned above.
		return ioutil.ReadFile(filePath)
	})
}

// Run sets the handler to handle a given request.
//
//    Server.Expect(httpmock.MethodGet, "/path").
//		Run(func(*http.Request) ([]byte, error) {
//			return []byte("hello world!"), nil
//		})
func (r *Request) Run(handle func(r *http.Request) ([]byte, error)) *Request {
	r.lock()
	defer r.unlock()

	r.run = handle

	return r
}

// handle runs the HTTP request.
func (r *Request) handle(w http.ResponseWriter, req *http.Request, defaultHeaders map[string]string) error {
	// Block if specified.
	if r.waitFor != nil {
		<-r.waitFor
	} else {
		time.Sleep(r.waitTime)
	}

	body, err := r.run(req)
	if err != nil {
		_ = FailResponse(w, err.Error()) // nolint: errcheck

		return err
	}

	for key, val := range mergeHeaders(r.responseHeader, defaultHeaders) {
		w.Header().Set(key, val)
	}

	w.WriteHeader(r.responseCode)

	_, err = w.Write(body)

	return err
}

// Once indicates that the mock should only return the value once.
//
//    Server.Expect(http.MethodGet, "/path").
//    	Return("hello world!").
//    	Once()
func (r *Request) Once() *Request {
	return r.Times(1)
}

// Twice indicates that the mock should only return the value twice.
//
//    Server.Expect(http.MethodGet, "/path").
//    	Return("hello world!").
//    	Twice()
func (r *Request) Twice() *Request {
	return r.Times(2)
}

// UnlimitedTimes indicates that the mock should return the value at least once and there is no max limit in the number
// of return.
//
//    Server.Expect(http.MethodGet, "/path").
//    	Return("hello world!").
//    	UnlimitedTimes()
func (r *Request) UnlimitedTimes() *Request {
	return r.Times(0)
}

// Times indicates that the mock should only return the indicated number of times.
//
//    Server.Expect(http.MethodGet, "/path").
//    	Return("hello world!").
//    	Times(5)
func (r *Request) Times(i int) *Request {
	r.lock()
	defer r.unlock()
	r.repeatability = i

	return r
}

// WaitUntil sets the channel that will block the mocked return until its closed
// or a message is received.
//
//    Server.Expect(http.MethodGet, "/path").
//    	WaitUntil(time.After(time.Second)).
//    	Return("hello world!")
// nolint: unparam
func (r *Request) WaitUntil(w <-chan time.Time) *Request {
	r.lock()
	defer r.unlock()
	r.waitFor = w

	return r
}

// After sets how long to block until the call returns.
//
//    Server.Expect(http.MethodGet, "/path").
//    	After(time.Second).
//    	Return("hello world!")
// nolint: unparam
func (r *Request) After(d time.Duration) *Request {
	r.lock()
	defer r.unlock()
	r.waitTime = d

	return r
}
