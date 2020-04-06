package utils

import (
	"net/http"
)

// could have additional string parameter to write the type of error
func InternalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}

