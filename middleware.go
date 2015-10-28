package main

import (
	"net/http"
)

func textHtml(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		h.ServeHTTP(w, r)
	})
}
