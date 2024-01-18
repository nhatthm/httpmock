package httpmock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"go.nhat.io/wait"

	"go.nhat.io/httpmock/matcher"
	"go.nhat.io/httpmock/must"
	"go.nhat.io/httpmock/planner"
	"go.nhat.io/httpmock/value"
)

// Expectation sets the expectations for a http request.
//
// nolint: interfacebloat
type Expectation interface {
	// WithHeader sets an expected header of the given request.
	//
	//	Server.Expect(httpmock.MethodGet, "/path").
	//		WithHeader("foo", "bar")
	WithHeader(header string, value any) Expectation
	// WithHeaders sets a list of expected headers of the given request.
	//
	//	Server.Expect(httpmock.MethodGet, "/path").
	//		WithHeaders(map[string]any{"foo": "bar"})
	WithHeaders(headers map[string]any) Expectation
	// WithBody sets the expected body of the given request. It could be []byte, string, fmt.Stringer, or a Matcher.
	//
	//	Server.Expect(httpmock.MethodGet, "/path").
	//		WithBody("hello world!")
	WithBody(body any) Expectation
	// WithBodyf formats according to a format specifier and use it as the expected body of the given request.
	//
	//	Server.Expect(httpmock.MethodGet, "/path").
	//		WithBodyf("hello %s", "john)
	WithBodyf(format string, args ...any) Expectation
	// WithBodyJSON marshals the object and use it as the expected body of the given request.
	//
	//	Server.Expect(httpmock.MethodGet, "/path").
	//		WithBodyJSON(map[string]string{"foo": "bar"})
	//
	WithBodyJSON(v any) Expectation

	// ReturnCode sets the response code.
	//
	//	Server.Expect(httpmock.MethodGet, "/path").
	//		ReturnCode(httpmock.StatusBadRequest)
	ReturnCode(code int) Expectation
	// ReturnHeader sets a response header.
	//
	//	Server.Expect(httpmock.MethodGet, "/path").
	//		ReturnHeader("foo", "bar")
	ReturnHeader(header, value string) Expectation
	// ReturnHeaders sets a list of response headers.
	//
	//	Server.Expect(httpmock.MethodGet, "/path").
	//		ReturnHeaders(httpmock.Header{"foo": "bar"})
	ReturnHeaders(headers Header) Expectation
	// Return sets the result to return to client.
	//
	//	Server.Expect(httpmock.MethodGet, "/path").
	//		Return("hello world!")
	Return(v any) Expectation
	// Returnf formats according to a format specifier and use it as the result to return to client.
	//
	//	Server.Expect(httpmock.MethodGet, "/path").
	//		Returnf("hello %s", "john")
	Returnf(format string, args ...any) Expectation
	// ReturnJSON marshals the object using json.Marshal and uses it as the result to return to client.
	//
	//	Server.Expect(httpmock.MethodGet, "/path").
	//		ReturnJSON(map[string]string{"foo": "bar"})
	ReturnJSON(body any) Expectation
	// ReturnFile reads the file using ioutil.ReadFile and uses it as the result to return to client.
	//
	//	Server.Expect(httpmock.MethodGet, "/path").
	//		ReturnFile("resources/fixtures/response.txt")
	ReturnFile(filePath string) Expectation
	// Run sets the handler to handle a given request.
	//
	//	   Server.Expect(httpmock.MethodGet, "/path").
	//			Run(func(*http.Request) ([]byte, error) {
	//				return []byte("hello world!"), nil
	//			})
	Run(handle func(r *http.Request) ([]byte, error)) Expectation

	// Once indicates that the mock should only return the value once.
	//
	//	Server.Expect(http.MethodGet, "/path").
	//		Return("hello world!").
	//		Once()
	Once() Expectation
	// Twice indicates that the mock should only return the value twice.
	//
	//	Server.Expect(http.MethodGet, "/path").
	//		Return("hello world!").
	//		Twice()
	Twice() Expectation
	// UnlimitedTimes indicates that the mock should return the value at least once and there is no max limit in the
	// number of return.
	//
	//	Server.Expect(http.MethodGet, "/path").
	//		Return("hello world!").
	//		UnlimitedTimes()
	UnlimitedTimes() Expectation
	// Times indicates that the mock should only return the indicated number of times.
	//
	//	Server.Expect(http.MethodGet, "/path").
	//		Return("hello world!").
	//		Times(5)
	Times(i uint) Expectation

	// WaitUntil sets the channel that will block the mocked return until its closed
	// or a message is received.
	//
	//	Server.Expect(http.MethodGet, "/path").
	//		WaitUntil(time.After(time.Second)).
	//		Return("hello world!")
	WaitUntil(w <-chan time.Time) Expectation
	// After sets how long to block until the call returns.
	//
	//	Server.Expect(http.MethodGet, "/path").
	//		After(time.Second).
	//		Return("hello world!")
	After(d time.Duration) Expectation
}

// ExpectationHandler handles the expectation.
type ExpectationHandler interface {
	Handle(w http.ResponseWriter, r *http.Request, defaultHeaders map[string]string) error
}

var (
	_ Expectation         = (*requestExpectation)(nil)
	_ planner.Expectation = (*requestExpectation)(nil)
)

// requestExpectation is an expectation.
type requestExpectation struct {
	locker sync.Locker
	waiter wait.Waiter

	// requestMethod is the expected HTTP requestMethod of the given request.
	requestMethod string
	// requestURIMatcher is the expected HTTP request URI of the given request.
	// The uri does not need to be exactly same but satisfies the matcher.
	requestURIMatcher matcher.Matcher
	// requestHeaderMatcher is a list of expected headers of the given request.
	requestHeaderMatcher matcher.HeaderMatcher
	// requestBodyMatcher is the expected body of the given request.
	requestBodyMatcher *matcher.BodyMatcher

	// responseCode is the response code when the request is handled.
	responseCode int
	// responseHeader is a list of response headers to be sent to client when the request is handled.
	responseHeader Header

	handle func(r *http.Request) ([]byte, error)

	fulfilledTimes uint
	repeatTimes    uint
}

func (e *requestExpectation) lock() {
	e.locker.Lock()
}

func (e *requestExpectation) unlock() {
	e.locker.Unlock()
}

func (e *requestExpectation) Method() string {
	e.lock()
	defer e.unlock()

	return e.requestMethod
}

func (e *requestExpectation) URIMatcher() matcher.Matcher {
	e.lock()
	defer e.unlock()

	return e.requestURIMatcher
}

func (e *requestExpectation) HeaderMatcher() matcher.HeaderMatcher {
	e.lock()
	defer e.unlock()

	return e.requestHeaderMatcher
}

func (e *requestExpectation) BodyMatcher() *matcher.BodyMatcher {
	e.lock()
	defer e.unlock()

	return e.requestBodyMatcher
}

func (e *requestExpectation) RemainTimes() uint {
	e.lock()
	defer e.unlock()

	return e.repeatTimes
}

func (e *requestExpectation) Fulfilled() {
	e.lock()
	defer e.unlock()

	if e.repeatTimes > 0 {
		e.repeatTimes--
	}

	e.fulfilledTimes++
}

func (e *requestExpectation) FulfilledTimes() uint {
	e.lock()
	defer e.unlock()

	return e.fulfilledTimes
}

// WithHeader sets an expected header of the given request.
//
//	Server.Expect(httpmock.MethodGet, "/path").
//		WithHeader("foo", "bar")
//
//nolint:unparam
func (e *requestExpectation) WithHeader(header string, val any) Expectation {
	e.lock()
	defer e.unlock()

	if e.requestHeaderMatcher == nil {
		e.requestHeaderMatcher = matcher.HeaderMatcher{}
	}

	e.requestHeaderMatcher[header] = matcher.Match(val)

	return e
}

// WithHeaders sets a list of expected headers of the given request.
//
//	Server.Expect(httpmock.MethodGet, "/path").
//		WithHeaders(map[string]any{"foo": "bar"})
func (e *requestExpectation) WithHeaders(headers map[string]any) Expectation {
	for header, val := range headers {
		e.WithHeader(header, val)
	}

	return e
}

// WithBody sets the expected body of the given request. It could be []byte, string, fmt.Stringer, or a Matcher.
//
//	Server.Expect(httpmock.MethodGet, "/path").
//		WithBody("hello world!")
func (e *requestExpectation) WithBody(body any) Expectation {
	e.lock()
	defer e.unlock()

	e.requestBodyMatcher = matchBody(body)

	return e
}

// WithBodyf formats according to a format specifier and use it as the expected body of the given request.
//
//	Server.Expect(httpmock.MethodGet, "/path").
//		WithBodyf("hello %s", "john)
func (e *requestExpectation) WithBodyf(format string, args ...any) Expectation {
	return e.WithBody(fmt.Sprintf(format, args...))
}

// WithBodyJSON marshals the object and use it as the expected body of the given request.
//
//	Server.Expect(httpmock.MethodGet, "/path").
//		WithBodyJSON(map[string]string{"foo": "bar"})
//
// nolint:unparam
func (e *requestExpectation) WithBodyJSON(v any) Expectation {
	body, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return e.WithBody(matcher.JSON(string(body)))
}

// ReturnCode sets the response code.
//
//	Server.Expect(httpmock.MethodGet, "/path").
//		ReturnCode(httpmock.StatusBadRequest)
func (e *requestExpectation) ReturnCode(code int) Expectation {
	e.lock()
	defer e.unlock()

	e.responseCode = code

	return e
}

// ReturnHeader sets a response header.
//
//	Server.Expect(httpmock.MethodGet, "/path").
//		ReturnHeader("foo", "bar")
func (e *requestExpectation) ReturnHeader(header, value string) Expectation {
	e.lock()
	defer e.unlock()

	if e.responseHeader == nil {
		e.responseHeader = map[string]string{}
	}

	e.responseHeader[header] = value

	return e
}

// ReturnHeaders sets a list of response headers.
//
//	Server.Expect(httpmock.MethodGet, "/path").
//		ReturnHeaders(httpmock.Header{"foo": "bar"})
func (e *requestExpectation) ReturnHeaders(headers Header) Expectation {
	e.lock()
	defer e.unlock()

	e.responseHeader = headers

	return e
}

// Return sets the result to return to client.
//
//	Server.Expect(httpmock.MethodGet, "/path").
//		Return("hello world!")
func (e *requestExpectation) Return(v any) Expectation {
	body := []byte(value.String(v))

	return e.Run(func(*http.Request) ([]byte, error) {
		return body, nil
	})
}

// Returnf formats according to a format specifier and use it as the result to return to client.
//
//	Server.Expect(httpmock.MethodGet, "/path").
//		Returnf("hello %s", "john")
func (e *requestExpectation) Returnf(format string, args ...any) Expectation {
	return e.Return(fmt.Sprintf(format, args...))
}

// ReturnJSON marshals the object using json.Marshal and uses it as the result to return to client.
//
//	Server.Expect(httpmock.MethodGet, "/path").
//		ReturnJSON(map[string]string{"foo": "bar"})
func (e *requestExpectation) ReturnJSON(body any) Expectation {
	return e.Run(func(*http.Request) ([]byte, error) {
		return json.Marshal(body)
	})
}

// ReturnFile reads the file using ioutil.ReadFile and uses it as the result to return to client.
//
//	Server.Expect(httpmock.MethodGet, "/path").
//		ReturnFile("resources/fixtures/response.txt")
//
// nolint:unparam
func (e *requestExpectation) ReturnFile(filePath string) Expectation {
	filePath = filepath.Join(".", filepath.Clean(filePath))

	_, err := os.Stat(filePath)
	must.NotFail(err)

	return e.Run(func(*http.Request) ([]byte, error) {
		// nolint:gosec // filePath is cleaned above.
		return os.ReadFile(filePath)
	})
}

// Run sets the handler to handle a given request.
//
//	   Server.Expect(httpmock.MethodGet, "/path").
//			Run(func(*http.Request) ([]byte, error) {
//				return []byte("hello world!"), nil
//			})
func (e *requestExpectation) Run(handle func(r *http.Request) ([]byte, error)) Expectation {
	e.lock()
	defer e.unlock()

	e.handle = handle

	return e
}

// Once indicates that the mock should only return the value once.
//
//	Server.Expect(http.MethodGet, "/path").
//		Return("hello world!").
//		Once()
func (e *requestExpectation) Once() Expectation {
	return e.Times(1)
}

// Twice indicates that the mock should only return the value twice.
//
//	Server.Expect(http.MethodGet, "/path").
//		Return("hello world!").
//		Twice()
func (e *requestExpectation) Twice() Expectation {
	return e.Times(2)
}

// UnlimitedTimes indicates that the mock should return the value at least once and there is no max limit in the number
// of return.
//
//	Server.Expect(http.MethodGet, "/path").
//		Return("hello world!").
//		UnlimitedTimes()
func (e *requestExpectation) UnlimitedTimes() Expectation {
	return e.Times(0)
}

// Times indicates that the mock should only return the indicated number of times.
//
//	Server.Expect(http.MethodGet, "/path").
//		Return("hello world!").
//		Times(5)
func (e *requestExpectation) Times(i uint) Expectation {
	e.lock()
	defer e.unlock()

	e.repeatTimes = i

	return e
}

// WaitUntil sets the channel that will block the mocked return until its closed
// or a message is received.
//
//	Server.Expect(http.MethodGet, "/path").
//		WaitUntil(time.After(time.Second)).
//		Return("hello world!")
//
// nolint: unparam
func (e *requestExpectation) WaitUntil(w <-chan time.Time) Expectation {
	e.lock()
	defer e.unlock()

	e.waiter = wait.ForSignal(w)

	return e
}

// After sets how long to block until the call returns.
//
//	Server.Expect(http.MethodGet, "/path").
//		After(time.Second).
//		Return("hello world!")
//
// nolint: unparam
func (e *requestExpectation) After(d time.Duration) Expectation {
	e.lock()
	defer e.unlock()

	e.waiter = wait.ForDuration(d)

	return e
}

// Handle handles the HTTP request.
func (e *requestExpectation) Handle(w http.ResponseWriter, req *http.Request, defaultHeaders map[string]string) error {
	e.lock()
	defer e.unlock()

	if err := e.waiter.Wait(req.Context()); err != nil {
		return err
	}

	body, err := e.handle(req)
	if err != nil {
		_ = FailResponse(w, err.Error()) // nolint: errcheck

		return err
	}

	for key, val := range mergeHeaders(e.responseHeader, defaultHeaders) {
		w.Header().Set(key, val)
	}

	w.WriteHeader(e.responseCode)

	_, err = w.Write(body)

	return err
}

// newRequestExpectation creates a new request expectation.
func newRequestExpectation(method string, requestURI any) *requestExpectation {
	return &requestExpectation{
		locker:            &sync.Mutex{},
		requestMethod:     method,
		responseCode:      http.StatusOK,
		requestURIMatcher: matcher.Match(requestURI),
		repeatTimes:       0,
		waiter:            wait.NoWait,
		handle: func(r *http.Request) ([]byte, error) {
			return nil, nil
		},
	}
}

func matchBody(v any) *matcher.BodyMatcher {
	switch v := v.(type) {
	case matcher.Matcher,
		func() matcher.Matcher,
		*regexp.Regexp:
		return matcher.Body(v)
	}

	return matcher.Body(value.String(v))
}

// mergeHeaders merges a list of headers with some defaults. If a default header appears in the given headers, it
// will not be merged, no matter what the value is.
func mergeHeaders(headers, defaultHeaders Header) Header {
	result := make(Header, len(headers)+len(defaultHeaders))

	for header, val := range defaultHeaders {
		result[textproto.CanonicalMIMEHeaderKey(header)] = val
	}

	for header, val := range headers {
		result[textproto.CanonicalMIMEHeaderKey(header)] = val
	}

	return result
}
