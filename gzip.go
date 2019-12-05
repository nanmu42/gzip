package gzip

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handler implement gzip compression for gin and net/http
type handler struct {
	// writerPool keeps gzip writer
	// writerPool sync.Pool

}

// Handle implement gin's middleware
func (h *handler) Handle(c *gin.Context) {
	panic("implement me")
}

// ServeHTTP implement http.Handler
func (h *handler) ServeHTTP(http.ResponseWriter, *http.Request) {
	panic("implement me")
}
