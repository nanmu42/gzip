package gzip

import (
	"net/http"
	"path"
	"strings"
)

// filter decide whether or not to compress response judging by request
type RequestFilter interface {
	// ShouldCompress decide whether or not to compress response,
	// judging by request
	ShouldCompress(req *http.Request) bool
}

// CommonCaseFilter judge via common easy criteria like
// http method, accept-encoding header, etc.
type CommonCaseFilter struct{}

// ShouldCompress implements RequestFilter interface
func (c *CommonCaseFilter) ShouldCompress(req *http.Request) bool {
	switch true {
	case req.Method == http.MethodHead,
		req.Header.Get("Upgrade") != "",
		!strings.Contains(req.Header.Get("Accept-Encoding"), "gzip"):
		return false
	}

	return true
}

// ExtensionFilter judge via the extension in path
//
// Omit this filter if you want to compress all extension.
type ExtensionFilter struct {
	Exts Set
}

// ShouldCompress implements RequestFilter interface
func (e *ExtensionFilter) ShouldCompress(req *http.Request) bool {
	return e.Exts.Contains(path.Ext(req.URL.Path))
}

// defaultExtensions is the list of default extensions for which to enable gzip.
// original source:
// https://github.com/caddyserver/caddy/blob/7fa90f08aee0861187236b2fbea16b4fa69c5a28/caddyhttp/gzip/requestfilter.go#L32
var defaultExtensions = []string{"", ".txt", ".htm", ".html", ".css", ".php", ".js", ".json",
	".md", ".mdown", ".xml", ".svg", ".go", ".cgi", ".py", ".pl", ".aspx", ".asp", ".m3u", ".m3u8", ".wasm"}

// DefaultExtensionFilter permits
func DefaultExtensionFilter() *ExtensionFilter {
	var exts = make(Set)

	for _, item := range defaultExtensions {
		exts.Add(item)
	}

	return &ExtensionFilter{Exts: exts}
}
