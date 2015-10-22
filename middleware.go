package main

import (
	"net/http"
)

func HtmlText(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
