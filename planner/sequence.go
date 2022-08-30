package planner

import (
	"net/http"
	"sync"

	"go.nhat.io/httpmock/request"
)

var _ Planner = (*sequence)(nil)

type sequence struct {
	expectations []*request.Request

	mu sync.Mutex
}

func (s *sequence) IsEmpty() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.expectations) == 0
}

func (s *sequence) Expect(expect *request.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.expectations = append(s.expectations, expect)
}

func (s *sequence) Plan(req *http.Request) (*request.Request, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := MatchRequest(s.expectations[0], req); err != nil {
		return nil, err
	}

	expected, expectations := nextExpectations(s.expectations)
	s.expectations = expectations

	return expected, nil
}

func (s *sequence) Remain() []*request.Request {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.expectations
}

func (s *sequence) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.expectations = nil
}

// Sequence creates a new Planner that matches the request sequentially.
func Sequence() Planner {
	return &sequence{}
}

func nextExpectations(expectedRequests []*request.Request) (*request.Request, []*request.Request) {
	r := expectedRequests[0]
	t := request.Repeatability(r)

	if t == 0 {
		return r, expectedRequests
	}

	if t > 0 {
		request.SetRepeatability(r, t-1)

		if request.Repeatability(r) > 0 {
			return r, expectedRequests
		}
	}

	return r, expectedRequests[1:]
}
