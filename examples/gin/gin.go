package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nanmu42/gzip"
)

func main() {
	g := gin.Default()

	g.Use(gzip.DefaultHandler().Gin)

	g.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code": 0,
			"msg":  "hello",
			"data": "GET /short and GET /long to have a try!",
		})
	})

	// short response will not be compressed
	g.GET("/short", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code": 0,
			"msg":  "This content is not long enough to be compressed.",
			"data": "short!",
		})
	})

	// long response that will be compressed by gzip
	g.GET("/long", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code": 0,
			"msg":  "This content is compressed",
			"data": fmt.Sprintf("l%sng!", strings.Repeat("o", 1000)),
		})
	})

	g.GET("/204", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	g.GET("/304", func(c *gin.Context) {
		c.Status(http.StatusNotModified)
	})

	const port = 3000

	log.Printf("Service is litsenning on port %d...", port)
	log.Println(g.Run(fmt.Sprintf(":%d", port)))
}
