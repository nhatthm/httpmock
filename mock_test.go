package httpmock_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/httpmock"
)

type TestingT struct {
	strings.Builder
	clean func()
}

func (t *TestingT) Errorf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(t, format, args...)
}

func (t *TestingT) FailNow() {
	panic("failed")
}

func (t *TestingT) Cleanup(clean func()) {
	t.clean = clean
}

func T() *TestingT {
	return &TestingT{
		clean: func() {},
	}
}

func TestMock(t *testing.T) {
	t.Parallel()

	s := httpmock.New(func(s *Server) {
		s.ExpectPost("/create").
			WithHeader("Authorization", "Bearer token").
			WithBody(`{"foo":"bar"}`).
			ReturnCode(http.StatusCreated).
			Return(`{"id":1,"foo":"bar"}`)
	})(t)

	defer s.Close()

	code, _, body, _ := httpmock.DoRequest(t,
		http.MethodPost,
		s.URL()+"/create",
		map[string]string{"Authorization": "Bearer token"},
		[]byte(`{"foo":"bar"}`),
	)

	expectedCode := http.StatusCreated
	expectedBody := []byte(`{"id":1,"foo":"bar"}`)

	assert.Equal(t, expectedCode, code)
	assert.Equal(t, expectedBody, body)
}
