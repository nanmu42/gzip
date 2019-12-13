package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const handlerTestSize = 256

func newGinInstance(payload []byte, middleware ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	g := gin.New()
	g.HandleMethodNotAllowed = true
	g.Use(middleware...)

	g.POST("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain; charset=utf8", payload)
	})

	return g
}

func newEchoGinInstance(payload []byte, middleware ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	g := gin.New()
	g.Use(middleware...)

	g.POST("/", func(c *gin.Context) {
		var buf bytes.Buffer

		_, _ = io.Copy(&buf, c.Request.Body)
		_, _ = buf.Write(payload)

		c.Data(http.StatusOK, "text/plain; charset=utf8", buf.Bytes())
	})

	return g
}

func newHTTPInstance(payload []byte, wrapper ...func(next http.Handler) http.Handler) http.Handler {
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf8")
		_, _ = w.Write(payload)
	})

	for _, wrap := range wrapper {
		handler = wrap(handler)
	}

	return handler
}

func newEchoHTTPInstance(payload []byte, wrapper ...func(next http.Handler) http.Handler) http.Handler {
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf8")

		var buf bytes.Buffer

		_, _ = io.Copy(&buf, r.Body)
		_, _ = buf.Write(payload)
		_, _ = w.Write(buf.Bytes())
	})

	for _, wrap := range wrapper {
		handler = wrap(handler)
	}

	return handler
}

type NopWriter struct {
	header http.Header
}

func NewNopWriter() *NopWriter {
	return &NopWriter{
		header: make(http.Header),
	}
}

func (n *NopWriter) Header() http.Header {
	return n.header
}

func (n *NopWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (n *NopWriter) WriteHeader(_ int) {
	// relax
}

func TestNewHandler_Checks(t *testing.T) {
	assert.NotPanics(t, func() {
		NewHandler(Config{
			CompressionLevel: 5,
			MinContentLength: 100,
		})
	})

	assert.Panics(t, func() {
		NewHandler(Config{
			CompressionLevel: -3,
			MinContentLength: 100,
		})
	})

	assert.Panics(t, func() {
		NewHandler(Config{
			CompressionLevel: 10,
			MinContentLength: 100,
		})
	})

	assert.Panics(t, func() {
		NewHandler(Config{
			CompressionLevel: 5,
			MinContentLength: 0,
		})
	})

	assert.Panics(t, func() {
		NewHandler(Config{
			CompressionLevel: 5,
			MinContentLength: -1,
		})
	})
}

func BenchmarkSoleGin_SmallPayload(b *testing.B) {
	var (
		g = newGinInstance(smallPayload)
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.ServeHTTP(w, r)
	}

	b.StopTimer()
	if encoding := w.Header().Get("Content-Encoding"); encoding != "" {
		b.Fatalf("Content-Encoding is not empty, but %s", encoding)
	}
}

func BenchmarkGinWithDefaultHandler_SmallPayload(b *testing.B) {
	var (
		g = newGinInstance(smallPayload, DefaultHandler().Gin)
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.ServeHTTP(w, r)
	}

	b.StopTimer()
	if encoding := w.Header().Get("Content-Encoding"); encoding != "" {
		b.Fatalf("Content-Encoding is not empty, but %s", encoding)
	}
}

func BenchmarkSoleGin_BigPayload(b *testing.B) {
	var (
		g = newGinInstance(bigPayload)
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.ServeHTTP(w, r)
	}

	b.StopTimer()
	if encoding := w.Header().Get("Content-Encoding"); encoding != "" {
		b.Fatalf("Content-Encoding is not empty, but %s", encoding)
	}
}

func BenchmarkGinWithDefaultHandler_BigPayload(b *testing.B) {
	var (
		g = newGinInstance(bigPayload, DefaultHandler().Gin)
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.ServeHTTP(w, r)
	}

	b.StopTimer()
	if encoding := w.Header().Get("Content-Encoding"); encoding != "gzip" {
		b.Fatalf("Content-Encoding is not gzip, but %q", encoding)
	}
}

func TestSoloGinHandler(t *testing.T) {
	var (
		g = newGinInstance(bigPayload)
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "gzip")

	g.ServeHTTP(w, r)

	assert.Empty(t, w.Header().Get("Content-Encoding"))
}

func TestGinWithDefaultHandler(t *testing.T) {
	var (
		g = newEchoGinInstance(bigPayload, DefaultHandler().Gin)
	)

	for i := 0; i < handlerTestSize; i++ {
		var seq = strconv.Itoa(i)
		t.Run(seq, func(t *testing.T) {
			t.Parallel()

			var (
				w = httptest.NewRecorder()
				r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(seq))
			)

			r.Header.Set("Accept-Encoding", "gzip")
			g.ServeHTTP(w, r)

			result := w.Result()
			require.EqualValues(t, http.StatusOK, result.StatusCode)
			require.Equal(t, "gzip", result.Header.Get("Content-Encoding"))

			reader, err := gzip.NewReader(result.Body)
			require.NoError(t, err)
			body, err := ioutil.ReadAll(reader)
			require.NoError(t, err)
			require.True(t, bytes.HasPrefix(body, []byte(seq)))
		})
	}
}

func TestGinWithDefaultHandler_404(t *testing.T) {
	var (
		g = newGinInstance(bigPayload, DefaultHandler().Gin)
		r = httptest.NewRequest(http.MethodPost, "/404", nil)
		w = httptest.NewRecorder()
	)

	r.Header.Set("Accept-Encoding", "gzip")

	g.ServeHTTP(w, r)

	result := w.Result()

	assert.EqualValues(t, http.StatusNotFound, result.StatusCode)
	assert.Equal(t, "404 page not found", w.Body.String())
}

func TestGinWithDefaultHandler_405(t *testing.T) {
	var (
		g = newGinInstance(bigPayload, DefaultHandler().Gin)
		r = httptest.NewRequest(http.MethodPatch, "/", nil)
		w = httptest.NewRecorder()
	)

	r.Header.Set("Accept-Encoding", "gzip")

	g.ServeHTTP(w, r)

	result := w.Result()

	assert.EqualValues(t, http.StatusMethodNotAllowed, result.StatusCode)
	assert.Equal(t, "405 method not allowed", w.Body.String())
}

func TestHTTPWithDefaultHandler_404(t *testing.T) {
	var (
		g = newHTTPInstance(bigPayload, DefaultHandler().WrapHandler)
		r = httptest.NewRequest(http.MethodPost, "/404", nil)
		w = httptest.NewRecorder()
	)

	mux := http.NewServeMux()
	mux.Handle("/somewhere", g)

	r.Header.Set("Accept-Encoding", "gzip")

	mux.ServeHTTP(w, r)

	result := w.Result()

	assert.EqualValues(t, http.StatusNotFound, result.StatusCode)
	assert.Equal(t, "404 page not found\n", w.Body.String())
}

func TestSoloHTTP(t *testing.T) {
	var (
		g = newHTTPInstance(bigPayload)
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "gzip")

	g.ServeHTTP(w, r)

	assert.Empty(t, w.Header().Get("Content-Encoding"))
}

func TestHTTPWithDefaultHandler(t *testing.T) {
	var (
		g = newEchoHTTPInstance(bigPayload, DefaultHandler().WrapHandler)
	)

	for i := 0; i < handlerTestSize; i++ {
		var seq = strconv.Itoa(i)
		t.Run(seq, func(t *testing.T) {
			t.Parallel()

			var (
				w = httptest.NewRecorder()
				r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(seq))
			)

			r.Header.Set("Accept-Encoding", "gzip")
			g.ServeHTTP(w, r)

			result := w.Result()
			require.EqualValues(t, http.StatusOK, result.StatusCode)
			require.Equal(t, "gzip", result.Header.Get("Content-Encoding"))

			reader, err := gzip.NewReader(result.Body)
			require.NoError(t, err)
			body, err := ioutil.ReadAll(reader)
			require.NoError(t, err)
			require.True(t, bytes.HasPrefix(body, []byte(seq)))
		})
	}
}

func TestHTTPWithDefaultHandler_TinyPayload_WriteTwice(t *testing.T) {
	var (
		handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf8")
			_, _ = io.WriteString(w, "part 1\n")
			_, _ = io.WriteString(w, "part 2\n")
		})
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = httptest.NewRecorder()
	)

	r.Header.Set("Accept-Encoding", "gzip")
	handler = DefaultHandler().WrapHandler(handler)

	handler.ServeHTTP(w, r)

	result := w.Result()

	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.Empty(t, result.Header.Get("Vary"))
	assert.Empty(t, result.Header.Get("Content-Encoding"))
	assert.Equal(t, "part 1\npart 2\n", w.Body.String())
}

func TestHTTPWithDefaultHandler_TinyPayload_WriteThreeTimes(t *testing.T) {
	var (
		handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf8")
			_, _ = io.WriteString(w, "part 1\n")
			_, _ = io.WriteString(w, "part 2\n")
			_, _ = io.WriteString(w, "part 3\n")
		})
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = httptest.NewRecorder()
	)

	r.Header.Set("Accept-Encoding", "gzip")
	handler = DefaultHandler().WrapHandler(handler)

	handler.ServeHTTP(w, r)

	result := w.Result()

	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.Empty(t, result.Header.Get("Vary"))
	assert.Empty(t, result.Header.Get("Content-Encoding"))
	assert.Equal(t, "part 1\npart 2\npart 3\n", w.Body.String())
}
