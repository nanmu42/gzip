package gzip

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/klauspost/compress/gzip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const minContentLength = 100

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
	assert.EqualValues(t, minContentLength, cap(wrapper.bodyBuffer))
	assert.EqualValues(t, 0, len(wrapper.bodyBuffer))
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

	wrapper.FinishWriting()
	result := recorder.Result()

	assert.EqualValues(t, http.StatusNotImplemented, result.StatusCode)
}

func Test_writerWrapper_WriteHeader_ShouldNotCompress(t *testing.T) {
	wrapper, recorder := newWrapper()
	wrapper.shouldCompress = false

	wrapper.WriteHeader(http.StatusBadRequest)
	wrapper.FinishWriting()
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
	wrapper.FinishWriting()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusConflict, result.StatusCode)
}

func Test_writerWrapper_Write_big(t *testing.T) {
	require.Greater(t, len(bigPayload), minContentLength)

	wrapper, recorder := newWrapper()

	_, err := wrapper.Write(bigPayload)
	assert.NoError(t, err)
	wrapper.FinishWriting()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.True(t, wrapper.shouldCompress)
	assert.True(t, wrapper.responseHeaderChecked)
	assert.True(t, wrapper.bodyBigEnough)
	assert.EqualValues(t, minContentLength, cap(wrapper.bodyBuffer))
	assert.EqualValues(t, 0, len(wrapper.bodyBuffer))

	reader, err := gzip.NewReader(result.Body)
	assert.NoError(t, err)
	body, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, bigPayload, body)
}

func Test_writerWrapper_Write_big_part_by_part_and_reset(t *testing.T) {
	const partial = 10

	require.Greater(t, len(bigPayload), minContentLength)
	require.Less(t, partial, minContentLength)

	wrapper, recorder := newWrapper()
	assert.EqualValues(t, wrapper.Status(), 0)
	assert.EqualValues(t, wrapper.Size(), 0)
	assert.False(t, wrapper.Written())

	_, err := wrapper.Write(bigPayload[:partial])
	assert.NoError(t, err)
	assert.True(t, wrapper.shouldCompress)
	assert.True(t, wrapper.responseHeaderChecked)
	assert.False(t, wrapper.bodyBigEnough)
	assert.EqualValues(t, wrapper.Status(), http.StatusOK)
	assert.EqualValues(t, wrapper.Size(), partial)
	assert.True(t, wrapper.Written())

	_, err = wrapper.Write(bigPayload[partial:])
	assert.NoError(t, err)
	assert.EqualValues(t, wrapper.Status(), http.StatusOK)
	assert.EqualValues(t, wrapper.Size(), len(bigPayload))
	assert.True(t, wrapper.Written())

	wrapper.FinishWriting()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.True(t, wrapper.shouldCompress)
	assert.True(t, wrapper.responseHeaderChecked)
	assert.True(t, wrapper.bodyBigEnough)
	assert.EqualValues(t, minContentLength, cap(wrapper.bodyBuffer))
	assert.EqualValues(t, partial, len(wrapper.bodyBuffer))

	wrapper.Reset(nil)
	assert.EqualValues(t, minContentLength, cap(wrapper.bodyBuffer))
	assert.EqualValues(t, 0, len(wrapper.bodyBuffer))
	assert.EqualValues(t, wrapper.Status(), 0)
	assert.EqualValues(t, wrapper.Size(), 0)
	assert.False(t, wrapper.Written())

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
	wrapper.FinishWriting()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.True(t, wrapper.shouldCompress)
	assert.True(t, wrapper.responseHeaderChecked)
	assert.True(t, wrapper.bodyBigEnough)

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
	wrapper.FinishWriting()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.False(t, wrapper.shouldCompress)
	assert.True(t, wrapper.responseHeaderChecked)
	assert.False(t, wrapper.bodyBigEnough)
}

func Test_writerWrapper_Write_big_filter_no_yes_no(t *testing.T) {
	wrapper, recorder := newWrapper(DummyResFilter(false), DummyResFilter(true), DummyResFilter(false))

	_, err := wrapper.Write(bigPayload)
	assert.NoError(t, err)
	wrapper.FinishWriting()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.False(t, wrapper.shouldCompress)
	assert.True(t, wrapper.responseHeaderChecked)
	assert.False(t, wrapper.bodyBigEnough)
}

func Test_writerWrapper_Write_big_filter_all_no(t *testing.T) {
	wrapper, recorder := newWrapper(DummyResFilter(false), DummyResFilter(false), DummyResFilter(false))

	_, err := wrapper.Write(bigPayload)
	assert.NoError(t, err)
	wrapper.FinishWriting()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.False(t, wrapper.shouldCompress)
	assert.True(t, wrapper.responseHeaderChecked)
	assert.False(t, wrapper.bodyBigEnough)
}

func Test_writerWrapper_Write_small(t *testing.T) {
	assert.Less(t, len(smallPayload), minContentLength)

	wrapper, recorder := newWrapper()

	_, err := wrapper.Write(smallPayload)
	assert.NoError(t, err)
	wrapper.FinishWriting()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.False(t, wrapper.shouldCompress)
	assert.True(t, wrapper.responseHeaderChecked)
	assert.False(t, wrapper.bodyBigEnough)

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
	wrapper.FinishWriting()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.True(t, wrapper.shouldCompress)
	assert.True(t, wrapper.responseHeaderChecked)
	assert.True(t, wrapper.bodyBigEnough)

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

func Test_writerWrapper_Write_content_type_sniff(t *testing.T) {
	assert.Greater(t, len(bigPayload), minContentLength)

	wrapper, recorder := newWrapper(DummyResFilter(true))

	var err error

	_, err = wrapper.Write([]byte("<html>"))
	assert.NoError(t, err)
	_, err = wrapper.Write(bigPayload)
	assert.NoError(t, err)
	_, err = wrapper.Write([]byte("</html>"))
	assert.NoError(t, err)

	wrapper.FinishWriting()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.True(t, wrapper.shouldCompress)
	assert.True(t, wrapper.responseHeaderChecked)
	assert.True(t, wrapper.bodyBigEnough)
	assert.Equal(t, "text/html; charset=utf-8", result.Header.Get("Content-Type"))

	reader, err := gzip.NewReader(result.Body)
	assert.NoError(t, err)
	_, err = ioutil.ReadAll(reader)
	assert.NoError(t, err)
}

func Test_writerWrapper_Write_content_type_no_sniff(t *testing.T) {
	const contentType = "text/special; charset=utf-8"

	assert.Greater(t, len(bigPayload), minContentLength)

	wrapper, recorder := newWrapper(DummyResFilter(true))

	var err error
	wrapper.Header().Set("Content-Type", contentType)
	_, err = wrapper.Write(bigPayload)
	assert.NoError(t, err)

	wrapper.FinishWriting()

	result := recorder.Result()
	assert.EqualValues(t, http.StatusOK, result.StatusCode)
	assert.True(t, wrapper.shouldCompress)
	assert.True(t, wrapper.responseHeaderChecked)
	assert.True(t, wrapper.bodyBigEnough)
	assert.Equal(t, contentType, result.Header.Get("Content-Type"))

	reader, err := gzip.NewReader(result.Body)
	assert.NoError(t, err)
	body, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, bigPayload, body)
}

func Test_writeWrapper_does_not_change_status_code_after_204(t *testing.T) {
	wrapper, recorder := newWrapper()

	wrapper.WriteHeader(http.StatusNoContent)
	_, _ = wrapper.Write([]byte("something"))

	assert.Equal(t, http.StatusNoContent, recorder.Code)
}