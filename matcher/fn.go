package matcher

import "go.nhat.io/matcher/v2"

var _ matcher.Matcher = (*FnMatcher)(nil)

// FnMatcher is a matcher that call itself.
type FnMatcher struct {
	match    func(actual any) (bool, error)
	expected func() string
}

// Match satisfies the matcher.Matcher interface.
func (f FnMatcher) Match(actual any) (bool, error) {
	return f.match(actual)
}

// Expected satisfies the matcher.Matcher interface.
func (f FnMatcher) Expected() string {
	return f.expected()
}

// Fn creates a new FnMatcher matcher.
func Fn(expected string, match func(actual any) (bool, error)) FnMatcher {
	return FnMatcher{
		match: match,
		expected: func() string {
			return expected
		},
	}
}
