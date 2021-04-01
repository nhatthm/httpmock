package httpmock

import (
	"net/http"

	"github.com/stretchr/testify/assert"
)

// RequestMatcherOption configures RequestMatcherConfig.
type RequestMatcherOption func(c *RequestMatcherConfig)

// RequestMatcher matches a request with one of the expectations.
type RequestMatcher func(r *http.Request, expectations []*Request) (*Request, []*Request, error)

// RequestMatcherConfig is config of RequestMatcher.
type RequestMatcherConfig struct {
	matchURI    URIMatcher
	matchBody   BodyMatcher
	matchHeader HeaderMatcher
}

// URIMatcher matches an URI with an expectation.
type URIMatcher func(expected, url string) bool

// BodyMatcher matches a body with an expectation.
type BodyMatcher func(expected, body []byte) bool

// HeaderMatcher matches header with an expectation.
type HeaderMatcher func(expected, header string) bool

// DefaultRequestMatcher instantiates a SequentialRequestMatcher with ExactURIMatcher, ExactHeaderMatcher
// and ExactBodyMatcher.
func DefaultRequestMatcher() RequestMatcher {
	return SequentialRequestMatcher(
		WithExactURIMatcher(),
		WithExactHeaderMatcher(),
		WithExactBodyMatcher(),
	)
}

// ConfigureRequestMatcher configures ConfigureRequestMatcher with RequestMatcherOption.
func ConfigureRequestMatcher(options ...RequestMatcherOption) *RequestMatcherConfig {
	m := &RequestMatcherConfig{}

	for _, o := range options {
		o(m)
	}

	if m.matchURI == nil {
		m.matchURI = ExactURIMatcher()
	}

	if m.matchHeader == nil {
		m.matchHeader = ExactHeaderMatcher()
	}

	if m.matchBody == nil {
		m.matchBody = ExactBodyMatcher()
	}

	return m
}

// SequentialRequestMatcher matches a request in sequence.
func SequentialRequestMatcher(options ...RequestMatcherOption) RequestMatcher {
	m := ConfigureRequestMatcher(options...)

	return func(r *http.Request, expectedRequests []*Request) (*Request, []*Request, error) {
		expected := expectedRequests[0]

		if expected.Method != r.Method {
			return nil, nil, MatcherError(expected, r,
				"method %q expected, %q received", expected.Method, r.Method,
			)
		}

		if !m.matchURI(expected.RequestURI, r.RequestURI) {
			return nil, nil, MatcherError(expected, r,
				"request uri %q expected, %q received", expected.RequestURI, r.RequestURI,
			)
		}

		for header, expectedValue := range expected.RequestHeader {
			value := r.Header.Get(header)
			if !m.matchHeader(expectedValue, value) {
				return nil, nil, MatcherError(expected, r,
					"header %q with value %q expected, %q received", header, expectedValue, value,
				)
			}
		}

		if expected.RequestBody != nil {
			body, err := GetBody(r)
			if err != nil {
				return nil, nil, MatcherError(expected, r, err.Error())
			}

			if !m.matchBody(expected.RequestBody, body) {
				return nil, nil, MatcherError(expected, r, "expected request body: %q, received: %q", string(expected.RequestBody), string(body))
			}
		}

		if expected.Repeatability > 0 {
			expected.Repeatability--

			if expected.Repeatability > 0 {
				return expected, expectedRequests, nil
			}
		}

		return expected, expectedRequests[1:], nil
	}
}

// WithExactURIMatcher sets URIMatcher to ExactURIMatcher.
func WithExactURIMatcher() RequestMatcherOption {
	return func(c *RequestMatcherConfig) {
		c.matchURI = ExactURIMatcher()
	}
}

// WithExactHeaderMatcher sets HeaderMatcher to ExactHeaderMatcher.
func WithExactHeaderMatcher() RequestMatcherOption {
	return func(c *RequestMatcherConfig) {
		c.matchHeader = ExactHeaderMatcher()
	}
}

// WithExactBodyMatcher sets BodyMatcher to ExactBodyMatcher.
func WithExactBodyMatcher() RequestMatcherOption {
	return func(c *RequestMatcherConfig) {
		c.matchBody = ExactBodyMatcher()
	}
}

// ExactURIMatcher matches an url by checking if it is equal to the expectation.
func ExactURIMatcher() URIMatcher {
	return func(url, expected string) bool {
		return assert.ObjectsAreEqual(expected, url)
	}
}

// ExactHeaderMatcher matches a header by checking if it is equal to the expectation.
func ExactHeaderMatcher() HeaderMatcher {
	return func(expected, header string) bool {
		return assert.ObjectsAreEqual(expected, header)
	}
}

// ExactBodyMatcher matches a body by checking if it is equal to the expectation.
func ExactBodyMatcher() BodyMatcher {
	return func(expected, body []byte) bool {
		return assert.ObjectsAreEqual(expected, body)
	}
}
