package gzip

import (
	"net/http"
	"testing"
)

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
			contentTypeHeader(""),
			false,
		},
		{
			contentTypeHeader("application/json; charset=utf8"),
			true,
		},
		{
			contentTypeHeader("application/json"),
			true,
		},
		{
			contentTypeHeader("application/xml; charset=utf8"),
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
