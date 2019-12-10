package gzip

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// These constants are copied from the gzip package
const (
	NoCompression      = gzip.NoCompression
	BestSpeed          = gzip.BestSpeed
	BestCompression    = gzip.BestCompression
	DefaultCompression = gzip.DefaultCompression
	HuffmanOnly        = gzip.HuffmanOnly
)

// Config is used in Handler initialization
type Config struct {
	// gzip compression level to use,
	// valid value: -2 ~ 9.
	//
	// see https://golang.org/pkg/compress/gzip/#NewWriterLevel
	CompressionLevel int
	// Minimum content length to trigger gzip,
	// the unit is in byte.
	//
	// Content length is obtained in response's header,
	// and len(data) of http.ResponseWriter.Write(data []byte)'s first calling
	// if header["Content-Length"] is not available.
	MinContentLength int64
	// Filters are applied in the sequence here
	RequestFilter []RequestFilter
	// Filters are applied in the sequence here
	ResponseHeaderFilter []ResponseHeaderFilter
}

// Handler implement gzip compression for gin and net/http
type Handler struct {
	compressionLevel     int
	minContentLength     int64
	requestFilter        []RequestFilter
	responseHeaderFilter []ResponseHeaderFilter
	gzipWriterPool       sync.Pool
}

// NewHandler initialized a costumed gzip handler to take care of response compression.
//
// config must not be modified after calling on NewHandler()
func NewHandler(config Config) *Handler {
	if config.CompressionLevel < HuffmanOnly || config.CompressionLevel > BestCompression {
		panic(fmt.Sprintf("gzip: invalid CompressionLevel: %d", config.CompressionLevel))
	}
	if config.MinContentLength <= 0 {
		panic(fmt.Sprintf("gzip: invalid MinContentLength: %d", config.MinContentLength))
	}

	handler := Handler{
		compressionLevel:     config.CompressionLevel,
		minContentLength:     config.MinContentLength,
		requestFilter:        config.RequestFilter,
		responseHeaderFilter: config.ResponseHeaderFilter,
		gzipWriterPool: sync.Pool{
			New: func() interface{} {
				writer, _ := gzip.NewWriterLevel(ioutil.Discard, config.CompressionLevel)
				return writer
			}},
	}

	return &handler
}

var defaultConfig = Config{
	CompressionLevel: 6,
	MinContentLength: 256,
	RequestFilter: []RequestFilter{
		NewCommonRequestFilter(),
		DefaultExtensionFilter(),
	},
	ResponseHeaderFilter: []ResponseHeaderFilter{
		NewSkipCompressedFilter(),
		DefaultContentTypeFilter(),
	},
}

// DefaultHandler creates a gzip handler to take care of response compression,
// with meaningful preset.
func DefaultHandler() *Handler {
	return NewHandler(defaultConfig)
}

func (h *Handler) getGzipWriter() *gzip.Writer {
	return h.gzipWriterPool.Get().(*gzip.Writer)
}

func (h *Handler) putGzipWriter(w *gzip.Writer) {
	if w == nil {
		return
	}

	_ = w.Close()
	w.Reset(ioutil.Discard)
	h.gzipWriterPool.Put(w)
}

type ginGzipWriter struct {
	*writerWrapper
	gin.ResponseWriter
}

var _ gin.ResponseWriter = (*ginGzipWriter)(nil)

// WriteString implements interface gin.ResponseWriter
func (g *ginGzipWriter) WriteString(s string) (int, error) {
	return g.writerWrapper.Write([]byte(s))
}

// Write implements interface gin.ResponseWriter
func (g *ginGzipWriter) Write(data []byte) (int, error) {
	return g.writerWrapper.Write(data)
}

// WriteHeader implements interface gin.ResponseWriter
func (g *ginGzipWriter) WriteHeader(code int) {
	g.writerWrapper.WriteHeader(code)
}

// WriteHeader implements interface gin.ResponseWriter
func (g *ginGzipWriter) Header() http.Header {
	return g.writerWrapper.Header()
}

// Flush implements http.Flusher
func (g *ginGzipWriter) Flush() {
	g.writerWrapper.Flush()
}

// Gin implement gin's middleware
func (h *Handler) Gin(c *gin.Context) {
	var shouldCompress = true

	for _, filter := range h.requestFilter {
		shouldCompress = filter.ShouldCompress(c.Request)
		if !shouldCompress {
			break
		}
	}

	if shouldCompress {
		wrapper := newWriterWrapper(h.responseHeaderFilter, h.minContentLength, c.Writer, h.getGzipWriter, h.putGzipWriter)
		c.Writer = &ginGzipWriter{
			ResponseWriter: c.Writer,
			writerWrapper:  wrapper,
		}
		defer wrapper.CleanUp()
	}

	c.Next()
}

// WrapHandler wraps a http.Handler, returning its gzip-enabled version
func (h *Handler) WrapHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var shouldCompress = true

		for _, filter := range h.requestFilter {
			shouldCompress = filter.ShouldCompress(r)
			if !shouldCompress {
				break
			}
		}

		if shouldCompress {
			wrapper := newWriterWrapper(h.responseHeaderFilter, h.minContentLength, w, h.getGzipWriter, h.putGzipWriter)
			w = wrapper
			defer wrapper.CleanUp()
		}

		next.ServeHTTP(w, r)
	})
}
