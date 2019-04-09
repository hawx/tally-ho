package data

import (
	"testing"
	"time"

	"hawx.me/code/assert"
)

func TestCreateAndUpdate(t *testing.T) {
	assert := assert.New(t)

	store, err := Open("file::memory:", fakeURLFactory{})
	if !assert.Nil(err) {
		return
	}
	defer store.Close()

	if err = store.Create("some-id", map[string][]interface{}{
		"name":     {"A post"},
		"category": {"cool", "post"},
		"author":   {"someone"},
	}); !assert.Nil(err) {
		return
	}

	if post, err := store.Get("some-id"); assert.Nil(err) {
		assert.Equal(map[string][]interface{}{
			"name":     {"A post"},
			"category": {"cool", "post"},
			"author":   {"someone"},
		}, post)
	}

	if err = store.Update("some-id",
		map[string][]interface{}{"name": {"what post"}},
		map[string][]interface{}{"category": {"?"}},
		map[string][]interface{}{"author": {"someone"}, "category": {"cool"}},
	); !assert.Nil(err) {
		return
	}

	if post, err := store.Get("some-id"); assert.Nil(err) {
		updated, _ := time.Parse(time.RFC3339, post["updated"][0].(string))
		delete(post, "updated")

		assert.Equal(map[string][]interface{}{
			"name":     {"what post"},
			"category": {"?", "post"},
		}, post)

		assert.WithinDuration(updated, time.Now(), time.Second)
	}
}
