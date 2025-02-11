package format

import (
	"fmt"
	"io"
	"net/http"
	"sort"

	"go.nhat.io/httpmock/matcher"
)

const indent = "    "

// ExpectedRequest formats an expected request.
func ExpectedRequest(w io.Writer, method string, uri matcher.Matcher, header matcher.HeaderMatcher, body *matcher.BodyMatcher) {
	ExpectedRequestTimes(w, method, uri, header, body, 0, 0)
}

// ExpectedRequestTimes formats an expected request with total and remaining calls.
func ExpectedRequestTimes(w io.Writer, method string, uri matcher.Matcher, header matcher.HeaderMatcher, body *matcher.BodyMatcher, totalCalls, remainingCalls int) {
	expectedHeader := map[string]any(nil)
	if header != nil {
		expectedHeader = make(map[string]any, len(header))

		for header, m := range header {
			expectedHeader[header] = m
		}
	}

	formatRequestTimes(w, method, uri.Expected(), expectedHeader, body, totalCalls, remainingCalls)
}

// HTTPRequest formats a request.
func HTTPRequest(w io.Writer, method, uri string, header http.Header, body []byte) {
	expectedHeader := map[string]any(nil)
	if header != nil {
		expectedHeader = make(map[string]any, len(header))

		for key := range header {
			expectedHeader[key] = header.Get(key)
		}
	}

	formatRequestTimes(w, method, uri, expectedHeader, body, 0, 0)
}

func formatRequestTimes(w io.Writer, method string, uri any, header map[string]any, body any, totalCalls, remainingCalls int) {
	_, _ = fmt.Fprintf(w, "%s %s", method, formatValueInline(uri)) //nolint: errcheck

	if remainingCalls > 0 && (totalCalls != 0 || remainingCalls != 1) {
		_, _ = fmt.Fprintf(w, " (called: %d time(s), remaining: %d time(s))", totalCalls, remainingCalls) //nolint: errcheck
	}

	_, _ = fmt.Fprintln(w) //nolint: errcheck

	if len(header) > 0 {
		_, _ = fmt.Fprintf(w, "%swith header:\n", indent) //nolint: errcheck

		keys := make([]string, len(header))
		i := 0

		for k := range header {
			keys[i] = k
			i++
		}

		sort.Strings(keys)

		for _, key := range keys {
			_, _ = fmt.Fprintf(w, "%s%s%s: %s\n", indent, indent, key, formatValueInline(header[key])) //nolint: errcheck
		}
	}

	if body != nil {
		bodyStr := formatValue(body)

		if bodyStr != "" {
			_, _ = fmt.Fprintf(w, "%swith body%s\n", indent, formatType(body)) //nolint: errcheck
			_, _ = fmt.Fprintf(w, "%s%s%s\n", indent, indent, bodyStr)         //nolint: errcheck
		}
	}
}

func formatValueInline(v any) string {
	if v == nil {
		return "<nil>"
	}

	switch m := v.(type) {
	case matcher.ExactMatcher,
		matcher.FnMatcher,
		[]byte,
		string:
		return formatValue(v)

	case matcher.Callback:
		return formatValue(m.Matcher())

	case *matcher.BodyMatcher:
		return formatValue(m.Matcher())

	case matcher.Matcher:
		return fmt.Sprintf("%T(%q)", v, m.Expected())

	default:
		panic("unknown value type")
	}
}

func formatType(v any) string {
	if isNil(v) {
		return ""
	}

	switch m := v.(type) {
	case matcher.ExactMatcher,
		matcher.FnMatcher,
		[]byte,
		string:
		return ""

	case matcher.Callback:
		return formatType(m.Matcher())

	case *matcher.BodyMatcher:
		return formatType(m.Matcher())

	default:
		return fmt.Sprintf(" using %T", v)
	}
}

// nolint: cyclop
func formatValue(v any) string {
	if v == nil {
		return "<nil>"
	}

	switch m := v.(type) {
	case []byte:
		return string(m)

	case string:
		return m

	case matcher.Callback:
		return formatValue(m.Matcher())

	case matcher.FnMatcher:
		if e := m.Expected(); e != "" {
			return e
		}

		return "matches custom expectation"

	case *matcher.BodyMatcher:
		if m == nil {
			return ""
		}

		return m.Expected()

	case matcher.Matcher:
		return m.Expected()

	default:
		panic("unknown value type")
	}
}
