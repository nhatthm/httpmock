package request

import "net/textproto"

// Header is an alias of a string map.
//
// Deprecated: the package will be removed in the future.
type Header = map[string]string

// mergeHeaders merges a list of headers with some defaults. If a default header appears in the given headers, it
// will not be merged, no matter what the value is.
func mergeHeaders(headers, defaultHeaders Header) Header {
	result := make(Header, len(headers)+len(defaultHeaders))

	for header, value := range defaultHeaders {
		result[textproto.CanonicalMIMEHeaderKey(header)] = value
	}

	for header, value := range headers {
		result[textproto.CanonicalMIMEHeaderKey(header)] = value
	}

	return result
}
