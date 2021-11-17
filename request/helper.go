package request

import (
	"net/http"

	"github.com/nhatthm/httpmock/matcher"
)

// Method returns the method of the expectation.
func Method(req *Request) string {
	return req.method
}

// URIMatcher returns the uri matcher of the expectation.
func URIMatcher(req *Request) matcher.Matcher {
	return req.requestURI
}

// HeaderMatcher returns the header matcher of the expectation.
func HeaderMatcher(req *Request) matcher.HeaderMatcher {
	return req.requestHeader
}

// BodyMatcher returns the body matcher of the expectation.
func BodyMatcher(req *Request) *matcher.BodyMatcher {
	return req.requestBody
}

// Repeatability gets the repeatability of the expectation.
func Repeatability(r *Request) int {
	return r.repeatability
}

// SetRepeatability sets the repeatability of the expectation.
func SetRepeatability(r *Request, i int) {
	r.repeatability = i
}

// CountCall records a call to the expectation.
func CountCall(r *Request) {
	r.totalCalls++
}

// NumCalls returns the number of times the expectation was called.
func NumCalls(r *Request) int {
	return r.totalCalls
}

// Handle handles the incoming request using the expectation.
func Handle(r *Request, w http.ResponseWriter, req *http.Request, defaultHeaders map[string]string) error {
	return r.handle(w, req, defaultHeaders)
}
