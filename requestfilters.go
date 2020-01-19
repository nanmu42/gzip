package gzip

import (
	"net/http"
	"path"
	"strings"

	"github.com/signalsciences/ac/acascii"
)

// RequestFilter decide whether or not to compress response judging by request
type RequestFilter interface {
	// ShouldCompress decide whether or not to compress response,
	// judging by request
	ShouldCompress(req *http.Request) bool
}

// interface guards
var (
	_ RequestFilter = (*CommonRequestFilter)(nil)
	_ RequestFilter = (*ExtensionFilter)(nil)
)

// CommonRequestFilter judge via common easy criteria like
// http method, accept-encoding header, etc.
type CommonRequestFilter struct{}

// NewCommonRequestFilter ...
func NewCommonRequestFilter() *CommonRequestFilter {
	return &CommonRequestFilter{}
}

// ShouldCompress implements RequestFilter interface
func (c *CommonRequestFilter) ShouldCompress(req *http.Request) bool {
	return req.Method != http.MethodHead &&
		req.Method != http.MethodOptions &&
		req.Header.Get("Upgrade") == "" &&
		strings.Contains(req.Header.Get("Accept-Encoding"), "gzip")
}

// ExtensionFilter judge via the extension in path
//
// Omit this filter if you want to compress all extension.
type ExtensionFilter struct {
	Exts       *acascii.Matcher
	AllowEmpty bool
}

// NewExtensionFilter returns a extension or panics
func NewExtensionFilter(extensions []string) *ExtensionFilter {
	var (
		exts       = make([]string, 0, len(extensions))
		allowEmpty bool
	)

	for _, item := range extensions {
		if item == "" {
			allowEmpty = true
			continue
		}
		exts = append(exts, item)
	}

	return &ExtensionFilter{
		Exts:       acascii.MustCompileString(exts),
		AllowEmpty: allowEmpty,
	}
}

// ShouldCompress implements RequestFilter interface
func (e *ExtensionFilter) ShouldCompress(req *http.Request) bool {
	ext := path.Ext(req.URL.Path)
	if ext == "" {
		return e.AllowEmpty
	}
	return e.Exts.MatchString(ext)
}

// defaultExtensions is the list of default extensions for which to enable gzip.
// original source:
// https://github.com/caddyserver/caddy/blob/7fa90f08aee0861187236b2fbea16b4fa69c5a28/caddyhttp/gzip/requestfilter.go#L32
var defaultExtensions = []string{"", ".txt", ".htm", ".html", ".css", ".php", ".js", ".json",
	".md", ".mdown", ".xml", ".svg", ".go", ".cgi", ".py", ".pl", ".aspx", ".asp", ".m3u", ".m3u8", ".wasm"}

// DefaultExtensionFilter permits
func DefaultExtensionFilter() *ExtensionFilter {
	return NewExtensionFilter(defaultExtensions)
}
