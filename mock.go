package httpmock

import (
	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/httpmock/test"
)

// Mocker is a function that applies expectations to the mocked server.
type Mocker func(t test.T) *Server

// MockServer creates a mocked server.
func MockServer(mocks ...func(s *Server)) *Server {
	s := NewServer()

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
//   		ReturnCode(httpmock.StatusCreated).
//   		Return(`{"id":1,"foo":"bar"}`)
//   })(t)
//
//   code, _, body, _ := httpmock.DoRequest(t,
//   	httpmock.MethodPost,
//   	s.URL()+"/create",
//   	map[string]string{"Authorization": "Bearer token"},
//   	[]byte(`{"foo":"bar"}`),
//   )
//
//   expectedCode := httpmock.StatusCreated
//   expectedBody := []byte(`{"id":1,"foo":"bar"}`)
//
//   assert.Equal(t, expectedCode, code)
//   assert.Equal(t, expectedBody, body)
func New(mocks ...func(s *Server)) Mocker {
	return func(t test.T) *Server {
		s := MockServer(mocks...).WithTest(t)

		t.Cleanup(func() {
			assert.NoError(t, s.ExpectationsWereMet())
			s.Close()
		})

		return s
	}
}
