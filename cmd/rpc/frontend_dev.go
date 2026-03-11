//go:build dev

package main

import "net/http"

func frontendHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "frontend served by vite dev server in dev mode", http.StatusNotFound)
	})
}
