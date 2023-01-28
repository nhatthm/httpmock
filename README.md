> ⚠️ From `v0.9.0`, the project will be rebranded to `go.nhat.io/httpmock`. `v.8.x` is the last version with `github.com/nhatthm/httpmock`.

# HTTP Mock for Golang

[![GitHub Releases](https://img.shields.io/github/v/release/nhatthm/httpmock)](https://github.com/nhatthm/httpmock/releases/latest)
[![Build Status](https://github.com/nhatthm/httpmock/actions/workflows/test.yaml/badge.svg)](https://github.com/nhatthm/httpmock/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/nhatthm/httpmock/branch/master/graph/badge.svg?token=eTdAgDE2vR)](https://codecov.io/gh/nhatthm/httpmock)
[![Go Report Card](https://goreportcard.com/badge/go.nhat.io/httpmock)](https://goreportcard.com/report/go.nhat.io/httpmock)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/go.nhat.io/httpmock)
[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://www.paypal.com/donate/?hosted_button_id=PJZSGJN57TDJY)

**httpmock** is a mock library implementing [httptest.Server](https://golang.org/pkg/net/http/httptest/#NewServer) to
support HTTP behavioral tests.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Install](#install)
- [Usage](#usage)
- [Match a value](#match-a-value)
    - [Exact](#exact)
    - [Regexp](#regexp)
    - [JSON](#json)
    - [Custom Matcher](#custom-matcher)
- [Expect a request](#expect-a-request)
    - [Request URI](#request-uri)
    - [Request Body](#request-body)
    - [Request Header](#request-header)
    - [Response Code](#response-code)
    - [Response Header](#response-header)
    - [Response Body](#response-body)
- [Execution Plan](#execution-plan)
- [Examples](#examples)

## Prerequisites

- `Go >= 1.18`

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Install

```bash
go get go.nhat.io/httpmock
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Usage

In a nutshell, the `httpmock.Server` is wrapper around [`httptest.Server`](https://pkg.go.dev/net/http/httptest#Server).
It provides extremely powerful methods to write complex expectations and test scenarios.

For creating a basic server, you can use `httpmock.NewServer()`. It starts a new HTTP server, and you can write your
expectations right away.

However, if you use it in a test (with a `t *testing.T`), and want to stop the test when an error occurs (for example,
unexpected requests, can't read request body, etc...), use `Server.WithTest(t)`. At the end of the test, you can
use `Server.ExpectationsWereMet() error` to check if the server serves all the expectation and there is nothing left.
The approach is similar to [`stretchr/testify`](https://github.com/stretchr/testify#mock-package). Also, you need to
close the server with `Server.Close()`. Luckily, you don't have to do that for every test, there is `httpmock.New()`
method to start a new server, call `Server.ExpectationsWereMet()` and close the server at the end of the test,
automatically.

For example:

```go
package main

import (
	"testing"

	"go.nhat.io/httpmock"
)

func TestSimple(t *testing.T) {
	srv := httpmock.New(func(s *httpmock.Server) {
		s.ExpectGet("/").
			Return("hello world!")
	})(t)

	// Your request and assertions.
	// The server is ready at `srv.URL()`
}
```

After starting the server, you can use `Server.URL()` to get the address of the server.

For test table approach, you can use the `Server.Mocker`, example:

```go
package main

import (
	"testing"

	"go.nhat.io/httpmock"
)

func TestSimple(t *testing.T) {
	testCases := []struct {
		scenario   string
		mockServer httpmock.Mocker
		// other input and expectations.
	}{
		{
			scenario: "some scenario",
			mockServer: httpmock.New(func(s *httpmock.Server) {
				s.ExpectGet("/").
					Return("hello world!")
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			srv := tc.mockServer(t)

			// Your request and assertions.
		})
	}
}
```

Further reading:

- [Match a value](#match-a-value)
- [Expect a request](#expect-a-request)
- [Execution Plan](#execution-plan)

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Match a value

`httpmock` is using [`go.nhat.io/matcher`](https://go.nhat.io/matcher/v2) for matching values and that
makes `httpmock` more powerful and convenient than ever. When writing expectations for the header or the payload, you
can use any kind of matchers for your needs.

For example, the `Request.WithHeader(header string, value interface{})` means you expect a header that matches a value,
you can put any of these into the `value`:

|          Type          | Explanation                            | Example                                               |
|:----------------------:|:---------------------------------------|:------------------------------------------------------|
| `string`<br/>`[]byte`  | Match the exact string, case-sensitive | `.WithHeader("locale", "en-US")`                      |
|    `*regexp.Regexp`    | Match using `regexp.Regex.MatchString` | `.WithHeader("locale", regexp.MustCompile("^en-"))`   |
| `matcher.RegexPattern` | Match using `regexp.Regex.MatchString` | `.WithHeader("locale", matcher.RegexPattern("^en-"))` |

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Exact

`matcher.Exact` matches a value by
using [`testify/assert.ObjectsAreEqual()`](https://github.com/stretchr/testify/assert).

|             Matcher             |       Input       | Result  |
|:-------------------------------:|:-----------------:|:-------:|
|    `matcher.Exact("en-US")`     |     `"en-US"`     | `true`  |
|    `matcher.Exact("en-US")`     |     `"en-us"`     | `false` |
| `matcher.Exact([]byte("en-US))` | `[]byte("en-US")` | `true`  |
| `matcher.Exact([]byte("en-US))` |     `"en-US"`     | `false` |

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Regexp

`matcher.Regex` and `matcher.RegexPattern` match a value by
using [`Regexp.MatchString`](https://pkg.go.dev/regexp#Regexp.MatchString). `matcher.Regex`
expects a `*regexp.Regexp` while `matcher.RegexPattern` expects only a regexp pattern. However, in the end, they are the
same because `nhatthm/go-matcher` creates a new `*regexp.Regexp` from the pattern using `regexp.MustCompile(pattern)`.

Notice, if the given input is not a `string` or `[]byte`, the matcher always fails.

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### JSON

`matcher.JSON` matches a value by using [`swaggest/assertjson.FailNotEqual`](https://github.com/swaggest/assertjson).
The matcher will marshal the input if it is not a `string` or a `[]byte`, and then check against the expectation. For
example, the expectation is ``matcher.JSON(`{"message": "hello"}`)``

These inputs match that expectation:

- `{"message":"hello"}` (notice there is no space after the `:` and it still matches)
- ``[]byte(`{"message":"hello"}`)``
- `map[string]string{"message": "hello"}`
- Or any objects that produce the same JSON object after calling `json.Marshal()`

You could also ignore some fields that you don't want to match. For example, the expectation
is ``matcher.JSON(`{"name": "John Doe"}`)``.If you match it with `{"name": "John Doe", "message": "hello"}`, that will
fail because the `message` is unexpected. Therefore,
use ``matcher.JSON(`{"name": "John Doe", "message": "<ignore-diff>"}`)``

The `"<ignore-diff>"` can be used against any data types, not just the `string`. For example, `{"id": "<ignore-diff>"}`
and `{"id": 42}` is a match.

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Custom Matcher

You can use your own matcher as long as it implements
the [`matcher.Matcher`](https://github.com/nhatthm/go-matcher/blob/master/matcher.go#L12-L15) interface.

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Expect a request

### Request URI

Use the `Server.Expect(method string, requestURI interface{})`, or `Server.Expect[METHOD](requestURI interface{})` to
start a new expectation. You can put a `string`, a `[]byte` or a [`matcher.Matcher`](#match-a-value) for
the `requestURI`. If the `value` is a `string` or a `[]byte`, the URI is checked by using the [`matcher.Exact`](#exact).

For example:

```go
package main

import (
	"testing"

	"go.nhat.io/httpmock"
)

func TestSimple(t *testing.T) {
	srv := httpmock.New(func(s *httpmock.Server) {
		s.ExpectGet("/").
			Return("hello world!")
	})(t)

	// Your request and assertions.
}

```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Request Header

To check whether the header of the incoming request matches some values. You can use:

- `Request.WithHeader(key string, value interface{})`: to match a single header.
- `Request.WithHeaders(header map[string]interface{})`: to match multiple headers.

The `value` could be `string`, `[]byte`, or a [`matcher.Matcher`](#match-a-value). If the `value` is a `string` or
a `[]byte`, the header is checked by using the [`matcher.Exact`](#exact).

For example:

```go
package main

import (
	"testing"

	"go.nhat.io/httpmock"
)

func TestSimple(t *testing.T) {
	srv := httpmock.New(func(s *httpmock.Server) {
		s.ExpectGet("/").
			WithHeader("Authorization", httpmock.RegexPattern("^Bearer "))
	})(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Request Body

There are several ways to match a request body:

- `WithBody(body interface{})`: The expected body can be a `string`, a `[]byte` or a [`matcher.Matcher`](#match-a-value)
  . If it is a `string` or a `[]byte`, the request body is checked by [`matched.Exact`](#exact).
- `WithBodyf(format string, args ...interface{})`: Old school `fmt.Sprintf()` call, the request body is checked
  by [`matched.Exact`](#exact) with the result from `fmt.Sprintf()`.
- `WithBodyJSON(body interface{})`: The expected body will be marshaled using `json.Marshal()` and the request body is
  checked by [`matched.JSON`](#json).

For example:

```go
package main

import (
	"testing"

	"go.nhat.io/httpmock"
)

func TestSimple(t *testing.T) {
	srv := httpmock.New(func(s *httpmock.Server) {
		s.ExpectPost("/users").
			WithBody(httpmock.JSON(`{"id": 42}`))
	})(t)

	// Your request and assertions.
}
```

or

```go
package main

import (
	"testing"

	"go.nhat.io/httpmock"
)

func TestSimple(t *testing.T) {
	srv := httpmock.New(func(s *httpmock.Server) {
		s.ExpectPost("/users").
			WithBodyJSON(map[string]interface{}{"id": 42})
	})(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Response Code

By default, the response code is `200`. You can change it by using `ReturnCode(code int)`

For example:

```go
package main

import (
	"testing"

	"go.nhat.io/httpmock"
)

func TestSimple(t *testing.T) {
	srv := httpmock.New(func(s *httpmock.Server) {
		s.ExpectPost("/users").
			ReturnCode(httpmock.StatusCreated)
	})(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Response Header

To send a header to client, there are 2 options:

- `ReturnHeader(key, value string)`: Send a single header.
- `ReturnHeaders(header map[string]string)`: Send multiple headers.

Of course the header is not sent right away when you write the expectation but later on when the request is handled.

For example:

```go
package main

import (
	"testing"

	"go.nhat.io/httpmock"
)

func TestSimple(t *testing.T) {
	srv := httpmock.New(func(s *httpmock.Server) {
		s.ExpectGet("/").
			ReturnHeader("Content-Type", "application/json").
			Return(`{"id": 42}`)
	})(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Response Body

There are several ways to create a response for the request

| Method                                        | Explanation                                                               | Example                                                                                |
|:----------------------------------------------|:--------------------------------------------------------------------------|:---------------------------------------------------------------------------------------|
| `Return(v string,bytes,fmt.Stringer)`         | Nothing fancy, the response is the given string                           | `Return("hello world")`                                                                |
| `Returnf(format string, args ...interface{})` | Same as `Return()`, but with support for formatting using `fmt.Sprintf()` | `Returnf("hello %s", "world")`                                                         |
| `ReturnJSON(v interface{})`                   | The response is the result of `json.Marshal(v)`                           | `ReturnJSON(map[string]string{"name": "john"})`                                        |
| `ReturnFile(path string)`                     | The response is the content of given file, read by `io.ReadFile()`        | `ReturnFile("resources/fixtures/result.json")`                                         |
| `Run(func(r *http.Request) ([]byte, error))`  | Custom Logic                                                              | [See the example](https://github.com/nhatthm/httpmock/blob/master/example_test.go#L44) |

For example:

```go
package main

import (
	"testing"

	"go.nhat.io/httpmock"
)

func TestSimple(t *testing.T) {
	srv := httpmock.New(func(s *httpmock.Server) {
		s.ExpectGet("/").
			Return("hello world")
	})(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Execution Plan

The mocked HTTP server is created with the `go.nhat.io/httpmock/planner.Sequence()` by default, and it matches
incoming requests sequentially. You can easily change this behavior to match your application execution by implementing
the `planner.Planner` interface.

```go
package planner

import (
	"net/http"

	"go.nhat.io/httpmock/request"
)

type Planner interface {
	// IsEmpty checks whether the planner has no expectation.
	IsEmpty() bool
	// Expect adds a new expectation.
	Expect(expect *request.Request)
	// Plan decides how a request matches an expectation.
	Plan(req *http.Request) (*request.Request, error)
	// Remain returns remain expectations.
	Remain() []*request.Request
	// Reset removes all the expectations.
	Reset()
}
```

Then use it with `Server.WithPlanner(newPlanner)` (see
the [`ExampleMockServer_alwaysFailPlanner`](https://github.com/nhatthm/httpmock/blob/master/example_test.go#L94))

When the `Server.Expect()`, or `Server.Expect[METHOD]()` is called, the mocked server will prepare a request and sends
it to the planner. If there is an incoming request, the server will call `Planner.PLan()` to find the expectation that
matches the request and executes it.

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Examples

```go
package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.nhat.io/httpmock"
)

func TestSimple(t *testing.T) {
	mockServer := httpmock.New(func(s *httpmock.Server) {
		s.Expect(httpmock.MethodGet, "/").
			Return("hello world!")
	})

	s := mockServer(t)

	code, _, body, _ := httpmock.DoRequest(t, httpmock.MethodGet, s.URL()+"/", nil, nil)

	expectedCode := httpmock.StatusOK
	expectedBody := []byte(`hello world!`)

	assert.Equal(t, expectedCode, code)
	assert.Equal(t, expectedBody, body)

	// Success
}

func TestCustomResponse(t *testing.T) {
	mockServer := httpmock.New(func(s *httpmock.Server) {
		s.Expect(httpmock.MethodPost, "/create").
			WithHeader("Authorization", "Bearer token").
			WithBody(`{"name":"John Doe"}`).
			After(time.Second).
			ReturnCode(httpmock.StatusCreated).
			ReturnJSON(map[string]interface{}{
				"id":   1,
				"name": "John Doe",
			})
	})

	s := mockServer(t)

	requestHeader := map[string]string{"Authorization": "Bearer token"}
	requestBody := []byte(`{"name":"John Doe"}`)
	code, _, body, _ := httpmock.DoRequestWithTimeout(t, httpmock.MethodPost, s.URL()+"/create", requestHeader, requestBody, time.Second)

	expectedCode := httpmock.StatusCreated
	expectedBody := []byte(`{"id":1,"name":"John Doe"}`)

	assert.Equal(t, expectedCode, code)
	assert.Equal(t, expectedBody, body)

	// Success
}

func TestExpectationsWereNotMet(t *testing.T) {
	mockServer := httpmock.New(func(s *httpmock.Server) {
		s.Expect(httpmock.MethodGet, "/").
			Return("hello world!")

		s.Expect(httpmock.MethodPost, "/create").
			WithHeader("Authorization", "Bearer token").
			WithBody(`{"name":"John Doe"}`).
			After(time.Second).
			ReturnJSON(map[string]interface{}{
				"id":   1,
				"name": "John Doe",
			})
	})

	s := mockServer(t)

	code, _, body, _ := httpmock.DoRequest(t, httpmock.MethodGet, s.URL()+"/", nil, nil)

	expectedCode := httpmock.StatusOK
	expectedBody := []byte(`hello world!`)

	assert.Equal(t, expectedCode, code)
	assert.Equal(t, expectedBody, body)

	// The test fails with
	// Error:      	Received unexpected error:
	//             	there are remaining expectations that were not met:
	//             	- POST /create
	//             	    with header:
	//             	        Authorization: Bearer token
	//             	    with body
	//             	        {"name":"John Doe"}
}
```

[See more examples](https://github.com/nhatthm/httpmock/blob/master/example_test.go)

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Donation

If this project help you reduce time to develop, you can give me a cup of coffee :)

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Paypal donation

[![paypal](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/donate/?hosted_button_id=PJZSGJN57TDJY)

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;or scan this

<img src="https://user-images.githubusercontent.com/1154587/113494222-ad8cb200-94e6-11eb-9ef3-eb883ada222a.png" width="147px" />

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)
