package matcher

import "github.com/nhatthm/go-matcher"

var _ matcher.Matcher = (*FnMatcher)(nil)

// FnMatcher is a matcher that call itself.
type FnMatcher struct {
	match    func(actual interface{}) (bool, error)
	expected func() string
}

// Match satisfies the matcher.Matcher interface.
func (f FnMatcher) Match(actual interface{}) (bool, error) {
	return f.match(actual)
}

// Expected satisfies the matcher.Matcher interface.
func (f FnMatcher) Expected() string {
	return f.expected()
}

// Fn creates a new FnMatcher matcher.
func Fn(expected string, match func(actual interface{}) (bool, error)) FnMatcher {
	return FnMatcher{
		match: match,
		expected: func() string {
			return expected
		},
	}
}
