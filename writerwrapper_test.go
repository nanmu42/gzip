package gzip

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var bigPayload = []byte(`Four score and seven years ago our fathers brought forth on this continent, a new nation, conceived in Liberty, and dedicated to the proposition that all men are created equal.

Now we are engaged in a great civil war, testing whether that nation, or any nation so conceived and so dedicated, can long endure. We are met on a great battle-field of that war. We have come to dedicate a portion of that field, as a final resting place for those who here gave their lives that that nation might live. It is altogether fitting and proper that we should do this.

But, in a larger sense, we can not dedicate -- we can not consecrate -- we can not hallow -- this ground. The brave men, living and dead, who struggled here, have consecrated it, far above our poor power to add or detract. The world will little note, nor long remember what we say here, but it can never forget what they did here. It is for us the living, rather, to be dedicated here to the unfinished work which they who fought here have thus far so nobly advanced. It is rather for us to be here dedicated to the great task remaining before us -- that from these honored dead we take increased devotion to that cause for which they gave the last full measure of devotion -- that we here highly resolve that these dead shall not have died in vain -- that this nation, under God, shall have a new birth of freedom -- and that government of the people, by the people, for the people, shall not perish from the earth.`)

var smallPayload = []byte(`Chancellor on brink of second bailout for banks`)

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(ioutil.Discard)
	}}

func getGzipWriter() *gzip.Writer {
	return gzipWriterPool.Get().(*gzip.Writer)
}

func putGzipWriter(w *gzip.Writer) {
	if w == nil {
		return
	}

	_ = w.Close()
	w.Reset(ioutil.Discard)
	gzipWriterPool.Put(w)
}

type DummyResFilter bool

func (d DummyResFilter) ShouldCompress(_ http.Header) bool {
	return bool(d)
}

const minContentLength = 100

func newWrapper(filters ...ResponseHeaderFilter) (*writerWrapper, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	return newWriterWrapper(
		filters,
		minContentLength,
		recorder,
		getGzipWriter,
		putGzipWriter,
	), recorder
}

func Test_writerWrapper_Flush(t *testing.T) {
	wrapper, recorder := newWrapper()
	wrapper.Flush()
	assert.True(t, recorder.Flushed)
}

func TestNewWriterWrapper_ShouldCompress_True(t *testing.T) {
	wrapper := newWriterWrapper(
		nil,
		minContentLength,
		nil,
		getGzipWriter,
		putGzipWriter,
	)

	assert.True(t, wrapper.shouldCompress)
}

func Test_writerWrapper_Header(t *testing.T) {
	const (
		key   = "hi"
		value = "I am here!"
	)

	wrapper, recorder := newWrapper()

	recorder.Header().Set(key, value)
	wrapper.OriginWriter = recorder

	assert.Equal(t, value, wrapper.Header().Get(key))
}

func Test_writerWrapper_WriteHeader_Twice(t *testing.T) {
	wrapper, recorder := newWrapper()

	wrapper.WriteHeader(http.StatusBadRequest)
	wrapper.WriteHeader(http.StatusNotImplemented)

	wrapper.CleanUp()
	result := recorder.Result()

	assert.EqualValues(t, http.StatusBadRequest, result.StatusCode)
}

func Test_writerWrapper_WriteHeader_ShouldNotCompress(t *testing.T) {
	wrapper, recorder := newWrapper()
	wrapper.shouldCompress = false

	wrapper.WriteHeader(http.StatusBadRequest)
	wrapper.CleanUp()
	result := recorder.Result()

	assert.EqualValues(t, http.StatusBadRequest, result.StatusCode)
}

func Test_writerWrapper_WriteHeader_ShouldNotCompress_StatusCode(t *testing.T) {
	wrapper, _ := newWrapper()

	wrapper.WriteHeader(http.StatusNoContent)
	assert.False(t, wrapper.shouldCompress)

	wrapper, _ = newWrapper()
	wrapper.WriteHeader(http.StatusNotModified)
	assert.False(t, wrapper.shouldCompress)
}

func Test_writerWrapper_WriteHeader_filter_yes(t *testing.T) {
	wrapper, _ := newWrapper(DummyResFilter(true))

	wrapper.WriteHeader(http.StatusOK)
	assert.True(t, wrapper.shouldCompress)
}

func Test_writerWrapper_Write_after_WriteHeader(t *testing.T) {
	wrapper, recorder := newWrapper()

	wrapper.WriteHeader(http.StatusConflict)
	_, err := wrapper.Write(bigPayload)
	assert.NoError(t, err)
	wrapper.CleanUp()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusConflict, result.StatusCode)
}

func Test_writerWrapper_Write_big(t *testing.T) {
	assert.Greater(t, len(bigPayload), minContentLength)

	wrapper, recorder := newWrapper()

	_, err := wrapper.Write(bigPayload)
	assert.NoError(t, err)
	wrapper.CleanUp()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.True(t, wrapper.shouldCompress)
	assert.True(t, wrapper.didFirstWrite)

	reader, err := gzip.NewReader(result.Body)
	assert.NoError(t, err)
	body, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, bigPayload, body)
}

func Test_writerWrapper_Write_big_all_yes(t *testing.T) {
	assert.Greater(t, len(bigPayload), minContentLength)

	wrapper, recorder := newWrapper(DummyResFilter(true), DummyResFilter(true), DummyResFilter(true))

	_, err := wrapper.Write(bigPayload)
	assert.NoError(t, err)
	wrapper.CleanUp()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.True(t, wrapper.shouldCompress)
	assert.True(t, wrapper.didFirstWrite)

	reader, err := gzip.NewReader(result.Body)
	assert.NoError(t, err)
	body, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, bigPayload, body)
}

func Test_writerWrapper_Write_big_filter_yes_no_yes(t *testing.T) {
	wrapper, recorder := newWrapper(DummyResFilter(true), DummyResFilter(false), DummyResFilter(true))

	_, err := wrapper.Write(bigPayload)
	assert.NoError(t, err)
	wrapper.CleanUp()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.False(t, wrapper.shouldCompress)
	assert.True(t, wrapper.didFirstWrite)
}

func Test_writerWrapper_Write_big_filter_no_yes_no(t *testing.T) {
	wrapper, recorder := newWrapper(DummyResFilter(false), DummyResFilter(true), DummyResFilter(false))

	_, err := wrapper.Write(bigPayload)
	assert.NoError(t, err)
	wrapper.CleanUp()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.False(t, wrapper.shouldCompress)
	assert.True(t, wrapper.didFirstWrite)
}

func Test_writerWrapper_Write_big_filter_all_no(t *testing.T) {
	wrapper, recorder := newWrapper(DummyResFilter(false), DummyResFilter(false), DummyResFilter(false))

	_, err := wrapper.Write(bigPayload)
	assert.NoError(t, err)
	wrapper.CleanUp()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.False(t, wrapper.shouldCompress)
	assert.True(t, wrapper.didFirstWrite)
}

func Test_writerWrapper_Write_small(t *testing.T) {
	assert.Less(t, len(smallPayload), minContentLength)

	wrapper, recorder := newWrapper()

	_, err := wrapper.Write(smallPayload)
	assert.NoError(t, err)
	wrapper.CleanUp()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.False(t, wrapper.shouldCompress)
	assert.True(t, wrapper.didFirstWrite)

	body, err := ioutil.ReadAll(result.Body)
	assert.NoError(t, err)
	assert.Equal(t, smallPayload, body)
}

func Test_writerWrapper_Write_small_with_bigContentLength(t *testing.T) {
	assert.Less(t, len(smallPayload), minContentLength)

	wrapper, recorder := newWrapper()
	recorder.Header().Set("Content-Length", strconv.Itoa(minContentLength+1))
	recorder.Header().Set("ETag", "12345")

	_, err := wrapper.Write(smallPayload)
	assert.NoError(t, err)
	wrapper.CleanUp()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.True(t, wrapper.shouldCompress)
	assert.True(t, wrapper.didFirstWrite)

	reader, err := gzip.NewReader(result.Body)
	assert.NoError(t, err)
	body, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, smallPayload, body)

	assert.Empty(t, result.Header.Get("Content-Length"))
	assert.Equal(t, "gzip", result.Header.Get("Content-Encoding"))
	assert.Equal(t, "Accept-Encoding", result.Header.Get("Vary"))
	assert.Equal(t, "W/12345", result.Header.Get("ETag"))
}
