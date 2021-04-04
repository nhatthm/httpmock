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
//   s := httpmock.New(func(s *Server) {
//   	s.ExpectPost("/create").
//   		WithHeader("Authorization", "Bearer token").
//   		WithBody(`{"foo":"bar"}`).
//   		ReturnCode(http.StatusCreated).
//   		Return(`{"id":1,"foo":"bar"}`)
//   })(t)
//
//   code, _, body, _ := httpmock.DoRequest(t,
//   	http.MethodPost,
//   	s.URL()+"/create",
//   	map[string]string{"Authorization": "Bearer token"},
//   	[]byte(`{"foo":"bar"}`),
//   )
//
//   expectedCode := http.StatusCreated
//   expectedBody := []byte(`{"id":1,"foo":"bar"}`)
//
//   assert.Equal(t, expectedCode, code)
//   assert.Equal(t, expectedBody, body)
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
