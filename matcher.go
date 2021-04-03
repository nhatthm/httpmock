package httpmock

import (
	"net/http"

	"github.com/stretchr/testify/assert"
)

// RequestMatcherOption configures RequestMatcherConfig.
type RequestMatcherOption func(c *RequestMatcherConfig)

// RequestMatcher matches a request with one of the expectations.
type RequestMatcher func(t TestingT, r *http.Request, expectations []*Request) (*Request, []*Request, error)

// RequestMatcherConfig is config of RequestMatcher.
type RequestMatcherConfig struct {
	matchURI    URIMatcher
	matchBody   BodyMatcher
	matchHeader HeaderMatcher
}

// URIMatcher matches an URI with an expectation.
type URIMatcher func(t TestingT, expected, url string) bool

// BodyMatcher matches a body with an expectation.
type BodyMatcher func(t TestingT, expected, body []byte) bool

// HeaderMatcher matches header with an expectation.
type HeaderMatcher func(t TestingT, expected, header string) bool

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

	return func(t TestingT, r *http.Request, expectedRequests []*Request) (*Request, []*Request, error) {
		expected := expectedRequests[0]

		if expected.Method != r.Method {
			return nil, nil, MatcherError(expected, r,
				"method %q expected, %q received", expected.Method, r.Method,
			)
		}

		if !m.matchURI(t, expected.RequestURI, r.RequestURI) {
			return nil, nil, MatcherError(expected, r,
				"request uri %q expected, %q received", expected.RequestURI, r.RequestURI,
			)
		}

		for header, expectedValue := range expected.RequestHeader {
			value := r.Header.Get(header)
			if !m.matchHeader(t, expectedValue, value) {
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

			if !m.matchBody(t, expected.RequestBody, body) {
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

// WithURIMatcher sets URIMatcher.
func WithURIMatcher(m URIMatcher) RequestMatcherOption {
	return func(c *RequestMatcherConfig) {
		c.matchURI = m
	}
}

// WithHeaderMatcher sets HeaderMatcher.
func WithHeaderMatcher(m HeaderMatcher) RequestMatcherOption {
	return func(c *RequestMatcherConfig) {
		c.matchHeader = m
	}
}

// WithBodyMatcher sets BodyMatcher.
func WithBodyMatcher(m BodyMatcher) RequestMatcherOption {
	return func(c *RequestMatcherConfig) {
		c.matchBody = m
	}
}

// WithExactURIMatcher sets URIMatcher to ExactURIMatcher.
func WithExactURIMatcher() RequestMatcherOption {
	return WithURIMatcher(ExactURIMatcher())
}

// WithExactHeaderMatcher sets HeaderMatcher to ExactHeaderMatcher.
func WithExactHeaderMatcher() RequestMatcherOption {
	return WithHeaderMatcher(ExactHeaderMatcher())
}

// WithExactBodyMatcher sets BodyMatcher to ExactBodyMatcher.
func WithExactBodyMatcher() RequestMatcherOption {
	return WithBodyMatcher(ExactBodyMatcher())
}

// ExactURIMatcher matches an url by checking if it is equal to the expectation.
func ExactURIMatcher() URIMatcher {
	return func(_ TestingT, url, expected string) bool {
		return assert.ObjectsAreEqual(expected, url)
	}
}

// ExactHeaderMatcher matches a header by checking if it is equal to the expectation.
func ExactHeaderMatcher() HeaderMatcher {
	return func(_ TestingT, expected, header string) bool {
		return assert.ObjectsAreEqual(expected, header)
	}
}

// ExactBodyMatcher matches a body by checking if it is equal to the expectation.
func ExactBodyMatcher() BodyMatcher {
	return func(_ TestingT, expected, body []byte) bool {
		return assert.ObjectsAreEqual(expected, body)
	}
}
