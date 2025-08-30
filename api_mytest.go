package main

import (
	"net/http"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	testpath := r.PathValue("chirpID")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(testpath))
}
