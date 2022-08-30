package request

import (
	"regexp"

	"go.nhat.io/httpmock/matcher"
	"go.nhat.io/httpmock/value"
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
