package value

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// String returns the string value of the given object.
func String(v any) string {
	switch v := v.(type) {
	case []byte:
		return string(v)

	case string:
		return v

	case fmt.Stringer:
		return v.String()
	}

	panic(ErrUnsupportedDataType)
}

// GetBody returns request body and lets it re-readable.
func GetBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	err = r.Body.Close()
	if err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewReader(body))

	return body, err
}
