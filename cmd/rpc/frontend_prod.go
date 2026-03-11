//go:build !dev

package main

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:frontend_dist
var frontendDistFS embed.FS

func frontendHandler() http.Handler {
	sub, err := fs.Sub(frontendDistFS, "frontend_dist")
	if err != nil {
		panic("failed to create sub filesystem: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fs.Stat(sub, path); err != nil {
			if os.IsNotExist(err) {
				r.URL.Path = "/"
			}
		} else {
			ext := filepath.Ext(path)
			switch ext {
			case ".js":
				w.Header().Set("Content-Type", "application/javascript")
			case ".css":
				w.Header().Set("Content-Type", "text/css")
			case ".svg":
				w.Header().Set("Content-Type", "image/svg+xml")
			}
		}

		fileServer.ServeHTTP(w, r)
	})
}
