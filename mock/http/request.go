package http

import (
	"bytes"
	"context"
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

	// StatusOK is an alias of http.StatusOK.
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

// RequestBuilder is a builder that constructs a http request.
type RequestBuilder struct {
	r *http.Request
}

// BuildRequest builds a Request.
// nolint: revive
func BuildRequest() *RequestBuilder {
	return &RequestBuilder{
		r: httptest.NewRequest(http.MethodGet, `/`, newReader(new(bytes.Buffer), nil, nil)),
	}
}

func (b *RequestBuilder) clone() *RequestBuilder {
	return &RequestBuilder{
		r: b.r.Clone(context.Background()),
	}
}

// WithMethod sets the http method.
func (b *RequestBuilder) WithMethod(method string) *RequestBuilder {
	result := b.clone()
	result.r.Method = method

	return result
}

// WithURI sets the request uri.
func (b *RequestBuilder) WithURI(uri string) *RequestBuilder {
	result := b.clone()
	result.r.RequestURI = uri

	return result
}

// WithHeader sets the request header.
func (b *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	result := b.clone()
	result.r.Header.Set(key, value)

	return result
}

// WithBody sets the request body.
func (b *RequestBuilder) WithBody(body string) *RequestBuilder {
	result := b.clone()
	result.r.Body.(*reader).upstream = strings.NewReader(body) //nolint: errcheck

	return result
}

// WithBodyReadError sets the request body that returns an error while reading.
func (b *RequestBuilder) WithBodyReadError(err error) *RequestBuilder {
	result := b.clone()
	result.r.Body.(*reader).readErr = err //nolint: errcheck

	return result
}

// WithBodyCloseError sets the request body that returns an error while closing.
func (b *RequestBuilder) WithBodyCloseError(err error) *RequestBuilder {
	result := b.clone()
	result.r.Body.(*reader).closeErr = err //nolint: errcheck

	return result
}

// Build returns the request.
func (b *RequestBuilder) Build() *http.Request {
	return b.r.Clone(context.Background())
}
