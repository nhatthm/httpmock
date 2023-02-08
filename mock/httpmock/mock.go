package httpmock

import (
	"testing"

	"go.nhat.io/httpmock"
)

//go:generate mockery --name expectation --structname Expectation --output . --outpkg httpmock --filename expectation.go
type expectation interface { //nolint: unused
	httpmock.Expectation
	httpmock.ExpectationHandler
}

// ExpectationMocker is Expectation mocker.
type ExpectationMocker func(tb testing.TB) *Expectation

// NoMockExpectation is no mock Expectation.
var NoMockExpectation = Mock()

// Mock creates Expectation mock with cleanup to ensure all the expectations are met.
func Mock(mocks ...func(e *Expectation)) ExpectationMocker {
	return func(tb testing.TB) *Expectation {
		tb.Helper()

		e := NewExpectation(tb)

		for _, m := range mocks {
			m(e)
		}

		return e
	}
}
