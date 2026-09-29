package httpmock

import (
	"go.nhat.io/httpmock"
)

type expectation interface { //nolint: unused
	httpmock.Expectation
	httpmock.ExpectationHandler
}
