package gzip

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/klauspost/compress/gzip"
)

// These constants are copied from the gzip package
const (
	NoCompression      = gzip.NoCompression
	BestSpeed          = gzip.BestSpeed
	BestCompression    = gzip.BestCompression
	DefaultCompression = gzip.DefaultCompression
	HuffmanOnly        = gzip.HuffmanOnly
	// Stateless will do compression but without maintaining any state
	// between Write calls, so long running responses will not take memory.
	// There will be no memory kept between Write calls,
	// but compression and speed will be suboptimal.
	// Because of this, the size of actual Write calls will affect output size.
	Stateless = gzip.StatelessCompression
)

// Config is used in Handler initialization
type Config struct {
	// gzip compression level to use,
	// valid value: -3 => 9.
	//
	// see https://golang.org/pkg/compress/gzip/#NewWriterLevel
	CompressionLevel int
	// Minimum content length to trigger gzip,
	// the unit is in byte.
	//
	// When `Content-Length` is not available, handler may buffer your writes to
	// decide if its big enough to do a meaningful compression.
	// A high `MinContentLength` may bring memory overhead,
	// although the handler tries to be smart by reusing buffers
	// and testing if `len(data)` of the first
	// `http.ResponseWriter.Write(data []byte)` calling suffices or not.
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
	wrapperPool          sync.Pool
}

// NewHandler initialized a costumed gzip handler to take care of response compression.
//
// config must not be modified after calling on NewHandler()
func NewHandler(config Config) *Handler {
	if config.CompressionLevel < Stateless || config.CompressionLevel > BestCompression {
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
	}

	handler.gzipWriterPool.New = func() interface{} {
		writer, _ := gzip.NewWriterLevel(ioutil.Discard, handler.compressionLevel)
		return writer
	}
	handler.wrapperPool.New = func() interface{} {
		return newWriterWrapper(handler.responseHeaderFilter, handler.minContentLength, nil, handler.getGzipWriter, handler.putGzipWriter)
	}

	return &handler
}

var defaultConfig = Config{
	CompressionLevel: 6,
	MinContentLength: 1 * 1024,
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

func (h *Handler) getWriteWrapper() *writerWrapper {
	return h.wrapperPool.Get().(*writerWrapper)
}

func (h *Handler) putWriteWrapper(w *writerWrapper) {
	if w == nil {
		return
	}

	w.FinishWriting()
	w.OriginWriter = nil
	h.wrapperPool.Put(w)
}

type ginGzipWriter struct {
	wrapper      *writerWrapper
	originWriter gin.ResponseWriter
}

// interface guard
var _ gin.ResponseWriter = (*ginGzipWriter)(nil)

func (g *ginGzipWriter) WriteHeaderNow() {
	g.wrapper.WriteHeaderNow()
}

func (g *ginGzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return g.originWriter.Hijack()
}

func (g *ginGzipWriter) CloseNotify() <-chan bool {
	return g.originWriter.CloseNotify()
}

func (g *ginGzipWriter) Status() int {
	return g.wrapper.Status()
}

func (g *ginGzipWriter) Size() int {
	return g.wrapper.Size()
}

func (g *ginGzipWriter) Written() bool {
	return g.wrapper.Written()
}

func (g *ginGzipWriter) Pusher() http.Pusher {
	// TODO: not sure how to implement gzip for HTTP2
	return nil
}

// WriteString implements interface gin.ResponseWriter
func (g *ginGzipWriter) WriteString(s string) (int, error) {
	return g.wrapper.Write([]byte(s))
}

// Write implements interface gin.ResponseWriter
func (g *ginGzipWriter) Write(data []byte) (int, error) {
	return g.wrapper.Write(data)
}

// WriteHeader implements interface gin.ResponseWriter
func (g *ginGzipWriter) WriteHeader(code int) {
	g.wrapper.WriteHeader(code)
}

// WriteHeader implements interface gin.ResponseWriter
func (g *ginGzipWriter) Header() http.Header {
	return g.wrapper.Header()
}

// Flush implements http.Flusher
func (g *ginGzipWriter) Flush() {
	g.wrapper.Flush()
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
		wrapper := h.getWriteWrapper()
		wrapper.Reset(c.Writer)
		originWriter := c.Writer
		c.Writer = &ginGzipWriter{
			originWriter: c.Writer,
			wrapper:      wrapper,
		}
		defer func() {
			h.putWriteWrapper(wrapper)
			c.Writer = originWriter
		}()
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
			wrapper := h.getWriteWrapper()
			wrapper.Reset(w)
			originWriter := w
			w = wrapper
			defer func() {
				h.putWriteWrapper(wrapper)
				w = originWriter
			}()
		}

		next.ServeHTTP(w, r)
	})
}
