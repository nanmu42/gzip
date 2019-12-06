package gzip

import (
	"net/http"
	"strconv"
	"testing"
)

func TestContentLengthFilter_ShouldCompress(t *testing.T) {
	const min = 20

	tests := []struct {
		name   string
		header http.Header
		want   bool
	}{
		{
			"no content length",
			make(http.Header),
			true,
		},
		{
			"invalid content length",
			http.Header{"Content-Length": []string{"-1"}},
			false,
		},
		{
			"small content length",
			http.Header{"Content-Length": []string{strconv.Itoa(min - 1)}},
			false,
		},
		{
			"enough content length",
			http.Header{"Content-Length": []string{strconv.Itoa(min)}},
			true,
		},
		{
			"big content length",
			http.Header{"Content-Length": []string{strconv.Itoa(min + 100)}},
			true,
		},
	}

	c := NewContentLengthFilter(min)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.ShouldCompress(tt.header); got != tt.want {
				t.Errorf("ShouldCompress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkipCompressedFilter_ShouldCompress(t *testing.T) {
	tests := []struct {
		name   string
		header http.Header
		want   bool
	}{
		{
			"should pass",
			make(http.Header),
			true,
		},
		{
			"gzip Content-Encoding",
			http.Header{"Content-Encoding": []string{"gzip"}},
			false,
		},
		{
			"br Content-Encoding",
			http.Header{"Content-Encoding": []string{"br"}},
			false,
		},
		{
			"complex Content-Encoding",
			http.Header{"Content-Encoding": []string{"deflate, gzip"}},
			false,
		},
		{
			"br Transfer-Encoding",
			http.Header{"Transfer-Encoding": []string{"br"}},
			false,
		},
		{
			"gzip Transfer-Encoding",
			http.Header{"Transfer-Encoding": []string{"gzip"}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SkipCompressedFilter{}
			if got := s.ShouldCompress(tt.header); got != tt.want {
				t.Errorf("ShouldCompress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContentTypeFilter_ShouldCompress(t *testing.T) {
	tests := []struct {
		header http.Header
		want   bool
	}{
		{
			contentTypeHeader("application/json; chatset=utf8"),
			true,
		},
		{
			contentTypeHeader("application/xml; chatset=utf8"),
			true,
		},
		{
			contentTypeHeader("image/png"),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.header.Get("Content-Type"), func(t *testing.T) {
			e := DefaultContentTypeFilter()
			if got := e.ShouldCompress(tt.header); got != tt.want {
				t.Errorf("ShouldCompress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func contentTypeHeader(contentType string) http.Header {
	return http.Header{"Content-Type": []string{contentType}}
}
