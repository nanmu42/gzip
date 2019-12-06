package gzip

import (
	"net/http"
	"net/url"
	"testing"
)

func TestCommonCaseFilter_ShouldCompress(t *testing.T) {
	tests := []struct {
		name string
		req  *http.Request
		want bool
	}{
		{
			name: "Good request",
			req:  &http.Request{Method: http.MethodPost, Header: map[string][]string{"Accept-Encoding": {"gzip"}}},
			want: true,
		},
		{
			name: "HEAD request",
			req:  &http.Request{Method: http.MethodHead, Header: map[string][]string{"Accept-Encoding": {"gzip"}}},
			want: false,
		},
		{
			name: "HTTP2 upgrade request",
			req:  &http.Request{Method: http.MethodPost, Header: map[string][]string{"Accept-Encoding": {"gzip"}, "Upgrade": {"http2"}}},
			want: false,
		},
		{
			name: "Not accepting gzip request",
			req:  &http.Request{Method: http.MethodPost},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCommonRequestFilter()
			if got := c.ShouldCompress(tt.req); got != tt.want {
				t.Errorf("ShouldCompress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtensionFilter_ShouldCompress(t *testing.T) {
	tests := []struct {
		name string
		req  *http.Request
		want bool
	}{
		{
			name: "no ext",
			req:  &http.Request{URL: mustParseURL("https://example.com/hello"), Method: http.MethodPost, Header: map[string][]string{"Accept-Encoding": {"gzip"}}},
			want: true,
		},
		{
			name: "txt",
			req:  &http.Request{URL: mustParseURL("https://example.com/a.txt"), Method: http.MethodPost, Header: map[string][]string{"Accept-Encoding": {"gzip"}}},
			want: true,
		},
		{
			name: "md",
			req:  &http.Request{URL: mustParseURL("https://example.com/a.txt.md"), Method: http.MethodPost, Header: map[string][]string{"Accept-Encoding": {"gzip"}}},
			want: true,
		},
		{
			name: "png",
			req:  &http.Request{URL: mustParseURL("https://example.com/a.exe.png"), Method: http.MethodPost, Header: map[string][]string{"Accept-Encoding": {"gzip"}}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := DefaultExtensionFilter()
			if got := e.ShouldCompress(tt.req); got != tt.want {
				t.Errorf("ShouldCompress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mustParseURL(rawurl string) (URL *url.URL) {
	URL, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}

	return
}
