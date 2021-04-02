package httpmock

import (
	"github.com/stretchr/testify/assert"
)

// TestingT is an interface wrapper around *testing.T.
type TestingT interface {
	Errorf(format string, args ...interface{})
	FailNow()
	Cleanup(func())
}

// Mocker is a function that applies expectations to the mocked server.
type Mocker func(t TestingT) *Server

// MockServer creates a mocked server.
func MockServer(t TestingT, mocks ...func(s *Server)) *Server {
	s := NewServer(t)

	for _, m := range mocks {
		m(s)
	}

	return s
}

// New creates a mocker server with expectations and assures that ExpectationsWereMet() is called.
//
//   mockServer := httpmock.New(func(s *Server) {
//   	s.ExpectPost("/created").
//   		WithHeader("Authorization", "Bearer token").
//   		WithBody(`{"foo":"bar"}`).
//   		ReturnCode(http.StatusCreated).
//   		Return(`{"id":1,"foo":"bar"}`)
//   })
//   server := mockServer(t)
func New(mocks ...func(s *Server)) Mocker {
	return func(t TestingT) *Server {
		s := MockServer(t, mocks...)

		t.Cleanup(func() {
			assert.NoError(t, s.ExpectationsWereMet())
			s.Close()
		})

		return s
	}
}
