package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/nanmu42/gzip"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeString(w, "GET /short and /long to have a try!")
	})
	mux.HandleFunc("/short", func(w http.ResponseWriter, r *http.Request) {
		writeString(w, "This content is not long enough to be compressed.")
	})
	mux.HandleFunc("/long", func(w http.ResponseWriter, r *http.Request) {
		writeString(w, fmt.Sprintf("This content is compressed: l%sng!", strings.Repeat("o", 1000)))
	})

	const port = 3001

	log.Printf("Service is litsenning on port %d...", port)
	log.Println(http.ListenAndServe(fmt.Sprintf(":%d", port), gzip.DefaultHandler().WrapHandler(mux)))
}

func writeString(w http.ResponseWriter, payload string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf8")
	_, _ = io.WriteString(w, payload+"\n")
}
