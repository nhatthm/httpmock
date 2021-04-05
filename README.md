# HTTP Mock for Golang

[![Build Status](https://github.com/nhatthm/httpmock/actions/workflows/test.yaml/badge.svg)](https://github.com/nhatthm/httpmock/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/nhatthm/httpmock/branch/master/graph/badge.svg?token=eTdAgDE2vR)](https://codecov.io/gh/nhatthm/httpmock)
[![Go Report Card](https://goreportcard.com/badge/github.com/nhatthm/httpmock)](https://goreportcard.com/report/github.com/nhatthm/httpmock)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/nhatthm/httpmock)
[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://www.paypal.com/donate/?hosted_button_id=PJZSGJN57TDJY)

**httpmock** is a mock library implementing [httptest.Server](https://golang.org/pkg/net/http/httptest/#NewServer) to support HTTP behavioral tests.

## Prerequisites

- `Go >= 1.14`

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

## Donation

If this project help you reduce time to develop, you can give me a cup of coffee :)

### Paypal donation

[![paypal](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/donate/?hosted_button_id=PJZSGJN57TDJY)

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;or scan this

<img src="https://user-images.githubusercontent.com/1154587/113494222-ad8cb200-94e6-11eb-9ef3-eb883ada222a.png" width="147px" />
