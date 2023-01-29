package planner

import "go.nhat.io/httpmock/matcher"

// Expectation is an interface that represents an expectation.
//
//go:generate mockery --name Expectation --output ../mock/planner --outpkg planner --filename expectation.go
type Expectation interface {
	Method() string
	URIMatcher() matcher.Matcher
	HeaderMatcher() matcher.HeaderMatcher
	BodyMatcher() *matcher.BodyMatcher
	RemainTimes() uint
	Fulfilled()
	FulfilledTimes() uint
}
