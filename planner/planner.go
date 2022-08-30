package planner

import (
	"net/http"

	"go.nhat.io/httpmock/request"
)

// Planner or Request Execution Planner is in charge of selecting the right expectation for a given request.
type Planner interface {
	// IsEmpty checks whether the planner has no expectation.
	IsEmpty() bool
	// Expect adds a new expectation.
	Expect(expect *request.Request)
	// Plan decides how a request matches an expectation.
	Plan(req *http.Request) (*request.Request, error)
	// Remain returns remain expectations.
	Remain() []*request.Request
	// Reset removes all the expectations.
	Reset()
}
