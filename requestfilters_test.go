package gzip

import (
	"net/http"
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
			c := &CommonCaseFilter{}
			if got := c.ShouldCompress(tt.req); got != tt.want {
				t.Errorf("ShouldCompress() = %v, want %v", got, tt.want)
			}
		})
	}
}
