package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func isPlainText(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return strings.HasPrefix(ct, ContentType)
}

func RespondJSON(rw http.ResponseWriter, status int, data any) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(status)
	json.NewEncoder(rw).Encode(data)
}

func RespondJSONOK(rw http.ResponseWriter, data any) {
	RespondJSON(rw, http.StatusOK, data)
}

func RespondText(rw http.ResponseWriter, status int, message string) {
	rw.Header().Set("Content-Type", ContentTypePlainText)
	rw.WriteHeader(status)
	fmt.Fprint(rw, message)
}

func RespondTextOK(rw http.ResponseWriter, message string) {
	RespondText(rw, http.StatusOK, message)
}
