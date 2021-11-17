package planner_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/nhatthm/httpmock/mock/http"
	"github.com/nhatthm/httpmock/mock/planner"
	"github.com/nhatthm/httpmock/request"
)

func TestPlanner_IsEmpty(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		mock     planner.Mocker
		expected bool
	}{
		{
			scenario: "not empty",
			mock: planner.Mock(func(p *planner.Planner) {
				p.On("IsEmpty").Return(false)
			}),
		},
		{
			scenario: "empty",
			mock: planner.Mock(func(p *planner.Planner) {
				p.On("IsEmpty").Return(true)
			}),
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.mock(t).IsEmpty())
		})
	}
}

func TestPlanner_Expect(t *testing.T) {
	t.Parallel()

	p := planner.Mock(func(p *planner.Planner) {
		p.On("Expect", (*request.Request)(nil))
	})(t)

	p.Expect((*request.Request)(nil))
}

func TestPlanner_Plan(t *testing.T) {
	t.Parallel()

	r := http.BuildRequest().Build()

	testCases := []struct {
		scenario       string
		mock           planner.Mocker
		expectedResult *request.Request
		expectedError  error
	}{
		{
			scenario: "error",
			mock: planner.Mock(func(p *planner.Planner) {
				p.On("Plan", mock.Anything).
					Return(nil, errors.New("plan error"))
			}),
			expectedError: errors.New("plan error"),
		},
		{
			scenario: "nil and no error",
			mock: planner.Mock(func(p *planner.Planner) {
				p.On("Plan", mock.Anything).
					Return(nil, nil)
			}),
		},
		{
			scenario: "nil interface and no error",
			mock: planner.Mock(func(p *planner.Planner) {
				p.On("Plan", mock.Anything).
					Return((*request.Request)(nil), nil)
			}),
		},
		{
			scenario: "success",
			mock: planner.Mock(func(p *planner.Planner) {
				p.On("Plan", r).
					Return(&request.Request{}, nil)
			}),
			expectedResult: &request.Request{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			req, err := tc.mock(t).Plan(r)

			assert.Equal(t, tc.expectedResult, req)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestPlanner_Remain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		mock     planner.Mocker
		expected []*request.Request
	}{
		{
			scenario: "nil",
			mock: planner.Mock(func(p *planner.Planner) {
				p.On("Remain").Return(nil)
			}),
		},
		{
			scenario: "nil slice",
			mock: planner.Mock(func(p *planner.Planner) {
				p.On("Remain").Return(([]*request.Request)(nil))
			}),
		},
		{
			scenario: "not nil",
			mock: planner.Mock(func(p *planner.Planner) {
				p.On("Remain").Return([]*request.Request{})
			}),
			expected: []*request.Request{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.mock(t).Remain())
		})
	}
}

func TestPlanner_Reset(t *testing.T) {
	t.Parallel()

	p := planner.Mock(func(p *planner.Planner) {
		p.On("Reset")
	})(t)

	p.Reset()
}
