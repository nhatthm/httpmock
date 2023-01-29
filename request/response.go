package request

import (
	"fmt"
	"net/http"
)

// FailResponse responds a failure to client.
//
// Deprecated: the package will be removed in the future.
func FailResponse(w http.ResponseWriter, format string, args ...any) error {
	body := fmt.Sprintf(format, args...)

	w.WriteHeader(http.StatusInternalServerError)

	_, err := w.Write([]byte(body))

	return err
}
