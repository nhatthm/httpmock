package request_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/httpmock/matcher"
	"github.com/nhatthm/httpmock/mock/http"
	"github.com/nhatthm/httpmock/request"
)

func TestMethod(t *testing.T) {
	t.Parallel()

	r := request.NewRequest(&sync.Mutex{}, http.MethodGet, "/")

	assert.Equal(t, http.MethodGet, request.Method(r))
}

func TestURIMatcher(t *testing.T) {
	t.Parallel()

	r := request.NewRequest(&sync.Mutex{}, http.MethodGet, "/")

	assert.Equal(t, matcher.Exact(`/`), request.URIMatcher(r))
}

func TestHeaderMatcher(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario    string
		mockRequest func(*request.Request)
		expected    matcher.HeaderMatcher
	}{
		{
			scenario: "no matcher",
		},
		{
			scenario: "has matcher",
			mockRequest: func(r *request.Request) {
				r.WithHeader("Authorization", "Bearer token")
			},
			expected: matcher.HeaderMatcher{
				"Authorization": matcher.Match("Bearer token"),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			r := request.NewRequest(&sync.Mutex{}, http.MethodGet, "/")

			if tc.mockRequest != nil {
				tc.mockRequest(r)
			}

			assert.Equal(t, tc.expected, request.HeaderMatcher(r))
		})
	}
}

func TestBodyMatcher_Match(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		body     string
		expected bool
	}{
		{
			scenario: "mismatched",
			body:     "foobar",
		},
		{
			scenario: "matched",
			body:     "payload",
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			r := request.NewRequest(&sync.Mutex{}, http.MethodGet, "/").
				WithBody(`payload`)

			matched, err := request.BodyMatcher(r).Match(http.BuildRequest().
				WithBody(tc.body).Build(),
			)

			assert.Equal(t, tc.expected, matched)
			assert.NoError(t, err)
		})
	}
}

func TestRepeatability(t *testing.T) {
	t.Parallel()

	r := request.NewRequest(&sync.Mutex{}, http.MethodGet, "/")

	assert.Equal(t, 0, request.Repeatability(r))

	request.SetRepeatability(r, 1)

	assert.Equal(t, 1, request.Repeatability(r))
}

func TestCalls(t *testing.T) {
	t.Parallel()

	r := request.NewRequest(&sync.Mutex{}, http.MethodGet, "/")

	assert.Equal(t, 0, request.NumCalls(r))

	request.CountCall(r)

	assert.Equal(t, 1, request.NumCalls(r))
}

func TestHandle(t *testing.T) {
	t.Parallel()

	r := request.NewRequest(&sync.Mutex{}, http.MethodGet, "/").
		ReturnCode(200).
		Return(`payload`)

	w := http.MockResponseWriter(func(w *http.ResponseWriter) {
		w.On("WriteHeader", 200)

		w.On("Write", []byte(`payload`)).
			Return(7, nil)
	})(t)

	err := request.Handle(r, w, nil, nil)

	assert.NoError(t, err)
}
