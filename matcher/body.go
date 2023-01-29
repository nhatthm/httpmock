package matcher

import (
	"net/http"

	"go.nhat.io/matcher/v2"

	"go.nhat.io/httpmock/value"
)

const initActual = "<could not decode>"

var _ matcher.Matcher = (*BodyMatcher)(nil)

// BodyMatcher matches a body of a request.
type BodyMatcher struct {
	matcher matcher.Matcher
	actual  string
}

// Matcher returns the underlay matcher.
func (m *BodyMatcher) Matcher() matcher.Matcher {
	return m.matcher
}

// Match satisfies matcher.Matcher interface.
func (m *BodyMatcher) Match(in any) (bool, error) {
	m.actual = initActual

	actual, err := value.GetBody(in.(*http.Request))
	if err != nil {
		return false, err
	}

	m.actual = string(actual)

	return m.matcher.Match(m.actual)
}

// Actual returns the decoded input.
func (m BodyMatcher) Actual() string {
	return m.actual
}

// Expected returns the expectation.
func (m BodyMatcher) Expected() string {
	return m.matcher.Expected()
}

// Body initiates a new body matcher.
func Body(v any) *BodyMatcher {
	return &BodyMatcher{
		matcher: matcher.Match(v),
	}
}
