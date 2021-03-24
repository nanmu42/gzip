package gin

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	gzippit "github.com/nanmu42/gzip/v2"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const handlerTestSize = 256

var bigPayload = []byte(`Four score and seven years ago our fathers brought forth on this continent, a new nation, conceived in Liberty, and dedicated to the proposition that all men are created equal.

Now we are engaged in a great civil war, testing whether that nation, or any nation so conceived and so dedicated, can long endure. We are met on a great battle-field of that war. We have come to dedicate a portion of that field, as a final resting place for those who here gave their lives that that nation might live. It is altogether fitting and proper that we should do this.

But, in a larger sense, we can not dedicate -- we can not consecrate -- we can not hallow -- this ground. The brave men, living and dead, who struggled here, have consecrated it, far above our poor power to add or detract. The world will little note, nor long remember what we say here, but it can never forget what they did here. It is for us the living, rather, to be dedicated here to the unfinished work which they who fought here have thus far so nobly advanced. It is rather for us to be here dedicated to the great task remaining before us -- that from these honored dead we take increased devotion to that cause for which they gave the last full measure of devotion -- that we here highly resolve that these dead shall not have died in vain -- that this nation, under God, shall have a new birth of freedom -- and that government of the people, by the people, for the people, shall not perish from the earth.`)

var smallPayload = []byte(`Chancellor on brink of second bailout for banks`)

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

func BenchmarkSoleGin_SmallPayload(b *testing.B) {
	var (
		g = newGinInstance(smallPayload)
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	h := map[string][]string(w.header)
	for i := 0; i < b.N; i++ {
		// Delete header between calls.
		for k := range h {
			delete(h, k)
		}
		g.ServeHTTP(w, r)
	}

	b.StopTimer()
	if encoding := w.Header().Get("Content-Encoding"); encoding != "" {
		b.Fatalf("Content-Encoding is not empty, but %s", encoding)
	}
}

func BenchmarkGinWithDefaultHandler_SmallPayload(b *testing.B) {
	var (
		g = newGinInstance(smallPayload, Adapt(gzippit.DefaultHandler()))
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	h := map[string][]string(w.header)
	for i := 0; i < b.N; i++ {
		// Delete header between calls.
		for k := range h {
			delete(h, k)
		}
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
	h := map[string][]string(w.header)
	for i := 0; i < b.N; i++ {
		// Delete header between calls.
		for k := range h {
			delete(h, k)
		}
		g.ServeHTTP(w, r)
	}

	b.StopTimer()
	if encoding := w.Header().Get("Content-Encoding"); encoding != "" {
		b.Fatalf("Content-Encoding is not empty, but %s", encoding)
	}
}

func BenchmarkGinWithDefaultHandler_BigPayload(b *testing.B) {
	var (
		g = newGinInstance(bigPayload, Adapt(gzippit.DefaultHandler()))
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	h := map[string][]string(w.header)
	for i := 0; i < b.N; i++ {
		// Delete header between calls.
		for k := range h {
			delete(h, k)
		}
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
		g = newEchoGinInstance(bigPayload, Adapt(gzippit.DefaultHandler()))
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
			body, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.True(t, bytes.HasPrefix(body, []byte(seq)))
		})
	}
}

func TestGinWithLevelsHandler(t *testing.T) {
	for i := gzippit.Stateless; i < 10; i++ {
		var seq = "level_" + strconv.Itoa(i)
		i := i
		t.Run(seq, func(t *testing.T) {
			g := newEchoGinInstance(bigPayload, Adapt(gzippit.NewHandler(gzippit.Config{
				CompressionLevel: i,
				MinContentLength: 1,
			})))

			var (
				w = httptest.NewRecorder()
				r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(seq))
			)

			r.Header.Set("Accept-Encoding", "gzip")
			g.ServeHTTP(w, r)

			result := w.Result()
			require.EqualValues(t, http.StatusOK, result.StatusCode)
			require.Equal(t, "gzip", result.Header.Get("Content-Encoding"))
			comp, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			reader, err := gzip.NewReader(bytes.NewReader(comp))
			require.NoError(t, err)
			body, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.True(t, bytes.HasPrefix(body, []byte(seq)))
			t.Logf("%s: compressed %d => %d", seq, len(body), len(comp))
		})
	}
}

func TestGinWithDefaultHandler_404(t *testing.T) {
	var (
		g = newGinInstance(bigPayload, Adapt(gzippit.DefaultHandler()))
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
		g = newGinInstance(bigPayload, Adapt(gzippit.DefaultHandler()))
		r = httptest.NewRequest(http.MethodPatch, "/", nil)
		w = httptest.NewRecorder()
	)

	r.Header.Set("Accept-Encoding", "gzip")

	g.ServeHTTP(w, r)

	result := w.Result()

	assert.EqualValues(t, http.StatusMethodNotAllowed, result.StatusCode)
	assert.Equal(t, "405 method not allowed", w.Body.String())
}

func TestGinCORSMiddleware(t *testing.T) {
	var (
		g = newGinInstance(bigPayload, Adapt(gzippit.DefaultHandler()), corsMiddleware)
		r = httptest.NewRequest(http.MethodOptions, "/", nil)
		w = httptest.NewRecorder()
	)

	g.ServeHTTP(w, r)
	result := w.Result()

	assert.EqualValues(t, http.StatusNoContent, result.StatusCode)
	assert.Equal(t, "*", result.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "POST", result.Header.Get("Access-Control-Allow-Methods"))
	assert.EqualValues(t, 0, w.Body.Len())
}

func TestGinCORSMiddlewareWithDummyConfig(t *testing.T) {
	var (
		g = newGinInstance(bigPayload, Adapt(gzippit.NewHandler(gzippit.Config{
			CompressionLevel:     gzippit.DefaultCompression,
			MinContentLength:     100,
			RequestFilter:        nil,
			ResponseHeaderFilter: nil,
		})), corsMiddleware)
		r = httptest.NewRequest(http.MethodOptions, "/", nil)
		w = httptest.NewRecorder()
	)

	g.ServeHTTP(w, r)
	result := w.Result()

	assert.EqualValues(t, http.StatusNoContent, result.StatusCode)
	assert.Equal(t, "*", result.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "POST", result.Header.Get("Access-Control-Allow-Methods"))
	assert.EqualValues(t, 0, w.Body.Len())
}

// corsMiddleware allows CORS request
func corsMiddleware(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST")

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}
