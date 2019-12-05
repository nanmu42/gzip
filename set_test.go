package gzip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	var set = make(Set)

	assert.False(t, set.Contains("a"))
	set.Add("a")
	assert.True(t, set.Contains("a"))

	assert.False(t, set.Contains("b"))
	set.Add("b")
	assert.True(t, set.Contains("b"))

	assert.True(t, set.ContainsFunc(func(s string) bool {
		return s == "a"
	}))
	assert.True(t, set.ContainsFunc(func(s string) bool {
		return s == "b"
	}))
	assert.False(t, set.ContainsFunc(func(s string) bool {
		return s == "c"
	}))

	set.Remove("a")
	assert.False(t, set.Contains("a"))
	set.Remove("b")
	assert.False(t, set.Contains("b"))

	assert.False(t, set.Contains("c"))
	set.Remove("c")
	assert.False(t, set.Contains("c"))
}
