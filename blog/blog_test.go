package blog

import (
	"database/sql"
	"testing"
	"time"

	"hawx.me/code/assert"
)

func TestCreate(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")
	if !assert.Nil(t, err) {
		return
	}

	blog, err := New(Options{
		BaseURL:  "http://example.com/weblog/",
		BasePath: "/wwwroot/weblog/",
		Db:       db,
		WebPath:  "../web",
	})
	if !assert.Nil(t, err) {
		return
	}

	if !assert.Nil(t, blog.SetNextPage("some page")) {
		return
	}

	t.Run("with no name or mp-slug value", func(t *testing.T) {
		data, err := blog.Create(map[string][]interface{}{})

		if !assert.Nil(t, err) {
			return
		}
		assert.NotEmpty(t, data["uid"][0])
		assert.Equal(t, "some page", data["hx-page"][0])
		assert.Equal(t, "http://example.com/weblog/some-page/"+data["uid"][0].(string)+"/", data["url"][0])

		published, err := time.Parse(time.RFC3339, data["published"][0].(string))
		if assert.Nil(t, err) {
			assert.WithinDuration(t, time.Now(), published, time.Second)
		}
	})

	t.Run("with name", func(t *testing.T) {
		data, err := blog.Create(map[string][]interface{}{
			"name": {"this is my post"},
		})

		if !assert.Nil(t, err) {
			return
		}
		assert.NotEmpty(t, data["uid"][0])
		assert.Equal(t, "some page", data["hx-page"][0])
		assert.Equal(t, "http://example.com/weblog/some-page/this-is-my-post/", data["url"][0])

		published, err := time.Parse(time.RFC3339, data["published"][0].(string))
		if assert.Nil(t, err) {
			assert.WithinDuration(t, time.Now(), published, time.Second)
		}
	})

	t.Run("with name and mp-slug", func(t *testing.T) {
		data, err := blog.Create(map[string][]interface{}{
			"name":    {"ignored"},
			"mp-slug": {"please-use-this"},
		})

		if !assert.Nil(t, err) {
			return
		}
		assert.NotEmpty(t, data["uid"][0])
		assert.Equal(t, "some page", data["hx-page"][0])
		assert.Equal(t, "http://example.com/weblog/some-page/please-use-this/", data["url"][0])

		published, err := time.Parse(time.RFC3339, data["published"][0].(string))
		if assert.Nil(t, err) {
			assert.WithinDuration(t, time.Now(), published, time.Second)
		}
	})
}
