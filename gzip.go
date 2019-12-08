package gzip

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler implement gzip compression for gin and net/http
type Handler struct {

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
