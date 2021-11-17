package planner

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/httpmock/request"
)

func TestNextExpectations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario        string
		requests        []*request.Request
		expectedRequest *request.Request
		expectedResult  []*request.Request
	}{
		{
			scenario: "unlimited",
			requests: []*request.Request{
				newExpectWithTimes(0),
			},
			expectedRequest: newExpectWithTimes(0),
			expectedResult: []*request.Request{
				newExpectWithTimes(0),
			},
		},
		{
			scenario: "limited",
			requests: []*request.Request{
				newExpectWithTimes(2),
			},
			expectedRequest: newExpectWithTimes(1),
			expectedResult: []*request.Request{
				newExpectWithTimes(1),
			},
		},
		{
			scenario: "finished",
			requests: []*request.Request{
				newExpectWithTimes(1),
			},
			expectedRequest: newExpectWithTimes(0),
			expectedResult:  []*request.Request{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			req, result := nextExpectations(tc.requests)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedRequest, req)
		})
	}
}
