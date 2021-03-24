package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	adaptor "github.com/nanmu42/gzip/adaptors/gin/v2"
	"github.com/nanmu42/gzip/v2"
)

func main() {
	g := gin.Default()

	g.Use(adaptor.Adapt(gzip.DefaultHandler()))

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

	const port = 3000

	log.Printf("Service is litsenning on port %d...", port)
	log.Println(g.Run(fmt.Sprintf(":%d", port)))
}
