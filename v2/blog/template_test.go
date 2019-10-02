package blog

import (
	"testing"

	"hawx.me/code/assert"
)

func TestGet(t *testing.T) {
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

	value, ok := get(m, "missing")
	assert.False(ok)

	value, ok = get(m, "empty")
	assert.False(ok)

	value, ok = get(m, "simple")
	assert.True(ok)
	assert.Equal("a string", value)

	value, ok = get(m, "map.key")
	assert.True(ok)
	assert.Equal("a map", value)

	value, ok = get(m, "map.missing")
	assert.False(ok)

	value, ok = get(m, "nested.key")
	assert.True(ok)
	assert.Equal("a nested", value)
}
