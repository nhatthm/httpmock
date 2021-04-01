package httpmock

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
)

// GetBody returns request body and lets it re-readable.
func GetBody(r *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	err = r.Body.Close()
	if err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewReader(body))

	return body, err
}

// mergeHeaders merges a list of headers with some defaults. If a default header appears in the given headers, it
// will not be merged, no matter what the value is.
func mergeHeaders(headers, defaultHeaders map[string]string) map[string]string {
	for header, value := range defaultHeaders {
		if _, ok := headers[header]; !ok {
			headers[header] = value
		}
	}

	return headers
}

func formatRequest(w io.Writer, method, uri string, header Header, body []byte) {
	formatRequestTimes(w, method, uri, header, body, 0, 0)
}

func formatRequestTimes(w io.Writer, method, uri string, header Header, body []byte, totalCalls, remainingCalls int) {
	_, _ = fmt.Fprintf(w, "%s %s", method, uri)

	if remainingCalls > 0 {
		_, _ = fmt.Fprintf(w, " (called: %d time(s), remaining: %d time(s))", totalCalls, remainingCalls)
	}

	_, _ = fmt.Fprintln(w)

	if len(header) > 0 {
		_, _ = fmt.Fprintf(w, "%swith header:\n", indent)

		keys := make([]string, len(header))
		i := 0

		for k := range header {
			keys[i] = k
			i++
		}

		sort.Strings(keys)

		for _, key := range keys {
			_, _ = fmt.Fprintf(w, "%s%s%s: %s\n", indent, indent, key, header[key])
		}
	}

	if body != nil {
		bodyStr := string(body)

		if bodyStr != "" {
			_, _ = fmt.Fprintf(w, "%swith body:\n", indent)
			_, _ = fmt.Fprintf(w, "%s%s%s\n", indent, indent, string(body))
		}
	}
}
