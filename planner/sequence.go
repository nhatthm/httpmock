package planner

import (
	"net/http"
	"sync"
)

var _ Planner = (*sequence)(nil)

type sequence struct {
	expectations []Expectation

	mu sync.Mutex
}

func (s *sequence) IsEmpty() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.expectations) == 0
}

func (s *sequence) Expect(e Expectation) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.expectations = append(s.expectations, e)
}

func (s *sequence) Plan(req *http.Request) (Expectation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := MatchRequest(s.expectations[0], req); err != nil {
		return nil, err
	}

	expected, expectations := nextInSequence(s.expectations)
	s.expectations = expectations

	return expected, nil
}

func (s *sequence) Remain() []Expectation {
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

func nextInSequence(expectedRequests []Expectation) (Expectation, []Expectation) {
	r := expectedRequests[0]

	if trackRepeatable(r) {
		return r, expectedRequests
	}

	return r, expectedRequests[1:]
}
