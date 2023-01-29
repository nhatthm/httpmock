package httpmock

import (
	"errors"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.nhat.io/httpmock/matcher"
	"go.nhat.io/httpmock/mock/http"
)

func TestRequestExpectation_WithHeader(t *testing.T) {
	t.Parallel()

	r := &requestExpectation{locker: &sync.Mutex{}, requestHeaderMatcher: matcher.HeaderMatcher{}}
	r.WithHeader("foo", "bar")

	assert.Equal(t, matcher.HeaderMatcher{"foo": matcher.Exact("bar")}, r.requestHeaderMatcher)

	r.WithHeader("john", "doe")

	assert.Equal(t, matcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, r.requestHeaderMatcher)
}

func TestRequestExpectation_WithHeaders(t *testing.T) {
	t.Parallel()

	e := newRequestExpectation(MethodGet, "/")
	e.WithHeaders(map[string]any{"foo": "bar"})

	assert.Equal(t, matcher.HeaderMatcher{"foo": matcher.Exact("bar")}, e.requestHeaderMatcher)

	e.WithHeader("john", "doe")

	assert.Equal(t, matcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, e.requestHeaderMatcher)
}

func TestRequestExpectation_WithBody(t *testing.T) {
	t.Parallel()

	const body = `{"id":42}`

	testCases := []struct {
		scenario       string
		expect         any
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

			e := newRequestExpectation(MethodGet, "/")

			e.WithBody(tc.expect)

			matched, err := e.requestBodyMatcher.Match(http.BuildRequest().
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

func TestRequestExpectation_WithBody_Panic(t *testing.T) {
	t.Parallel()

	expected := `unsupported data type`

	assert.PanicsWithError(t, expected, func() {
		(&requestExpectation{locker: &sync.Mutex{}}).
			WithBody(42)
	})
}

func TestRequestExpectation_WithBodyf(t *testing.T) {
	t.Parallel()

	r := &requestExpectation{locker: &sync.Mutex{}}

	r.WithBodyf("hello %s", "john")

	matched, err := r.requestBodyMatcher.Match(http.BuildRequest().
		WithBody(`hello john`).Build(),
	)

	assert.True(t, matched)
	assert.NoError(t, err)
}

func TestRequestExpectation_WithBodyJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario    string
		body        any
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

			r := &requestExpectation{locker: &sync.Mutex{}}

			if tc.expectPanic {
				assert.Panics(t, func() {
					r.WithBodyJSON(tc.body)
				})
			} else {
				r.WithBodyJSON(tc.body)

				matched, err := r.requestBodyMatcher.Match(http.BuildRequest().
					WithBody(`{"foo":"bar"}`).Build(),
				)

				assert.True(t, matched)
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequestExpectation_ReturnCode(t *testing.T) {
	t.Parallel()

	r := &requestExpectation{locker: &sync.Mutex{}}
	r.ReturnCode(StatusCreated)

	assert.Equal(t, StatusCreated, r.responseCode)
}

func TestRequestExpectation_ReturnHeader(t *testing.T) {
	t.Parallel()

	r := &requestExpectation{locker: &sync.Mutex{}, responseHeader: map[string]string{}}
	r.ReturnHeader("foo", "bar")

	assert.Equal(t, map[string]string{"foo": "bar"}, r.responseHeader)

	r.ReturnHeader("john", "doe")

	assert.Equal(t, map[string]string{"foo": "bar", "john": "doe"}, r.responseHeader)
}

func TestRequestExpectation_ReturnHeaders(t *testing.T) {
	t.Parallel()

	r := &requestExpectation{locker: &sync.Mutex{}}
	r.ReturnHeaders(map[string]string{"foo": "bar"})

	assert.Equal(t, map[string]string{"foo": "bar"}, r.responseHeader)

	r.ReturnHeader("john", "doe")

	assert.Equal(t, map[string]string{"foo": "bar", "john": "doe"}, r.responseHeader)
}

func TestRequestExpectation_Return(t *testing.T) {
	t.Parallel()

	t.Helper()

	testCases := []struct {
		scenario     string
		body         any
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

			e := newRequestExpectation(MethodGet, "/")

			if tc.expectPanic {
				assert.Panics(t, func() {
					e.Return(tc.body)
				})
			} else {
				e.ReturnCode(StatusOK).
					Return(tc.body)

				w := http.MockResponseWriter(func(w *http.ResponseWriter) {
					w.On("WriteHeader", StatusOK)

					w.On("Write", tc.expectedBody).
						Return(0, nil)
				})(t)

				err := e.Handle(w, http.BuildRequest().Build(), nil)

				assert.NoError(t, err)
			}
		})
	}
}

func TestRequestExpectation_Returnf(t *testing.T) {
	t.Parallel()

	w := http.MockResponseWriter(func(w *http.ResponseWriter) {
		w.On("WriteHeader", StatusOK)

		w.On("Write", []byte(`hello john`)).
			Return(0, nil)
	})(t)

	e := newRequestExpectation(MethodGet, "/")

	e.ReturnCode(StatusOK).
		Returnf("hello %s", "john")

	err := e.Handle(w, http.BuildRequest().Build(), nil)

	assert.NoError(t, err)
}

func TestRequestExpectation_ReturnJSON(t *testing.T) {
	t.Parallel()

	w := http.MockResponseWriter(func(w *http.ResponseWriter) {
		w.On("WriteHeader", StatusOK)

		w.On("Write", []byte(`{"foo":"bar"}`)).
			Return(0, nil)
	})(t)

	e := newRequestExpectation(MethodGet, "/")

	e.ReturnCode(StatusOK).
		ReturnJSON(map[string]string{"foo": "bar"})

	err := e.Handle(w, http.BuildRequest().Build(), nil)

	assert.NoError(t, err)
}

func TestRequestExpectation_ReturnFile(t *testing.T) {
	t.Parallel()

	w := http.MockResponseWriter(func(w *http.ResponseWriter) {
		w.On("WriteHeader", StatusOK)

		w.On("Write", []byte("hello world!\n")).
			Return(0, nil)
	})(t)

	e := newRequestExpectation(MethodGet, "/")

	// File does not exist.
	assert.Panics(t, func() {
		e.ReturnFile("foo")
	})

	e.ReturnCode(StatusOK).
		ReturnFile("resources/fixtures/response.txt")

	err := e.Handle(w, http.BuildRequest().Build(), nil)

	assert.NoError(t, err)
}

func TestRequestExpectation_Handle_Success(t *testing.T) {
	t.Parallel()

	responseHeader := http.Header{}

	w := http.MockResponseWriter(func(w *http.ResponseWriter) {
		w.On("Header").Return(responseHeader)
		w.On("WriteHeader", 200)

		w.On("Write", []byte(`{"id":42}`)).
			Return(9, nil)
	})(t)

	e := newRequestExpectation(MethodGet, "/")

	e.ReturnCode(StatusOK).
		ReturnHeader("Content-Type", "application/json").
		ReturnHeader("Content-Length", "9").
		Return(`{"id":42}`)

	defaultHeaders := Header{
		"Content-Type": "text/plain",
	}

	err := e.Handle(w, http.BuildRequest().Build(), defaultHeaders)

	expectedHeader := http.Header{
		"Content-Type":   {"application/json"},
		"Content-Length": {"9"},
	}

	assert.Equal(t, expectedHeader, responseHeader)
	assert.NoError(t, err)
}

func TestRequestExpectation_Handle_RunError(t *testing.T) {
	t.Parallel()

	w := http.MockResponseWriter(func(w *http.ResponseWriter) {
		w.On("WriteHeader", 500)

		w.On("Write", []byte(`run error`)).
			Return(0, nil)
	})(t)

	e := newRequestExpectation(MethodGet, "/")

	e.ReturnCode(StatusOK).
		Run(func(*http.Request) ([]byte, error) {
			return nil, errors.New("run error")
		})

	actual := e.Handle(w, http.BuildRequest().Build(), nil)
	expected := errors.New("run error")

	assert.Equal(t, expected, actual)
}

func TestRequestExpectation_Handle_WriteError(t *testing.T) {
	t.Parallel()

	w := http.MockResponseWriter(func(w *http.ResponseWriter) {
		w.On("WriteHeader", 200)

		w.On("Write", []byte(nil)).
			Return(0, errors.New("write error"))
	})(t)

	e := newRequestExpectation(MethodGet, "/")

	e.ReturnCode(StatusOK)

	actual := e.Handle(w, http.BuildRequest().Build(), nil)
	expected := errors.New("write error")

	assert.Equal(t, expected, actual)
}

func TestRequestExpectation_Once(t *testing.T) {
	t.Parallel()

	e := newRequestExpectation(MethodGet, "/")
	e.Once()

	assert.Equal(t, uint(1), e.RemainTimes())
}

func TestRequestExpectation_Twice(t *testing.T) {
	t.Parallel()

	e := newRequestExpectation(MethodGet, "/")
	e.Twice()

	assert.Equal(t, uint(2), e.RemainTimes())
}

func TestRequestExpectation_Times(t *testing.T) {
	t.Parallel()

	e := newRequestExpectation(MethodGet, "/")
	e.Times(20)

	assert.Equal(t, uint(20), e.RemainTimes())
}

func TestRequestExpectation_UnlimitedTimes(t *testing.T) {
	t.Parallel()

	e := newRequestExpectation(MethodGet, "/")
	e.UnlimitedTimes()

	assert.Equal(t, uint(0), e.RemainTimes())
}

func TestRequestExpectation_Wait(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond

	testCases := []struct {
		scenario string
		mock     func(e *requestExpectation)
	}{
		{
			scenario: "chan",
			mock: func(r *requestExpectation) {
				r.WaitUntil(time.After(duration))
			},
		},
		{
			scenario: "sleep",
			mock: func(r *requestExpectation) {
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

			e := newRequestExpectation(MethodGet, "/")

			tc.mock(e)

			startTime := time.Now()

			err := e.Handle(w, http.BuildRequest().Build(), nil)

			endTime := time.Now()

			assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
			assert.NoError(t, err)
		})
	}
}

func TestMergeHeaders(t *testing.T) {
	t.Parallel()

	headers := Header{
		"Authorization": "Bearer token",
	}

	defaultHeaders := Header{
		"Authorization": "Bearer foobar",
		"Content-Type":  "application/json",
	}

	actual := mergeHeaders(headers, defaultHeaders)
	expected := Header{
		"Authorization": "Bearer token",
		"Content-Type":  "application/json",
	}

	assert.Equal(t, expected, actual)
}
