package gzip

import (
	"fmt"
	"io"
	"net/http"
	"sync"

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
	// see https://pkg.go.dev/github.com/klauspost/compress/gzip
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
		writer, _ := gzip.NewWriterLevel(io.Discard, handler.compressionLevel)
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
	w.Reset(io.Discard)
	h.gzipWriterPool.Put(w)
}

// GetWriteWrapper provides a *writerWrapper,
// which must be later returned to the pool by PutWriteWrapper().
//
// This method should only be used for building framework adaptors.
func (h *Handler) GetWriteWrapper() *WriterWrapper {
	return h.wrapperPool.Get().(*WriterWrapper)
}

// PutWriteWrapper puts provided *writerWrapper back to the pool.
// User must not hold the reference of a returned *writerWrapper.
func (h *Handler) PutWriteWrapper(w *WriterWrapper) {
	if w == nil {
		return
	}

	w.FinishWriting()
	w.OriginWriter = nil
	h.wrapperPool.Put(w)
}

// ShouldCompress decide whether or not to compress response, judging by request
//
// This method should only be used for building framework adaptors.
func (h *Handler) ShouldCompress(request *http.Request) (shouldCompress bool) {
	shouldCompress = true

	for _, filter := range h.requestFilter {
		shouldCompress = filter.ShouldCompress(request)
		if !shouldCompress {
			break
		}
	}

	return
}

// WrapHandler wraps a http.Handler, returning its gzip-enabled version
func (h *Handler) WrapHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.ShouldCompress(r) {
			wrapper := h.GetWriteWrapper()
			wrapper.Reset(w)
			originWriter := w
			w = wrapper
			defer func() {
				h.PutWriteWrapper(wrapper)
				w = originWriter
			}()
		}

		next.ServeHTTP(w, r)
	})
}
