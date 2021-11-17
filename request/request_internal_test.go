package request

import (
	"errors"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/httpmock/matcher"
	"github.com/nhatthm/httpmock/mock/http"
)

func TestRequest_WithHeader(t *testing.T) {
	t.Parallel()

	r := &Request{locker: &sync.Mutex{}, requestHeader: matcher.HeaderMatcher{}}
	r.WithHeader("foo", "bar")

	assert.Equal(t, matcher.HeaderMatcher{"foo": matcher.Exact("bar")}, r.requestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, matcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, r.requestHeader)
}

func TestRequest_WithHeaders(t *testing.T) {
	t.Parallel()

	r := &Request{locker: &sync.Mutex{}}
	r.WithHeaders(map[string]interface{}{"foo": "bar"})

	assert.Equal(t, matcher.HeaderMatcher{"foo": matcher.Exact("bar")}, r.requestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, matcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, r.requestHeader)
}

func TestRequest_WithBody(t *testing.T) {
	t.Parallel()

	const body = `{"id":42}`

	testCases := []struct {
		scenario       string
		expect         interface{}
		body           string
		expectedResult bool
		expectedError  string
	}{
		{
			scenario:       "[]byte matched",
			expect:         []byte(body),
			body:           body,
			expectedResult: true,
		},
		{
			scenario:       "string matched",
			expect:         body,
			body:           body,
			expectedResult: true,
		},
		{
			scenario:       "fmt.Stringer",
			expect:         time.UTC,
			body:           `UTC`,
			expectedResult: true,
		},
		{
			scenario: "json mismatched",
			expect:   matcher.JSON(`{"id": 1}`),
			body:     body,
		},
		{
			scenario:       "json matched",
			expect:         matcher.JSON(`{"id": 42}`),
			body:           body,
			expectedResult: true,
		},
		{
			scenario: "regex mismatched",
			expect:   regexp.MustCompile(`{"id":\d}`),
			body:     body,
		},
		{
			scenario:       "regex matched",
			expect:         regexp.MustCompile(`{"id":\d+}`),
			body:           body,
			expectedResult: true,
		},
		{
			scenario:       "not empty",
			expect:         matcher.IsNotEmpty(),
			body:           body,
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			r := &Request{locker: &sync.Mutex{}}

			r.WithBody(tc.expect)

			matched, err := r.requestBody.Match(http.BuildRequest().
				WithBody(tc.body).Build(),
			)

			assert.Equal(t, tc.expectedResult, matched)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRequest_WithBody_Panic(t *testing.T) {
	t.Parallel()

	expected := `unsupported data type`

	assert.PanicsWithError(t, expected, func() {
		(&Request{locker: &sync.Mutex{}}).
			WithBody(42)
	})
}

func TestRequest_WithBodyf(t *testing.T) {
	t.Parallel()

	r := &Request{locker: &sync.Mutex{}}

	r.WithBodyf("hello %s", "john")

	matched, err := r.requestBody.Match(http.BuildRequest().
		WithBody(`hello john`).Build(),
	)

	assert.True(t, matched)
	assert.NoError(t, err)
}

func TestRequest_WithBodyJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario    string
		body        interface{}
		expectPanic bool
	}{
		{
			scenario: "success",
			body:     map[string]string{"foo": "bar"},
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

			r := &Request{locker: &sync.Mutex{}}

			if tc.expectPanic {
				assert.Panics(t, func() {
					r.WithBodyJSON(tc.body)
				})
			} else {
				r.WithBodyJSON(tc.body)

				matched, err := r.requestBody.Match(http.BuildRequest().
					WithBody(`{"foo":"bar"}`).Build(),
				)

				assert.True(t, matched)
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequest_ReturnCode(t *testing.T) {
	t.Parallel()

	r := &Request{locker: &sync.Mutex{}}
	r.ReturnCode(http.StatusCreated)

	assert.Equal(t, http.StatusCreated, r.responseCode)
}

func TestRequest_ReturnHeader(t *testing.T) {
	t.Parallel()

	r := &Request{locker: &sync.Mutex{}, responseHeader: map[string]string{}}
	r.ReturnHeader("foo", "bar")

	assert.Equal(t, map[string]string{"foo": "bar"}, r.responseHeader)

	r.ReturnHeader("john", "doe")

	assert.Equal(t, map[string]string{"foo": "bar", "john": "doe"}, r.responseHeader)
}

func TestRequest_ReturnHeaders(t *testing.T) {
	t.Parallel()

	r := &Request{locker: &sync.Mutex{}}
	r.ReturnHeaders(map[string]string{"foo": "bar"})

	assert.Equal(t, map[string]string{"foo": "bar"}, r.responseHeader)

	r.ReturnHeader("john", "doe")

	assert.Equal(t, map[string]string{"foo": "bar", "john": "doe"}, r.responseHeader)
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

			r := &Request{locker: &sync.Mutex{}}

			if tc.expectPanic {
				assert.Panics(t, func() {
					r.Return(tc.body)
				})
			} else {
				r.ReturnCode(http.StatusOK).
					Return(tc.body)

				w := http.MockResponseWriter(func(w *http.ResponseWriter) {
					w.On("WriteHeader", http.StatusOK)

					w.On("Write", tc.expectedBody).
						Return(0, nil)
				})(t)

				err := r.handle(w, nil, nil)

				assert.NoError(t, err)
			}
		})
	}
}

func TestRequest_Returnf(t *testing.T) {
	t.Parallel()

	w := http.MockResponseWriter(func(w *http.ResponseWriter) {
		w.On("WriteHeader", http.StatusOK)

		w.On("Write", []byte(`hello john`)).
			Return(0, nil)
	})(t)

	r := &Request{locker: &sync.Mutex{}}

	r.ReturnCode(http.StatusOK).
		Returnf("hello %s", "john")

	err := r.handle(w, nil, nil)

	assert.NoError(t, err)
}

func TestRequest_ReturnJSON(t *testing.T) {
	t.Parallel()

	w := http.MockResponseWriter(func(w *http.ResponseWriter) {
		w.On("WriteHeader", http.StatusOK)

		w.On("Write", []byte(`{"foo":"bar"}`)).
			Return(0, nil)
	})(t)

	r := &Request{locker: &sync.Mutex{}}

	r.ReturnCode(http.StatusOK).
		ReturnJSON(map[string]string{"foo": "bar"})

	err := r.handle(w, nil, nil)

	assert.NoError(t, err)
}

func TestRequest_ReturnFile(t *testing.T) {
	t.Parallel()

	w := http.MockResponseWriter(func(w *http.ResponseWriter) {
		w.On("WriteHeader", http.StatusOK)

		w.On("Write", []byte("hello world!\n")).
			Return(0, nil)
	})(t)

	r := &Request{locker: &sync.Mutex{}}

	// File does not exist.
	assert.Panics(t, func() {
		r.ReturnFile("foo")
	})

	r.ReturnCode(http.StatusOK).
		ReturnFile("../resources/fixtures/response.txt")

	err := r.handle(w, nil, nil)

	assert.NoError(t, err)
}

func TestRequest_Handle_Success(t *testing.T) {
	t.Parallel()

	responseHeader := http.Header{}

	w := http.MockResponseWriter(func(w *http.ResponseWriter) {
		w.On("Header").Return(responseHeader)
		w.On("WriteHeader", 200)

		w.On("Write", []byte(`{"id":42}`)).
			Return(9, nil)
	})(t)

	r := &Request{locker: &sync.Mutex{}}

	r.ReturnCode(http.StatusOK).
		ReturnHeader("Content-Type", "application/json").
		ReturnHeader("Content-Length", "9").
		Return(`{"id":42}`)

	defaultHeaders := Header{
		"Content-Type": "text/plain",
	}

	err := r.handle(w, nil, defaultHeaders)

	expectedHeader := http.Header{
		"Content-Type":   {"application/json"},
		"Content-Length": {"9"},
	}

	assert.Equal(t, expectedHeader, responseHeader)
	assert.NoError(t, err)
}

func TestRequest_Handle_RunError(t *testing.T) {
	t.Parallel()

	w := http.MockResponseWriter(func(w *http.ResponseWriter) {
		w.On("WriteHeader", 500)

		w.On("Write", []byte(`run error`)).
			Return(0, nil)
	})(t)

	r := &Request{locker: &sync.Mutex{}}

	r.ReturnCode(http.StatusOK).
		Run(func(*http.Request) ([]byte, error) {
			return nil, errors.New("run error")
		})

	actual := r.handle(w, nil, nil)
	expected := errors.New("run error")

	assert.Equal(t, expected, actual)
}

func TestRequest_Handle_WriteError(t *testing.T) {
	t.Parallel()

	w := http.MockResponseWriter(func(w *http.ResponseWriter) {
		w.On("WriteHeader", 200)

		w.On("Write", []byte(nil)).
			Return(0, errors.New("write error"))
	})(t)

	r := NewRequest(&sync.Mutex{}, http.MethodGet, "/")

	r.ReturnCode(http.StatusOK)

	actual := r.handle(w, nil, nil)
	expected := errors.New("write error")

	assert.Equal(t, expected, actual)
}

func TestRequest_Once(t *testing.T) {
	t.Parallel()

	r := &Request{locker: &sync.Mutex{}}
	r.Once()

	assert.Equal(t, 1, r.repeatability)
}

func TestRequest_Twice(t *testing.T) {
	t.Parallel()

	r := &Request{locker: &sync.Mutex{}}
	r.Twice()

	assert.Equal(t, 2, r.repeatability)
}

func TestRequest_Times(t *testing.T) {
	t.Parallel()

	r := &Request{locker: &sync.Mutex{}}
	r.Times(20)

	assert.Equal(t, 20, r.repeatability)
}

func TestRequest_UnlimitedTimes(t *testing.T) {
	t.Parallel()

	r := &Request{locker: &sync.Mutex{}, repeatability: 1}
	r.UnlimitedTimes()

	assert.Equal(t, 0, r.repeatability)
}

func TestRequest_WaitUntil(t *testing.T) {
	t.Parallel()

	r := &Request{locker: &sync.Mutex{}}
	ch := time.After(time.Second)

	r.WaitUntil(ch)

	assert.Equal(t, ch, r.waitFor)
}

func TestRequest_WaitTime(t *testing.T) {
	t.Parallel()

	r := &Request{locker: &sync.Mutex{}}

	r.After(time.Second)

	assert.Equal(t, time.Second, r.waitTime)
}

func TestRequest_Wait(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond

	testCases := []struct {
		scenario string
		mock     func(*Request)
	}{
		{
			scenario: "chan",
			mock: func(r *Request) {
				r.WaitUntil(time.After(duration))
			},
		},
		{
			scenario: "sleep",
			mock: func(r *Request) {
				r.After(duration)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			w := http.MockResponseWriter(func(w *http.ResponseWriter) {
				w.On("WriteHeader", 200)

				w.On("Write", []byte(nil)).
					Return(0, nil)
			})(t)

			r := NewRequest(&sync.Mutex{}, http.MethodGet, "/").
				ReturnCode(200)

			tc.mock(r)

			startTime := time.Now()

			err := r.handle(w, nil, nil)

			endTime := time.Now()

			assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
			assert.NoError(t, err)
		})
	}
}
