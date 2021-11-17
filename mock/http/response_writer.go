package http

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ResponseWriterMocker is ResponseWriter mocker.
type ResponseWriterMocker func(tb testing.TB) *ResponseWriter

// NoMockResponseWriter is no mock ResponseWriter.
var NoMockResponseWriter = MockResponseWriter()

var _ http.ResponseWriter = (*ResponseWriter)(nil)

// ResponseWriter is a http.ResponseWriter.
type ResponseWriter struct {
	mock.Mock
}

// Header satisfies http.ResponseWriter interface.
func (r *ResponseWriter) Header() http.Header {
	result := r.Called().Get(0)

	if result == nil {
		return nil
	}

	if v, ok := result.(map[string][]string); ok {
		return v
	}

	return result.(http.Header)
}

// Write satisfies http.ResponseWriter interface.
func (r *ResponseWriter) Write(bytes []byte) (int, error) {
	result := r.Called(bytes)

	return result.Int(0), result.Error(1)
}

// WriteHeader satisfies http.ResponseWriter interface.
func (r *ResponseWriter) WriteHeader(statusCode int) {
	r.Called(statusCode)
}

// mockResponseWriter mocks http.ResponseWriter interface.
func mockResponseWriter(mocks ...func(w *ResponseWriter)) *ResponseWriter {
	w := &ResponseWriter{}

	for _, m := range mocks {
		m(w)
	}

	return w
}

// MockResponseWriter creates ResponseWriter mock with cleanup to ensure all the expectations are met.
func MockResponseWriter(mocks ...func(w *ResponseWriter)) ResponseWriterMocker {
	return func(tb testing.TB) *ResponseWriter {
		tb.Helper()

		w := mockResponseWriter(mocks...)

		tb.Cleanup(func() {
			assert.True(tb, w.Mock.AssertExpectations(tb))
		})

		return w
	}
}
