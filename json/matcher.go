package json

import (
	"github.com/nhatthm/httpmock"
	"github.com/swaggest/assertjson"
)

// RequestMatcher instantiates a httpmock.SequentialRequestMatcher with httpmock.ExactURIMatcher,
// httpmock.ExactHeaderMatcher and BodyMatcher.
func RequestMatcher() httpmock.RequestMatcher {
	return httpmock.SequentialRequestMatcher(
		httpmock.WithExactURIMatcher(),
		httpmock.WithExactHeaderMatcher(),
		WithBodyMatcher(),
	)
}

// WithBodyMatcher sets httpmock.BodyMatcher to JSONBodyMatcher.
func WithBodyMatcher() httpmock.RequestMatcherOption {
	return httpmock.WithBodyMatcher(BodyMatcher())
}

// BodyMatcher matches a json by checking if it is equal to the expectation.
func BodyMatcher() httpmock.BodyMatcher {
	return func(_ httpmock.TestingT, expected, body []byte) bool {
		return assertjson.FailNotEqual(expected, body) == nil
	}
}
