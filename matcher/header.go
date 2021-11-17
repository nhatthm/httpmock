package matcher

import (
	"fmt"
	"net/http"
)

// HeaderMatcher matches the header values.
type HeaderMatcher map[string]Matcher

// Match matches the header in context.
func (m HeaderMatcher) Match(header http.Header) error {
	if len(m) == 0 {
		return nil
	}

	for h, m := range m {
		value := header.Get(h)

		matched, err := m.Match(value)
		if err != nil {
			return fmt.Errorf("could not match header: %w", err)
		}

		if !matched {
			return fmt.Errorf("header %q with value %q expected, %q received", h, m.Expected(), value) // nolint: goerr113
		}
	}

	return nil
}
