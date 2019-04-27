package data

import (
	"database/sql"
	"testing"

	// register sqlite3 for database/sql
	_ "github.com/mattn/go-sqlite3"

	"hawx.me/code/assert"
)

func TestCreateAndUpdate(t *testing.T) {
	assert := assert.New(t)

	db, err := sql.Open("sqlite3", "file::memory:")
	if !assert.Nil(err) {
		return
	}

	store, err := Open(db, fakeURLFactory{})
	if !assert.Nil(err) {
		return
	}
	defer store.Close()

	if err = store.Create("some-id", map[string][]interface{}{
		"name":     {"A post"},
		"category": {"cool", "post"},
		"author":   {"someone"},
		"url":      {"some-id"},
	}); !assert.Nil(err) {
		return
	}

	if post, err := store.GetByURL("some-id"); assert.Nil(err) {
		assert.Equal(map[string][]interface{}{
			"name":     {"A post"},
			"category": {"cool", "post"},
			"author":   {"someone"},
			"url":      {"some-id"},
		}, post)
	}

	if err = store.Update("some-id",
		map[string][]interface{}{"name": {"what post"}},
		map[string][]interface{}{"category": {"?"}},
		map[string][]interface{}{"author": {"someone"}, "category": {"cool"}},
	); !assert.Nil(err) {
		return
	}

	if post, err := store.GetByURL("some-id"); assert.Nil(err) {
		// updated, _ := time.Parse(time.RFC3339, post["updated"][0].(string))
		// delete(post, "updated")

		assert.Equal(map[string][]interface{}{
			"name":     {"what post"},
			"category": {"?", "post"},
			"url":      {"some-id"},
		}, post)

		// assert.WithinDuration(updated, time.Now(), time.Second)
	}
}
