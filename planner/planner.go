package planner

import (
	"net/http"
)

// Planner or Request Execution Planner is in charge of selecting the right expectation for a given request.
//
//go:generate mockery --name Planner --output ../mock/planner --outpkg planner --filename planner.go
type Planner interface {
	// IsEmpty checks whether the planner has no expectation.
	IsEmpty() bool
	// Expect adds a new expectation.
	Expect(e Expectation)
	// Plan decides how a request matches an expectation.
	Plan(req *http.Request) (Expectation, error)
	// Remain returns remain expectations.
	Remain() []Expectation
	// Reset removes all the expectations.
	Reset()
}
