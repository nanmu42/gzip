package gzip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
