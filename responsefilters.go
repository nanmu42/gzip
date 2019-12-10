package gzip

import (
	"net/http"

	"github.com/signalsciences/ac/acascii"
)

// ResponseHeaderFilter decide whether or not to compress response
// judging by response header
type ResponseHeaderFilter interface {
	// ShouldCompress decide whether or not to compress response,
	// judging by response header
	ShouldCompress(header http.Header) bool
}

// interface guards
var (
	_ ResponseHeaderFilter = (*SkipCompressedFilter)(nil)
	_ ResponseHeaderFilter = (*ContentTypeFilter)(nil)
)

// SkipCompressedFilter judges whether content has been
// already compressed
type SkipCompressedFilter struct{}

// NewSkipCompressedFilter ...
func NewSkipCompressedFilter() *SkipCompressedFilter {
	return &SkipCompressedFilter{}
}

// ShouldCompress implements ResponseHeaderFilter interface
//
// Content-Encoding: https://tools.ietf.org/html/rfc2616#section-3.5
func (s *SkipCompressedFilter) ShouldCompress(header http.Header) bool {
	return header.Get("Content-Encoding") == "" && header.Get("Transfer-Encoding") == ""
}

// ContentTypeFilter judge via the response content type
//
// Omit this filter if you want to compress all content type.
type ContentTypeFilter struct {
	Types      *acascii.Matcher
	AllowEmpty bool
}

// NewContentTypeFilter ...
func NewContentTypeFilter(types []string) *ContentTypeFilter {
	var (
		nonEmpty   = make([]string, 0, len(types))
		allowEmpty bool
	)

	for _, item := range types {
		if item == "" {
			allowEmpty = true
			continue
		}
		nonEmpty = append(nonEmpty, item)
	}

	return &ContentTypeFilter{
		Types:      acascii.MustCompileString(nonEmpty),
		AllowEmpty: allowEmpty,
	}
}

// ShouldCompress implements RequestFilter interface
func (e *ContentTypeFilter) ShouldCompress(header http.Header) bool {
	contentType := header.Get("Content-Type")

	if contentType == "" {
		return e.AllowEmpty
	}

	return e.Types.MatchString(contentType)
}

// defaultContentType is the list of default content types for which to enable gzip.
// original source:
// https://support.cloudflare.com/hc/en-us/articles/200168396-What-will-Cloudflare-compress-
var defaultContentType = []string{"text/html", "text/richtext", "text/plain", "text/css", "text/x-script", "text/x-component", "text/x-java-source", "text/x-markdown", "application/javascript", "application/x-javascript", "text/javascript", "text/js", "image/x-icon", "application/x-perl", "application/x-httpd-cgi", "text/xml", "application/xml", "application/xml+rss", "application/json", "multipart/bag", "multipart/mixed", "application/xhtml+xml", "font/ttf", "font/otf", "font/x-woff", "image/svg+xml", "application/vnd.ms-fontobject", "application/ttf", "application/x-ttf", "application/otf", "application/x-otf", "application/truetype", "application/opentype", "application/x-opentype", "application/font-woff", "application/eot", "application/font", "application/font-sfnt", "application/wasm"}

// DefaultContentTypeFilter permits
func DefaultContentTypeFilter() *ContentTypeFilter {
	return NewContentTypeFilter(defaultContentType)
}
