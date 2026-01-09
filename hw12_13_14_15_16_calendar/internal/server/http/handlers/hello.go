package httphandlers

import (
	"fmt"
	"net/http"
)

func Hello(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := fmt.Fprintln(w, "Hello, World!"); err != nil {
		return
	}
}
