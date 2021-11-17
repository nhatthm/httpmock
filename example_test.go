package httpmock_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/stretchr/testify/mock"

	"github.com/nhatthm/httpmock"
	"github.com/nhatthm/httpmock/matcher"
	plannerMock "github.com/nhatthm/httpmock/mock/planner"
	"github.com/nhatthm/httpmock/must"
)

func ExampleMockServer_simple() {
	srv := httpmock.MockServer(func(s *httpmock.Server) {
		s.ExpectGet("/hi").
			Return(`hello world`)
	})

	requestURI := srv.URL() + "/hi"
	req, err := http.NewRequestWithContext(context.Background(), httpmock.MethodGet, requestURI, nil)
	must.NotFail(err)

	resp, err := http.DefaultClient.Do(req)
	must.NotFail(err)

	defer resp.Body.Close() // nolint: errcheck

	output, err := io.ReadAll(resp.Body)
	must.NotFail(err)

	fmt.Println(resp.Status)
	fmt.Println(string(output))

	// Output:
	// 200 OK
	// hello world
}

func ExampleMockServer_customHandle() {
	srv := httpmock.MockServer(func(s *httpmock.Server) {
		s.ExpectGet(matcher.RegexPattern(`^/uri.*`)).
			WithHeader("Authorization", "Bearer token").
			Run(func(r *http.Request) ([]byte, error) {
				return []byte(r.RequestURI), nil
			})
	})

	requestURI := srv.URL() + "/uri?a=1&b=2"
	req, err := http.NewRequestWithContext(context.Background(), httpmock.MethodGet, requestURI, nil)
	must.NotFail(err)

	req.Header.Set("Authorization", "Bearer token")

	resp, err := http.DefaultClient.Do(req)
	must.NotFail(err)

	defer resp.Body.Close() // nolint: errcheck

	output, err := io.ReadAll(resp.Body)
	must.NotFail(err)

	fmt.Println(resp.Status)
	fmt.Println(string(output))

	// Output:
	// 200 OK
	// /uri?a=1&b=2
}

func ExampleMockServer_expectationsWereNotMet() {
	srv := httpmock.MockServer(func(s *httpmock.Server) {
		s.ExpectGet("/hi").
			Return(`hello world`)

		s.ExpectGet("/pay").Twice().
			Return(`paid`)
	})

	err := srv.ExpectationsWereMet()

	fmt.Println(err.Error())

	// Output:
	// there are remaining expectations that were not met:
	// - GET /hi
	// - GET /pay (called: 0 time(s), remaining: 2 time(s))
}

func ExampleMockServer_alwaysFailPlanner() {
	srv := httpmock.MockServer(func(s *httpmock.Server) {
		p := &plannerMock.Planner{}

		p.On("IsEmpty").Return(false)
		p.On("Expect", mock.Anything)
		p.On("Plan", mock.Anything).
			Return(nil, errors.New("always fail"))

		s.WithPlanner(p)

		s.ExpectGet("/hi").
			Run(func(r *http.Request) ([]byte, error) {
				panic(`this never happens`)
			})
	})

	requestURI := srv.URL() + "/hi"
	req, err := http.NewRequestWithContext(context.Background(), httpmock.MethodGet, requestURI, nil)
	must.NotFail(err)

	resp, err := http.DefaultClient.Do(req)
	must.NotFail(err)

	defer resp.Body.Close() // nolint: errcheck

	output, err := io.ReadAll(resp.Body)
	must.NotFail(err)

	fmt.Println(resp.Status)
	fmt.Println(string(output))

	// Output:
	// 500 Internal Server Error
	// always fail
}
