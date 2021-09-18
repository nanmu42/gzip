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

	mux.HandleFunc("/204", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/wrong204", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		writeString(w, fmt.Sprintf("This content is compressed: l%sng!", strings.Repeat("o", 1000)))
	})

	mux.HandleFunc("/304", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotModified)
	})

	mux.HandleFunc("/wrong304", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotModified)
		writeString(w, fmt.Sprintf("This content is compressed: l%sng!", strings.Repeat("o", 1000)))
	})

	const port = 3001

	log.Printf("Service is litsenning on port %d...", port)
	log.Println(http.ListenAndServe(fmt.Sprintf(":%d", port), gzip.DefaultHandler().WrapHandler(mux)))
}

func writeString(w http.ResponseWriter, payload string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf8")
	_, err := io.WriteString(w, payload+"\n")
	if err != nil {
		fmt.Printf("wrting body: %s\n", err)
	}
}
