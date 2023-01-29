package planner

// import (
// 	"testing"
//
// 	"github.com/stretchr/testify/assert"
//
// 	"go.nhat.io/httpmock/request"
// )
//
// func TestNextInSequence(t *testing.T) {
// 	t.Parallel()
//
// 	testCases := []struct {
// 		scenario       string
// 		requests       []Expectation
// 		expectedResult Expectation
// 		expectedRemain []Expectation
// 	}{
// 		{
// 			scenario: "unlimited",
// 			requests: []Expectation{
// 				newExpectationWithTimes(0),
// 			},
// 			expectedResult: newExpectationWithTimes(0),
// 			expectedRemain: []Expectation{
// 				newExpectationWithTimes(0),
// 			},
// 		},
// 		{
// 			scenario: "limited",
// 			requests: []Expectation{
// 				newExpectationWithTimes(2),
// 			},
// 			expectedResult: newExpectationWithTimes(1),
// 			expectedRemain: []Expectation{
// 				newExpectationWithTimes(1),
// 			},
// 		},
// 		{
// 			scenario: "finished",
// 			requests: []Expectation{
// 				newExpectationWithTimes(1),
// 			},
// 			expectedResult: newExpectationWithTimes(0),
// 			expectedRemain: []Expectation{},
// 		},
// 	}
//
// 	for _, tc := range testCases {
// 		tc := tc
// 		t.Run(tc.scenario, func(t *testing.T) {
// 			t.Parallel()
//
// 			req, result := nextInSequence(tc.requests)
//
// 			assert.Equal(t, tc.expectedRemain, result)
// 			assert.Equal(t, tc.expectedResult, req)
// 		})
// 	}
// }
//
// func newExpectationWithTimes(i int) Expectation {
// 	r := &request.Request{}
//
// 	request.SetRepeatability(r, i)
//
// 	return r
// }
