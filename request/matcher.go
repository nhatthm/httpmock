package request

import (
	"regexp"

	"github.com/nhatthm/httpmock/matcher"
	"github.com/nhatthm/httpmock/value"
)

func matchBody(v interface{}) *matcher.BodyMatcher {
	switch v := v.(type) {
	case matcher.Matcher,
		func() matcher.Matcher,
		*regexp.Regexp:
		return matcher.Body(v)
	}

	return matcher.Body(value.String(v))
}
