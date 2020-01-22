**English** | [中文](https://github.com/nanmu42/gzip/blob/master/README.Chinese.md)

# Gzip Middleware for Go

[![GoDoc](https://godoc.org/github.com/nanmu42/gzip?status.svg)](https://godoc.org/github.com/nanmu42/gzip)
[![Build status](https://github.com/nanmu42/gzip/workflows/build/badge.svg)](https://github.com/nanmu42/gzip/actions)
[![codecov](https://codecov.io/gh/nanmu42/gzip/branch/master/graph/badge.svg)](https://codecov.io/gh/nanmu42/gzip)
[![Lint status](https://github.com/nanmu42/gzip/workflows/golangci-lint/badge.svg)](https://github.com/nanmu42/gzip/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/nanmu42/gzip)](https://goreportcard.com/report/github.com/nanmu42/gzip)

An out-of-the-box, also customizable gzip middleware for [Gin](https://github.com/gin-gonic/gin) and [net/http](https://golang.org/pkg/net/http/).

# Examples

## Gin

```go
import github.com/nanmu42/gzip

func main() {
	g := gin.Default()

    // use default settings
	g.Use(gzip.DefaultHandler().Gin)

	g.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code": 0,
			"msg":  "hello",
			"data": fmt.Sprintf("l%sng!", strings.Repeat("o", 1000)),
		})
	})

	log.Println(g.Run(fmt.Sprintf(":%d", 3000)))
}
```

## net/http

```go
import github.com/nanmu42/gzip

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeString(w, fmt.Sprintf("This content is compressed: l%sng!", strings.Repeat("o", 1000)))
	})

    // wrap http.Handler using default settings
	log.Println(http.ListenAndServe(fmt.Sprintf(":%d", 3001), gzip.DefaultHandler().WrapHandler(mux)))
}

func writeString(w http.ResponseWriter, payload string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf8")
	_, _ = io.WriteString(w, payload+"\n")
}
```

## Customize Handler

Handler can be customized during initialization:

```go
import github.com/nanmu42/gzip

handler := gzip.NewHandler(gzip.Config{
    // gzip compression level to use
	CompressionLevel: 6,
    // minimum content length to trigger gzip, the unit is in byte.
	MinContentLength: 256,
    // RequestFilter decide whether or not to compress response judging by request.
    // Filters are applied in the sequence here.
	RequestFilter: []RequestFilter{
	    NewCommonRequestFilter(),
	    DefaultExtensionFilter(),
	},
    // ResponseHeaderFilter decide whether or not to compress response
    // judging by request
	ResponseHeaderFilter: []ResponseHeaderFilter{
		NewSkipCompressedFilter(),
		DefaultContentTypeFilter(),
	},
})
```

`RequestFilter` and `ResponseHeaderFilter` are interfaces.
You may define one that specially suits your need.

# Performance

This middleware is fine-tuned so that little overhead is added to your program. See [benchmarks](https://github.com/nanmu42/gzip/blob/master/docs/benchmarks.md).

# Limitation

* You should always provide a `Content-Type` in http response's header, though handler guesses by `http.DetectContentType()`as makeshift;
* When `Content-Length` is not available, handler may buffer your writes to decide if its big enough to do a meaningful compression. A high `MinContentLength` may bring memory overhead, although the handler tries to be smart by reusing buffers and testing if `len(data)` of the first `http.ResponseWriter.Write(data []byte)` calling suffices or not.

# Status: v0

API not stabilized, things may breaks.

You are welcome to try gzip on your test/unimportant environments.

Pull requests and issues are welcome.

# Acknowledgement

During the development of this work, the author took following works/materials as reference:

* https://github.com/caddyserver/caddy/tree/master/caddyhttp/gzip (Apache License 2.0)
* https://github.com/gin-contrib/gzip (MIT License)
* https://blog.cloudflare.com/results-experimenting-brotli/
* https://support.cloudflare.com/hc/en-us/articles/200168396-What-will-Cloudflare-compress-

# License

```
MIT License
Copyright (c) 2019 LI Zhennan

Caddy is licensed under the Apache License
Copyright 2015 Light Code Labs, LLC
```