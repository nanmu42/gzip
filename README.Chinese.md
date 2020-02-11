[English](https://github.com/nanmu42/gzip/blob/master/README.md) | **中文**

# gzip

[![GoDoc](https://godoc.org/github.com/nanmu42/gzip?status.svg)](https://godoc.org/github.com/nanmu42/gzip)
[![Build status](https://github.com/nanmu42/gzip/workflows/build/badge.svg)](https://github.com/nanmu42/gzip/actions)
[![codecov](https://codecov.io/gh/nanmu42/gzip/branch/master/graph/badge.svg)](https://codecov.io/gh/nanmu42/gzip)
[![Lint status](https://github.com/nanmu42/gzip/workflows/golangci-lint/badge.svg)](https://github.com/nanmu42/gzip/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/nanmu42/gzip)](https://goreportcard.com/report/github.com/nanmu42/gzip)

一个开箱即用，可定制，适用于[Gin](https://github.com/gin-gonic/gin)和[net/http](https://golang.org/pkg/net/http/)的gzip中间件。

# 使用示例

默认设置开箱即用，可以满足大部分场景。

## Gin

```go
import github.com/nanmu42/gzip

func main() {
	g := gin.Default()
	
    // 使用默认设定
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

    // 使用默认设定
	log.Println(http.ListenAndServe(fmt.Sprintf(":%d", 3001), gzip.DefaultHandler().WrapHandler(mux)))
}

func writeString(w http.ResponseWriter, payload string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf8")
	_, _ = io.WriteString(w, payload+"\n")
}
```

## 定制`Handler`

在创建`Handler`时，可以定制参数以满足你的需要：

```go
import github.com/nanmu42/gzip

handler := gzip.NewHandler(gzip.Config{
    // gzip压缩等级
	CompressionLevel: 6,
    // 触发gzip的最小body体积，单位：byte
	MinContentLength: 1024,
    // 请求过滤器基于请求来判断是否对这条请求的返回启用gzip，
    // 过滤器按其定义顺序执行，下同。
	RequestFilter: []RequestFilter{
	    NewCommonRequestFilter(),
	    DefaultExtensionFilter(),
	},
    // 返回header过滤器基于返回的header判断是否对这条请求的返回启用gzip
	ResponseHeaderFilter: []ResponseHeaderFilter{
		NewSkipCompressedFilter(),
		DefaultContentTypeFilter(),
	},
})
```

`RequestFilter` 和 `ResponseHeaderFilter` 是 interface.
你可以实现你自己的过滤器。

# 效率

本中间件经过了性能调优，以确保高效运行，[查看benchmark](https://github.com/nanmu42/gzip/blob/master/docs/benchmarks.md)。

# 局限性

* 你应该总是在返回中提供`Content-Type`。虽然Handler会在`Content-Type`缺失时使用`http.DetectContentType()`进行猜测，但是效果并没有那么好；
* 返回的`Content-Length` 缺失时，Handler可能会缓冲返回的报文数据以决定报文是否大到值得进行压缩，如果`MinContentLength`设置得太大，这个过程可能会带来内存压力。Handler针对这个情况做了一些优化，例如查看`http.ResponseWriter.Write(data []byte)`在首次调用时的 `len(data)`，以及资源复用。

# 项目状态：Beta

API基本稳定，但仍可能变更。

代码可用于测试环境，但需要格外关注其表现。

欢迎在测试环境或不重要的环境使用本项目。

欢迎提PR和Issue.

# 致谢

在本项目的开发中，作者参考了下列项目和资料：

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