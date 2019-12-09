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

// Handle implement gin's middleware
func (h *Handler) Gin(c *gin.Context) {
	panic("implement me")
}

// ServeHTTP implement http.Handler
func (h *Handler) WrapHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("implement me")
	})
}
