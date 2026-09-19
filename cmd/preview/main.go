// Preview renders the real frontend templates without starting background jobs,
// connecting to databases, or registering bot webhooks.
package main

import (
	"bytes"
	"flag"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type previewData struct {
	URI, WSHost, S3Domain      string
	Count                      []string
	Keywords, Authors, PushSum []struct {
		Board, Word string
		Count       int
	}
}

func renderPage(root, name string, data interface{}) ([]byte, error) {
	t, err := template.ParseGlob(filepath.Join(root, "public", "*.html"))
	if err != nil {
		return nil, err
	}
	t, err = t.ParseGlob(filepath.Join(root, "public", "tpls", "*.tpl"))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, name, data)
	return buf.Bytes(), err
}

func main() {
	addr := flag.String("addr", "127.0.0.1:9091", "local preview listen address")
	root := flag.String("root", ".", "repository root")
	flag.Parse()
	routes := map[string]string{"/": "telegram", "/telegram": "telegram", "/docs": "docs", "/top": "top", "/messenger": "messenger"}
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.FileServer(http.Dir(filepath.Join(*root, "public"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		page, ok := routes[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		body, err := renderPage(*root, page+".html", previewData{URI: page})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(body)
	})
	log.Printf("Frontend preview: http://%s (no database or bot connections)", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
