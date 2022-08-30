package planner

import (
	"net/http"

	"go.nhat.io/httpmock/request"
)

// MatchRequest checks whether a request is matched.
func MatchRequest(expected *request.Request, actual *http.Request) error {
	if err := MatchMethod(expected, actual); err != nil {
		return err
	}

	if err := MatchURI(expected, actual); err != nil {
		return err
	}

	if err := MatchHeader(expected, actual); err != nil {
		return err
	}

	if err := MatchBody(expected, actual); err != nil {
		return err
	}

	return nil
}

// MatchMethod matches the method of a given request.
func MatchMethod(expected *request.Request, actual *http.Request) (err error) {
	expectedMethod := request.Method(expected)

	if expectedMethod != actual.Method {
		return NewError(expected, actual,
			"method %q expected, %q received", expectedMethod, actual.Method,
		)
	}

	return nil
}

// MatchURI matches the URI of a given request.
func MatchURI(expected *request.Request, actual *http.Request) (err error) {
	uri := request.URIMatcher(expected)

	defer func() {
		if p := recover(); p != nil {
			err = NewError(expected, actual,
				"could not match request uri: %s", recovered(p),
			)
		}
	}()

	matched, err := uri.Match(actual.RequestURI)
	if err != nil {
		return NewError(expected, actual,
			"could not match request uri: %s", err.Error(),
		)
	}

	if !matched {
		return NewError(expected, actual,
			"request uri %q expected, %q received", uri.Expected(), actual.RequestURI,
		)
	}

	return nil
}

// MatchHeader matches the header of a given request.
func MatchHeader(expected *request.Request, actual *http.Request) (err error) {
	header := request.HeaderMatcher(expected)
	if len(header) == 0 {
		return nil
	}

	defer func() {
		if p := recover(); p != nil {
			err = NewError(expected, actual,
				"could not match header: %s", recovered(p),
			)
		}
	}()

	if err := header.Match(actual.Header); err != nil {
		return NewError(expected, actual, err.Error())
	}

	return nil
}

// MatchBody matches the payload of a given request.
func MatchBody(expected *request.Request, actual *http.Request) (err error) {
	m := request.BodyMatcher(expected)
	if m == nil {
		return nil
	}

	defer func() {
		if p := recover(); p != nil {
			err = NewError(expected, actual,
				"could not match body: %s", recovered(p),
			)
		}
	}()

	matched, err := m.Match(actual)
	if err != nil {
		return NewError(expected, actual,
			"could not match body: %s", err.Error(),
		)
	}

	if !matched {
		if e := m.Expected(); e != "" {
			return NewError(expected, actual, "expected request body: %s, received: %s", m.Expected(), m.Actual())
		}

		return NewError(expected, actual, "body does not match expectation, received: %s", m.Actual())
	}

	return nil
}
