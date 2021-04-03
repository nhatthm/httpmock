# HTTP Mock for Golang

[![Build Status](https://github.com/nhatthm/httpmock/actions/workflows/test.yaml/badge.svg)](https://github.com/nhatthm/httpmock/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/nhatthm/httpmock/branch/master/graph/badge.svg?token=eTdAgDE2vR)](https://codecov.io/gh/nhatthm/httpmock)

**httpmock** is a mock library implementing [httptest.Server](https://golang.org/pkg/net/http/httptest/#NewServer) to support HTTP behavioral tests.

## Install

```bash
go get github.com/nhatthm/httpmock
```

## Examples

```go
func Test_Simple(t *testing.T) {
	mockServer := httpmock.New(func(s *httpmock.Server) {
		s.Expect(http.MethodGet, "/").
			Return("hello world!")
	})

	s := mockServer(t)

	code, _, body, _ := httpmock.DoRequest(t, http.MethodGet, s.URL+"/", nil, nil)

	expectedCode := http.StatusOK
	expectedBody := []byte(`hello world!`)

	assert.Equal(t, expectedCode, code)
	assert.Equal(t, expectedBody, body)
  
	// Success
}

func Test_CustomResponse(t *testing.T) {
	mockServer := httpmock.New(func(s *httpmock.Server) {
		s.Expect(http.MethodPost, "/create").
			WithHeader("Authorization", "Bearer token").
			WithBody(`{"name":"John Doe"}`).
			After(time.Second).
			ReturnCode(http.StatusCreated)
			ReturnJSON(map[string]interface{}{
				"id":   1,
				"name": "John Doe",
			})
	})

	s := mockServer(t)

	requestHeader := map[string]string{"Authorization": "Bearer token"}
	requestBody := []byte(`{"name":"John Doe"}`)
	code, _, body, _ := httpmock.DoRequestWithTimeout(t, http.MethodPost, s.URL()+"/create", requestHeader, requestBody, time.Second)

	expectedCode := http.StatusCreated
	expectedBody := []byte(`{"id":1,"name":"John Doe"}`)

	assert.Equal(t, expectedCode, code)
	assert.Equal(t, expectedBody, body)

	// Success
}

func Test_ExpectationsWereNotMet(t *testing.T) {
	mockServer := httpmock.New(func(s *httpmock.Server) {
		s.Expect(http.MethodGet, "/").
			Return("hello world!")

		s.Expect(http.MethodPost, "/create").
			WithHeader("Authorization", "Bearer token").
			WithBody(`{"name":"John Doe"}`).
			After(time.Second).
			ReturnJSON(map[string]interface{}{
				"id":   1,
				"name": "John Doe",
			})
	})

	s := mockServer(t)

	code, _, body, _ := httpmock.DoRequest(t, http.MethodGet, s.URL()+"/", nil, nil)

	expectedCode := http.StatusOK
	expectedBody := []byte(`hello world!`)

	assert.Equal(t, expectedCode, code)
	assert.Equal(t, expectedBody, body)
  
	// The test fails with
	// Error:      	Received unexpected error:
	//             	there are remaining expectations that were not met:
	//             	- POST /create
	//             	    with header:
	//             	        Authorization: Bearer token
	//             	    with body:
	//             	        {"name":"John Doe"}
}
```
