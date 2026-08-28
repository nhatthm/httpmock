package request

import (
	"net/http"

	"go.nhat.io/httpmock/matcher"
)

// Method returns the method of the expectation.
//
// Deprecated: Use [go.nhat.io/httpmock/planner.Expectation.Method] instead.
func Method(req *Request) string {
	return req.method
}

// URIMatcher returns the uri matcher of the expectation.
//
// Deprecated: Use [go.nhat.io/httpmock/planner.Expectation.URIMatcher] instead.
func URIMatcher(req *Request) matcher.Matcher {
	return req.requestURI
}

// HeaderMatcher returns the header matcher of the expectation.
//
// Deprecated: Use [go.nhat.io/httpmock/planner.Expectation.HeaderMatcher] instead.
func HeaderMatcher(req *Request) matcher.HeaderMatcher {
	return req.requestHeader
}

// BodyMatcher returns the body matcher of the expectation.
//
// Deprecated: Use [go.nhat.io/httpmock/planner.Expectation.BodyMatcher] instead.
func BodyMatcher(req *Request) *matcher.BodyMatcher {
	return req.requestBody
}

// Repeatability gets the repeatability of the expectation.
//
// Deprecated: Use [go.nhat.io/httpmock/planner.Expectation.RemainTimes] instead.
func Repeatability(r *Request) int {
	return r.repeatability
}

// SetRepeatability sets the repeatability of the expectation.
//
// Deprecated: the package will be removed in the future.
func SetRepeatability(r *Request, i int) {
	r.repeatability = i
}

// CountCall records a call to the expectation.
//
// Deprecated: Use [go.nhat.io/httpmock/planner.Expectation.Fulfilled] instead.
func CountCall(r *Request) {
	r.totalCalls++
}

// NumCalls returns the number of times the expectation was called.
//
// Deprecated: Use [go.nhat.io/httpmock/planner.Expectation.FulfilledTimes] instead.
func NumCalls(r *Request) int {
	return r.totalCalls
}

// Handle handles the incoming request using the expectation.
//
// Deprecated: the package will be removed in the future.
func Handle(r *Request, w http.ResponseWriter, req *http.Request, defaultHeaders map[string]string) error {
	return r.handle(w, req, defaultHeaders)
}
