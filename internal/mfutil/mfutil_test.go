package mfutil

import (
	"testing"

	"hawx.me/code/assert"
)

func TestSafeGet(t *testing.T) {
	assert := assert.New(t)

	m := map[string][]interface{}{
		"empty":  {},
		"simple": {"a string"},
		"map": {
			map[string]interface{}{
				"key": "a map",
			},
		},
		"nested": {
			map[string][]interface{}{
				"key": {"a nested"},
			},
		},
	}

	value, ok := SafeGet(m, "missing")
	assert.False(ok)

	value, ok = SafeGet(m, "empty")
	assert.False(ok)

	value, ok = SafeGet(m, "simple")
	assert.True(ok)
	assert.Equal("a string", value)

	value, ok = SafeGet(m, "map.key")
	assert.True(ok)
	assert.Equal("a map", value)

	value, ok = SafeGet(m, "map.missing")
	assert.False(ok)

	value, ok = SafeGet(m, "nested.key")
	assert.True(ok)
	assert.Equal("a nested", value)
}
