package request

import (
	"fmt"
	"net/http"
)

// FailResponse responds a failure to client.
func FailResponse(w http.ResponseWriter, format string, args ...interface{}) error {
	body := fmt.Sprintf(format, args...)

	w.WriteHeader(http.StatusInternalServerError)

	_, err := w.Write([]byte(body))

	return err
}
