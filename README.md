# HTTP Mock for Golang

[![Build Status](https://github.com/nhatthm/httpmock/actions/workflows/test.yaml/badge.svg)](https://github.com/nhatthm/httpmock/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/nhatthm/httpmock/branch/master/graph/badge.svg?token=eTdAgDE2vR)](https://codecov.io/gh/nhatthm/httpmock)

*httpmock* is a mock library implementing [httptest.Server](https://golang.org/pkg/net/http/httptest/#NewServer) to support HTTP behavioral tests.

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

	code, _, body, _ := request(t, s.URL(), http.MethodGet, "/", nil, nil, 0)

	expectedCode := http.StatusOK
	expectedBody := []byte(`hello world!`)

	assert.Equal(t, expectedCode, code)
	assert.Equal(t, expectedBody, body)\
  
	// Success
}

func Test_CustomResponse(t *testing.T) {
	mockServer := httpmock.New(func(s *httpmock.Server) {
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

	requestHeader := map[string]string{"Authorization": "Bearer token"}
	requestBody := []byte(`{"name":"John Doe"}`)
	code, _, body, _ := request(t, s.URL(), http.MethodPost, "/create", requestHeader, requestBody, time.Second)

	expectedCode := http.StatusOK
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

	code, _, body, _ := request(t, s.URL(), http.MethodGet, "/", nil, nil, 0)

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

func request(
	t *testing.T,
	baseURL string,
	method, uri string,
	headers map[string]string,
	body []byte,
	waitTime time.Duration,
) (int, map[string]string, []byte, time.Duration) {
	t.Helper()

	var reqBody io.Reader

	if body != nil {
		reqBody = strings.NewReader(string(body))
	}

	req, err := http.NewRequest(method, baseURL+uri, reqBody)
	require.NoError(t, err, "could not create a new request")

	for header, value := range headers {
		req.Header.Set(header, value)
	}

	timeout := waitTime + time.Second
	client := http.Client{Timeout: timeout}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	require.NoError(t, err, "could not make a request to mocked server")

	respCode := resp.StatusCode
	respHeaders := map[string]string(nil)

	if len(resp.Header) > 0 {
		respHeaders = map[string]string{}

		for header := range resp.Header {
			respHeaders[header] = resp.Header.Get(header)
		}
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err, "could not read response body")

	err = resp.Body.Close()
	require.NoError(t, err, "could not close response body")

	return respCode, respHeaders, respBody, elapsed
}
```
