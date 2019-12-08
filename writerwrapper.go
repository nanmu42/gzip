package gzip

import (
	"net/http"
	"strconv"
	"strings"
)

// writerWrapper wraps the originalHandler
// to test whether to gzip and gzip the body if applicable.
type writerWrapper struct {
	// header filter are applied by its sequence
	filters []ResponseHeaderFilter
	// compress or not
	shouldCompress bool
	// is header already flushed?
	headerFlushed bool
	didFirstWrite bool
	statusCode    int
	// min content length to enable compress
	minContentLength int64
	originWriter     http.ResponseWriter
}

// interface guard
var _ http.ResponseWriter = (*writerWrapper)(nil)
var _ http.Flusher = (*writerWrapper)(nil)

func (w *writerWrapper) headerWritten() bool {
	return w.statusCode != 0
}

func (w *writerWrapper) contentLengthFromHeader() *int64 {
	cl := w.Header().Get("Content-Length")
	if cl == "" {
		return nil
	}

	length, _ := strconv.ParseInt(cl, 10, 64)
	return &length
}

// Header implements http.ResponseWriter
func (w *writerWrapper) Header() http.Header {
	return w.originWriter.Header()
}

// Write implements http.ResponseWriter
func (w *writerWrapper) Write(data []byte) (int, error) {
	if !w.headerWritten() {
		w.WriteHeader(http.StatusOK)
	}

	// use origin handler directly
	if !w.shouldCompress {
		w.flushHeader()
		return w.originWriter.Write(data)
	}

	if !w.didFirstWrite {
		// first time to meet data
		contentLength := w.contentLengthFromHeader()
		contentType := w.Header().Get("")
	}
}

// WriteHeader implements http.ResponseWriter
//
// WriteHeader does not really calls originalHandler's WriteHeader,
// and the calling will actually be handler by flushHeader().
func (w *writerWrapper) WriteHeader(statusCode int) {
	if w.headerWritten() {
		return
	}

	w.statusCode = statusCode

	if !w.shouldCompress {
		return
	}

	if statusCode == http.StatusNoContent ||
		statusCode == http.StatusNotModified {
		w.shouldCompress = false
		return
	}

	if len(w.filters) == 0 {
		return
	}

	header := w.Header()
	for _, filter := range w.filters {
		w.shouldCompress = filter.ShouldCompress(header)
		if !w.shouldCompress {
			return
		}
	}
}

// flushHeader must always be called and called at last
// with WriteHeader() and/or Write() already called
func (w *writerWrapper) flushHeader() {
	if w.headerFlushed {
		return
	}

	// if neither WriteHeader() or Write() are called,
	// do nothing
	if !w.headerWritten() {
		w.headerFlushed = true
		return
	}

	if w.shouldCompress {
		header := w.Header()
		header.Del("Content-Length")
		header.Set("Content-Encoding", "gzip")
		header.Add("Vary", "Accept-Encoding")
		originalEtag := w.Header().Get("ETag")
		if originalEtag != "" && !strings.HasPrefix(originalEtag, "W/") {
			w.Header().Set("ETag", "W/"+originalEtag)
		}
	}

	w.originWriter.WriteHeader(w.statusCode)

	w.headerFlushed = true
}

// Flush implements http.Flusher
func (w *writerWrapper) Flush() {
	w.flushHeader()
	if flusher, ok := w.originWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
