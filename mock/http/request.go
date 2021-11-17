package http

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

// Request is an alias of http.Request.
type Request = http.Request

// Header is an alias of http.Header.
type Header = http.Header

const (
	// MethodGet is an alias of http.MethodGet.
	MethodGet = http.MethodGet
	// MethodPost is an alias of http.MethodPost.
	MethodPost = http.MethodPost

	// StatusOK is an lias of http.StatusOK.
	StatusOK = http.StatusOK
	// StatusCreated is an lias of http.StatusCreated.
	StatusCreated = http.StatusCreated
)

type reader struct {
	upstream io.Reader
	readErr  error
	closeErr error
}

func (r *reader) Read(p []byte) (n int, err error) {
	if r.readErr != nil {
		return 0, r.readErr
	}

	return r.upstream.Read(p)
}

func (r *reader) Close() error {
	return r.closeErr
}

func newReader(upstream io.Reader, readErr, closeErr error) *reader {
	return &reader{
		upstream: upstream,
		readErr:  readErr,
		closeErr: closeErr,
	}
}

type requestBuilder struct {
	r *http.Request
}

// BuildRequest builds a Request.
// nolint: revive
func BuildRequest() *requestBuilder {
	return &requestBuilder{
		r: httptest.NewRequest(http.MethodGet, `/`, newReader(new(bytes.Buffer), nil, nil)),
	}
}

func (b *requestBuilder) WithMethod(method string) *requestBuilder {
	b.r.Method = method

	return b
}

func (b *requestBuilder) WithURI(uri string) *requestBuilder {
	b.r.RequestURI = uri

	return b
}

func (b *requestBuilder) WithHeader(key, value string) *requestBuilder {
	b.r.Header.Set(key, value)

	return b
}

func (b *requestBuilder) WithBody(body string) *requestBuilder {
	b.r.Body.(*reader).upstream = strings.NewReader(body)

	return b
}

func (b *requestBuilder) WithBodyReadError(err error) *requestBuilder {
	b.r.Body.(*reader).readErr = err

	return b
}

func (b *requestBuilder) WithBodyCloseError(err error) *requestBuilder {
	b.r.Body.(*reader).closeErr = err

	return b
}

func (b *requestBuilder) Build() *http.Request {
	return b.r
}
