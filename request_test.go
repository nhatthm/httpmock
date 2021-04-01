package httpmock

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRequest_WithHeader(t *testing.T) {
	t.Parallel()

	r := &Request{parent: &Server{}, RequestHeader: Header{}}
	r.WithHeader("foo", "bar")

	assert.Equal(t, Header{"foo": "bar"}, r.RequestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, Header{"foo": "bar", "john": "doe"}, r.RequestHeader)
}

func TestRequest_WithHeaders(t *testing.T) {
	t.Parallel()

	r := &Request{parent: &Server{}}
	r.WithHeaders(Header{"foo": "bar"})

	assert.Equal(t, Header{"foo": "bar"}, r.RequestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, Header{"foo": "bar", "john": "doe"}, r.RequestHeader)
}

func TestRequest_WithBody(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario     string
		body         interface{}
		expectedBody []byte
		expectPanic  bool
	}{
		{
			scenario:     "body is []bytes",
			body:         []byte(`body`),
			expectedBody: []byte(`body`),
		},
		{
			scenario:     "body is string",
			body:         `body`,
			expectedBody: []byte(`body`),
		},
		{
			scenario:     "body is fmt.Stringer",
			body:         time.UTC,
			expectedBody: []byte(`UTC`),
		},
		{
			scenario:    "body has unexpected data type",
			body:        42,
			expectPanic: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			r := &Request{parent: &Server{}}

			if tc.expectPanic {
				assert.Panics(t, func() {
					r.WithBody(tc.body)
				})
			} else {
				r.WithBody(tc.body)

				assert.Equal(t, tc.expectedBody, r.RequestBody)
			}
		})
	}
}

func TestRequest_WithBodyf(t *testing.T) {
	t.Parallel()

	r := &Request{parent: &Server{}}
	expectedBody := []byte(`hello john`)

	r.WithBodyf("hello %s", "john")

	assert.Equal(t, expectedBody, r.RequestBody)
}

func TestRequest_WithBodyJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario     string
		body         interface{}
		expectedBody []byte
		expectPanic  bool
	}{
		{
			scenario:     "success",
			body:         map[string]string{"foo": "bar"},
			expectedBody: []byte(`{"foo":"bar"}`),
		},
		{
			scenario:    "panic",
			body:        make(chan struct{}, 1),
			expectPanic: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			r := &Request{parent: &Server{}}

			if tc.expectPanic {
				assert.Panics(t, func() {
					r.WithBodyJSON(tc.body)
				})
			} else {
				r.WithBodyJSON(tc.body)

				assert.Equal(t, tc.expectedBody, r.RequestBody)
			}
		})
	}
}

func TestRequest_ReturnCode(t *testing.T) {
	t.Parallel()

	r := &Request{parent: &Server{}}
	r.ReturnCode(http.StatusCreated)

	assert.Equal(t, http.StatusCreated, r.StatusCode)
}

func TestRequest_ReturnHeader(t *testing.T) {
	t.Parallel()

	r := &Request{parent: &Server{}, ResponseHeader: Header{}}
	r.ReturnHeader("foo", "bar")

	assert.Equal(t, Header{"foo": "bar"}, r.ResponseHeader)

	r.ReturnHeader("john", "doe")

	assert.Equal(t, Header{"foo": "bar", "john": "doe"}, r.ResponseHeader)
}

func TestRequest_ReturnHeaders(t *testing.T) {
	t.Parallel()

	r := &Request{parent: &Server{}}
	r.ReturnHeaders(Header{"foo": "bar"})

	assert.Equal(t, Header{"foo": "bar"}, r.ResponseHeader)

	r.ReturnHeader("john", "doe")

	assert.Equal(t, Header{"foo": "bar", "john": "doe"}, r.ResponseHeader)
}

func TestRequest_Return(t *testing.T) {
	t.Parallel()

	t.Helper()

	testCases := []struct {
		scenario     string
		body         interface{}
		expectedBody []byte
		expectPanic  bool
	}{
		{
			scenario:     "body is []bytes",
			body:         []byte(`body`),
			expectedBody: []byte(`body`),
		},
		{
			scenario:     "body is string",
			body:         `body`,
			expectedBody: []byte(`body`),
		},
		{
			scenario:     "body is fmt.Stringer",
			body:         time.UTC,
			expectedBody: []byte(`UTC`),
		},
		{
			scenario:    "body has unexpected data type",
			body:        42,
			expectPanic: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			r := &Request{parent: &Server{}}

			if tc.expectPanic {
				assert.Panics(t, func() {
					r.Return(tc.body)
				})
			} else {
				r.Return(tc.body)
				result, err := r.Do(nil)

				assert.Equal(t, tc.expectedBody, result)
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequest_Returnf(t *testing.T) {
	t.Parallel()

	r := &Request{parent: &Server{}}
	result, err := r.Returnf("hello %s", "john").Do(nil)

	expectedBody := []byte(`hello john`)

	assert.Equal(t, expectedBody, result)
	assert.NoError(t, err)
}

func TestRequest_ReturnJSON(t *testing.T) {
	t.Parallel()

	r := &Request{parent: &Server{}}
	r.ReturnJSON(map[string]string{"foo": "bar"})

	result, err := r.Do(nil)

	assert.Equal(t, `{"foo":"bar"}`, string(result))
	assert.NoError(t, err)
}

func TestRequest_ReturnFile(t *testing.T) {
	t.Parallel()

	r := &Request{parent: &Server{}}

	// File does not exist.
	assert.Panics(t, func() {
		r.ReturnFile("foo")
	})

	r.ReturnFile("resources/fixtures/response.txt")

	result, err := r.Do(nil)

	assert.Equal(t, `hello world!`, strings.TrimSpace(string(result)))
	assert.NoError(t, err)
}

func TestRequest_Once(t *testing.T) {
	t.Parallel()

	r := &Request{parent: &Server{}}
	r.Once()

	assert.Equal(t, 1, r.Repeatability)
}

func TestRequest_Twice(t *testing.T) {
	t.Parallel()

	r := &Request{parent: &Server{}}
	r.Twice()

	assert.Equal(t, 2, r.Repeatability)
}

func TestRequest_Times(t *testing.T) {
	t.Parallel()

	r := &Request{parent: &Server{}}
	r.Times(20)

	assert.Equal(t, 20, r.Repeatability)
}

func TestRequest_WaitUntil(t *testing.T) {
	t.Parallel()

	r := &Request{parent: &Server{}}
	ch := time.After(time.Second)

	r.WaitUntil(ch)

	assert.Equal(t, ch, r.waitFor)
}

func TestRequest_WaitTime(t *testing.T) {
	t.Parallel()

	r := &Request{parent: &Server{}}

	r.After(time.Second)

	assert.Equal(t, time.Second, r.waitTime)
}
