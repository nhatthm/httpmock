package planner

import (
	"github.com/nhatthm/httpmock/request"
)

func newExpectWithTimes(i int) *request.Request {
	r := &request.Request{}

	request.SetRepeatability(r, i)

	return r
}
