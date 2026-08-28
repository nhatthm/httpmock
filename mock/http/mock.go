package http

import "testing"

// ResponseWriterMocker is ResponseWriter mocker.
type ResponseWriterMocker func(tb testing.TB) *ResponseWriter

// NopResponseWriter is no mock ResponseWriter.
var NopResponseWriter = MockResponseWriter()

// MockResponseWriter creates ResponseWriter mock with cleanup to ensure all the expectations are met.
func MockResponseWriter(mocks ...func(w *ResponseWriter)) ResponseWriterMocker { //nolint: revive
	return func(tb testing.TB) *ResponseWriter {
		tb.Helper()

		w := NewResponseWriter(tb)

		for _, m := range mocks {
			m(w)
		}

		return w
	}
}
