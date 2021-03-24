package gin

import (
	"bufio"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nanmu42/gzip/v2"
)

// Adapt adapts gzip handler into gin's middleware
func Adapt(handler *gzip.Handler) func(*gin.Context) {
	return func(c *gin.Context) {
		if !handler.ShouldCompress(c.Request) {
			// c.Next() is not necessary here.
			return
		}

		wrapper := handler.GetWriteWrapper()
		wrapper.Reset(c.Writer)
		originWriter := c.Writer
		c.Writer = &ginGzipWriter{
			originWriter: c.Writer,
			wrapper:      wrapper,
		}
		defer func() {
			handler.PutWriteWrapper(wrapper)
			c.Writer = originWriter
		}()

		c.Next()
	}
}

type ginGzipWriter struct {
	wrapper      *gzip.WriterWrapper
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
